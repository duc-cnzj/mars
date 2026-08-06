package nsq

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/application"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/project"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/plugins/wssender"
	gonsq "github.com/nsqio/go-nsq"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// ephemeralBroadcastRoom 广播房间的 NSQ ephemeral 通道名，消费端离线即销毁。
const ephemeralBroadcastRoom = wssender.BroadcastRoom + "#ephemeral"

// nsqSenderName 插件注册名。
var nsqSenderName = "ws_sender_nsq"

// nsqProducer 抽象 go-nsq producer 的最小方法面，便于测试注入 fake。
type nsqProducer interface {
	Ping() error
	Stop()
	Publish(topic string, body []byte) error
}

// nsqConsumer 抽象 go-nsq consumer 的最小方法面。
// 不包含 SetLoggerForLevel：其参数为 go-nsq 未导出的 logger 类型，无法跨包表达。
type nsqConsumer interface {
	AddHandler(gonsq.Handler)
	ConnectToNSQD(addr string) error
	ConnectToNSQLookupd(addr string) error
	Stop()
	StopChan() <-chan int
}

// consumerWrapper 包装 *gonsq.Consumer，把 StopChan 字段以方法形式暴露给 nsqConsumer。
type consumerWrapper struct {
	*gonsq.Consumer
}

// StopChan 返回底层 consumer 的停止通道。
func (c *consumerWrapper) StopChan() <-chan int {
	return c.Consumer.StopChan
}

// newProducer 创建 producer 的测试缝：默认走 go-nsq 真实实现。
var newProducer = func(addr string, cfg *gonsq.Config) (nsqProducer, error) {
	return gonsq.NewProducer(addr, cfg)
}

// newConsumer 创建 consumer 的测试缝：默认走 go-nsq 真实实现并包装为 nsqConsumer。
var newConsumer = func(topic, channel string, cfg *gonsq.Config) (nsqConsumer, error) {
	c, err := gonsq.NewConsumer(topic, channel, cfg)
	if err != nil {
		return nil, err
	}
	return &consumerWrapper{Consumer: c}, nil
}

func init() {
	dr := &nsqSender{}
	application.RegisterPlugin(dr.Name(), dr)
}

// getNsqProjectEventRoom 返回命名空间 pod 事件频道的 NSQ ephemeral 通道名。
func getNsqProjectEventRoom[T int64 | int](nsID T) string {
	return wssender.GetProjectPodEventRoom(nsID) + "#ephemeral"
}

// nsqSender 是 NSQ 版 WsSender：持有共享 producer 与默认配置，按需创建连接。
type nsqSender struct {
	producer    nsqProducer
	cfg         *gonsq.Config
	lookupdAddr string
	addr        string
	logger      mlog.Logger
	db          *ent.Client
}

// Name 返回插件名 ws_sender_nsq。
func (n *nsqSender) Name() string {
	return nsqSenderName
}

