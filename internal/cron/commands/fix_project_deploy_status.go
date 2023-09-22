package commands

import (
	"github.com/duc-cnzj/mars-client/v4/types"
	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/cron"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"gorm.io/gorm"
)

func init() {
	cron.Register(func(manager contracts.CronManager, app contracts.ApplicationInterface) {
		manager.NewCommand("fix_project_deploy_status", fixDeployStatus).EveryTwoMinutes()
	})
}

// fixDeployStatus 当 project helm 状态为异常的时候，自动去查询状态并且修复它(当人工手动把 helm 恢复时)
func fixDeployStatus() error {
	var projects []models.Project
	app.DB().Preload("Namespace", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Where(`deploy_status in ?`, []types.Deploy{types.Deploy_StatusFailed, types.Deploy_StatusUnknown}).Find(&projects)
	for _, project := range projects {
		status := app.Helmer().ReleaseStatus(project.Name, project.Namespace.Name)
		if status != types.Deploy_StatusFailed && status != types.Deploy_StatusUnknown {
			app.DB().Model(&project).Updates(map[string]any{
				"deploy_status": status,
			})
		}
	}
	return nil
}
