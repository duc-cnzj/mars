package plugins

import (
	"encoding/json"
	"sync"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
)

var wsSenderOnce sync.Once

type TO uint8

const (
	ToSelf TO = iota
	ToAll
	ToOthers
)

type WsResponse struct {
	// 有可能同一个用户同时部署两个环境, 必须要有 slug 区分
	Slug   string `json:"slug"`
	Type   string `json:"type"`
	Result string `json:"result"`
	Data   string `json:"data"`
	End    bool   `json:"end"`
	Uid    string `json:"uid"`
	ID     string `json:"id"`

	To TO `json:"to"`
}

func (r *WsResponse) EncodeToBytes() []byte {
	marshal, _ := json.Marshal(&r)
	return marshal
}
func (r *WsResponse) EncodeToString() string {
	marshal, _ := json.Marshal(&r)
	return string(marshal)
}

type WsSender interface {
	New(uid, id string) PubSub
}

type PubSub interface {
	Info() interface{}
	Uid() string
	ID() string
	ToSelf(*WsResponse) error
	ToAll(*WsResponse) error
	ToOthers(*WsResponse) error
	Subscribe() <-chan string
	Close() error
}

func GetWsSender() WsSender {
	p := app.App().GetPluginByName(app.Config().WsSenderPlugin.Name)
	wsSenderOnce.Do(func() {
		if err := p.Initialize(); err != nil {
			panic(err)
		}
		app.App().RegisterAfterShutdownFunc(func(applicationInterface contracts.ApplicationInterface) {
			p.Destroy()
		})
	})

	return p.(WsSender)
}
