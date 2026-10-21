package cronjob

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/deploy"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/locker"
	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/dustin/go-humanize"
	"github.com/samber/lo"
	"go.opentelemetry.io/otel"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// cronjobTracerName 是定时任务用例的 OpenTelemetry tracer 名称（取 module path），
// 让系统发起的周期工作（如休眠对账）在 trace 面板与其触发的下游调用串联成一条链：
// 对账本身无请求上下文，需在入口开根 span，下游 data 层 repo span 才能挂到同一条 trace 上。
const cronjobTracerName = "github.com/duc-cnzj/mars/v6/internal/cronjob"

// Tasks 是定时任务用例集合：7 个任务由注册表 Registry 枚举声明（CronTask 值），
// 内部仅依赖各 biz 端口/uploader/helm/git 完成周期清理、状态修复与缓存预热，是
// 纯应用层用例，不触碰任何基础设施门面（DB/K8sClient/Config 均经端口访问）。
type Tasks struct {
	timer             timer.Timer
	logger            mlog.Logger
	event             biz.EventRepo
	cfg               *config.Config
	up                uploader.Uploader
	helm              biz.HelmerRepo
	gitRepo           biz.GitRepo
	getCerts          func() (name, key, crt string)
	k8sRepo           biz.K8sRepo
	nsRepo            biz.NamespaceRepo
	repoRepo          biz.RepoRepo
	fileRepo          biz.FileRepo
	projectRepo       biz.ProjectRepo
	hibernateRepo     biz.HibernateRepo
	hibernateDeployer biz.HibernateDeployer
	// agentActivity 判定 agent 沙箱是否仍在活跃窗口内（空闲关闭对账用）。
	agentActivity biz.AgentActivity
	// locker 让空闲缩容与唤醒/部署互斥：共用部署全局锁的项目级键。
	locker locker.Locker
	// hibernateWindowRepo 解析仓库引用的全局休眠时间窗（按名字查启止时段）。
	hibernateWindowRepo biz.HibernateWindowRepo
	// warnedMissingWindow 记录已告警过的悬空窗口名：悬空引用是对账每轮都会遇到的预期态，
	// 去重后同一窗名只告警一次，既保留可观测性又避免每 2 分钟刷屏淹没真信号。
	warnedMissingWindow sync.Map
}

// PluginDeps 定时任务用例的插件能力集合。字段为惰性闭包：wire 期构造
// （插件未加载），触发时实时解析已加载插件，无需服务器启动前二次刷新。
type PluginDeps struct {
	// GetCerts 返回域名插件提供的 TLS 证书名称、密钥与证书。
	GetCerts func() (name, key, crt string)
}

// NewTasks 构造 Tasks 用例，依赖全部经端口注入，不含基础设施实现。
// deps 为惰性插件闭包集合（wire 期由组合根捕获 PluginManager 注入）。
func NewTasks(
	tm timer.Timer,
	logger mlog.Logger,
	cfg *config.Config,
	fileRepo biz.FileRepo,
	projectRepo biz.ProjectRepo,
	repoRepo biz.RepoRepo,
	nsRepo biz.NamespaceRepo,
	k8sRepo biz.K8sRepo,
	event biz.EventRepo,
	up uploader.Uploader,
	helm biz.HelmerRepo,
	gitRepo biz.GitRepo,
	hibernateRepo biz.HibernateRepo,
	hibernateDeployer biz.HibernateDeployer,
	hibernateWindowRepo biz.HibernateWindowRepo,
	agentActivity biz.AgentActivity,
	locker locker.Locker,
	deps *PluginDeps,
) *Tasks {
	return &Tasks{
		timer:               tm,
		fileRepo:            fileRepo,
		logger:              logger.WithModule("app/cron"),
		event:               event,
		cfg:                 cfg,
		up:                  up,
		helm:                helm,
		gitRepo:             gitRepo,
		getCerts:            deps.GetCerts,
		k8sRepo:             k8sRepo,
		nsRepo:              nsRepo,
		repoRepo:            repoRepo,
		projectRepo:         projectRepo,
		hibernateRepo:       hibernateRepo,
		hibernateDeployer:   hibernateDeployer,
		hibernateWindowRepo: hibernateWindowRepo,
		agentActivity:       agentActivity,
		locker:              locker,
	}
}

