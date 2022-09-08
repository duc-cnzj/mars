package wssender

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/models"
	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/duc-cnzj/mars/internal/contracts"

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
		return errors.New("redisSender need addr, password, db args")
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

func (p *redisSender) New(uid, id string) contracts.PubSub {
	ctx, cancelFunc := context.WithCancel(context.TODO())

	ch := make(chan []byte, messageChSize)

	pem := &podEventManagers{
		ch:           ch,
		id:           id,
		rds:          p.rds,
		channels:     make(map[string]struct{}),
		pubSub:       p.rds.Subscribe(context.TODO()),
		pidSelectors: make(map[int64][]labels.Selector),
	}

	return &rdsPubSub{
		done:                      ctx,
		doneFunc:                  cancelFunc,
		ch:                        ch,
		rds:                       p.rds,
		uid:                       uid,
		id:                        id,
		wsPubSub:                  p.rds.Subscribe(context.TODO()),
		ProjectPodEventSubscriber: pem,
		ProjectPodEventPublisher:  pem,
	}
}

type rdsPubSub struct {
	rds *redis.Client
	uid string
	id  string
	ch  chan []byte

	wsPubSub *redis.PubSub

	done     context.Context
	doneFunc func()

	contracts.ProjectPodEventSubscriber
	contracts.ProjectPodEventPublisher
}

type projectEventObj struct {
	Channel     string  `json:"channel"`
	NamespaceID int64   `json:"namespace_id"`
	Pod         *v1.Pod `json:"pod"`
}

func getRedisProjectEventRoom[T int64 | int](nsID T) string {
	return fmt.Sprintf("project-pod-events-%d", nsID)
}

type podEventManagers struct {
	id     string
	uid    string
	rds    *redis.Client
	pubSub *redis.PubSub

	ch chan []byte

	mu       sync.RWMutex
	channels map[string]struct{}

	pmu          sync.RWMutex
	pidSelectors map[int64][]labels.Selector
}

func (p *podEventManagers) Publish(nsID int64, pod *v1.Pod) error {
	channel := getRedisProjectEventRoom(nsID)
	marshal, _ := json.Marshal(&projectEventObj{
		Channel:     channel,
		NamespaceID: nsID,
		Pod:         pod,
	})
	p.rds.Publish(context.TODO(), channel, marshal)
	return nil
}

func (p *podEventManagers) Join(projectID int64) error {
	var pmodel models.Project
	if err := app.DB().First(&pmodel, projectID).Error; err != nil {
		return err
	}

	channel := getRedisProjectEventRoom(pmodel.NamespaceId)
	if err := p.pubSub.Subscribe(context.TODO(), channel); err != nil {
		return err
	}
	func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.channels[channel] = struct{}{}
	}()

	func() {
		p.pmu.Lock()
		defer p.pmu.Unlock()
		var selectors []labels.Selector
		for _, s := range pmodel.GetPodSelectors() {
			parse, _ := labels.Parse(s)
			selectors = append(selectors, parse)
		}
		p.pidSelectors[projectID] = selectors
	}()

	return nil
}

func (p *podEventManagers) Leave(nsID int64, projectID int64) error {
	channel := getRedisProjectEventRoom(nsID)
	if err := p.pubSub.Unsubscribe(context.TODO(), channel); err != nil {
		return err
	}
	func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		delete(p.channels, channel)
	}()
	func() {
		p.pmu.Lock()
		defer p.pmu.Unlock()
		delete(p.pidSelectors, projectID)
	}()

	return nil
}

func (p *podEventManagers) Run(ctx context.Context) error {
	defer p.pubSub.Close()
	ch := p.pubSub.Channel()
	for {
		select {
		case <-ctx.Done():
			mlog.Debug("podEventManagers exit")
			return nil
		case data, ok := <-ch:
			if !ok {
				mlog.Debug("podEventManagers ch closed")
				return nil
			}
			fn := func() bool {
				p.mu.RLock()
				defer p.mu.RUnlock()
				_, ok := p.channels[data.Channel]
				return ok
			}
			if !fn() {
				continue
			}

			var obj projectEventObj
			if err := json.Unmarshal([]byte(data.Payload), &obj); err != nil {
				mlog.Error(err)
			}
			func() {
				p.pmu.RLock()
				defer p.pmu.RUnlock()
				for pid, selectors := range p.pidSelectors {
					func() {
						for _, selector := range selectors {
							if selector.Matches(labels.Set(obj.Pod.Labels)) {
								marshal, _ := proto.Marshal(&websocket_pb.WsProjectPodEventResponse{
									Metadata: &websocket_pb.Metadata{
										Id:     p.id,
										Uid:    p.uid,
										Type:   websocket_pb.Type_ProjectPodEvent,
										End:    true,
										Result: websocket_pb.ResultType_Success,
										To:     plugins.ToSelf,
									},
									ProjectId: pid,
								})
								p.ch <- marshal
								return
							}
						}
					}()
				}
			}()
		}
	}
}

func (p *rdsPubSub) ID() string {
	return p.id
}

func (p *rdsPubSub) Uid() string {
	return p.uid
}

func (p *rdsPubSub) Info() any {
	return "<unknown>"
}

func (p *rdsPubSub) Close() error {
	mlog.Debugf("[Websocket]: Closed, uid: %v id: %v", p.uid, p.id)
	p.doneFunc()
	return nil
}

func (p *rdsPubSub) ToSelf(wsResponse contracts.WebsocketMessage) error {
	p.rds.Publish(context.TODO(), p.id, plugins.ProtoToMessage(wsResponse, websocket_pb.To_ToSelf, p.id).Marshal())
	return nil
}

func (p *rdsPubSub) ToAll(wsResponse contracts.WebsocketMessage) error {
	p.rds.Publish(context.TODO(), BroadcastRoom, plugins.ProtoToMessage(wsResponse, websocket_pb.To_ToAll, p.id).Marshal())
	return nil
}

func (p *rdsPubSub) ToOthers(wsResponse contracts.WebsocketMessage) error {
	p.rds.Publish(context.TODO(), BroadcastRoom, plugins.ProtoToMessage(wsResponse, websocket_pb.To_ToOthers, p.id).Marshal())
	return nil
}

func (p *rdsPubSub) Subscribe() <-chan []byte {
	p.wsPubSub.Subscribe(context.TODO(), p.id, BroadcastRoom)
	channel := p.wsPubSub.Channel()
	mlog.Debugf("[Websocket]: Subscribe Start id: %v channels: %v %s", p.id, p.id, BroadcastRoom)
	go func() {
		for {
			select {
			case msg, ok := <-channel:
				if !ok {
					p.wsPubSub.Close()
					p.doneFunc()
					return
				}
				message, _ := plugins.DecodeMessage([]byte(msg.Payload))
				switch message.To {
				case plugins.ToSelf:
					fallthrough
				case plugins.ToAll:
					p.ch <- message.Data
				case plugins.ToOthers:
					if message.ID != p.id {
						p.ch <- message.Data
					}
				}
			case <-p.done.Done():
				p.wsPubSub.Close()
				p.doneFunc()
				mlog.Debugf("[Websocket]: redis channel closed, uid: %s, id: %v", p.uid, p.id)
				return
			}
		}
	}()

	return p.ch
}
