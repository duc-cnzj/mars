package nsq

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v5/websocket"
	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/ent"
	"github.com/duc-cnzj/mars/v5/internal/ent/project"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/plugins/wssender"
	gonsq "github.com/nsqio/go-nsq"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const ephemeralBroadcastRoom = wssender.BroadcastRoom + "#ephemeral"

var nsqSenderName = "ws_sender_nsq"

func init() {
	dr := &nsqSender{}
	application.RegisterPlugin(dr.Name(), dr)
}

func getNsqProjectEventRoom[T int64 | int](nsID T) string {
	return wssender.GetProjectPodEventRoom(nsID) + "#ephemeral"
}

type nsqSender struct {
	producer    *gonsq.Producer
	cfg         *gonsq.Config
	lookupdAddr string
	addr        string
	logger      mlog.Logger
	db          *ent.Client
}

func (n *nsqSender) Name() string {
	return nsqSenderName
}

func (n *nsqSender) Initialize(app application.App, args map[string]any) (err error) {
	n.db = app.DB()
	n.cfg = gonsq.NewConfig()
	// 坑:
	// 当多个nsqd服务都有相同的topic的时候，consumer要修改默认设置config.MaxInFlight才能连接
	// 本地 k8s 搭建 nsq 集群时，访问 lookupd 返回的是集群内部的 ip，不通的
	// 必须 <= MessageChSize (1000)，否则 handler SendOrDrop 会批量丢消息
	n.cfg.MaxInFlight = 1000
	n.cfg.LookupdPollInterval = 3 * time.Second

	// 超时配置
	if v, ok := args["msg_timeout"]; ok {
		if timeout, ok := v.(int); ok && timeout > 0 {
			n.cfg.MsgTimeout = time.Duration(timeout) * time.Second
		}
	}
	if v, ok := args["dial_timeout"]; ok {
		if timeout, ok := v.(int); ok && timeout > 0 {
			n.cfg.DialTimeout = time.Duration(timeout) * time.Second
		}
	}
	if v, ok := args["read_timeout"]; ok {
		if timeout, ok := v.(int); ok && timeout > 0 {
			n.cfg.ReadTimeout = time.Duration(timeout) * time.Second
		}
	}
	if v, ok := args["write_timeout"]; ok {
		if timeout, ok := v.(int); ok && timeout > 0 {
			n.cfg.WriteTimeout = time.Duration(timeout) * time.Second
		}
	}
	if v, ok := args["heartbeat_interval"]; ok {
		if interval, ok := v.(int); ok && interval > 0 {
			n.cfg.HeartbeatInterval = time.Duration(interval) * time.Second
		}
	}

	n.logger = app.Logger().WithModule("plugins/ws_sender_nsq")
	if s, ok := args["addr"]; ok {
		n.logger.Debugf("[NSQ]: addr '%v'", s)
		n.addr = s.(string)
	} else {
		err = errors.New("[nsq]: add not exits")
		return
	}
	if s, ok := args["lookupd_addr"]; ok {
		n.logger.Debugf("[NSQ]: lookupd_addr '%v'", s)
		n.lookupdAddr = s.(string)
	}
	p, err := gonsq.NewProducer(n.addr, n.cfg)
	if err != nil {
		return err
	}
	setLogLevel(n.logger, p)
	if err = p.Ping(); err != nil {
		p.Stop()
		return err
	}
	n.producer = p
	n.logger.Info("[Plugin]: " + n.Name() + " plugin Initialize...")
	return
}

func (n *nsqSender) Destroy() error {
	n.producer.Stop()
	n.logger.Info("[Plugin]: " + n.Name() + " plugin Destroy...")
	return nil
}

func (n *nsqSender) New(uid, id string) application.PubSub {
	return &nsq{
		logger:       n.logger,
		addr:         n.addr,
		lookupdAddr:  n.lookupdAddr,
		cfg:          n.cfg,
		uid:          uid,
		id:           id,
		db:           n.db,
		consumers:    map[string]*gonsq.Consumer{},
		producer:     n.producer,
		msgCh:        make(chan []byte, wssender.MessageChSize),
		eventMsgCh:   make(chan []byte, wssender.MessageChSize),
		channels:     map[string]struct{}{},
		pidSelectors: map[int32][]labels.Selector{},
	}
}