// ReconcileAgentIdle 关闭空闲的 AI Agent 沙箱：遍历 mars-agent 空间下的 agent 项目，
// 把活跃窗口已过期且仍在运行的沙箱缩容到 0。活跃窗口由 LLM/MCP/配置入口的 Touch
// 写入（TTL=ai.idle_timeout）。ai.idle_timeout 为负（禁用）时直接返回；空间不存在
// 视为尚未启用 AI。单个项目的活跃度/缩容失败只记日志并跳过，下一轮重试。
func (repo *Tasks) ReconcileAgentIdle() error {
	if repo.cfg.AgentIdleTimeout() <= 0 {
		return nil
	}
	// 对账无请求上下文，入口开根 span 让下游 repo 调用挂到同一条 trace 上。
	ctx, span := otel.Tracer(cronjobTracerName).Start(context.Background(), "cronjob/ReconcileAgentIdle")
	defer span.End()

	ns, err := repo.nsRepo.FindByName(ctx, biz.AgentNamespace)
	if err != nil {
		if errs.IsNotFound(err) {
			return nil
		}
		return err
	}
	projects, err := repo.projectRepo.ListAllProjectBriefs(ctx)
	if err != nil {
		return err
	}
	for _, proj := range projects {
		if proj.Namespace == nil || proj.Namespace.Name != ns.Name ||
			!strings.HasPrefix(proj.Name, biz.AgentProjectPrefix) {
			continue
		}
		repo.closeIdleAgent(ctx, ns, proj)
	}
	return nil
}

// closeIdleAgent 处理单个 agent 项目：活跃中、未部署或已缩容都直接返回；
// 缩容失败只记日志，交由下一轮对账重试。
func (repo *Tasks) closeIdleAgent(ctx context.Context, ns *biz.Namespace, proj *biz.Project) {
	active, err := repo.agentActivity.IsActive(ctx, proj.Name)
	if err != nil {
		repo.logger.Warningf("[agent-idle]: 查询项目 %s 活跃度失败，跳过本轮: %v", proj.Name, err)
		return
	}
	if active {
		return
	}
	name := biz.AgentDeploymentName(proj.Name)
	dep, err := repo.k8sRepo.GetDeployment(ns.Name, name)
	if err != nil {
		if !errs.IsNotFound(err) {
			repo.logger.Warningf("[agent-idle]: 查询项目 %s 的沙箱 Deployment 失败: %v", proj.Name, err)
		}
		return
	}
	if dep.Spec.Replicas != nil && *dep.Spec.Replicas == 0 {
		return
	}
	// 与唤醒/部署共用同一把项目级锁，把「复查活跃度 + 缩容」整体串行化：
	// 仅做二次复查仍留有检查与 Scale 之间的窗口，并发唤醒时可能刚唤醒又被关掉。
	// 抢不到锁说明同一项目正在唤醒或部署，跳过本轮即可。
	release, acquired := repo.locker.RenewalAcquire(
		deploy.GetSlugName(ns.ID, proj.Name), deploy.LockTTLSeconds, deploy.LockRenewalSeconds)
	if !acquired {
		return
	}
	defer release()

	active, err = repo.agentActivity.IsActive(ctx, proj.Name)
	if err != nil {
		repo.logger.Warningf("[agent-idle]: 缩容前复查项目 %s 活跃度失败，跳过本轮: %v", proj.Name, err)
		return
	}
	if active {
		return
	}
	if err := repo.k8sRepo.ScaleDeployment(ctx, ns.Name, name, 0); err != nil {
		repo.logger.Warningf("[agent-idle]: 缩容项目 %s 的沙箱失败: %v", proj.Name, err)
		return
	}
	metrics.AgentSandboxTransitionTotal.WithLabelValues(metrics.AgentSandboxActionSleep).Inc()
	// 系统发起的缩容会直接影响用户沙箱可用性，与唤醒轮换令牌一样留审计痕迹
	// （唤醒侧在 services 层记录，这里以 system 身份记录同一条状态变更）。
	// 操作人恒为 system，沙箱归属用户由项目创建者标识（与部署/唤醒令牌绑定的身份一致）。
	repo.event.AuditLogWithRequest(types.EventActionType_Update, "system", "",
		fmt.Sprintf("AI Agent 沙箱空闲缩容到 0（用户 %s）", proj.Creator),
		map[string]any{"project_id": proj.ID, "project_name": proj.Name, "creator": proj.Creator})
	repo.logger.Infof("[agent-idle]: 项目 %s（用户 %s）空闲超过 %s，沙箱已缩容到 0", proj.Name, proj.Creator, repo.cfg.AgentIdleTimeout())
}

// CacheAllBranches 并发拉取全部启用 git 仓库的分支并写入缓存（最多 8 个 goroutine）。
func (repo *Tasks) CacheAllBranches() error {
	defer func(t time.Time) {
		repo.logger.Debug("CacheAllBranches done", repo.timer.Since(t))
	}(repo.timer.Now())

	var wg = &sync.WaitGroup{}
	all, err := repo.repoRepo.All(context.Background(), &biz.AllRepoRequest{Enabled: lo.ToPtr(true), NeedGitRepo: lo.ToPtr(true)})
	if err != nil {
		return err
	}
	goroutineNum := len(all)
	if len(all) > 10 {
		goroutineNum = 8
	}
	wg.Add(goroutineNum)
	ch := make(chan int32, 100)
	for i := 0; i < goroutineNum; i++ {
		go func() {
			defer wg.Done()
			for id := range ch {
				repo.gitRepo.AllBranches(context.Background(), int(id), true)
			}
		}()
	}
	for _, it := range lo.UniqBy(all, func(item *biz.Repo) int32 { return item.GitProjectID }) {
		ch <- it.GitProjectID
	}
	close(ch)
	wg.Wait()

	return nil
}

