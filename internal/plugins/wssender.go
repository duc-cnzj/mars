package plugins

//go:generate mockgen -destination ../mock/mock_wssender.go -package mock github.com/duc-cnzj/mars/v4/internal/plugins WsSender

import (
	"context"
	"sync"

	v1 "k8s.io/api/core/v1"

	websocket_pb "github.com/duc-cnzj/mars-client/v4/websocket"
	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
)

var wsSenderOnce sync.Once

const (
	ToSelf   = websocket_pb.To_ToSelf
	ToAll    = websocket_pb.To_ToAll
	ToOthers = websocket_pb.To_ToOthers
)

var _ contracts.WebsocketMessage = (*websocket_pb.WsMetadataResponse)(nil)

type WsSender interface {
	contracts.PluginInterface

	New(uid, id string) contracts.PubSub
}

func GetWsSender() WsSender {
	pcfg := app.Config().WsSenderPlugin
	p := app.App().GetPluginByName(pcfg.Name)
	wsSenderOnce.Do(func() {
		if err := p.Initialize(pcfg.GetArgs()); err != nil {
			panic(err)
		}
		app.App().RegisterAfterShutdownFunc(func(applicationInterface contracts.ApplicationInterface) {
			p.Destroy()
		})
	})

	return p.(WsSender)
}

type EmptyPubSub struct{}

func (e *EmptyPubSub) Join(projectID int64) error {
	return nil
}

func (e *EmptyPubSub) Leave(nsID int64, projectID int64) error {
	return nil
}

func (e *EmptyPubSub) Run(ctx context.Context) error {
	return nil
}

func (e *EmptyPubSub) Publish(nsID int64, pod *v1.Pod) error {
	return nil
}

func (e *EmptyPubSub) Info() any {
	return nil
}

func (e *EmptyPubSub) Uid() string {
	return ""
}

func (e *EmptyPubSub) ID() string {
	return ""
}

func (e *EmptyPubSub) ToSelf(message contracts.WebsocketMessage) error {
	return nil
}

func (e *EmptyPubSub) ToAll(message contracts.WebsocketMessage) error {
	return nil
}

func (e *EmptyPubSub) ToOthers(message contracts.WebsocketMessage) error {
	return nil
}

func (e *EmptyPubSub) Subscribe() <-chan []byte {
	return nil
}

func (e *EmptyPubSub) Close() error {
	return nil
}
