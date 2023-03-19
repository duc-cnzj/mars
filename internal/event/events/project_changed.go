package events

import (
	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/models"
)

const EventProjectChanged contracts.Event = "project_changed"

type ProjectChangedData struct {
	Project *models.Project

	Username string
}

func init() {
	Register(EventProjectChanged, HandleProjectChanged)
}

func HandleProjectChanged(data any, e contracts.Event) error {
	if changedData, ok := data.(*ProjectChangedData); ok {
		last := &models.Changelog{}
		app.DB().
			Select("config", "id", "version", "git_project_id", "project_id", "git_commit").
			Where("`project_id` = ?", changedData.Project.ID).
			Order("`id` desc").
			First(&last)
		gp := models.GitProject{}
		app.DB().Select("id", "git_project_id").Where("`git_project_id` = ?", changedData.Project.GitProjectId).First(&gp)
		var (
			configChanged bool
			version       int = 1
		)
		if last != nil {
			if last.Config != changedData.Project.Config || last.GitCommit != changedData.Project.GitCommit {
				configChanged = true
			}
			version = last.Version + 1
		}
		if err := app.DB().Create(&models.Changelog{
			Version:          version,
			Username:         changedData.Username,
			Manifest:         changedData.Project.Manifest,
			Config:           changedData.Project.Config,
			ConfigType:       changedData.Project.ConfigType,
			GitBranch:        changedData.Project.GitBranch,
			GitCommit:        changedData.Project.GitCommit,
			DockerImage:      changedData.Project.DockerImage,
			EnvValues:        changedData.Project.EnvValues,
			ExtraValues:      changedData.Project.ExtraValues,
			FinalExtraValues: changedData.Project.FinalExtraValues,
			GitCommitWebUrl:  changedData.Project.GitCommitWebUrl,
			GitCommitTitle:   changedData.Project.GitCommitTitle,
			GitCommitAuthor:  changedData.Project.GitCommitAuthor,
			GitCommitDate:    changedData.Project.GitCommitDate,
			ConfigChanged:    configChanged,
			ProjectID:        changedData.Project.ID,
			GitProjectID:     gp.ID,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
