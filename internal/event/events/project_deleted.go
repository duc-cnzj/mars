package events

import (
	websocket_pb "github.com/duc-cnzj/mars-client/v4/websocket"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
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
