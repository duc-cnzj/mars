package cronjob

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"maps"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/dustin/go-humanize"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

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
	// hibernateWindowRepo 解析仓库引用的全局休眠时间窗（按名字查启止时段）。
	hibernateWindowRepo biz.HibernateWindowRepo
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
	}
}

// CacheAllBranches 并发拉取全部启用 git 仓库的分支并写入缓存（最多 8 个 goroutine）。
func (repo *Tasks) CacheAllBranches() error {
	defer func(t time.Time) {
		repo.logger.Debug("CacheAllBranches done", repo.timer.Since(t))
	}(repo.timer.Now())

	var wg = &sync.WaitGroup{}
	all, err := repo.repoRepo.All(context.TODO(), &biz.AllRepoRequest{Enabled: lo.ToPtr(true), NeedGitRepo: lo.ToPtr(true)})
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
				repo.gitRepo.AllBranches(context.TODO(), int(id), true)
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
	namespaceList, err := repo.nsRepo.ListAll(context.TODO())
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
			secret, err := repo.k8sRepo.GetSecret(context.TODO(), ns.Name, secretName)
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
					if _, err := repo.k8sRepo.UpdateSecret(context.TODO(), ns.Name, secret.Name, secret); err != nil {
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
			secret, err := repo.k8sRepo.CreateDockerSecrets(context.TODO(), ns.Name, missing)
			if err == nil {
				logger.Warningf("[syncImagePullSecrets]: Missing %v", missing)
				if err := repo.nsRepo.UpdateImagePullSecrets(context.TODO(), ns.ID, append(ns.ImagePullSecrets, secret.Name)); err != nil {
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

	if err := repo.k8sRepo.DeleteSecret(context.TODO(), ns.Name, secretName); err != nil {
		logger.Warningf("[syncImagePullSecrets]: delete k8s secret '%s', err %v", secretName, err)
	}
	var newNsArray []string
	for _, name := range ns.ImagePullSecrets {
		if name != secretName {
			newNsArray = append(newNsArray, name)
		}
	}
	if err := repo.nsRepo.UpdateImagePullSecrets(context.TODO(), ns.ID, newNsArray); err != nil {
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
		allNamespaces, err := repo.nsRepo.ListAll(context.TODO())
		if err != nil {
			return err
		}
		for _, n := range allNamespaces {
			secret, err := repo.k8sRepo.GetSecret(context.TODO(), n.Name, secretName)
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
			_, err := repo.k8sRepo.UpdateSecret(context.TODO(), secret.Namespace, secret.Name, secret)
			if err == nil {
				repo.logger.Infof("[TLS]: namespace: %s, name %s updated", secret.Namespace, secret.Name)
			}
		}
	}
	return nil
}

// FixDeployStatus 用 helm 实测状态修复标记为失败/未知的项目。
func (repo *Tasks) FixDeployStatus() error {
	projects, err := repo.projectRepo.ListByDeployStatus(context.TODO(), types.Deploy_StatusFailed, types.Deploy_StatusUnknown)
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
			if _, err := repo.projectRepo.UpdateDeployStatus(context.TODO(), p.ID, status); err != nil {
				return err
			}
		}
	}
	return nil
}

// ReconcileHibernate 对账项目定时休眠：每 2 分钟推进一步状态机。
// 总闸：只接管项目级开关已开（Enabled=true）的记录——ListEnabled 已在查询侧过滤掉开关已关的
// 记录，故开关一关，对账即彻底放弃该项目，无论它当前处于什么状态（含已置 0 的休眠态）都不再做
// 任何修改，交由用户手动掌控；重新打开开关即恢复接管。
// 单趟遍历：不再分「清醒轮 / 休眠轮」两趟，每条记录按「是否在窗口内」与「是否真在休眠（状态
// sleeping 且副本字段全 0）」即时归位——窗口内非真休眠即休眠（含错位当轮重睡，无需等下一轮）、
// 窗口外真休眠即唤醒、窗口外错位则释放接管权。单个项目失败只记日志并跳过（下一轮重试），
// 不阻断本轮其它项目；列表查询失败才上抛。
func (repo *Tasks) ReconcileHibernate() error {
	now := repo.timer.Now()
	enabled, err := repo.hibernateRepo.ListEnabled(context.TODO())
	if err != nil {
		return err
	}
	for _, rec := range enabled {
		repo.reconcileHibernateProject(now, rec)
	}
	return nil
}

// reconcileHibernateProject 推进单条「开关已开」休眠记录的状态：加载项目与所属仓库配置，按
// 窗口内外与真实休眠态的四象限分派到休眠 / 唤醒 / 释放，或原样保持。加载失败由
// loadHibernateTarget 记日志并返回 ok=false，直接跳过（下一轮重试）。
func (repo *Tasks) reconcileHibernateProject(now time.Time, rec *biz.ProjectHibernate) {
	proj, cfg, ok := repo.loadHibernateTarget(rec.ProjectID)
	if !ok {
		return
	}
	// replicaPaths 只要仓库配了休眠就取：即便仓库级能力已关，判断「副本是否真全 0」仍需求值，
	// 以免把「已置 0 的休眠态」误判为错位。未配置休眠时为 nil，ReplicaValuesAllZero 恒真，
	// 退化为只看状态，保证此类项目仍能按快照被唤醒。
	var replicaPaths []string
	if cfg.Hibernate != nil {
		replicaPaths = cfg.Hibernate.ReplicaPaths
	}
	inWindow := false
	if cfg.Hibernate != nil && cfg.Hibernate.Enabled {
		if start, end, wok := repo.resolveHibernateWindow(cfg.Hibernate); wok {
			inWindow = biz.InHibernateWindow(now, start, end)
		}
	}
	actuallySleeping := rec.State == biz.HibernateStateSleeping &&
		biz.ReplicaValuesAllZero(proj.FinalExtraValues, replicaPaths)

	switch {
	case inWindow && actuallySleeping:
		// 窗口内且确在休眠：保持，不做任何改动。
		return
	case inWindow:
		// 清醒态命中窗口，或标记休眠但副本非全 0 的错位：本轮回休眠（错位时当前值成为新快照），
		// 与修正同一轮完成，不再等下一轮对账。
		repo.sleepHibernateProject(proj, rec, replicaPaths)
	case actuallySleeping:
		// 窗口外且确在休眠：按快照回填唤醒。
		repo.wakeHibernateProject(proj, rec)
	case rec.State == biz.HibernateStateSleeping:
		// 窗口外仍标记 sleeping 但副本非全 0：释放接管权（保留当前值）。
		repo.releaseHibernateProject(proj, rec)
	}
}

// sleepHibernateProject 执行单个项目的休眠：快照副本字段当前值并**先**落库「休眠态+快照」，
// 再把副本字段置 0 重部署。窗口判定与「仓库是否启用休眠」由调用方完成，本方法只负责落动作。
//
// 顺序是幂等的承重点：重部署会把置 0 后的副本值写回项目记录（deploy 管线 UpdateProject）。
// 若先部署后落库，一旦部署成功而落库失败/进程崩溃，下一轮对账读到的基座副本已被置 0，会据此
// 重拍出「全 0 快照」覆盖真快照，令原始副本数永久丢失（唤醒只能恢复成 0）。改为先落库后：
// 落库失败则本轮不部署、基座未变，下一轮重拍仍是原值；落库成功而部署失败，下一轮以同一快照
// 重试（基座同样未变）。两种失败都基于「基座未变」而自愈，故顺序不可颠倒。
func (repo *Tasks) sleepHibernateProject(proj *biz.Project, rec *biz.ProjectHibernate, replicaPaths []string) {
	base := proj.FinalExtraValues
	snapshot := biz.SnapshotReplicaValues(base, replicaPaths)
	target := biz.SleepExtraValues(base, replicaPaths)
	if _, err := repo.hibernateRepo.Upsert(context.TODO(), &biz.ProjectHibernate{
		ProjectID: proj.ID,
		Enabled:   rec.Enabled,
		State:     biz.HibernateStateSleeping,
		Snapshot:  snapshot,
	}); err != nil {
		repo.logger.Errorf("[hibernate]: 项目 %s 休眠状态落库失败: %v", proj.Name, err)
		return
	}
	if err := repo.hibernateDeployer.Redeploy(context.TODO(), proj, target); err != nil {
		repo.logger.Errorf("[hibernate]: 项目 %s 休眠重部署失败: %v", proj.Name, err)
	}
}

// wakeHibernateProject 执行单个项目的唤醒：按快照回填副本字段重部署、状态转 awake 并清空快照。
// 仅由对账在「窗口外且确在休眠（副本全 0）」时调用，故无需再判窗口。
func (repo *Tasks) wakeHibernateProject(proj *biz.Project, rec *biz.ProjectHibernate) {
	target := biz.WakeExtraValues(proj.FinalExtraValues, rec.Snapshot)
	if err := repo.hibernateDeployer.Redeploy(context.TODO(), proj, target); err != nil {
		repo.logger.Errorf("[hibernate]: 项目 %s 唤醒重部署失败: %v", proj.Name, err)
		return
	}
	if _, err := repo.hibernateRepo.Upsert(context.TODO(), &biz.ProjectHibernate{
		ProjectID: proj.ID,
		Enabled:   rec.Enabled,
		State:     biz.HibernateStateAwake,
	}); err != nil {
		repo.logger.Errorf("[hibernate]: 项目 %s 唤醒状态落库失败: %v", proj.Name, err)
	}
}

// releaseHibernateProject 释放单个项目的休眠接管权：状态转 awake 并清空快照、保留当前副本值，
// 不重部署。用于「窗口外仍标记 sleeping 但副本非全 0」的错位（关闭期间改过副本数后重开），
// 交回用户当前值掌控；窗口内出现同类错位则由休眠侧当轮重睡（见 reconcileHibernateProject）。
func (repo *Tasks) releaseHibernateProject(proj *biz.Project, rec *biz.ProjectHibernate) {
	repo.logger.Warningf("[hibernate]: 项目 %s 标记休眠中但副本字段非全 0，释放接管权并转清醒", proj.Name)
	if _, err := repo.hibernateRepo.Upsert(context.TODO(), &biz.ProjectHibernate{
		ProjectID: proj.ID,
		Enabled:   rec.Enabled,
		State:     biz.HibernateStateAwake,
	}); err != nil {
		repo.logger.Errorf("[hibernate]: 项目 %s 释放休眠接管权落库失败: %v", proj.Name, err)
	}
}

// resolveHibernateWindow 按仓库休眠配置引用的全局时间窗名解析启止时段。名字为空或窗口
// 不存在（被删/导出到未同步的环境）时返回 ok=false——调用方按「无窗口」处理：休眠侧跳过、
// 唤醒侧继续唤醒，避免悬空引用让项目永久卡在休眠态。
func (repo *Tasks) resolveHibernateWindow(h *mars.Hibernate) (string, string, bool) {
	if h == nil || strings.TrimSpace(h.WindowName) == "" {
		return "", "", false
	}
	win, err := repo.hibernateWindowRepo.GetByName(context.TODO(), h.WindowName)
	if err != nil {
		repo.logger.Warningf("[hibernate]: 解析休眠时间窗 %q 失败: %v", h.WindowName, err)
		return "", "", false
	}
	if win == nil {
		repo.logger.Warningf("[hibernate]: 引用的休眠时间窗 %q 不存在，跳过", h.WindowName)
		return "", "", false
	}
	return win.Start, win.End, true
}

// loadHibernateTarget 加载休眠对账所需的项目与其所属仓库配置；任一步失败记日志返回 ok=false，
// 由调用方跳过本轮（下一轮重试），避免单个坏项目阻断整轮对账。
// 项目已删（NotFound）时顺手清理孤儿休眠记录：项目可能经空间级联软删离场，那条路径不经过
// projectRepo.Delete，不清理会让本任务每轮都去 Show 一个已删项目、反复告警。
func (repo *Tasks) loadHibernateTarget(projectID int) (*biz.Project, *mars.Config, bool) {
	proj, err := repo.projectRepo.Show(context.TODO(), projectID)
	if err != nil {
		if errs.IsNotFound(err) {
			if derr := repo.hibernateRepo.Delete(context.TODO(), projectID); derr != nil {
				repo.logger.Warningf("[hibernate]: 清理项目 %d 的孤儿休眠记录失败: %v", projectID, derr)
			}
			return nil, nil, false
		}
		repo.logger.Warningf("[hibernate]: 加载项目 %d 失败: %v", projectID, err)
		return nil, nil, false
	}
	r, err := repo.repoRepo.Get(context.TODO(), proj.RepoID)
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
				if err := repo.fileRepo.DeleteRecord(context.TODO(), fileID); err != nil {
					repo.logger.Warningf("[CleanUploadFiles]: delete record %d err %v", fileID, err)
				}
				return true
			}
			return false
		}
	)

	files, err := repo.fileRepo.ListByCreatedAtRange(context.TODO(), startOfDay, endOfDay)
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
	_, err := repo.gitRepo.AllProjects(context.TODO(), true)
	return err
}

// CacheClusterBoard 每 30s 强制刷新集群看板快照缓存，避免读端高峰期触达全集群
// List。force=true 使 data 层跳过缓存读直接回填。失败即上抛，由 cron 记录，
// 下一轮自动重试。
func (repo *Tasks) CacheClusterBoard() error {
	_, err := repo.k8sRepo.ClusterBoard(context.TODO(), true)
	return err
}

// CacheResourceSnapshot 每 5m 强制刷新空间资源快照缓存（cron 节奏比 10min 缓存
// TTL 更密，读端总命中热缓存），避免读端高峰期触达全集群 List。force=true 使
// data 层跳过缓存读直接回填。失败即上抛，由 cron 记录，下一轮自动重试。
func (repo *Tasks) CacheResourceSnapshot() error {
	_, err := repo.k8sRepo.ResourceSnapshot(context.TODO(), true)
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
