package nsq

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/websocket"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/testutil"
	"github.com/golang/mock/gomock"
	gonsq "github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/labels"
)

var (
	skip = false
)

var (
	addr        string = os.Getenv("NSQ_ADDR")
	lookupdAddr string = os.Getenv("NSQ_LOOKUPD_ADDR")
)

func TestMain(t *testing.M) {
	setDefault := func(key *string, value string) {
		if *key == "" {
			*key = value
		}
	}
	setDefault(&addr, "127.0.0.1:4150")
	setDefault(&lookupdAddr, "127.0.0.1:4161")
	config := gonsq.NewConfig()
	p, err := gonsq.NewProducer(addr, config)
	if err != nil {
		skip = true
	}
	setLogLevel(p)
	if err = p.Ping(); err != nil {
		skip = true
	}
	p.Stop()
	code := t.Run()
	os.Exit(code)
}

func NewNsqProducer() *gonsq.Producer {
	config := gonsq.NewConfig()
	p, _ := gonsq.NewProducer(addr, config)
	setLogLevel(p)
	return p
}

func TestNsqSender_Destroy(t *testing.T) {
	if skip {
		t.Skip()
	}
	ns := &NsqSender{
		producer: NewNsqProducer(),
	}
	assert.Nil(t, ns.Destroy())
}

func TestNsqSender_Initialize(t *testing.T) {
	if skip {
		t.Skip()
	}
	ns := &NsqSender{}
	assert.Nil(t, ns.Initialize(map[string]any{
		"addr":        addr,
		"lookupdAddr": lookupdAddr,
	}))
	assert.NotNil(t, ns.producer)
	ns.producer.Stop()

	assert.Error(t, ns.Initialize(map[string]any{}))
	assert.Error(t, ns.Initialize(map[string]any{
		"addr": "xxx",
	}))
}

func TestNsqSender_Name(t *testing.T) {
	if skip {
		t.Skip()
	}
	assert.Equal(t, nsqSenderName, (&NsqSender{}).Name())
}

func TestNsqSender_New(t *testing.T) {
	if skip {
		t.Skip()
	}
	cfg := gonsq.NewConfig()
	sub := (&NsqSender{
		addr:        "xxx",
		lookupdAddr: "yyy",
		producer:    NewNsqProducer(),
		cfg:         cfg,
	}).New("uid", "id")
	s := sub.(*nsq)
	assert.Equal(t, "xxx", s.addr)
	assert.Equal(t, "yyy", s.lookupdAddr)
	assert.Equal(t, cfg, s.cfg)
	assert.Equal(t, "uid", s.uid)
	assert.Equal(t, "id", s.id)
	assert.NotNil(t, s.consumers)
	assert.NotNil(t, s.producer)
	assert.NotNil(t, s.msgCh)
	assert.NotNil(t, s.eventMsgCh)
	assert.NotNil(t, s.channels)
	assert.NotNil(t, s.pidSelectors)
	s.producer.Stop()
}

func Test_connect(t *testing.T) {
	if skip {
		t.Skip()
	}
	consumer, _ := gonsq.NewConsumer("test", "duc", gonsq.NewConfig())
	assert.Nil(t, connect(consumer, addr, "", nil))
	consumer2, _ := gonsq.NewConsumer("test", "duc", gonsq.NewConfig())
	assert.Nil(t, connect(consumer2, addr, lookupdAddr, nil))
	consumer.Stop()
	consumer2.Stop()
}

func Test_getNsqProjectEventRoom(t *testing.T) {
	if skip {
		t.Skip()
	}
	consumer, _ := gonsq.NewConsumer("test", "duc1", gonsq.NewConfig())
	connect(consumer, addr, "", nil)
	consumer2, _ := gonsq.NewConsumer("test", "duc2", gonsq.NewConfig())
	connect(consumer2, addr, "", nil)
	n := &nsq{
		addr: addr,
		consumers: map[string]*gonsq.Consumer{
			"a": consumer,
			"b": consumer2,
		},
	}
	n.Close()
	_, ok := <-consumer.StopChan
	assert.False(t, ok)
	consumer3, _ := gonsq.NewConsumer("test", "duc3", gonsq.NewConfig())
	connect(consumer3, addr, "", nil)
	consumer4, _ := gonsq.NewConsumer("test", "duc4", gonsq.NewConfig())
	connect(consumer4, addr, "", nil)
	n2 := &nsq{
		addr:        addr,
		lookupdAddr: lookupdAddr,
		consumers: map[string]*gonsq.Consumer{
			"a": consumer3,
			"b": consumer4,
		},
	}
	n2.Close()
	_, ok = <-consumer3.StopChan
	assert.False(t, ok)
}