type nsq struct {
	logger            mlog.Logger
	addr, lookupdAddr string
	cfg               *gonsq.Config
	uid, id           string
	db                *ent.Client

	consumersMu sync.RWMutex
	consumers   map[string]*gonsq.Consumer

	producer   *gonsq.Producer
	msgCh      chan []byte
	eventMsgCh chan []byte
	closeOnce  sync.Once

	mu       sync.RWMutex
	channels map[string]struct{}

	pMu          sync.RWMutex
	pidSelectors map[int32][]labels.Selector
}

type directHandler struct {
	ch  chan []byte
	log mlog.Logger
}

func (j *directHandler) HandleMessage(m *gonsq.Message) error {
	if m == nil || len(m.Body) == 0 {
		return nil
	}
	wssender.SendOrDrop(j.ch, m.Body, j.log, "direct")

	return nil
}

func (n *nsq) Join(projectID int64) error {
	pmodel, err := n.db.Project.Query().WithNamespace().Where(project.ID(int(projectID))).Only(context.TODO())
	if err != nil {
		return err
	}
	channel := getNsqProjectEventRoom(pmodel.Edges.Namespace.ID)

	consumer, err := gonsq.NewConsumer(channel, n.ephemeralID(), n.cfg)
	if err != nil {
		n.logger.Error(err)
		return err
	}
	if err := n.connect(consumer, n.addr, n.lookupdAddr, &directHandler{ch: n.eventMsgCh, log: n.logger}); err != nil {
		n.logger.Error(err)
		return err
	}

	n.consumersMu.Lock()
	n.consumers[channel] = consumer
	n.consumersMu.Unlock()

	n.mu.Lock()
	n.channels[channel] = struct{}{}
	n.mu.Unlock()

	n.pMu.Lock()
	var selectors []labels.Selector
	for _, s := range pmodel.PodSelectors {
		parse, err := labels.Parse(s)
		if err != nil {
			n.logger.Errorf("[NSQ] invalid pod selector %q: %v", s, err)
			continue
		}
		selectors = append(selectors, parse)
	}
	n.pidSelectors[int32(projectID)] = selectors
	n.pMu.Unlock()

	return nil
}

func (n *nsq) Leave(nsID int64, projectID int64) error {
	channel := getNsqProjectEventRoom(nsID)

	n.consumersMu.Lock()
	consumer, ok := n.consumers[channel]
	if ok {
		consumer.Stop()
		delete(n.consumers, channel)
	}
	n.consumersMu.Unlock()

	n.mu.Lock()
	delete(n.channels, channel)
	n.mu.Unlock()

	n.pMu.Lock()
	delete(n.pidSelectors, int32(projectID))
	n.pMu.Unlock()

	return nil
}

func (n *nsq) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case data, ok := <-n.eventMsgCh:
			if !ok {
				return errors.New("nsq event ch closed")
			}
			var obj wssender.ProjectPodEventObj
			if err := json.Unmarshal([]byte(data), &obj); err != nil {
				n.logger.Error(err)
				continue
			}

			n.mu.RLock()
			_, subscribed := n.channels[obj.Channel]
			n.mu.RUnlock()
			if !subscribed {
				continue
			}

			n.pMu.RLock()
			wssender.MatchSelectorsAndSend(n.msgCh, labels.Set(obj.Pod.Labels), n.pidSelectors, n.id, n.uid, n.logger)
			n.pMu.RUnlock()
		}
	}
}

func (n *nsq) Publish(nsID int64, pod *v1.Pod) error {
	room := getNsqProjectEventRoom(nsID)
	marshal, err := json.Marshal(&wssender.ProjectPodEventObj{
		NamespaceID: nsID,
		Pod:         pod,
		Channel:     room,
	})
	if err != nil {
		return err
	}
	return n.producer.Publish(room, marshal)
}

func (n *nsq) Info() any {
	return nil
}

func (n *nsq) Uid() string {
	return n.uid
}

func (n *nsq) ID() string {
	return n.id
}

func (n *nsq) ToSelf(response application.WebsocketMessage) error {
	return n.to(response, websocket_pb.To_ToSelf)
}

func (n *nsq) ToAll(response application.WebsocketMessage) error {
	return n.to(response, websocket_pb.To_ToAll)
}

func (n *nsq) ToOthers(response application.WebsocketMessage) error {
	return n.to(response, websocket_pb.To_ToOthers)
}

func (n *nsq) to(response application.WebsocketMessage, to websocket_pb.To) error {
	room := ephemeralBroadcastRoom
	if to == websocket_pb.To_ToSelf {
		room = n.ephemeralID()
	}
	return n.producer.Publish(room, wssender.ProtoToMessage(response, n.id, to).Marshal())
}

func (n *nsq) ephemeralID() string {
	return n.ID() + "#ephemeral"
}

