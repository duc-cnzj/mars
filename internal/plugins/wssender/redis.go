package wssender

import (
	"context"
	"encoding/json"
	"errors"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/go-redis/redis/v8"
)

var redisSenderName = "ws_sender_redis"

const BroadcastRoom = "all"

func init() {
	dr := &redisSender{}
	plugins.RegisterPlugin(dr.Name(), dr)
}

type redisSender struct {
	rds *redis.Client
}

func (p *redisSender) Name() string {
	return redisSenderName
}

func (p *redisSender) Initialize() error {
	args := app.Config().WsSenderPlugin.Args
	addr := args["addr"]
	pwd := args["password"]
	db := args["db"]

	if addr == nil || pwd == nil || db == nil {
		return errors.New("redisSender need addr, password, db args!")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr.(string),
		Password: pwd.(string),
		DB:       db.(int),
	})
	if err := rdb.Ping(context.TODO()).Err(); err != nil {
		return err
	}
	p.rds = rdb
	mlog.Info(p.Name() + " plugin Initialize...")
	return nil
}

func (p *redisSender) Destroy() error {
	p.rds.Close()
	mlog.Info(p.Name() + " plugin Destroy...")
	return nil
}

func (p *redisSender) New(uid, id string) plugins.PubSub {
	return &rdsPubSub{uid: uid, id: id, rds: p.rds, ch: make(chan string, messageChSize), close: make(chan struct{})}
}

type rdsPubSub struct {
	ch    chan string
	close chan struct{}
	rds   *redis.Client
	uid   string
	id    string
}

func (p *rdsPubSub) Close() error {
	mlog.Debugf("[Websocket] Closed, uid: %v id: %v", p.uid, p.id)
	close(p.close)
	return nil
}

func (p *rdsPubSub) Uid() string {
	return p.uid
}

func (p *rdsPubSub) ID() string {
	return p.id
}

func (p *rdsPubSub) ToSelf(wsResponse *plugins.WsResponse) error {
	mlog.Infof("ToSelf id: %s msg: %v ch len: %d", p.id, wsResponse.EncodeToString(), len(p.ch))
	wsResponse.To = plugins.ToSelf
	p.rds.Publish(context.TODO(), p.id, wsResponse.EncodeToBytes())
	return nil
}

func (p *rdsPubSub) ToAll(wsResponse *plugins.WsResponse) error {
	wsResponse.To = plugins.ToAll
	p.rds.Publish(context.TODO(), BroadcastRoom, wsResponse.EncodeToBytes())
	return nil
}

func (p *rdsPubSub) ToOthers(wsResponse *plugins.WsResponse) error {
	wsResponse.To = plugins.ToOthers
	p.rds.Publish(context.TODO(), BroadcastRoom, wsResponse.EncodeToBytes())
	return nil
}

func decodeMsg(msg string) (res *plugins.WsResponse) {
	json.Unmarshal([]byte(msg), &res)
	return res
}

func (p *rdsPubSub) Subscribe() <-chan string {
	ps := p.rds.Subscribe(context.TODO(), p.id, BroadcastRoom)
	channel := ps.Channel()
	mlog.Debugf("[Websocket] Subscribe Start id: %v channels: %v %s", p.id, p.id, BroadcastRoom)
	go func() {
		for {
			select {
			case msg, ok := <-channel:
				if !ok {
					ps.Close()
					p.Close()
					return
				}
				res := decodeMsg(msg.Payload)
				mlog.Debugf("[Websocket] receive msg %s", res.Data)
				switch res.To {
				case plugins.ToSelf:
					fallthrough
				case plugins.ToAll:
					p.ch <- res.EncodeToString()
				case plugins.ToOthers:
					if res.To == plugins.ToOthers && res.ID != p.id {
						p.ch <- res.EncodeToString()
					}
				}
			case <-p.close:
				ps.Close()
				mlog.Debugf("[Websocket] redis channel closed, uid: %s, id: %v", p.uid, p.id)
				return
			}
		}
	}()

	return p.ch
}