// Initialize 从 args 读取 nsq 地址与超时配置，创建 producer 并 Ping 验证连通性。
func (n *nsqSender) Initialize(app application.PluginApp, args map[string]any) (err error) {
	n.db = app.Data().DB()
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
	p, err := newProducer(n.addr, n.cfg)
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

// Destroy 停止共享 producer。
func (n *nsqSender) Destroy() error {
	n.producer.Stop()
	n.logger.Info("[Plugin]: " + n.Name() + " plugin Destroy...")
	return nil
}

// New 返回一个携带独立 consumer 集合与消息通道的 nsq 实例。
func (n *nsqSender) New(uid, id string) application.PubSub {
	return &nsq{
		logger:       n.logger,
		addr:         n.addr,
		lookupdAddr:  n.lookupdAddr,
		cfg:          n.cfg,
		uid:          uid,
		id:           id,
		db:           n.db,
		consumers:    map[string]nsqConsumer{},
		channelRefs:  map[string]int{},
		producer:     n.producer,
		msgCh:        make(chan []byte, wssender.MessageChSize),
		eventMsgCh:   make(chan []byte, wssender.MessageChSize),
		pidSelectors: map[int32][]labels.Selector{},
	}
}

// nsq 是单个连接对应的 PubSub：经 NSQ ephemeral 通道收发 websocket 与 pod 事件消息。
type nsq struct {
	logger            mlog.Logger
	addr, lookupdAddr string
	cfg               *gonsq.Config
	uid, id           string
	db                *ent.Client

	consumersMu sync.RWMutex
	consumers   map[string]nsqConsumer
	channelRefs map[string]int // channel 引用计数：同一 namespace 多个项目共享一个 consumer

	producer   nsqProducer
	msgCh      chan []byte
	eventMsgCh chan []byte
	closeOnce  sync.Once

	pMu          sync.RWMutex
	pidSelectors map[int32][]labels.Selector
}

// directHandler 将 pod 事件消息原样投递到通道，不做解码分发。
type directHandler struct {
	ch  chan []byte
	log mlog.Logger
}

// HandleMessage 将消息 body 投递到通道；空消息直接确认。
func (j *directHandler) HandleMessage(m *gonsq.Message) error {
	if m == nil || len(m.Body) == 0 {
		return nil
	}
	wssender.SendOrDrop(j.ch, m.Body, j.log, "direct")

	return nil
}

// Join 订阅项目所在命名空间的 ephemeral 频道（首个引用时创建 consumer），并登记选择器。
func (n *nsq) Join(projectID int64) error {
	pmodel, err := n.db.Project.Query().WithNamespace().Where(project.ID(int(projectID))).Only(context.TODO())
	if err != nil {
		return err
	}
	channel := getNsqProjectEventRoom(pmodel.Edges.Namespace.ID)

	// 同一 namespace 下多个项目共享同一个 channel：consumer 只在首个引用时创建，
	// 后续 Join 复用，避免覆盖旧 consumer 造成连接泄漏（与 redis channelRefs 同款引用计数）。
	n.consumersMu.Lock()
	if _, ok := n.consumers[channel]; !ok {
		consumer, err := newConsumer(channel, n.ephemeralID(), n.cfg)
		if err != nil {
			n.consumersMu.Unlock()
			n.logger.Error(err)
			return err
		}
		if err := n.connect(consumer, n.addr, n.lookupdAddr, &directHandler{ch: n.eventMsgCh, log: n.logger}); err != nil {
			consumer.Stop()
			n.consumersMu.Unlock()
			n.logger.Error(err)
			return err
		}
		n.consumers[channel] = consumer
	}
	n.channelRefs[channel]++
	n.consumersMu.Unlock()

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

// Leave 引用计数归零时停止并移除共享 consumer，同时移除项目选择器。
func (n *nsq) Leave(nsID int64, projectID int64) error {
	channel := getNsqProjectEventRoom(nsID)

	// 引用计数归零时才真正停止并移除共享 consumer，否则仅减计数。
	n.consumersMu.Lock()
	n.channelRefs[channel]--
	if n.channelRefs[channel] <= 0 {
		delete(n.channelRefs, channel)
		if consumer, ok := n.consumers[channel]; ok {
			consumer.Stop()
			delete(n.consumers, channel)
		}
	}
	n.consumersMu.Unlock()

	n.pMu.Lock()
	delete(n.pidSelectors, int32(projectID))
	n.pMu.Unlock()

	return nil
}

// Run 消费 pod 事件通道，按订阅选择器匹配并投递到消息通道；ctx 取消时退出。
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

			n.consumersMu.RLock()
			_, subscribed := n.channelRefs[obj.Channel]
			n.consumersMu.RUnlock()
			if !subscribed {
				continue
			}

			n.pMu.RLock()
			wssender.MatchSelectorsAndSend(n.msgCh, labels.Set(obj.Pod.Labels), n.pidSelectors, n.id, n.uid, n.logger)
			n.pMu.RUnlock()
		}
	}
}

// Publish 将 pod 事件序列化后发布到对应命名空间频道。
func (n *nsq) Publish(nsID int64, pod *v1.Pod) error {
	room := getNsqProjectEventRoom(nsID)
	// ProjectPodEventObj 的字段全部可序列化（nil Pod 序列化为 null），Marshal 恒不失败。
	marshal, _ := json.Marshal(&wssender.ProjectPodEventObj{
		NamespaceID: nsID,
		Pod:         pod,
		Channel:     room,
	})
	return n.producer.Publish(room, marshal)
}

// Info 返回空信息（NSQ 连接状态不可简单汇总）。
func (n *nsq) Info() any {
	return nil
}

// Uid 返回连接对应用户标识。
func (n *nsq) Uid() string {
	return n.uid
}

// ID 返回连接标识。
func (n *nsq) ID() string {
	return n.id
}

// ToSelf 发布到用户直连 ephemeral 通道，仅当前连接可收到。
func (n *nsq) ToSelf(response application.WebsocketMessage) error {
	return n.to(response, websocket_pb.To_ToSelf)
}

// ToAll 发布到广播 ephemeral 通道，全部订阅者均可收到。
func (n *nsq) ToAll(response application.WebsocketMessage) error {
	return n.to(response, websocket_pb.To_ToAll)
}

