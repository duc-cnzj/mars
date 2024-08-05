package repo

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/util/serialize"

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
	"github.com/dustin/go-humanize"
	"gopkg.in/yaml.v3"
)

type CronRepo interface {
	CleanUploadFiles() error
	FixDeployStatus() error
	DiskInfo() (int64, error)
	CacheAllGitProjects() error
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
}

func NewCronRepo(logger mlog.Logger, event EventRepo, data data.Data, up uploader.Uploader, helm HelmerRepo, gitRepo GitRepo, cronManager cron.Manager) CronRepo {
	cr := &cronRepo{logger: logger, event: event, data: data, up: up, helm: helm, gitRepo: gitRepo, cronManager: cronManager}

	cronManager.NewCommand("clean_upload_files", cr.CleanUploadFiles).DailyAt("2:00")
	cronManager.NewCommand("disk_info", func() error {
		_, err := cr.DiskInfo()
		return err
	}).EveryFifteenMinutes()
	cronManager.NewCommand("fix_project_deploy_status", cr.FixDeployStatus).EveryTwoMinutes()

	if data.Config().GitServerCached {
		cronManager.NewCommand("all_git_project_cache", cr.CacheAllGitProjects).EveryFiveMinutes()
		cronManager.NewCommand("all_branch_cache", cr.CacheAllBranches).EveryTwoMinutes()
	}

	return cr
}

func (repo *cronRepo) CacheAllBranches() error {
	//var (
	//	wg = &sync.WaitGroup{}
	//)
	//
	//db := application.DB()
	//enabledGitProjects := db.GitProject.Query().Where(gitproject.Enabled(true)).AllX(context.TODO())
	//goroutineNum := len(enabledGitProjects)
	//
	//if len(enabledGitProjects) > 10 {
	//	goroutineNum = 8
	//}
	//
	//ch := make(chan *ent.GitProject, goroutineNum)
	//gitServer := plugins.GetGitServer(application, application.Config().GitServerPlugin, application.Config().GitServerCached)
	//if server, ok := gitServer.(plugins.GitCacheServer); ok {
	//	for i := 0; i < goroutineNum; i++ {
	//		wg.Add(1)
	//		go func() {
	//			defer wg.Done()
	//			defer recovery.HandlePanic(application.IsDebug(), "[CRON]: all_branch_cache")
	//			for gitProject := range ch {
	//				err := server.ReCacheAllBranches(fmt.Sprintf("%d", gitProject.GitProjectID))
	//				application.Logger().Debugf("[CRON]: fetch AllBranches: '%s' '%d', err: '%v'", gitProject.Name, gitProject.GitProjectID, err)
	//			}
	//		}()
	//	}
	//	for i := range enabledGitProjects {
	//		ch <- enabledGitProjects[i]
	//	}
	//	close(ch)
	//	wg.Wait()
	//}

	return nil
}

func (repo *cronRepo) CacheAllGitProjects() error {
	//var gitServer plugins.GitServer = gitRepo.GetGitServer(application, application.Config().GitServerPlugin, application.Config().GitServerCached)
	//if cache, ok := gitRepo.(plugins.GitCacheServer); ok {
	//	return repo.gitRepo.ReCacheAllProjects()
	//}
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
