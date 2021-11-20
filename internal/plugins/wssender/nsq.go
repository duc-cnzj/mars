package wssender

import (
	"errors"

	"github.com/duc-cnzj/mars/internal/adapter"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/plugins"
	gonsq "github.com/nsqio/go-nsq"
)

const ephemeralBroadroom = BroadcastRoom + "#ephemeral"

var nsqSenderName = "ws_sender_nsq"

func init() {
	dr := &NsqSender{}
	plugins.RegisterPlugin(dr.Name(), dr)
}

type NsqSender struct {
	producer    *gonsq.Producer
	cfg         *gonsq.Config
	lookupdAddr string
	addr        string
}

func (n *NsqSender) Name() string {
	return nsqSenderName
}

func (n *NsqSender) Initialize(args map[string]interface{}) (err error) {
	n.cfg = gonsq.NewConfig()
	if s, ok := args["addr"]; ok {
		n.addr = s.(string)
	} else {
		err = errors.New("[nsq]: add not exits")
		return
	}
	if s, ok := args["lookupdAddr"]; ok {
		n.lookupdAddr = s.(string)
	}
	p, err := gonsq.NewProducer(n.addr, n.cfg)
	if err != nil {
		return err
	}
	setLogLevel(p)
	err = p.Ping()
	n.producer = p
	mlog.Info("[Plugin]: " + n.Name() + " plugin Initialize...")
	return
}

func (n *NsqSender) Destroy() error {
	n.producer.Stop()
	mlog.Info("[Plugin]: " + n.Name() + " plugin Destroy...")
	return nil
}

func (n *NsqSender) New(uid, id string) plugins.PubSub {
	return &nsq{id: id, uid: uid, cfg: n.cfg, producer: n.producer, addr: n.addr, lookupdAddr: n.lookupdAddr}
}

type nsq struct {
	addr, lookupdAddr string
	cfg               *gonsq.Config
	uid, id           string
	consumers         []*gonsq.Consumer
	producer          *gonsq.Producer
	msgCh             chan string
}

func (n *nsq) Info() interface{} {
	return nil
}

func (n *nsq) Uid() string {
	return n.uid
}

func (n *nsq) ID() string {
	return n.id
}

func (n *nsq) ToSelf(response *plugins.WsResponse) error {
	response.To = plugins.ToSelf
	return n.producer.Publish(n.ephemeralID(), response.EncodeToBytes())
}

func (n *nsq) ToAll(response *plugins.WsResponse) error {
	response.To = plugins.ToAll
	return n.producer.Publish(ephemeralBroadroom, response.EncodeToBytes())
}

func (n *nsq) ToOthers(response *plugins.WsResponse) error {
	response.To = plugins.ToOthers
	return n.producer.Publish(ephemeralBroadroom, response.EncodeToBytes())
}

func (n *nsq) ephemeralID() string {
	return n.ID() + "#ephemeral"
}

func (n *nsq) Subscribe() <-chan string {
	consumerAll, _ := gonsq.NewConsumer(ephemeralBroadroom, n.ephemeralID(), n.cfg)
	consumer, _ := gonsq.NewConsumer(n.ephemeralID(), n.ephemeralID(), n.cfg)
	setLogLevel(consumer)
	setLogLevel(consumerAll)
	n.consumers = []*gonsq.Consumer{consumer, consumerAll}

	ch := make(chan string, messageChSize)
	n.msgCh = ch
	h := &handler{msgCh: ch, id: n.id}
	consumer.AddHandler(h)
	consumerAll.AddHandler(h)
	if n.lookupdAddr != "" {
		consumerAll.ConnectToNSQLookupd(n.lookupdAddr)
		consumer.ConnectToNSQLookupd(n.lookupdAddr)
	} else {
		consumerAll.ConnectToNSQD(n.addr)
		consumer.ConnectToNSQD(n.addr)
	}

	return ch
}

func (n *nsq) Close() error {
	defer mlog.Debugf("[nsq]: id: %v closed", n.ID())
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
	msgCh chan string
}

func (h *handler) HandleMessage(m *gonsq.Message) error {
	if len(m.Body) == 0 {
		return nil
	}
	res := decodeMsg(string(m.Body))
	mlog.Debugf("[Websocket] receive msg %s", res.Data)
	switch res.To {
	case plugins.ToSelf:
		fallthrough
	case plugins.ToAll:
		h.msgCh <- res.EncodeToString()
	case plugins.ToOthers:
		if res.To == plugins.ToOthers && res.ID != h.id {
			h.msgCh <- res.EncodeToString()
		}
	}

	return nil
}

func setLogLevel(s interface{}) {
	if ss, ok := s.(*gonsq.Consumer); ok {
		ss.SetLoggerLevel(gonsq.LogLevelError)
		ss.SetLoggerForLevel(&adapter.NsqLoggerAdapter{}, gonsq.LogLevelError)
	}
	if ss, ok := s.(*gonsq.Producer); ok {
		ss.SetLoggerLevel(gonsq.LogLevelError)
		ss.SetLoggerForLevel(&adapter.NsqLoggerAdapter{}, gonsq.LogLevelError)
	}
}
