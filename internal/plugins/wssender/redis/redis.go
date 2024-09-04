package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	websocket_pb "github.com/duc-cnzj/mars/api/v5/websocket"
	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/ent"
	"github.com/duc-cnzj/mars/v5/internal/ent/project"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/plugins/wssender"
	"github.com/go-redis/redis/v8"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

var redisSenderName = "ws_sender_redis"

func init() {
	dr := &redisSender{}
	application.RegisterPlugin(dr.Name(), dr)
}

type redisSender struct {
	rds    *redis.Client
	logger mlog.Logger
	db     *ent.Client
}

func (p *redisSender) Name() string {
	return redisSenderName
}

func (p *redisSender) Initialize(app application.App, args map[string]any) error {
	addr := args["addr"]
	pwd := args["password"]
	db := args["db"]
	p.db = app.DB()

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
	p.logger.Info("[Plugin]: " + p.Name() + " plugin Initialize...")
	return nil
}

func (p *redisSender) Destroy() error {
	p.rds.Close()
	p.logger.Info("[Plugin]: " + p.Name() + " plugin Destroy...")
	return nil
}

func (p *redisSender) New(uid, id string) application.PubSub {
	ctx, cancelFunc := context.WithCancel(context.TODO())

	ch := make(chan []byte, wssender.MessageChSize)

	pem := &podEventManagers{
		logger:       p.logger.WithModule("plugins/ws_sender_redis"),
		db:           p.db,
		ch:           ch,
		id:           id,
		rds:          p.rds,
		channels:     make(map[string]struct{}),
		pubSub:       p.rds.Subscribe(context.TODO()),
		pidSelectors: make(map[int32][]labels.Selector),
	}

	return &rdsPubSub{
		logger:                    p.logger,
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
	rds    *redis.Client
	logger mlog.Logger
	uid    string
	id     string
	ch     chan []byte

	wsPubSub *redis.PubSub

	done     context.Context
	doneFunc func()

	application.ProjectPodEventSubscriber
	application.ProjectPodEventPublisher
}

type projectEventObj struct {
	Channel     string  `json:"channel"`
	NamespaceID int64   `json:"namespace_id"`
	Pod         *v1.Pod `json:"pod"`
}

func getRedisProjectEventRoom[T int64 | int](nsID T) string {
	return fmt.Sprintf("project-pod-events:%d", nsID)
}

type podEventManagers struct {
	db     *ent.Client
	logger mlog.Logger
	id     string
	uid    string
	rds    *redis.Client
	pubSub *redis.PubSub

	ch chan []byte

	mu       sync.RWMutex
	channels map[string]struct{}

	pmu          sync.RWMutex
	pidSelectors map[int32][]labels.Selector
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
	pmodel, err := p.db.Project.Query().WithNamespace().Where(project.ID(int(projectID))).Only(context.TODO())
	if err != nil {
		return err
	}

	channel := getRedisProjectEventRoom(pmodel.Edges.Namespace.ID)
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
		for _, s := range pmodel.PodSelectors {
			parse, _ := labels.Parse(s)
			selectors = append(selectors, parse)
		}
		p.pidSelectors[int32(projectID)] = selectors
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
		delete(p.pidSelectors, int32(projectID))
	}()

	return nil
}

func (p *podEventManagers) Run(ctx context.Context) error {
	defer p.pubSub.Close()
	ch := p.pubSub.Channel()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case data, ok := <-ch:
			if !ok {
				return errors.New("podEventManagers ch closed")
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
				p.logger.Error(err)
			}
			func() {
				p.pmu.RLock()
				defer p.pmu.RUnlock()
				for pid, selectors := range p.pidSelectors {
					func() {
						for _, selector := range selectors {
							if selector.Matches(labels.Set(obj.Pod.Labels)) {
								p.ch <- wssender.TransformToResponse(&websocket_pb.WsProjectPodEventResponse{
									Metadata: &websocket_pb.Metadata{
										Id:     p.id,
										Uid:    p.uid,
										Type:   websocket_pb.Type_ProjectPodEvent,
										End:    true,
										Result: websocket_pb.ResultType_Success,
										To:     websocket_pb.To_ToSelf,
									},
									ProjectId: pid,
								})
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
	p.logger.Debugf("[Websocket]: Closed, uid: %v id: %v", p.uid, p.id)
	p.doneFunc()
	return nil
}

func (p *rdsPubSub) ToSelf(wsResponse application.WebsocketMessage) error {
	return p.to(wsResponse, websocket_pb.To_ToSelf)
}

func (p *rdsPubSub) ToAll(wsResponse application.WebsocketMessage) error {
	return p.to(wsResponse, websocket_pb.To_ToAll)
}

func (p *rdsPubSub) ToOthers(wsResponse application.WebsocketMessage) error {
	return p.to(wsResponse, websocket_pb.To_ToOthers)
}

func (p *rdsPubSub) to(response application.WebsocketMessage, to websocket_pb.To) error {
	response.GetMetadata().To = to
	response.GetMetadata().Uid = p.uid
	response.GetMetadata().Id = p.id
	room := wssender.BroadcastRoom
	if to == websocket_pb.To_ToSelf {
		room = p.id
	}
	p.rds.Publish(context.TODO(), room, wssender.ProtoToMessage(response, p.id).Marshal())
	return nil
}

func (p *rdsPubSub) Subscribe() <-chan []byte {
	if err := p.wsPubSub.Subscribe(context.TODO(), p.id, wssender.BroadcastRoom); err != nil {
		p.logger.Fatal(err)
	}
	channel := p.wsPubSub.Channel()
	p.logger.Debugf("[Websocket]: Subscribe start id: %v channels: %v %s", p.id, p.id, wssender.BroadcastRoom)
	go func() {
		defer p.logger.HandlePanic("[PubSub]: Subscribe")
		for {
			select {
			case msg, ok := <-channel:
				if !ok {
					p.wsPubSub.Close()
					p.doneFunc()
					return
				}
				message, _ := wssender.DecodeMessage([]byte(msg.Payload))
				switch message.To {
				case websocket_pb.To_ToSelf:
					fallthrough
				case websocket_pb.To_ToAll:
					p.ch <- message.Data
				case websocket_pb.To_ToOthers:
					if message.ID != p.id {
						p.ch <- message.Data
					}
				}
			case <-p.done.Done():
				p.wsPubSub.Close()
				p.doneFunc()
				p.logger.Debugf("[Websocket]: redis channel closed, uid: %s, id: %v", p.uid, p.id)
				return
			}
		}
	}()

	return p.ch
}
