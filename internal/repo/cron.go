package repo

import (
	"context"
	"sync"
	"time"

	"github.com/samber/lo"

	"github.com/duc-cnzj/mars/v4/internal/application"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/cron"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/file"
	"github.com/duc-cnzj/mars/v4/internal/ent/namespace"
	"github.com/duc-cnzj/mars/v4/internal/ent/project"
	"github.com/duc-cnzj/mars/v4/internal/ent/schema/schematype"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/uploader"
	"github.com/duc-cnzj/mars/v4/internal/util/serialize"
	"github.com/dustin/go-humanize"
	"gopkg.in/yaml.v3"
)

type CronRepo interface {
	CleanUploadFiles() error
	FixDeployStatus() error
	DiskInfo() (int64, error)
	CacheAllBranches() error
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
	repoRepo    RepoImp
}

func NewCronRepo(logger mlog.Logger, repoRepo RepoImp, nsRepo NamespaceRepo, k8sRepo K8sRepo, pluginMgr application.PluginManger, event EventRepo, data data.Data, up uploader.Uploader, helm HelmerRepo, gitRepo GitRepo, cronManager cron.Manager) CronRepo {
	cr := &cronRepo{logger: logger, repoRepo: repoRepo, nsRepo: nsRepo, k8sRepo: k8sRepo, event: event, pluginMgr: pluginMgr, data: data, up: up, helm: helm, gitRepo: gitRepo, cronManager: cronManager}

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

// SyncDomainSecret 定期同步域名证书，比如配置文件发生变更，或者源证书发生变更
func (repo *cronRepo) SyncDomainSecret() error {
	var (
		changed        bool
		changedSecrets []*corev1.Secret

		k8sCli = repo.data.K8sClient()
	)
	secretName, tlsKey, tlsCrt := repo.pluginMgr.Domain().GetCerts()
	if secretName != "" && tlsKey != "" && tlsCrt != "" {
		all, err := repo.nsRepo.All(context.TODO())
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

func (repo *cronRepo) DiskInfo() (int64, error) {
	return repo.up.DirSize()
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
