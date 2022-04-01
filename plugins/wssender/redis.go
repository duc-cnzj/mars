package wssender

import (
	"context"
	"errors"

	websocket_pb "github.com/duc-cnzj/mars-client/v4/websocket"
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

func (p *redisSender) Initialize(args map[string]any) error {
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
	mlog.Info("[Plugin]: " + p.Name() + " plugin Initialize...")
	return nil
}

func (p *redisSender) Destroy() error {
	p.rds.Close()
	mlog.Info("[Plugin]: " + p.Name() + " plugin Destroy...")
	return nil
}

func (p *redisSender) New(uid, id string) plugins.PubSub {
	return &rdsPubSub{uid: uid, id: id, rds: p.rds, ch: make(chan []byte, messageChSize), close: make(chan struct{})}
}

type rdsPubSub struct {
	ch    chan []byte
	close chan struct{}
	rds   *redis.Client
	uid   string
	id    string
}

func (p *rdsPubSub) Close() error {
	mlog.Debugf("[Websocket]: Closed, uid: %v id: %v", p.uid, p.id)
	close(p.close)
	return nil
}

func (p *rdsPubSub) Info() any {
	return "<unknown>"
}
func (p *rdsPubSub) Uid() string {
	return p.uid
}

func (p *rdsPubSub) ID() string {
	return p.id
}

func (p *rdsPubSub) ToSelf(wsResponse plugins.WebsocketMessage) error {
	p.rds.Publish(context.TODO(), p.id, plugins.ProtoToMessage(wsResponse, websocket_pb.To_ToSelf, p.id).Marshal())
	return nil
}

func (p *rdsPubSub) ToAll(wsResponse plugins.WebsocketMessage) error {
	p.rds.Publish(context.TODO(), BroadcastRoom, plugins.ProtoToMessage(wsResponse, websocket_pb.To_ToAll, p.id).Marshal())
	return nil
}

func (p *rdsPubSub) ToOthers(wsResponse plugins.WebsocketMessage) error {
	p.rds.Publish(context.TODO(), BroadcastRoom, plugins.ProtoToMessage(wsResponse, websocket_pb.To_ToOthers, p.id).Marshal())
	return nil
}

func (p *rdsPubSub) Subscribe() <-chan []byte {
	ps := p.rds.Subscribe(context.TODO(), p.id, BroadcastRoom)
	channel := ps.Channel()
	mlog.Debugf("[Websocket]: Subscribe Start id: %v channels: %v %s", p.id, p.id, BroadcastRoom)
	go func() {
		for {
			select {
			case msg, ok := <-channel:
				if !ok {
					ps.Close()
					p.Close()
					return
				}
				message, _ := plugins.DecodeMessage([]byte(msg.Payload))
				switch message.To {
				case plugins.ToSelf:
					fallthrough
				case plugins.ToAll:
					p.ch <- message.Data
				case plugins.ToOthers:
					if message.To == plugins.ToOthers && message.ID != p.id {
						p.ch <- message.Data
					}
				}
			case <-p.close:
				ps.Close()
				mlog.Debugf("[Websocket]: redis channel closed, uid: %s, id: %v", p.uid, p.id)
				return
			}
		}
	}()

	return p.ch
}