// SyncImagePullSecrets 对每个 namespace 对账 imagePullSecrets：清理已失效的
// docker secret、同步过期的 registry 凭据、为缺失的 registry 创建新 secret。
// 全部 k8s 操作经 K8sRepo 端口，namespace 列表/回写经 NamespaceRepo 端口。
func (repo *Tasks) SyncImagePullSecrets() error {
	var (
		cfgImagePullSecrets = repo.cfg.ImagePullSecrets
		logger              = repo.logger
	)
	var serverMap = make(map[string]biz.DockerConfigEntry)
	for _, s := range cfgImagePullSecrets {
		serverMap[s.Server] = biz.DockerConfigEntry{
			Username: s.Username,
			Password: s.Password,
			Email:    s.Email,
			Auth:     base64.StdEncoding.EncodeToString([]byte(s.Username + ":" + s.Password)),
		}
	}
	namespaceList, err := repo.nsRepo.ListAll(context.Background())
	if err != nil {
		return err
	}
	for _, namespace := range namespaceList {
		var (
			checked = make(map[string]struct{})
			missing []string
			ns      = namespace
		)
		for _, secretName := range ns.ImagePullSecrets {
			secret, err := repo.k8sRepo.GetSecret(context.Background(), ns.Name, secretName)
			if err != nil {
				logger.Warningf("[syncImagePullSecrets]: error get secret '%s', err %v", secretName, err)
				if apierrors.IsNotFound(err) {
					ns = repo.deleteSecret(ns, secretName)
				}
				continue
			}
			if secret.Type == corev1.SecretTypeDockerConfigJson {
				var dockerJsonKeyData = secret.Data[corev1.DockerConfigJsonKey]
				res, err := biz.DecodeDockerConfigJSON(dockerJsonKeyData)
				if err != nil {
					logger.Warningf("[syncImagePullSecrets]: decode secret '%s', err %v", secretName, err)
					continue
				}
				var newConfigJson = biz.DockerConfigJSON{
					Auths:       map[string]biz.DockerConfigEntry{},
					HttpHeaders: map[string]string{},
				}
				for server, cfg := range serverMap {
					for s := range res.Auths {
						if server == s {
							newConfigJson.Auths[server] = cfg
							checked[server] = struct{}{}
							break
						}
					}
				}
				if len(newConfigJson.Auths) == 0 {
					ns = repo.deleteSecret(ns, secretName)
					continue
				}

				if !maps.Equal(newConfigJson.Auths, res.Auths) {
					logger.Warningf("[syncImagePullSecrets]: Find Diff, Auto Sync: '%s'", secretName)
					marshal, _ := json.Marshal(&newConfigJson)
					secret.Data[corev1.DockerConfigJsonKey] = marshal
					if _, err := repo.k8sRepo.UpdateSecret(context.Background(), ns.Name, secret.Name, secret); err != nil {
						logger.Warningf("[syncImagePullSecrets]: update secret '%s', err %v", secretName, err)
					}
				}
			}
		}

		for s := range serverMap {
			if _, ok := checked[s]; !ok {
				missing = append(missing, s)
			}
		}

		if len(missing) > 0 {
			secret, err := repo.k8sRepo.CreateDockerSecrets(context.Background(), ns.Name, missing)
			if err == nil {
				logger.Warningf("[syncImagePullSecrets]: Missing %v", missing)
				if err := repo.nsRepo.UpdateImagePullSecrets(context.Background(), ns.ID, append(ns.ImagePullSecrets, secret.Name)); err != nil {
					logger.Warningf("[syncImagePullSecrets]: update namespace '%s' imagePullSecrets, err %v", ns.Name, err)
				}
			}
		}
	}
	return nil
}

// deleteSecret 删除 k8s secret 并从 namespace 的 ImagePullSecrets 列表移除该名字，
// 返回更新后的 namespace（调用方沿用其 ImagePullSecrets 做后续回写）。
func (repo *Tasks) deleteSecret(ns *biz.Namespace, secretName string) *biz.Namespace {
	logger := repo.logger
	logger.Warningf("[syncImagePullSecrets]: DELETE: %s", secretName)

	if err := repo.k8sRepo.DeleteSecret(context.Background(), ns.Name, secretName); err != nil {
		logger.Warningf("[syncImagePullSecrets]: delete k8s secret '%s', err %v", secretName, err)
	}
	var newNsArray []string
	for _, name := range ns.ImagePullSecrets {
		if name != secretName {
			newNsArray = append(newNsArray, name)
		}
	}
	if err := repo.nsRepo.UpdateImagePullSecrets(context.Background(), ns.ID, newNsArray); err != nil {
		logger.Warningf("[syncImagePullSecrets]: update namespace '%s' imagePullSecrets, err %v", ns.Name, err)
	}
	ns.ImagePullSecrets = newNsArray
	return ns
}

