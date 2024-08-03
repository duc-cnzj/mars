package repo

import (
	"github.com/duc-cnzj/mars/v4/internal/application"
)

type WsRepo interface {
	New(uid, id string) application.PubSub
}

type wsRepo struct {
	ws application.WsSender
}

var _ WsRepo = (*wsRepo)(nil)

func (w *wsRepo) New(uid, id string) application.PubSub {
	return w.ws.New(uid, id)
}

func NewWsRepo(pl application.PluginManger) WsRepo {
	return &wsRepo{ws: pl.Ws()}
}