func Test_handler_HandleMessage(t *testing.T) {
	if skip {
		t.Skip()
	}
	h := &handler{msgCh: make(chan []byte, 10), id: "id"}
	assert.Nil(t, h.HandleMessage(nil))
	assert.Nil(t, h.HandleMessage(&gonsq.Message{Body: nil}))

	item := &websocket.WsMetadataResponse{}
	marshal, _ := proto.Marshal(item)
	assert.Nil(t, h.HandleMessage(&gonsq.Message{Body: plugins.ProtoToMessage(item, websocket.To_ToOthers, "xxx").Marshal()}))
	data := <-h.msgCh
	assert.Equal(t, marshal, data)

	assert.Nil(t, h.HandleMessage(&gonsq.Message{Body: plugins.ProtoToMessage(item, websocket.To_ToOthers, "id").Marshal()}))
	assert.Len(t, h.msgCh, 0)
	assert.Nil(t, h.HandleMessage(&gonsq.Message{Body: plugins.ProtoToMessage(item, websocket.To_ToOthers, "ccc").Marshal()}))
	assert.Len(t, h.msgCh, 1)
}

func Test_jsonHandler_HandleMessage(t *testing.T) {
	if skip {
		t.Skip()
	}
	h := &jsonHandler{
		ch: make(chan []byte, 10),
	}
	assert.Nil(t, h.HandleMessage(nil))
	assert.Nil(t, h.HandleMessage(&gonsq.Message{Body: nil}))

	assert.Nil(t, h.HandleMessage(&gonsq.Message{Body: []byte("aaa")}))
	assert.Equal(t, <-h.ch, []byte("aaa"))
}

func Test_nsq_ID(t *testing.T) {
	if skip {
		t.Skip()
	}
	ns := &nsq{id: "id"}
	assert.Equal(t, "id", ns.ID())
}

func Test_nsq_Info(t *testing.T) {
	if skip {
		t.Skip()
	}
	ns := &nsq{}
	assert.Equal(t, nil, ns.Info())
}

func Test_nsq_Join(t *testing.T) {
	if skip {
		t.Skip()
	}
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, f := testutil.SetGormDB(m, app)
	defer f()
	db.AutoMigrate(&models.Project{}, &models.Namespace{})
	ns := &models.Namespace{
		Name: "ns",
	}
	db.Create(ns)
	pmodel := &models.Project{
		Name:         "app",
		PodSelectors: "name=app,age=17|name=bpp,age=18",
		NamespaceId:  ns.ID,
	}
	assert.Nil(t, db.Create(pmodel).Error)
	n := &nsq{
		addr:         addr,
		cfg:          gonsq.NewConfig(),
		uid:          "uid",
		id:           "id",
		consumers:    map[string]*gonsq.Consumer{},
		msgCh:        make(chan []byte, 10),
		eventMsgCh:   make(chan []byte, 10),
		channels:     map[string]struct{}{},
		pidSelectors: map[int64][]labels.Selector{},
	}
	assert.Error(t, n.Join(10))
	assert.Nil(t, n.Join(int64(pmodel.ID)))
	channel := getNsqProjectEventRoom(pmodel.ID)
	assert.NotNil(t, n.consumers[channel])
	assert.NotNil(t, n.channels[channel])
	assert.NotNil(t, n.pidSelectors[int64(pmodel.ID)])
}

