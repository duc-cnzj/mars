package redis

import (
	"context"
	"encoding/json"
	"errors"
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

// ---------------------------------------------------------------------------
// Shared-subscriber redisSender (1 Redis PubSub connection for ALL ws messages)
// ---------------------------------------------------------------------------

type redisSender struct {
	rds    *redis.Client
	logger mlog.Logger
	db     *ent.Client

	// One shared PubSub connection — not N connections.
	wsPubSub *redis.PubSub
	msgCh    <-chan *redis.Message

	mu   sync.RWMutex
	subs map[string]*subEntry // id → local subscriber

	ctx    context.Context
	cancel context.CancelFunc
}

type subEntry struct {
	ch  chan []byte
	uid string
	id  string
}

func (p *redisSender) Name() string {
	return redisSenderName
}

func (p *redisSender) Initialize(app application.App, args map[string]any) error {
	addr, ok := args["addr"].(string)
	if !ok || addr == "" {
		return errors.New("redisSender need valid addr")
	}
	pwd, _ := args["password"].(string)
	db, _ := args["db"].(int)

	p.db = app.DB()

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       db,
	})
	if err := rdb.Ping(context.TODO()).Err(); err != nil {
		return err
	}
	p.rds = rdb

	p.ctx, p.cancel = context.WithCancel(context.TODO())
	p.subs = make(map[string]*subEntry)

	// One shared PubSub subscribes to the broadcast room.
	p.wsPubSub = p.rds.Subscribe(context.TODO())
	if err := p.wsPubSub.Subscribe(context.TODO(), wssender.BroadcastRoom); err != nil {
		return err
	}
	p.msgCh = p.wsPubSub.Channel()
	go p.dispatcher()

	p.logger.Info("[Plugin]: " + p.Name() + " plugin Initialize...")
	return nil
}

func (p *redisSender) Destroy() error {
	p.cancel()
	p.wsPubSub.Close()
	p.rds.Close()
	p.logger.Info("[Plugin]: " + p.Name() + " plugin Destroy...")
	return nil
}

// dispatcher reads from the shared PubSub channel and fans out to local subscribers.
// One goroutine for ALL connections instead of one goroutine PER connection.
func (p *redisSender) dispatcher() {
	defer p.logger.HandlePanic("[PubSub]: dispatcher")
	for {
		select {
		case <-p.ctx.Done():
			return
		case msg, ok := <-p.msgCh:
			if !ok {
				return
			}
			message, err := wssender.DecodeMessage([]byte(msg.Payload))
			if err != nil {
				p.logger.Error(err)
				continue
			}

			p.mu.RLock()
			switch message.To {
			case websocket_pb.To_ToSelf:
				if sub, ok := p.subs[message.ID]; ok {
					wssender.SendOrDrop(sub.ch, message.Data, p.logger, message.ID)
				}
			case websocket_pb.To_ToAll:
				for _, sub := range p.subs {
					wssender.SendOrDrop(sub.ch, message.Data, p.logger, sub.id)
				}
			case websocket_pb.To_ToOthers:
				for _, sub := range p.subs {
					if sub.id != message.ID {
						wssender.SendOrDrop(sub.ch, message.Data, p.logger, sub.id)
					}
				}
			}
			p.mu.RUnlock()
		}
	}
}

func (p *redisSender) New(uid, id string) application.PubSub {
	ctx, cancel := context.WithCancel(context.TODO())
	ch := make(chan []byte, wssender.MessageChSize)

	// Register with shared dispatcher.
	p.mu.Lock()
	p.subs[id] = &subEntry{ch: ch, uid: uid, id: id}
	p.mu.Unlock()

	// Subscribe shared connection to this user's direct channel (for cross-instance ToSelf).
	p.wsPubSub.Subscribe(context.TODO(), id)

	pem := &podEventManagers{
		logger:       p.logger.WithModule("plugins/ws_sender_redis"),
		db:           p.db,
		ch:           ch,
		id:           id,
		uid:          uid,
		rds:          p.rds,
		channelRefs:  make(map[string]int),
		pubSub:       p.rds.Subscribe(context.TODO()),
		pidSelectors: make(map[int32][]labels.Selector),
	}

	return &rdsPubSub{
		logger:   p.logger,
		done:     ctx,
		doneFunc: cancel,
		ch:       ch,
		rds:      p.rds,
		uid:      uid,
		id:       id,
		manager:  p,
		ProjectPodEventSubscriber: pem,
		ProjectPodEventPublisher:  pem,
	}
}

// ---------------------------------------------------------------------------
// rdsPubSub — thin wrapper, no per-connection Redis subscription
// ---------------------------------------------------------------------------

type rdsPubSub struct {
	logger   mlog.Logger
	rds      *redis.Client
	manager  *redisSender
	uid, id  string
	ch       chan []byte
	done     context.Context
	doneFunc func()
	closeOnce sync.Once

	application.ProjectPodEventSubscriber
	application.ProjectPodEventPublisher
}

