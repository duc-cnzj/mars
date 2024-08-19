package repo

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/cron"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/file"
	"github.com/duc-cnzj/mars/v4/internal/ent/namespace"
	"github.com/duc-cnzj/mars/v4/internal/ent/project"
	"github.com/duc-cnzj/mars/v4/internal/ent/schema/schematype"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/uploader"
	"github.com/duc-cnzj/mars/v4/internal/util/k8s"
	"github.com/duc-cnzj/mars/v4/internal/util/mars"
	"github.com/duc-cnzj/mars/v4/internal/util/serialize"
	"github.com/dustin/go-humanize"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CronRepo interface {
	CleanUploadFiles() error
	FixDeployStatus() error
	DiskInfo() (int64, error)
	CacheAllBranches() error
	SyncImagePullSecrets() error
}

var _ CronRepo = (*cronRepo)(nil)

type cronRepo struct {
	logger      mlog.Logger
	event       EventRepo
	data        data.Data
	up          uploader.Uploader
	helm        HelmerRepo
	gitRepo     GitRepo
	cronManager cron.Manager
	pluginMgr   application.PluginManger
	k8sRepo     K8sRepo
	nsRepo      NamespaceRepo
	repoRepo    RepoRepo
	cache       cache.Cache
	fileRepo    FileRepo
}

func NewCronRepo(logger mlog.Logger, fileRepo FileRepo, cache cache.Cache, repoRepo RepoRepo, nsRepo NamespaceRepo, k8sRepo K8sRepo, pluginMgr application.PluginManger, event EventRepo, data data.Data, up uploader.Uploader, helm HelmerRepo, gitRepo GitRepo, cronManager cron.Manager) CronRepo {
	cr := &cronRepo{
		fileRepo:    fileRepo,
		logger:      logger.WithModule("repo/cron"),
		event:       event,
		data:        data,
		up:          up,
		helm:        helm,
		gitRepo:     gitRepo,
		cronManager: cronManager,
		pluginMgr:   pluginMgr,
		k8sRepo:     k8sRepo,
		nsRepo:      nsRepo,
		repoRepo:    repoRepo,
		cache:       cache,
	}

	cronManager.NewCommand("clean_upload_files", cr.CleanUploadFiles).DailyAt("2:00")
	cronManager.NewCommand("disk_info", func() error {
		_, err := cr.DiskInfo()
		return err
	}).EveryFifteenMinutes()
	cronManager.NewCommand("fix_project_deploy_status", cr.FixDeployStatus).EveryTwoMinutes()

	if data.Config().GitServerCached {
		cronManager.NewCommand("all_branch_cache", cr.CacheAllBranches).EveryTwoMinutes()
	}

	cronManager.NewCommand("sync_domain_secret", cr.SyncDomainSecret).EveryMinute()

	logger.Warning(data.Config())
	if data.Config().KubeConfig != "" {
		logger.Warning("k8s cron job enabled")
		cronManager.NewCommand("sync_image_pull_secrets", cr.SyncImagePullSecrets).EveryFiveMinutes()
		cronManager.NewCommand("project_pod_event_listener", cr.ProjectPodEventListener).EveryFiveSeconds()
	}

	return cr
}

func (repo *cronRepo) CacheAllBranches() error {
	defer func(t time.Time) {
		repo.logger.Debug("CacheAllBranches done", time.Since(t))
	}(time.Now())

	var wg = &sync.WaitGroup{}
	all, err := repo.repoRepo.All(context.TODO(), &AllRepoRequest{Enabled: lo.ToPtr(true), NeedGitRepo: lo.ToPtr(true)})
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
			for {
				select {
				case id, ok := <-ch:
					if !ok {
						return
					}
					repo.gitRepo.AllBranches(context.TODO(), int(id))
				}
			}
		}()
	}
	for _, it := range all {
		ch <- it.GitProjectID
	}
	close(ch)
	wg.Wait()

	return nil
}