// SyncDomainSecret 把插件提供的 TLS 证书同步到全部 namespace：缺失时创建，
// 内容不一致时更新。k8s 读写经 K8sRepo 端口，namespace 遍历经 NamespaceRepo。
func (repo *Tasks) SyncDomainSecret() error {
	var (
		changed        bool
		changedSecrets []*corev1.Secret
	)
	secretName, tlsKey, tlsCrt := repo.getCerts()
	if secretName != "" && tlsKey != "" && tlsCrt != "" {
		allNamespaces, err := repo.nsRepo.ListAll(context.Background())
		if err != nil {
			return err
		}
		for _, n := range allNamespaces {
			secret, err := repo.k8sRepo.GetSecret(context.Background(), n.Name, secretName)
			if err != nil {
				if apierrors.IsNotFound(err) {
					repo.logger.Infof("[TLS]: Register secret namespace: %s, name %s.", n.Name, secretName)
					if _, err := repo.k8sRepo.AddTlsSecret(n.Name, secretName, tlsKey, tlsCrt); err != nil {
						repo.logger.Error(err)
					}
					continue
				}
				// 非 NotFound 错误（网络/权限等）时 secret 为 nil，落到下方
				// secret.Data 会 nil-deref panic。记日志后跳过本 namespace，
				// 下一轮同步重试。
				repo.logger.Errorf("[TLS]: get secret error, namespace: %s, name: %s: %v", n.Name, secretName, err)
				continue
			}
			if string(secret.Data["tls.crt"]) != tlsCrt || string(secret.Data["tls.key"]) != tlsKey {
				changed = true
				changedSecrets = append(changedSecrets, secret.DeepCopy())
			}
		}
	}

	if changed {
		sdata := map[string]string{
			"tls.key": tlsKey,
			"tls.crt": tlsCrt,
		}
		repo.logger.Warning("[TLS]: certs changed, updating...")
		for _, secret := range changedSecrets {
			secret.StringData = sdata
			_, err := repo.k8sRepo.UpdateSecret(context.Background(), secret.Namespace, secret.Name, secret)
			if err == nil {
				repo.logger.Infof("[TLS]: namespace: %s, name %s updated", secret.Namespace, secret.Name)
			}
		}
	}
	return nil
}

// FixDeployStatus 用 helm 实测状态修复标记为失败/未知的项目。
func (repo *Tasks) FixDeployStatus() error {
	projects, err := repo.projectRepo.ListByDeployStatus(context.Background(), types.Deploy_StatusFailed, types.Deploy_StatusUnknown)
	if err != nil {
		return err
	}
	for _, project := range projects {
		p := project
		// 所属空间为 nil 出现在「项目存活、其所属空间已软删」时：ListByDeployStatus 经
		// WithNamespace() 做边加载，而边加载走 SELECT、软删拦截器会把已软删的空间过滤掉，
		// 于是命名空间边为空（典型来源是项目级恢复与空间软删的并发竞态，见 data/project.go
		// 的 RestoreDeleted 注释）。
		//
		// 命中条件不是"deploy_status 恰好是列默认值 0"，而是**能进入本次遍历的行其状态本就
		// 已落在失败/未知两态**——级联软删不改 deploy_status，"已部署"的项目压根不会被
		// ListByDeployStatus 取出。不跳过就会 nil 解引用 panic：robfig cron 的 Recover 中间件
		// 兜得住进程，但本任务每轮都在同一行中断，其后的项目再也修不到，等于该任务失效。
		if p.Namespace == nil {
			continue
		}
		status := repo.helm.ReleaseStatus(p.Name, p.Namespace.Name)
		if status != types.Deploy_StatusFailed && status != types.Deploy_StatusUnknown {
			if _, err := repo.projectRepo.UpdateDeployStatus(context.Background(), p.ID, status); err != nil {
				return err
			}
		}
	}
	return nil
}

// hibernateReconcileWorkers 是休眠对账的并发度上限。单个项目的跃迁要同步跑完一次 helm
// 重部署（受 install_timeout 约束，默认 90s）；若共用同一时间窗的 K 个项目在窗口边界齐齐
// 跃迁，串行会叠加成 K×T 的长尾（如 K=20、T=30s → 约 10 分钟，远超 2 分钟调度周期），
// 且队尾项目被前序项目耗时之和拖累。固定 worker 池把长尾压到 ⌈K/并发度⌉×T，同时限制
// 并发 helm 数量，避免大量重部署同时冲击 apiserver/helm。取 5：一轮对账约等于一个调度
// 周期，又不至于并发过高惊群。
const hibernateReconcileWorkers = 5

