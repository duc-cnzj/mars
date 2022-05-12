package contracts

import (
	websocket_pb "github.com/duc-cnzj/mars-client/v4/websocket"
	"google.golang.org/protobuf/proto"
)

//go:generate mockgen -destination ../mock/mock_plugin.go -package mock github.com/duc-cnzj/mars/internal/contracts PluginInterface

type PluginInterface interface {
	Name() string
	Initialize(args map[string]any) error
	Destroy() error
}

type WebsocketMessage interface {
	proto.Message
	GetMetadata() *websocket_pb.Metadata
}