func Test_nsq_Leave(t *testing.T) {
	if skip {
		t.Skip()
	}
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, f := testutil.SetGormDB(m, app)
	defer f()
	db.AutoMigrate(&models.Project{}, &models.Namespace{})
	ns := &models.Namespace{
		Name: "ns",
	}
	db.Create(ns)
	pmodel := &models.Project{
		Name:         "app",
		PodSelectors: "name=app,age=17|name=bpp,age=18",
		NamespaceId:  ns.ID,
	}
	assert.Nil(t, db.Create(pmodel).Error)
	n := &nsq{
		addr:         addr,
		cfg:          gonsq.NewConfig(),
		uid:          "uid",
		id:           "id",
		consumers:    map[string]*gonsq.Consumer{},
		msgCh:        make(chan []byte, 10),
		eventMsgCh:   make(chan []byte, 10),
		channels:     map[string]struct{}{},
		pidSelectors: map[int64][]labels.Selector{},
	}
	assert.Nil(t, n.Join(int64(pmodel.ID)))

	assert.Nil(t, n.Leave(int64(ns.ID), int64(pmodel.ID)))

	assert.Len(t, n.pidSelectors[int64(pmodel.ID)], 0)
	assert.Len(t, n.channels, 0)
	assert.Len(t, n.consumers, 0)
}

func Test_nsq_Run(t *testing.T) {
	if skip {
		t.Skip()
	}
	ns := &NsqSender{
		producer: NewNsqProducer(),
		cfg:      gonsq.NewConfig(),
		addr:     addr,
	}
	n := ns.New("uid", "id").(*nsq)
	cancel, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	assert.Nil(t, n.Run(cancel))
	ns.Destroy()
}

func Test_nsq_Subscribe(t *testing.T) {
	if skip {
		t.Skip()
	}
	ns := &NsqSender{
		producer: NewNsqProducer(),
		cfg:      gonsq.NewConfig(),
		addr:     addr,
	}
	defer ns.Destroy()
	n := ns.New("uid", "id").(*nsq)
	ch := n.Subscribe()
	p := NewNsqProducer()
	wg := sync.WaitGroup{}
	wg.Add(1)
	cancel, cancelFunc := context.WithCancel(context.TODO())
	defer cancelFunc()
	go func() {
		wg.Done()
		n.Run(cancel)
	}()
	assert.Nil(t, p.Publish(n.ephemeralID(), []byte("aa")))
	assert.Nil(t, p.Publish(ephemeralBroadroom, []byte("bb")))
	assert.Len(t, n.consumers, 2)
	<-ch
	<-ch
	cancelFunc()
	wg.Wait()
	n.Close()
	p.Stop()
}

func Test_nsq_ToAll(t *testing.T) {
	if skip {
		t.Skip()
	}
	n := &nsq{
		id:       "id",
		producer: NewNsqProducer(),
	}
	assert.Nil(t, n.ToAll(&websocket.WsMetadataResponse{}))
	n.producer.Stop()
}

func Test_nsq_ToOthers(t *testing.T) {
	if skip {
		t.Skip()
	}
	n := &nsq{
		id:       "id",
		producer: NewNsqProducer(),
	}
	assert.Nil(t, n.ToOthers(&websocket.WsMetadataResponse{}))
	n.producer.Stop()
}

func Test_nsq_ToSelf(t *testing.T) {
	if skip {
		t.Skip()
	}
	n := &nsq{
		id:       "id",
		producer: NewNsqProducer(),
	}
	assert.Nil(t, n.ToSelf(&websocket.WsMetadataResponse{}))
	n.producer.Stop()
}

func Test_nsq_Uid(t *testing.T) {
	if skip {
		t.Skip()
	}
	ns := &nsq{uid: "uid"}
	assert.Equal(t, "uid", ns.Uid())
}

func Test_nsq_ephemeralID(t *testing.T) {
	if skip {
		t.Skip()
	}
	assert.Equal(t, "id#ephemeral", (&nsq{id: "id"}).ephemeralID())
}

func Test_setLogLevel(t *testing.T) {
	if skip {
		t.Skip()
	}
	p := NewNsqProducer()
	setLogLevel(p)
	c, _ := gonsq.NewConsumer("test", "duc3", gonsq.NewConfig())
	setLogLevel(c)
	c.Stop()
	p.Stop()
}