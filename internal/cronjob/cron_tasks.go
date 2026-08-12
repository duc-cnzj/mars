package cronjob

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"maps"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/config"
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
// 原常驻的 Pod 事件监听（ProjectPodEventListener）已归位 eventhandler 包，
// 此处只留真定时任务。
type Tasks struct {
	timer       timer.Timer
	logger      mlog.Logger
	event       biz.EventRepo
	cfg         *config.Config
	up          uploader.Uploader
	helm        biz.HelmerRepo
	gitRepo     biz.GitRepo
	getCerts    func() (name, key, crt string)
	k8sRepo     biz.K8sRepo
	nsRepo      biz.NamespaceRepo
	repoRepo    biz.RepoRepo
	fileRepo    biz.FileRepo
	projectRepo biz.ProjectRepo
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
	deps *PluginDeps,
) *Tasks {
	return &Tasks{
		timer:       tm,
		fileRepo:    fileRepo,
		logger:      logger.WithModule("app/cron"),
		event:       event,
		cfg:         cfg,
		up:          up,
		helm:        helm,
		gitRepo:     gitRepo,
		getCerts:    deps.GetCerts,
		k8sRepo:     k8sRepo,
		nsRepo:      nsRepo,
		repoRepo:    repoRepo,
		projectRepo: projectRepo,
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
		status := repo.helm.ReleaseStatus(p.Name, p.Namespace.Name)
		if status != types.Deploy_StatusFailed && status != types.Deploy_StatusUnknown {
			if _, err := repo.projectRepo.UpdateDeployStatus(context.TODO(), p.ID, status); err != nil {
				return err
			}
		}
	}
	return nil
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
