package plugins

import (
	"encoding/json"
	"sync"

	"google.golang.org/protobuf/proto"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	websocket_pb "github.com/duc-cnzj/mars/pkg/websocket"
)

var wsSenderOnce sync.Once

const (
	ToSelf   = websocket_pb.To_ToSelf
	ToAll    = websocket_pb.To_ToAll
	ToOthers = websocket_pb.To_ToOthers
)

type Message struct {
	Data []byte
	To   websocket_pb.To
	ID   string
}

func (m Message) Marshal() []byte {
	marshal, _ := json.Marshal(&m)
	return marshal
}

func DecodeMessage(data []byte) (msg Message, err error) {
	err = json.Unmarshal(data, &msg)
	return
}

func ProtoToMessage(m proto.Message, to websocket_pb.To, id string) Message {
	marshal, _ := proto.Marshal(m)

	return Message{
		Data: marshal,
		To:   to,
		ID:   id,
	}
}

type WsResponseMetadata = websocket_pb.WsResponseMetadata

type WsSender interface {
	New(uid, id string) PubSub
}

type PubSub interface {
	Info() interface{}
	Uid() string
	ID() string
	ToSelf(proto.Message) error
	ToAll(proto.Message) error
	ToOthers(proto.Message) error
	Subscribe() <-chan []byte
	Close() error
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