func (p *rdsPubSub) ID() string {
	return p.id
}

func (p *rdsPubSub) Uid() string {
	return p.uid
}

func (p *rdsPubSub) Info() any {
	p.manager.mu.RLock()
	defer p.manager.mu.RUnlock()
	return map[string]any{
		"subscribers": len(p.manager.subs),
		"id":          p.id,
	}
}

func (p *rdsPubSub) Close() error {
	p.logger.Debugf("[Websocket]: Closed, uid: %v id: %v", p.uid, p.id)

	p.closeOnce.Do(func() {
		// Remove from shared dispatcher first, so no more messages arrive.
		p.manager.mu.Lock()
		delete(p.manager.subs, p.id)
		p.manager.mu.Unlock()

		p.manager.wsPubSub.Unsubscribe(context.TODO(), p.id)
		p.doneFunc()
		close(p.ch)
	})
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

// to publishes to Redis. The shared dispatcher in redisSender handles fan-out.
// Does NOT mutate wsResponse.
func (p *rdsPubSub) to(response application.WebsocketMessage, to websocket_pb.To) error {
	room := wssender.BroadcastRoom
	if to == websocket_pb.To_ToSelf {
		room = p.id
	}
	return p.rds.Publish(context.TODO(), room, wssender.ProtoToMessage(response, p.id, to).Marshal()).Err()
}

// Subscribe returns the local channel. No per-connection Redis subscription.
func (p *rdsPubSub) Subscribe() <-chan []byte {
	return p.ch
}

// ---------------------------------------------------------------------------
// Pod event types (unchanged, still per-connection Redis subscription)
// ---------------------------------------------------------------------------

type podEventManagers struct {
	db     *ent.Client
	logger mlog.Logger
	id     string
	uid    string
	rds    *redis.Client
	pubSub *redis.PubSub

	ch chan []byte

	mu          sync.RWMutex
	channelRefs map[string]int // reference count for Join/Leave symmetry

	pmu          sync.RWMutex
	pidSelectors map[int32][]labels.Selector
}

func (p *podEventManagers) Publish(nsID int64, pod *v1.Pod) error {
	channel := wssender.GetProjectPodEventRoom(nsID)
	marshal, err := json.Marshal(&wssender.ProjectPodEventObj{
		Channel:     channel,
		NamespaceID: nsID,
		Pod:         pod,
	})
	if err != nil {
		return err
	}
	return p.rds.Publish(context.TODO(), channel, marshal).Err()
}

func (p *podEventManagers) Join(projectID int64) error {
	pmodel, err := p.db.Project.Query().WithNamespace().Where(project.ID(int(projectID))).Only(context.TODO())
	if err != nil {
		return err
	}

	channel := wssender.GetProjectPodEventRoom(pmodel.Edges.Namespace.ID)

	p.mu.Lock()
	p.channelRefs[channel]++
	if p.channelRefs[channel] == 1 {
		// First reference: actually subscribe.
		if err := p.pubSub.Subscribe(context.TODO(), channel); err != nil {
			p.channelRefs[channel]--
			p.mu.Unlock()
			return err
		}
	}
	p.mu.Unlock()

	p.pmu.Lock()
	var selectors []labels.Selector
	for _, s := range pmodel.PodSelectors {
		parse, err := labels.Parse(s)
		if err != nil {
			p.logger.Errorf("[Redis] invalid pod selector %q: %v", s, err)
			continue
		}
		selectors = append(selectors, parse)
	}
	p.pidSelectors[int32(projectID)] = selectors
	p.pmu.Unlock()

	return nil
}

func (p *podEventManagers) Leave(nsID int64, projectID int64) error {
	channel := wssender.GetProjectPodEventRoom(nsID)

	p.mu.Lock()
	p.channelRefs[channel]--
	if p.channelRefs[channel] <= 0 {
		delete(p.channelRefs, channel)
		if err := p.pubSub.Unsubscribe(context.TODO(), channel); err != nil {
			p.mu.Unlock()
			return err
		}
	}
	p.mu.Unlock()

	p.pmu.Lock()
	delete(p.pidSelectors, int32(projectID))
	p.pmu.Unlock()

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

			p.mu.RLock()
			_, subscribed := p.channelRefs[data.Channel]
			p.mu.RUnlock()
			if !subscribed {
				continue
			}

			var obj wssender.ProjectPodEventObj
			if err := json.Unmarshal([]byte(data.Payload), &obj); err != nil {
				p.logger.Error(err)
				continue
			}

			p.pmu.RLock()
			wssender.MatchSelectorsAndSend(p.ch, labels.Set(obj.Pod.Labels), p.pidSelectors, p.id, p.uid, p.logger)
			p.pmu.RUnlock()
		}
	}
}