// SyncImagePullSecrets
// 少了自动加上，更新了自动修改
// 自动同步包括
// 1. db image_pull_secrets 丢失(不会自动删除之前的 secret)
// 2. k8s secrets 丢失
// 3. 密码更新
// 4. 删除 config
// 4. 新增 config
func (repo *cronRepo) SyncImagePullSecrets() error {
	var (
		namespaceList       []*ent.Namespace
		cfgImagePullSecrets = repo.data.Config().ImagePullSecrets
		k8sClient           = repo.data.K8sClient()
		db                  = repo.data.DB()
		logger              = repo.logger
	)
	var serverMap = make(map[string]k8s.DockerConfigEntry)
	for _, s := range cfgImagePullSecrets {
		serverMap[s.Server] = k8s.DockerConfigEntry{
			Username: s.Username,
			Password: s.Password,
			Email:    s.Email,
			Auth:     base64.StdEncoding.EncodeToString([]byte(s.Username + ":" + s.Password)),
		}
	}
	namespaceList, _ = db.Namespace.Query().Select(
		namespace.FieldID,
		namespace.FieldName,
		namespace.FieldImagePullSecrets,
	).All(context.TODO())
	for _, namespace := range namespaceList {
		var (
			checked = make(map[string]struct{})
			missing config.DockerAuths
			ns      = namespace
		)
		for _, secretName := range ns.ImagePullSecrets {
			secret, err := k8sClient.SecretLister.Secrets(ns.Name).Get(secretName)
			if err != nil {
				logger.Warningf("[syncImagePullSecrets]: error get secret '%s', err %v", secretName, err)
				if apierrors.IsNotFound(err) {
					repo.deleteSecret(ns, secretName)
				}
				continue
			}
			if secret.Type == corev1.SecretTypeDockerConfigJson {
				var dockerJsonKeyData []byte = secret.Data[corev1.DockerConfigJsonKey]
				res, err := k8s.DecodeDockerConfigJSON(dockerJsonKeyData)
				if err != nil {
					logger.Warningf("[syncImagePullSecrets]: decode secret '%s', err %v", secretName, err)
					continue
				}
				var newConfigJson = k8s.DockerConfigJSON{
					Auths:       map[string]k8s.DockerConfigEntry{},
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
					repo.deleteSecret(ns, secretName)
					continue
				}

				if !reflect.DeepEqual(newConfigJson.Auths, res.Auths) {
					logger.Warningf("[syncImagePullSecrets]: Find Diff, Auto Sync: '%s'", secretName)
					marshal, _ := json.Marshal(&newConfigJson)
					secret.Data[corev1.DockerConfigJsonKey] = marshal
					k8sClient.Client.CoreV1().Secrets(ns.Name).Update(context.TODO(), secret, metav1.UpdateOptions{})
				}
			}
		}

		for s, cfg := range serverMap {
			if _, ok := checked[s]; !ok {
				missing = append(missing, &config.DockerAuth{
					Username: cfg.Username,
					Password: cfg.Password,
					Email:    cfg.Email,
					Server:   s,
				})
			}
		}

		if len(missing) > 0 {
			secret, err := k8s.CreateDockerSecrets(k8sClient.Client, ns.Name, missing)
			if err == nil {
				logger.Warningf("[syncImagePullSecrets]: Missing %v", missing)

				ns.Update().SetImagePullSecrets(append(ns.ImagePullSecrets, secret.Name)).SaveX(context.TODO())
			}
		}
	}
	return nil
}

func (repo *cronRepo) deleteSecret(ns *ent.Namespace, secretName string) {
	logger := repo.logger
	client := repo.data.K8sClient().Client
	logger.Warningf("[syncImagePullSecrets]: DELETE: %s", secretName)

	client.CoreV1().Secrets(ns.Name).Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	var newNsArray []string
	for _, name := range ns.ImagePullSecrets {
		if name != secretName {
			newNsArray = append(newNsArray, name)
		}
	}
	ns.Update().SetImagePullSecrets(newNsArray).SaveX(context.TODO())
}