// ReconcileHibernate 对账项目定时休眠：每 2 分钟推进一步状态机。
// 总闸：只接管项目级开关已开（Enabled=true）的记录——ListEnabled 已在查询侧过滤掉开关已关的
// 记录，故开关一关，对账即彻底放弃该项目，无论它当前处于什么状态（含已置零的休眠态）都不再做
// 任何修改，交由用户手动掌控；重新打开开关即恢复接管。
// 单趟遍历：不再分「清醒轮 / 休眠轮」两趟，每条记录按「是否在窗口内」与「是否真在休眠（状态
// sleeping 且副本字段均处休眠零值）」即时归位——窗口内非真休眠即休眠（含错位当轮重睡，无需等下一轮）、
// 窗口外真休眠即唤醒、窗口外错位则释放接管权。单个项目失败只记日志并跳过（下一轮重试），
// 不阻断本轮其它项目；列表查询失败才上抛。
// 并发：一趟内用固定 worker 池（hibernateReconcileWorkers）有界并发处理各项目——各项目互相
// 独立（不同 projectID，部署侧 GlobalLock 按 slug 分键互不冲突），并行不改变单项目状态机语义；
// 并发度封顶以避免同窗项目齐齐跃迁时全量重部署惊群。
func (repo *Tasks) ReconcileHibernate() error {
	// 对账无请求上下文，入口开根 span 并派生 ctx 贯穿全程：下游 repo/部署调用的 span 与日志
	// 因此挂到同一条 trace 上，可追溯「哪轮对账 → 哪个项目 → 哪次重部署」。
	ctx, span := otel.Tracer(cronjobTracerName).Start(context.Background(), "cronjob/ReconcileHibernate")
	defer span.End()
	now := repo.timer.Now()
	enabled, err := repo.hibernateRepo.ListEnabled(ctx)
	if err != nil {
		return err
	}
	if len(enabled) == 0 {
		return nil
	}

	workers := len(enabled)
	if workers > hibernateReconcileWorkers {
		workers = hibernateReconcileWorkers
	}
	// 缓冲容量设为全长：投递侧不因 worker 回收而阻塞（避免 worker 异常退出时投递侧死锁），
	// 未被消费的记录随本轮结束丢弃、下一轮重试。
	ch := make(chan *biz.ProjectHibernate, len(enabled))
	wg := &sync.WaitGroup{}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for rec := range ch {
				// 每个项目包一层独立 recover 帧：单项目 panic 只作废该项目、下一轮重试，
				// 不再像串行那样顺循环上抛拖垮整趟对账（串行时一个 panic 令本轮其余项目全部跳过）。
				//
				// 这里刻意不用 logger.HandlePanic：它会在 debug 模式把 panic 重新抛出，而本
				// worker 是独立 goroutine——robfig 的 Recover 只兜住它自己那个调度 goroutine，
				// 重抛的 panic 无人接住会直接崩掉整个进程（串行实现下的 panic 有 wrap 的
				// HandlePanicWithCallback 承接，故当时 debug 不会崩）。改用不重抛的 recover：
				// 记栈日志后继续消费下一项目，debug/prod 行为一致且都不崩进程。
				func() {
					defer func() {
						if r := recover(); r != nil {
							repo.logger.Errorf("[Panic]: [cron]: reconcileHibernateProject, err: %v --- [%s]", r, debug.Stack())
						}
					}()
					repo.reconcileHibernateProject(ctx, now, rec)
				}()
			}
		}()
	}
	for _, rec := range enabled {
		ch <- rec
	}
	close(ch)
	wg.Wait()

	return nil
}

