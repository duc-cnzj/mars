package events

import (
	websocket_pb "github.com/duc-cnzj/mars-client/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/plugins"
)

const EventNamespaceDeleted contracts.Event = "namespace_deleted"

type NamespaceDeletedData struct {
	NsModel *models.Namespace
}

func init() {
	Register(EventNamespaceDeleted, HandleNamespaceDeleted)
}

func HandleNamespaceDeleted(data any, e contracts.Event) error {
	plugins.GetWsSender().New("", "").ToAll(&websocket_pb.WsReloadProjectsResponse{
		Metadata:    &websocket_pb.Metadata{Type: websocket_pb.Type_ReloadProjects},
		NamespaceId: int64(data.(NamespaceDeletedData).NsModel.ID),
	})
	mlog.Debug("event handled: ", e.String())

	return nil
}
