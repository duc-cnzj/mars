package contracts

import (
	"github.com/duc-cnzj/mars-client/v4/websocket"
	"google.golang.org/protobuf/proto"
)

//go:generate mockgen -destination ../mock/mock_wssender_pubsub.go -package mock github.com/duc-cnzj/mars/internal/contracts PubSub

type WebsocketMessage interface {
	proto.Message
	GetMetadata() *websocket.Metadata
}

type PubSub interface {
	Info() any
	Uid() string
	ID() string
	ToSelf(WebsocketMessage) error
	ToAll(WebsocketMessage) error
	ToOthers(WebsocketMessage) error
	Subscribe() <-chan []byte
	Close() error
}