// reconcileHibernateProject 推进单条「开关已开」休眠记录的状态：加载项目与所属仓库配置，按
// 窗口内外与真实休眠态的四象限分派到休眠 / 唤醒 / 释放，或原样保持。加载失败由
// loadHibernateTarget 记日志并返回 ok=false，直接跳过（下一轮重试）。
func (repo *Tasks) reconcileHibernateProject(ctx context.Context, now time.Time, rec *biz.ProjectHibernate) {
	proj, cfg, ok := repo.loadHibernateTarget(ctx, rec.ProjectID)
	if !ok {
		return
	}
	// replicaPaths 只要仓库配了休眠就取：即便仓库级能力已关，判断「副本是否真处休眠零值」仍需求值，
	// 以免把「已置零的休眠态」误判为错位。未配置休眠时为 nil，ReplicaValuesAllZero 恒真，
	// 退化为只看状态，保证此类项目仍能按快照被唤醒。
	var replicaPaths []string
	if cfg.Hibernate != nil {
		replicaPaths = cfg.Hibernate.ReplicaPaths
	}
	inWindow := false
	if cfg.Hibernate != nil && cfg.Hibernate.Enabled {
		// 生效窗：项目自身选定的子集（启用时已保证非空），命中任一窗即算在窗内（并集语义）。
		// 空选择（历史/异常数据）不继承仓库全量——按「未选即未启用」处理：解析不到窗即不参与休眠。
		if wins, wok := repo.resolveHibernateWindows(ctx, rec.WindowNames); wok {
			inWindow = anyHibernateWindowHit(now, wins)
		}
	}
	actuallySleeping := rec.State == biz.HibernateStateSleeping &&
		biz.ReplicaValuesAllZero(proj.FinalExtraValues, cfg.Elements, replicaPaths)

	switch {
	case inWindow && actuallySleeping:
		// 窗口内且确在休眠：保持，不做任何改动。
		return
	case inWindow:
		// 清醒态命中窗口，或标记休眠但副本非处零值的错位：本轮回休眠（错位时当前值成为新快照），
		// 与修正同一轮完成，不再等下一轮对账。
		repo.sleepHibernateProject(ctx, proj, rec, cfg.Elements, replicaPaths)
	case actuallySleeping:
		// 窗口外且确在休眠：按快照回填唤醒。
		repo.wakeHibernateProject(ctx, proj, rec)
	case rec.State == biz.HibernateStateSleeping:
		// 窗口外仍标记 sleeping 但副本非处零值：释放接管权（保留当前值）。
		repo.releaseHibernateProject(ctx, proj, rec)
	}
}

// sleepHibernateProject 执行单个项目的休眠：快照副本字段当前值并**先**落库「休眠态+快照」，
// 再把副本字段置为各自的休眠零值重部署（数字类置 0、开关类关闭）。窗口判定与「仓库是否启用
// 休眠」由调用方完成，本方法只负责落动作。
//
// 顺序是幂等的承重点：重部署会把置零后的副本值写回项目记录（deploy 管线 UpdateProject）。
// 若先部署后落库，一旦部署成功而落库失败/进程崩溃，下一轮对账读到的基座副本已被置零，会据此
// 重拍出「全零快照」覆盖真快照，令原始副本数永久丢失（唤醒只能恢复成零）。改为先落库后：
// 落库失败则本轮不部署、基座未变，下一轮重拍仍是原值；落库成功而部署失败，下一轮以同一快照
// 重试（基座同样未变）。两种失败都基于「基座未变」而自愈，故顺序不可颠倒。
func (repo *Tasks) sleepHibernateProject(ctx context.Context, proj *biz.Project, rec *biz.ProjectHibernate, elements []*mars.Element, replicaPaths []string) {
	base := proj.FinalExtraValues
	snapshot := biz.SnapshotReplicaValues(base, replicaPaths)
	target := biz.SleepExtraValues(base, elements, replicaPaths)
	if _, err := repo.hibernateRepo.Upsert(ctx, &biz.ProjectHibernate{
		ProjectID:   proj.ID,
		Enabled:     rec.Enabled,
		State:       biz.HibernateStateSleeping,
		Snapshot:    snapshot,
		WindowNames: rec.WindowNames,
	}); err != nil {
		repo.logger.Errorf("[hibernate]: 项目 %s 休眠状态落库失败: %v", proj.Name, err)
		return
	}
	if err := repo.hibernateDeployer.Redeploy(ctx, proj, target); err != nil {
		repo.logger.Errorf("[hibernate]: 项目 %s 休眠重部署失败: %v", proj.Name, err)
		return
	}
	// 自动把副本置零是运维需要可追溯的状态变更，成功也留 Info（此前仅失败有痕）。
	repo.logger.Infof("[hibernate]: 项目 %s 命中休眠窗口，副本字段已置为休眠零值（快照 %d 个字段）", proj.Name, len(snapshot))
}

// wakeHibernateProject 执行单个项目的唤醒：按快照回填副本字段重部署、状态转 awake 并清空快照。
// 仅由对账在「窗口外且确在休眠（副本均处休眠零值）」时调用，故无需再判窗口。
func (repo *Tasks) wakeHibernateProject(ctx context.Context, proj *biz.Project, rec *biz.ProjectHibernate) {
	target := biz.WakeExtraValues(proj.FinalExtraValues, rec.Snapshot)
	if err := repo.hibernateDeployer.Redeploy(ctx, proj, target); err != nil {
		repo.logger.Errorf("[hibernate]: 项目 %s 唤醒重部署失败: %v", proj.Name, err)
		return
	}
	if _, err := repo.hibernateRepo.Upsert(ctx, &biz.ProjectHibernate{
		ProjectID:   proj.ID,
		Enabled:     rec.Enabled,
		State:       biz.HibernateStateAwake,
		WindowNames: rec.WindowNames,
	}); err != nil {
		repo.logger.Errorf("[hibernate]: 项目 %s 唤醒状态落库失败: %v", proj.Name, err)
		return
	}
	// 成功唤醒同样留 Info，供运维追溯自动回填的产生。
	repo.logger.Infof("[hibernate]: 项目 %s 已唤醒（按快照回填 %d 个副本字段）", proj.Name, len(rec.Snapshot))
}

