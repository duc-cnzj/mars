package events

import (
	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/plugins"
)

const EventProjectDeleted contracts.Event = "project_deleted"

func init() {
	Register(EventProjectDeleted, HandleProjectDeleted)
}

func HandleProjectDeleted(data any, e contracts.Event) error {
	project := data.(*models.Project)
	plugins.GetWsSender().New("", "").ToAll(&websocket_pb.WsReloadProjectsResponse{
		Metadata:    &websocket_pb.Metadata{Type: websocket_pb.Type_ReloadProjects},
		NamespaceId: int64(project.NamespaceId),
	})
	mlog.Debug("event handled: ", e.String(), data)

	return nil
}
