package contracts

import (
	"context"

	"github.com/duc-cnzj/mars-client/v4/websocket"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
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

type ProjectPodEventSubscriber interface {
	Join(projectID int64) error
	Leave(nsID int64, projectID int64) error
	Run(ctx context.Context) error
}

type ProjectPodEventPublisher interface {
	Publish(nsID int64, pod *corev1.Pod) error
}