// to 按投递目标选择广播或直连频道，发布序列化后的 Message。
func (n *nsq) to(response application.WebsocketMessage, to websocket_pb.To) error {
	room := ephemeralBroadcastRoom
	if to == websocket_pb.To_ToSelf {
		room = n.ephemeralID()
	}
	return n.producer.Publish(room, wssender.ProtoToMessage(response, n.id, to).Marshal())
}

// ephemeralID 返回当前连接的 ephemeral 通道名。
func (n *nsq) ephemeralID() string {
	return n.ID() + "#ephemeral"
}

// closedMsgCh 返回一个已关闭通道，用于订阅失败时让消费者立即退出。
func closedMsgCh() chan []byte {
	ch := make(chan []byte)
	close(ch)
	return ch
}

// Subscribe 创建广播与直连两个 consumer 并连接 NSQ，返回消息通道。
func (n *nsq) Subscribe() <-chan []byte {
	consumerAll, err := newConsumer(ephemeralBroadcastRoom, n.ephemeralID(), n.cfg)
	if err != nil {
		n.logger.Errorf("[NSQ] create broadcast consumer: %v", err)
		return closedMsgCh()
	}
	consumer, err := newConsumer(n.ephemeralID(), n.ephemeralID(), n.cfg)
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

// connect 注册 handler 并连接 consumer：有 lookupd 走 nsqlookupd，否则直连 nsqd。
func (n *nsq) connect(consumer nsqConsumer, addr, lookupdAddr string, h gonsq.Handler) error {
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

// Close 停止全部 consumer 并等待其 handler 退出，保证只执行一次。
func (n *nsq) Close() error {
	defer n.logger.Debugf("[nsq]: id: %v closed", n.ID())
	n.closeOnce.Do(func() {
		n.consumersMu.Lock()
		var consumers []nsqConsumer
		for _, c := range n.consumers {
			c.Stop()
			consumers = append(consumers, c)
		}
		n.consumersMu.Unlock()

		// Stop() 非阻塞——发送 CLS 后立即返回。
		// 等待每个 consumer 的 StopChan（全部 handler 退出后关闭），确保 handler
		// goroutine 不再向 msgCh 写数据。
		for _, c := range consumers {
			<-c.StopChan()
		}
		// 不要 close(n.msgCh)：Run goroutine 仍可能通过 MatchSelectorsAndSend 写 msgCh，
		// send-on-closed-channel 会 panic；消费者已监听 ctx.Done() 退出，不依赖 channel
		// 关闭信号。写者全部退出后 channel 由 GC 回收。
	})
	return nil
}

// handler 解码 websocket 消息并按投递目标分发到消息通道。
type handler struct {
	id     string
	msgCh  chan []byte
	logger mlog.Logger
}

// HandleMessage 解码并按 To 字段分发；空消息或解码失败直接确认。
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
	// To_ToOthers：跳过来源连接。ToOthers 方法已删除（无生产调用方），此分支保留为防御代码。
	case websocket_pb.To_ToOthers:
		if message.ID != h.id {
			wssender.SendOrDrop(h.msgCh, message.Data, h.logger, h.id)
		}
	}

	return nil
}

// setLogLevel 将 consumer/producer 的日志级别统一为 Error 并接入 mlog 适配器。
func setLogLevel(logger mlog.Logger, s any) {
	log := NewNsqLoggerAdapter(logger)
	switch ss := s.(type) {
	case *gonsq.Consumer:
		ss.SetLoggerLevel(gonsq.LogLevelError)
		ss.SetLoggerForLevel(log, gonsq.LogLevelError)
	case *gonsq.Producer:
		ss.SetLoggerLevel(gonsq.LogLevelError)
		ss.SetLoggerForLevel(log, gonsq.LogLevelError)
	case *consumerWrapper:
		// 包装器解开后递归设置底层真实 consumer 的日志级别。
		setLogLevel(logger, ss.Consumer)
	}
}

// NsqLoggerAdapter 把 go-nsq 日志转发到 mlog：TOPIC_NOT_FOUND 降级为 Debug。
type NsqLoggerAdapter struct {
	logger mlog.Logger
}

// NewNsqLoggerAdapter 构造 NsqLoggerAdapter。
func NewNsqLoggerAdapter(logger mlog.Logger) *NsqLoggerAdapter {
	return &NsqLoggerAdapter{logger: logger}
}

// Output 实现 nsq.logger 接口，将 go-nsq 日志写入 mlog。
func (n *NsqLoggerAdapter) Output(calldepth int, s string) error {
	if strings.Contains(s, "TOPIC_NOT_FOUND") {
		n.logger.Debug(s)
	} else {
		n.logger.Error(s)
	}
	return nil
}