// SyncDomainSecret 定期同步域名证书，比如配置文件发生变更，或者源证书发生变更
func (repo *cronRepo) SyncDomainSecret() error {
	var (
		changed        bool
		changedSecrets []*corev1.Secret

		k8sCli = repo.data.K8sClient()
	)
	secretName, tlsKey, tlsCrt := repo.pluginMgr.Domain().GetCerts()
	if secretName != "" && tlsKey != "" && tlsCrt != "" {
		all, err := repo.nsRepo.All(context.TODO(), &AllNamespaceInput{})
		if err != nil {
			repo.logger.Error(err)
			return err
		}
		for _, n := range all {
			secret, err := k8sCli.SecretLister.Secrets(n.Name).Get(secretName)
			if err == nil {
				if apierrors.IsNotFound(err) {
					repo.logger.Infof("[TLS]: Register secret namespace: %s, name %s.", n.Name, secretName)
					if _, err := repo.k8sRepo.AddTlsSecret(n.Name, secretName, tlsKey, tlsCrt); err != nil {
						repo.logger.Error(err)
					}
					continue
				}
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
			_, err := k8sCli.Client.CoreV1().Secrets(secret.Namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
			if err == nil {
				repo.logger.Infof("[TLS]: namespace: %s, name %s updated", secret.Namespace, secret.Name)
			}
		}
	}
	return nil
}

// FixDeployStatus 当 project helm 状态为异常的时候，自动去查询状态并且修复它(当人工手动把 helm 恢复时)
func (repo *cronRepo) FixDeployStatus() error {
	var db = repo.data.DB()
	projects := db.Project.Query().WithNamespace(func(query *ent.NamespaceQuery) {
		query.Select(namespace.FieldID, namespace.FieldName)
	}).Where(project.DeployStatusIn(types.Deploy_StatusFailed, types.Deploy_StatusUnknown)).AllX(context.TODO())

	var status types.Deploy
	for _, project := range projects {
		pp := project
		status = repo.helm.ReleaseStatus(pp.Name, pp.Edges.Namespace.Name)
		if status != types.Deploy_StatusFailed && status != types.Deploy_StatusUnknown {
			pp.Update().SetDeployStatus(status).SaveX(context.TODO())
		}
	}
	return nil
}

var DirSizeCacheSeconds = int((15 * time.Minute).Seconds())

func (repo *cronRepo) DiskInfo() (int64, error) {
	return repo.fileRepo.DiskInfo()
}

func int64ToByte(i int64) []byte {
	return []byte(fmt.Sprintf("%d", i))
}

func byteToInt64(remember []byte) int64 {
	atoi, _ := strconv.Atoi(string(remember))
	return int64(atoi)
}

// CleanUploadFiles
//
// 1. 删除在数据库中存在，但是磁盘里面没有的文件
// 2. 删除本地磁盘有的，但是不存在于数据库中的文件
//
// dangerous !
func (repo *cronRepo) CleanUploadFiles() error {
	var (
		filesMap = make(map[string]struct{})

		db            = repo.data.DB()
		clearList     = make(listFiles, 0)
		upldr         = repo.up
		localUploader = repo.up.LocalUploader()

		// 因为执行时间是凌晨 2:00 所以清除的前一天的文件
		yesterday  = time.Now().Add(-24 * time.Hour)
		startOfDay = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.Local)
		endOfDay   = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, time.Local)

		cleanFunc = func(up uploader.Uploader, db *ent.Client, fileID int, filePath string) bool {
			if !up.Exists(filePath) {
				db.File.DeleteOneID(fileID).ExecX(context.TODO())
				return true
			}
			return false
		}
	)

	files := db.File.Query().Where(file.CreatedAtGTE(startOfDay), file.CreatedAtLTE(endOfDay)).AllX(context.TODO())

	for _, f := range serialize.Serialize(files, ToFile) {
		var deleted bool
		switch f.UploadType {
		case upldr.Type():
			deleted = cleanFunc(upldr, db, f.ID, f.Path)
		case localUploader.Type():
			deleted = cleanFunc(localUploader, db, f.ID, f.Path)
		}
		if deleted {
			clearList = append(clearList, f)
		}
		filesMap[f.Path] = struct{}{}
	}

	// 删除本地磁盘有的，但是不存在于数据库中的文件
	fn := func(up uploader.Uploader, filesMap map[string]struct{}) error {
		directoryFiles, _ := up.AllDirectoryFiles("")

		for _, file := range directoryFiles {
			if file.LastModified().Before(endOfDay) && file.LastModified().After(startOfDay) {
				_, ok := filesMap[file.Path()]
				if !ok {
					clearList = append(clearList, &File{Path: file.Path(), HumanizeSize: humanize.Bytes(file.Size())})
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

type listFiles []*File

type item struct {
	Name string `yaml:"name"`
	Size string `yaml:"size"`
}

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

func (repo *cronRepo) ProjectPodEventListener() error {
	var cfg = repo.data.Config()
	var ws = repo.pluginMgr.Ws()
	db := repo.data.DB()
	logger := repo.logger
	ch := make(chan data.Obj[*corev1.Pod], 100)
	listener := "pod-watcher"
	namespacePublisher := ws.New("", "").(application.ProjectPodEventPublisher)
	podFanOut := repo.data.K8sClient().PodFanOut
	podFanOut.AddListener(listener, ch)

	logger.Warning("[PodEventListener]: start pod-watcher")

	defer logger.HandlePanic(listener)
	defer podFanOut.RemoveListener(listener)

	for {
		select {
		case obj, ok := <-ch:
			if !ok {
				repo.logger.Warning("[PodEventListener]: pod-watcher channel closed")
				return nil
			}
			switch obj.Type() {
			case data.Update:
				if obj.Old().Status.Phase != obj.Current().Status.Phase || containerStatusChanged(logger, obj.Old(), obj.Current()) {
					logger.Debugf("old: '%s' new '%s'", obj.Old().Status.Phase, obj.Current().Status.Phase)
					if ns, err := db.Namespace.Query().Where(namespace.NameEQ(mars.GetMarsNamespace(obj.Current().Namespace, cfg.NsPrefix))).Only(context.TODO()); err == nil {
						if err := namespacePublisher.Publish(int64(ns.ID), obj.Current()); err != nil {
							logger.Errorf("[PodEventListener]: %v", err)
						}
					}
				}
			case data.Add, data.Delete:
				logger.Warning("[PodEventListener]: add pod", obj.Type(), obj.Current().Name, obj.Current().Namespace)
				if ns, err := db.Namespace.Query().Where(namespace.NameEQ(mars.GetMarsNamespace(obj.Current().Namespace, cfg.NsPrefix))).Only(context.TODO()); err == nil {
					logger.Debugf("[PodEventListener]: pod '%v': '%s' '%s' '%d' '%s'", obj.Type(), obj.Current().Name, obj.Current().Namespace, ns.ID, obj.Current().Status.Phase)
					if err := namespacePublisher.Publish(int64(ns.ID), obj.Current()); err != nil {
						logger.Errorf("[PodEventListener]: %v", err)
					}
				}
			default:
			}
		}
	}
}

type watchContainerStatus struct {
	Ready bool
}

func containerStatusChanged(logger mlog.Logger, old *corev1.Pod, current *corev1.Pod) bool {
	if len(old.Status.ContainerStatuses) != len(current.Status.ContainerStatuses) {
		return true
	}
	var oldMap = map[string]watchContainerStatus{}
	for _, status := range old.Status.ContainerStatuses {
		oldMap[status.Name] = watchContainerStatus{
			Ready: status.Ready,
		}
	}
	var currentMap = map[string]watchContainerStatus{}
	for _, status := range current.Status.ContainerStatuses {
		currentMap[status.Name] = watchContainerStatus{
			Ready: status.Ready,
		}
	}

	for k, v := range currentMap {
		if b, ok := oldMap[k]; !(ok && b == v) {
			logger.Debugf("ContainerStatus old: %v current: %v", b, v)
			return true
		}
	}

	return false
}
