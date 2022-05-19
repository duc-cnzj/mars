package events

import (
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/models"
)

const EventProjectChanged contracts.Event = "project_changed"

type ProjectChangedData struct {
	Project *models.Project

	Manifest string
	Config   string
	Username string
}

func init() {
	Register(EventProjectChanged, HandleProjectChanged)
}

func HandleProjectChanged(data any, e contracts.Event) error {
	if changedData, ok := data.(*ProjectChangedData); ok {
		last := &models.Changelog{}
		app.DB().Select("config", "id", "version", "git_project_id", "project_id").Where("`project_id` = ?", changedData.Project.ID).Order("`id` desc").First(&last)
		gp := models.GitProject{}
		app.DB().Select("id", "git_project_id").Where("`git_project_id` = ?", changedData.Project.GitProjectId).First(&gp)
		var (
			configChanged bool
			version       uint8 = 1
		)
		if last != nil {
			if last.Config != changedData.Config {
				configChanged = true
			}
			version = last.Version + 1
		}
		if err := app.DB().Create(&models.Changelog{
			Version:       version,
			ConfigChanged: configChanged,
			Username:      changedData.Username,
			Manifest:      changedData.Manifest,
			Config:        changedData.Config,
			ProjectID:     changedData.Project.ID,
			GitProjectID:  gp.ID,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