// releaseHibernateProject 释放单个项目的休眠接管权：状态转 awake 并清空快照、保留当前副本值，
// 不重部署。用于「窗口外仍标记 sleeping 但副本非处休眠零值」的错位（关闭期间改过副本数后重开），
// 交回用户当前值掌控；窗口内出现同类错位则由休眠侧当轮重睡（见 reconcileHibernateProject）。
func (repo *Tasks) releaseHibernateProject(ctx context.Context, proj *biz.Project, rec *biz.ProjectHibernate) {
	repo.logger.Warningf("[hibernate]: 项目 %s 标记休眠中但副本字段非处休眠零值，释放接管权并转清醒", proj.Name)
	if _, err := repo.hibernateRepo.Upsert(ctx, &biz.ProjectHibernate{
		ProjectID:   proj.ID,
		Enabled:     rec.Enabled,
		State:       biz.HibernateStateAwake,
		WindowNames: rec.WindowNames,
	}); err != nil {
		repo.logger.Errorf("[hibernate]: 项目 %s 释放休眠接管权落库失败: %v", proj.Name, err)
	}
}

// resolveHibernateWindows 按名字列表逐个解析全局时间窗，返回全部可解析的窗。名字全空、
// 或全部名字都悬空（被删/导出到未同步的环境）时返回 ok=false——调用方按「无窗口」处理：
// 休眠侧跳过、唤醒侧继续唤醒，避免悬空引用让项目永久卡在休眠态。单个名字悬空逐个跳过
// （同名只告警一次，避免每轮刷屏），不牵连其余可解析的窗。
func (repo *Tasks) resolveHibernateWindows(ctx context.Context, names []string) (out []*biz.HibernateWindow, ok bool) {
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		win, err := repo.hibernateWindowRepo.GetByName(ctx, name)
		if err != nil {
			repo.logger.Warningf("[hibernate]: 解析休眠时间窗 %q 失败: %v", name, err)
			continue
		}
		if win == nil {
			// 悬空引用是预期态（窗口被删/跨环境未同步），对账每 2 分钟一轮、逐轮告警会刷屏淹没真信号；
			// 同一窗名只告警一次（去重），保留可观测性又去掉噪声。
			if _, loaded := repo.warnedMissingWindow.LoadOrStore(name, struct{}{}); !loaded {
				repo.logger.Warningf("[hibernate]: 引用的休眠时间窗 %q 不存在，跳过（同名不再重复告警）", name)
			}
			continue
		}
		out = append(out, win)
	}
	return out, len(out) > 0
}

// anyHibernateWindowHit 判定 now 是否落在任一已解析窗内（项目多选窗的并集语义）。
func anyHibernateWindowHit(now time.Time, wins []*biz.HibernateWindow) bool {
	for _, w := range wins {
		if biz.InHibernateWindow(now, w.Start, w.End) {
			return true
		}
	}
	return false
}

// loadHibernateTarget 加载休眠对账所需的项目与其所属仓库配置；任一步失败记日志返回 ok=false，
// 由调用方跳过本轮（下一轮重试），避免单个坏项目阻断整轮对账。
// 项目已删（NotFound）时顺手清理孤儿休眠记录：项目可能经空间级联软删离场，那条路径不经过
// projectRepo.Delete，不清理会让本任务每轮都去 Show 一个已删项目、反复告警。
func (repo *Tasks) loadHibernateTarget(ctx context.Context, projectID int) (*biz.Project, *mars.Config, bool) {
	proj, err := repo.projectRepo.Show(ctx, projectID)
	if err != nil {
		if errs.IsNotFound(err) {
			if derr := repo.hibernateRepo.Delete(ctx, projectID); derr != nil {
				repo.logger.Warningf("[hibernate]: 清理项目 %d 的孤儿休眠记录失败: %v", projectID, derr)
			}
			return nil, nil, false
		}
		repo.logger.Warningf("[hibernate]: 加载项目 %d 失败: %v", projectID, err)
		return nil, nil, false
	}
	r, err := repo.repoRepo.Get(ctx, proj.RepoID)
	if err != nil {
		repo.logger.Warningf("[hibernate]: 加载项目 %s 的仓库 %d 失败: %v", proj.Name, proj.RepoID, err)
		return nil, nil, false
	}
	return proj, r.GetMarsConfig(), true
}

