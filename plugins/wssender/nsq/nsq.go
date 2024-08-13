package nsq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/project"

	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/plugins/wssender"

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
	return fmt.Sprintf("project-pod-events-%d#ephemeral", nsID)
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
	n.cfg = gonsq.NewConfig()
	// 坑:
	// 当多个nsqd服务都有相同的topic的时候，consumer要修改默认设置config.MaxInFlight才能连接
	// 本地 k8s 搭建 nsq 集群时，访问 lookupd 返回的是集群内部的 ip，不通的
	n.cfg.MaxInFlight = 1000
	n.cfg.LookupdPollInterval = 3 * time.Second
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
	p, _ := gonsq.NewProducer(n.addr, n.cfg)
	setLogLevel(n.logger, p)
	err = p.Ping()
	if err != nil {
		return err
	}
	n.db = app.DB()
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
		db:           n.db,
		addr:         n.addr,
		lookupdAddr:  n.lookupdAddr,
		cfg:          n.cfg,
		uid:          uid,
		id:           id,
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

	mu       sync.RWMutex
	channels map[string]struct{}

	pMu          sync.RWMutex
	pidSelectors map[int32][]labels.Selector
}

type directHandler struct {
	ch chan []byte
}

func (j *directHandler) HandleMessage(m *gonsq.Message) error {
	if m == nil || len(m.Body) == 0 {
		return nil
	}
	j.ch <- m.Body

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
	if err := n.connect(consumer, n.addr, n.lookupdAddr, &directHandler{ch: n.eventMsgCh}); err != nil {
		n.logger.Error(err)
		return err
	}

	func() {
		n.consumersMu.Lock()
		defer n.consumersMu.Unlock()
		n.consumers[channel] = consumer
	}()

	func() {
		n.mu.Lock()
		defer n.mu.Unlock()
		n.channels[channel] = struct{}{}
	}()

	func() {
		n.pMu.Lock()
		defer n.pMu.Unlock()
		var selectors []labels.Selector
		for _, s := range pmodel.PodSelectors {
			parse, _ := labels.Parse(s)
			selectors = append(selectors, parse)
		}
		n.pidSelectors[int32(projectID)] = selectors
	}()
	return nil
}

func (n *nsq) Leave(nsID int64, projectID int64) error {
	channel := getNsqProjectEventRoom(nsID)

	func() {
		n.consumersMu.Lock()
		defer n.consumersMu.Unlock()
		consumer, ok := n.consumers[channel]
		if ok {
			consumer.Stop()
			delete(n.consumers, channel)
		}
	}()

	func() {
		n.mu.Lock()
		defer n.mu.Unlock()
		delete(n.channels, channel)
	}()
	func() {
		n.pMu.Lock()
		defer n.pMu.Unlock()
		delete(n.pidSelectors, int32(projectID))
	}()
	return nil
}

type projectEventObj struct {
	Channel     string  `json:"channel"`
	NamespaceID int64   `json:"namespace_id"`
	Pod         *v1.Pod `json:"pod"`
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
			var obj projectEventObj
			if err := json.Unmarshal([]byte(data), &obj); err != nil {
				n.logger.Error(err)
			}
			fn := func() bool {
				n.mu.RLock()
				defer n.mu.RUnlock()
				_, ok := n.channels[obj.Channel]
				return ok
			}
			if !fn() {
				continue
			}

			func() {
				n.pMu.RLock()
				defer n.pMu.RUnlock()
				for pid, selectors := range n.pidSelectors {
					func() {
						for _, selector := range selectors {
							if selector.Matches(labels.Set(obj.Pod.Labels)) {
								n.msgCh <- wssender.TransformToResponse(&websocket_pb.WsProjectPodEventResponse{
									Metadata: &websocket_pb.Metadata{
										Id:     n.id,
										Uid:    n.uid,
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

func (n *nsq) Publish(nsID int64, pod *v1.Pod) error {
	room := getNsqProjectEventRoom(nsID)
	marshal, _ := json.Marshal(&projectEventObj{
		NamespaceID: nsID,
		Pod:         pod,
		Channel:     room,
	})
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
	response.GetMetadata().To = to
	response.GetMetadata().Uid = n.uid
	response.GetMetadata().Id = n.id
	room := ephemeralBroadcastRoom
	if to == websocket_pb.To_ToSelf {
		room = n.ephemeralID()
	}
	return n.producer.Publish(room, wssender.ProtoToMessage(response, n.id).Marshal())
}

func (n *nsq) ephemeralID() string {
	return n.ID() + "#ephemeral"
}

func (n *nsq) Subscribe() <-chan []byte {
	consumerAll, _ := gonsq.NewConsumer(ephemeralBroadcastRoom, n.ephemeralID(), n.cfg)
	consumer, _ := gonsq.NewConsumer(n.ephemeralID(), n.ephemeralID(), n.cfg)
	h := &handler{msgCh: n.msgCh, id: n.id}
	n.connect(consumer, n.addr, n.lookupdAddr, h)
	n.connect(consumerAll, n.addr, n.lookupdAddr, h)

	n.consumersMu.Lock()
	defer n.consumersMu.Unlock()
	n.consumers = map[string]*gonsq.Consumer{
		ephemeralBroadcastRoom: consumerAll,
		n.ephemeralID():        consumer,
	}

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
	n.consumersMu.Lock()
	defer n.consumersMu.Unlock()
	for _, c := range n.consumers {
		c.Stop()
		if n.lookupdAddr != "" {
			c.DisconnectFromNSQLookupd(n.lookupdAddr)
		} else {
			c.DisconnectFromNSQD(n.addr)
		}
	}
	return nil
}

type handler struct {
	id    string
	msgCh chan []byte
}

func (h *handler) HandleMessage(m *gonsq.Message) error {
	if m == nil || len(m.Body) == 0 {
		return nil
	}
	message, _ := wssender.DecodeMessage(m.Body)
	switch message.To {
	case websocket_pb.To_ToSelf:
		fallthrough
	case websocket_pb.To_ToAll:
		h.msgCh <- message.Data
	case websocket_pb.To_ToOthers:
		if message.ID != h.id {
			h.msgCh <- message.Data
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