func closedMsgCh() chan []byte {
	ch := make(chan []byte)
	close(ch)
	return ch
}

func (n *nsq) Subscribe() <-chan []byte {
	consumerAll, err := gonsq.NewConsumer(ephemeralBroadcastRoom, n.ephemeralID(), n.cfg)
	if err != nil {
		n.logger.Errorf("[NSQ] create broadcast consumer: %v", err)
		return closedMsgCh()
	}
	consumer, err := gonsq.NewConsumer(n.ephemeralID(), n.ephemeralID(), n.cfg)
	if err != nil {
		n.logger.Errorf("[NSQ] create direct consumer: %v", err)
		consumerAll.Stop()
		return closedMsgCh()
	}
	handler := &handler{msgCh: n.msgCh, id: n.id, logger: n.logger}

	if err := n.connect(consumer, n.addr, n.lookupdAddr, handler); err != nil {
		n.logger.Errorf("[NSQ] connect direct consumer: %v", err)
		consumer.Stop()
		consumerAll.Stop()
		return closedMsgCh()
	}
	if err := n.connect(consumerAll, n.addr, n.lookupdAddr, handler); err != nil {
		n.logger.Errorf("[NSQ] connect broadcast consumer: %v", err)
		consumer.Stop()
		consumerAll.Stop()
		return closedMsgCh()
	}

	n.consumersMu.Lock()
	n.consumers[ephemeralBroadcastRoom] = consumerAll
	n.consumers[n.ephemeralID()] = consumer
	n.consumersMu.Unlock()

	return n.msgCh
}

func (n *nsq) connect(consumer *gonsq.Consumer, addr, lookupdAddr string, h gonsq.Handler) error {
	setLogLevel(n.logger, consumer)
	consumer.AddHandler(h)

	var err error
	if lookupdAddr != "" {
		err = consumer.ConnectToNSQLookupd(lookupdAddr)
	} else {
		err = consumer.ConnectToNSQD(addr)
	}

	return err
}

func (n *nsq) Close() error {
	defer n.logger.Debugf("[nsq]: id: %v closed", n.ID())
	n.closeOnce.Do(func() {
		n.consumersMu.Lock()
		var consumers []*gonsq.Consumer
		for _, c := range n.consumers {
			c.Stop()
			consumers = append(consumers, c)
		}
		n.consumersMu.Unlock()

		// Stop() is non-blocking — it sends CLS and returns immediately.
		// Wait for each consumer's StopChan (closed after all handlers exit)
		// so we don't close msgCh while a handler goroutine is still writing.
		for _, c := range consumers {
			<-c.StopChan
		}
		close(n.msgCh)
	})
	return nil
}

type handler struct {
	id     string
	msgCh  chan []byte
	logger mlog.Logger
}

func (h *handler) HandleMessage(m *gonsq.Message) error {
	if m == nil || len(m.Body) == 0 {
		return nil
	}
	message, err := wssender.DecodeMessage(m.Body)
	if err != nil {
		h.logger.Debugf("[NSQ] handler decode error: %v", err)
		return nil
	}
	switch message.To {
	case websocket_pb.To_ToSelf:
		fallthrough
	case websocket_pb.To_ToAll:
		wssender.SendOrDrop(h.msgCh, message.Data, h.logger, h.id)
	case websocket_pb.To_ToOthers:
		if message.ID != h.id {
			wssender.SendOrDrop(h.msgCh, message.Data, h.logger, h.id)
		}
	}

	return nil
}

func setLogLevel(logger mlog.Logger, s any) {
	log := NewNsqLoggerAdapter(logger)
	if ss, ok := s.(*gonsq.Consumer); ok {
		ss.SetLoggerLevel(gonsq.LogLevelError)
		ss.SetLoggerForLevel(log, gonsq.LogLevelError)
	}
	if ss, ok := s.(*gonsq.Producer); ok {
		ss.SetLoggerLevel(gonsq.LogLevelError)
		ss.SetLoggerForLevel(log, gonsq.LogLevelError)
	}
}

type NsqLoggerAdapter struct {
	logger mlog.Logger
}

func NewNsqLoggerAdapter(logger mlog.Logger) *NsqLoggerAdapter {
	return &NsqLoggerAdapter{logger: logger}
}

// Output impl nsq.logger
func (n *NsqLoggerAdapter) Output(calldepth int, s string) error {
	if strings.Contains(s, "TOPIC_NOT_FOUND") {
		n.logger.Debug(s)
	} else {
		n.logger.Error(s)
	}
	return nil
}
