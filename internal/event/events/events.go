package events

import (
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	websocket_pb "github.com/duc-cnzj/mars/pkg/websocket"
	v1 "k8s.io/api/core/v1"
)

var (
	EventNamespaceCreated contracts.Event = "namespace_created"
	EventNamespaceDeleted contracts.Event = "namespace_deleted"

	EventProjectDeleted contracts.Event = "project_deleted"
	EventProjectChanged contracts.Event = "project_changed"
)

type ProjectChangedData struct {
	Project *models.Project

	Manifest string
	Config   string
	Username string
}

type NamespaceCreatedData struct {
	NsModel  *models.Namespace
	NsK8sObj *v1.Namespace
}

type NamespaceDeletedData struct {
	NsModel *models.Namespace
}

func HandleNamespaceDeleted(data interface{}, e contracts.Event) error {
	plugins.GetWsSender().New("", "").ToAll(&plugins.WsResponseMetadata{Metadata: &websocket_pb.ResponseMetadata{Type: websocket_pb.Type_ReloadProjects}})
	mlog.Debug("event handled: ", e.String())

	return nil
}

func HandleProjectDeleted(data interface{}, e contracts.Event) error {
	plugins.GetWsSender().New("", "").ToAll(&plugins.WsResponseMetadata{Metadata: &websocket_pb.ResponseMetadata{Type: websocket_pb.Type_ReloadProjects}})
	mlog.Debug("event handled: ", e.String(), data)

	return nil
}

func HandleProjectChanged(data interface{}, e contracts.Event) error {
	if changedData, ok := data.(*ProjectChangedData); ok {
		last := &models.Changelog{}
		app.DB().Select("Config", "ID", "Version").Order("`id` desc").First(&last)
		gp := models.GitlabProject{}
		app.DB().Select("ID", "GitlabProjectId").Where("`gitlab_project_id` = ?", changedData.Project.GitlabProjectId).First(&gp)
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
		app.DB().Create(&models.Changelog{
			Version:         version,
			ConfigChanged:   configChanged,
			Username:        changedData.Username,
			Manifest:        changedData.Manifest,
			Config:          changedData.Config,
			ProjectID:       changedData.Project.ID,
			GitlabProjectID: gp.ID,
		})
	}
	return nil
}
