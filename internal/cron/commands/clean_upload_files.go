package commands

import (
	"github.com/duc-cnzj/mars-client/v4/types"
	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/cron"
	"github.com/duc-cnzj/mars/v4/internal/event/events"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/models"

	"github.com/dustin/go-humanize"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	"time"
)

func init() {
	cron.Register(func(manager contracts.CronManager, app contracts.ApplicationInterface) {
		manager.NewCommand("clean_upload_files", cleanUploadFiles).DailyAt("2:00")
	})
}

type listFiles []*types.FileModel

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

// cleanUploadFiles
//
// 1. 删除在数据库中存在，但是磁盘里面没有的文件
// 2. 删除本地磁盘有的，但是不存在于数据库中的文件
//
// dangerous !
func cleanUploadFiles() error {
	var (
		files []models.File

		filesMap = make(map[string]struct{})

		clearList     = make(listFiles, 0)
		uploader      = app.Uploader()
		localUploader = app.LocalUploader()

		// 因为执行时间是凌晨 2:00 所以清除的前一天的文件
		yesterday     = time.Now().Add(-24 * time.Hour)
		startOfDay    = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.Local)
		startOfDayStr = startOfDay.Format("2006-01-02 15:04:05")
		endOfDay      = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, time.Local)
		endOfDayStr   = endOfDay.Format("2006-01-02 15:04:05")

		cleanFunc = func(up contracts.Uploader, db *gorm.DB, fileID int, filePath string) bool {
			if !up.Exists(filePath) {
				db.Delete(&models.File{ID: fileID})
				return true
			}
			return false
		}
	)

	// 删除在数据库中存在，但是磁盘里面没有的文件
	app.DB().Where("`created_at` between ? AND ?", startOfDayStr, endOfDayStr).FindInBatches(&files, 100, func(tx *gorm.DB, batch int) error {
		for _, f := range files {
			var deleted bool
			switch f.UploadType {
			case uploader.Type():
				deleted = cleanFunc(uploader, tx, f.ID, f.Path)
			case localUploader.Type():
				deleted = cleanFunc(localUploader, tx, f.ID, f.Path)
			}
			if deleted {
				clearList = append(clearList, f.ProtoTransform())
			}
			filesMap[f.Path] = struct{}{}
		}
		return nil
	})

	// 删除本地磁盘有的，但是不存在于数据库中的文件
	fn := func(up contracts.Uploader, filesMap map[string]struct{}) error {
		directoryFiles, _ := up.AllDirectoryFiles("")

		for _, file := range directoryFiles {
			if file.LastModified().Before(endOfDay) && file.LastModified().After(startOfDay) {
				_, ok := filesMap[file.Path()]
				if !ok {
					clearList = append(clearList, &types.FileModel{Path: file.Path(), HumanizeSize: humanize.Bytes(file.Size())})
					if err := up.Delete(file.Path()); err != nil {
						mlog.Error(err)
					}
				}
			}
		}
		return nil
	}
	var ups = []contracts.Uploader{app.LocalUploader()}
	if app.Uploader().Type() != contracts.Local {
		ups = append(ups, app.Uploader())
	}
	for _, up := range ups {
		fn(up, filesMap)
	}

	localUploader.RemoveEmptyDir()
	events.AuditLog("cronjob", types.EventActionType_Delete, "删除未被记录的文件", clearList, nil)
	return nil
}