// DiskInfo 返回文件仓库的磁盘用量。
func (repo *Tasks) DiskInfo() (int64, error) {
	return repo.fileRepo.DiskInfo(true)
}

// CleanUploadFiles 清理昨日文件：删除物理文件已不在存储的孤儿记录，
// 并删除目录中未被数据库记录的游离文件，最后写入删除审计。
func (repo *Tasks) CleanUploadFiles() error {
	var (
		filesMap = make(map[string]struct{})

		clearList     = make(listFiles, 0)
		upldr         = repo.up
		localUploader = repo.up.LocalUploader()

		yesterday  = repo.timer.Now().Add(-24 * time.Hour)
		startOfDay = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.Local)
		endOfDay   = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, time.Local)

		cleanFunc = func(up uploader.Uploader, fileID int, filePath string) bool {
			if !up.Exists(filePath) {
				if err := repo.fileRepo.DeleteRecord(context.Background(), fileID); err != nil {
					repo.logger.Warningf("[CleanUploadFiles]: delete record %d err %v", fileID, err)
				}
				return true
			}
			return false
		}
	)

	files, err := repo.fileRepo.ListByCreatedAtRange(context.Background(), startOfDay, endOfDay)
	if err != nil {
		return err
	}
	for _, f := range files {
		var deleted bool
		switch f.UploadType {
		case upldr.Type():
			deleted = cleanFunc(upldr, f.ID, f.Path)
		case localUploader.Type():
			deleted = cleanFunc(localUploader, f.ID, f.Path)
		}
		if deleted {
			clearList = append(clearList, f)
		}
		filesMap[f.Path] = struct{}{}
	}

	fn := func(up uploader.Uploader, filesMap map[string]struct{}) error {
		directoryFiles, _ := up.AllDirectoryFiles("")

		for _, file := range directoryFiles {
			if file.LastModified().Before(endOfDay) && file.LastModified().After(startOfDay) {
				_, ok := filesMap[file.Path()]
				if !ok {
					clearList = append(clearList, &biz.File{Path: file.Path(), HumanizeSize: humanize.Bytes(file.Size())})
					if err := up.Delete(file.Path()); err != nil {
						repo.logger.Error(err)
					}
				}
			}
		}
		return nil
	}
	var ups = []uploader.Uploader{localUploader}
	if upldr.Type() != schematype.Local {
		ups = append(ups, upldr)
	}
	for _, up := range ups {
		fn(up, filesMap)
	}

	localUploader.RemoveEmptyDir()
	repo.event.AuditLogWithChange(
		types.EventActionType_Delete,
		"system",
		"",
		"删除未被记录的文件",
		clearList,
		nil,
	)
	return nil
}

type listFiles []*biz.File

type item struct {
	Name string `yaml:"name"`
	Size string `yaml:"size"`
}

// PrettyYaml 把文件列表格式化为 name/size 的 YAML 文本。
func (l listFiles) PrettyYaml() string {
	var items = make([]item, 0, len(l))
	for _, f := range l {
		items = append(items, item{
			Name: f.Path,
			Size: f.HumanizeSize,
		})
	}
	marshal, _ := yaml.Marshal(items)
	return string(marshal)
}

// CacheAllProjects 拉取全部 git 项目并写入缓存。
func (repo *Tasks) CacheAllProjects() error {
	repo.logger.Info("CacheAllProjects")
	_, err := repo.gitRepo.AllProjects(context.Background(), true)
	return err
}

// CacheClusterBoard 每 30s 强制刷新集群看板快照缓存，避免读端高峰期触达全集群
// List。force=true 使 data 层跳过缓存读直接回填。失败即上抛，由 cron 记录，
// 下一轮自动重试。
func (repo *Tasks) CacheClusterBoard() error {
	_, err := repo.k8sRepo.ClusterBoard(context.Background(), true)
	return err
}

// CacheResourceSnapshot 每 5m 强制刷新空间资源快照缓存（cron 节奏比 10min 缓存
// TTL 更密，读端总命中热缓存），避免读端高峰期触达全集群 List。force=true 使
// data 层跳过缓存读直接回填。失败即上抛，由 cron 记录，下一轮自动重试。
func (repo *Tasks) CacheResourceSnapshot() error {
	_, err := repo.k8sRepo.ResourceSnapshot(context.Background(), true)
	return err
}

// CacheClusterInfo 每 30s 强制刷新集群信息统计缓存（cron 节奏比 10min 缓存 TTL
// 更密，读端总命中热缓存），避免顶栏 ClusterInfo 首次请求触达全集群 List（该缓存
// 原无 cron 预热，重启后首个请求必冷）。force=true 使 data 层跳过缓存读直接回填。
// 失败即上抛，由 cron 记录，下一轮自动重试。
func (repo *Tasks) CacheClusterInfo() error {
	_, err := repo.k8sRepo.RefreshClusterInfo()
	return err
}
