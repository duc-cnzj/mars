package nsq

import (
	"context"
	"os"
	"sync"
	"testing"

	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/duc-cnzj/mars-client/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/testutil"
	"github.com/duc-cnzj/mars/v4/plugins/wssender"
	"github.com/golang/mock/gomock"
	gonsq "github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/assert"
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
	config := newNsqConfig()
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
	config := newNsqConfig()
	p, _ := gonsq.NewProducer(addr, config)
	setLogLevel(p)
	return p
}

func TestNsqSender_Destroy(t *testing.T) {
	if skip {
		t.Skip()
	}
	ns := &nsqSender{
		producer: NewNsqProducer(),
	}
	assert.Nil(t, ns.Destroy())
}

func TestNsqSender_Initialize(t *testing.T) {
	if skip {
		t.Skip()
	}
	ns := &nsqSender{}
	assert.Nil(t, ns.Initialize(map[string]any{
		"addr":         addr,
		"lookupd_addr": lookupdAddr,
	}))
	assert.Equal(t, ns.lookupdAddr, lookupdAddr)
	assert.Equal(t, ns.addr, addr)
	assert.NotNil(t, ns.producer)
	ns.Destroy()
	assert.Error(t, ns.Initialize(map[string]any{}))
	assert.Error(t, ns.Initialize(map[string]any{
		"addr": "xxx",
	}))
}

func TestNsqSender_Name(t *testing.T) {
	if skip {
		t.Skip()
	}
	assert.Equal(t, nsqSenderName, (&nsqSender{}).Name())
}

func TestNsqSender_New(t *testing.T) {
	if skip {
		t.Skip()
	}
	cfg := newNsqConfig()
	sub := (&nsqSender{
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
	consumer, _ := gonsq.NewConsumer("test", "duc", newNsqConfig())
	assert.Nil(t, connect(consumer, addr, "", &noneHandler{}))
	consumer2, _ := gonsq.NewConsumer("test", "duc", newNsqConfig())
	assert.Nil(t, connect(consumer2, addr, lookupdAddr, &noneHandler{}))
	consumer.Stop()
	consumer2.Stop()
}

type noneHandler struct{}

func (n *noneHandler) HandleMessage(message *gonsq.Message) error {
	return nil
}

func Test_getNsqProjectEventRoom(t *testing.T) {
	if skip {
		t.Skip()
	}
	consumer, _ := gonsq.NewConsumer("test", "duc1", newNsqConfig())
	connect(consumer, addr, "", &noneHandler{})
	consumer2, _ := gonsq.NewConsumer("test", "duc2", newNsqConfig())
	connect(consumer2, addr, "", &noneHandler{})
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
	consumer3, _ := gonsq.NewConsumer("test", "duc3", newNsqConfig())
	connect(consumer3, addr, "", &noneHandler{})
	consumer4, _ := gonsq.NewConsumer("test", "duc4", newNsqConfig())
	connect(consumer4, addr, "", &noneHandler{})
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

	item := &websocket.WsMetadataResponse{
		Metadata: &websocket.Metadata{To: websocket.To_ToOthers},
	}
	marshal := wssender.TransformToResponse(item)
	assert.Nil(t, h.HandleMessage(&gonsq.Message{Body: wssender.ProtoToMessage(item, "xxx").Marshal()}))
	data := <-h.msgCh
	assert.Equal(t, marshal, data)

	assert.Nil(t, h.HandleMessage(&gonsq.Message{Body: wssender.ProtoToMessage(item, "id").Marshal()}))
	assert.Len(t, h.msgCh, 0)
	assert.Nil(t, h.HandleMessage(&gonsq.Message{Body: wssender.ProtoToMessage(item, "ccc").Marshal()}))
	assert.Len(t, h.msgCh, 1)
}

func Test_directHandler_HandleMessage(t *testing.T) {
	if skip {
		t.Skip()
	}
	h := &directHandler{
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
		cfg:          newNsqConfig(),
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
		cfg:          newNsqConfig(),
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
	assert.Nil(t, n.Close())
}

func Test_nsq_Run(t *testing.T) {
	if skip {
		t.Skip()
	}
	ns := &nsqSender{
		producer: NewNsqProducer(),
		cfg:      newNsqConfig(),
		addr:     addr,
	}
	n := ns.New("uid", "id").(*nsq)
	cancel, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	assert.Nil(t, n.Run(cancel))
	assert.Nil(t, ns.Destroy())
}

func Test_nsq_Subscribe(t *testing.T) {
	if skip {
		t.Skip()
	}
	ns := &nsqSender{
		producer: NewNsqProducer(),
		cfg:      newNsqConfig(),
		addr:     addr,
	}
	n := ns.New("uid", "id").(*nsq)
	ch := n.Subscribe()
	p := ns.producer
	wg := sync.WaitGroup{}
	wg.Add(1)
	cancel, cancelFunc := context.WithCancel(context.TODO())
	defer cancelFunc()
	go func() {
		defer wg.Done()
		n.Run(cancel)
	}()
	assert.Nil(t, p.Publish(n.ephemeralID(), []byte("aa")))
	assert.Nil(t, p.Publish(ephemeralBroadcastRoom, []byte("bb")))
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
	nss := &nsqSender{
		producer: NewNsqProducer(),
		cfg:      newNsqConfig(),
		addr:     addr,
	}
	n := nss.New("uid", "id").(*nsq)
	n2 := nss.New("uid-others", "id-others").(*nsq)
	defer n.Close()
	defer n2.Close()
	subscribe1 := n.Subscribe()
	subscribe2 := n2.Subscribe()
	assert.Nil(t, n.ToAll(&websocket.WsMetadataResponse{
		Metadata: &websocket.Metadata{
			Message: "to all",
		},
	}))
	defer n.producer.Stop()
	data1 := <-subscribe1
	data2 := <-subscribe2
	assert.Equal(t, websocket.To_ToAll, decodeMsg(data1).Metadata.To)
	assert.Equal(t, "id", decodeMsg(data1).Metadata.Id)
	assert.Equal(t, "uid", decodeMsg(data1).Metadata.Uid)
	assert.Equal(t, "to all", decodeMsg(data1).Metadata.Message)

	assert.Equal(t, websocket.To_ToAll, decodeMsg(data2).Metadata.To)
	assert.Equal(t, "id", decodeMsg(data2).Metadata.Id)
	assert.Equal(t, "uid", decodeMsg(data2).Metadata.Uid)
	assert.Equal(t, "to all", decodeMsg(data2).Metadata.Message)
}

func Test_nsq_ToOthers(t *testing.T) {
	if skip {
		t.Skip()
	}
	nss := &nsqSender{
		producer: NewNsqProducer(),
		cfg:      newNsqConfig(),
		addr:     addr,
	}
	n := nss.New("uid", "id").(*nsq)
	n2 := nss.New("uid-others", "id-others").(*nsq)
	defer n.Close()
	defer n2.Close()
	subscribe := n2.Subscribe()
	assert.Nil(t, n.ToOthers(&websocket.WsMetadataResponse{
		Metadata: &websocket.Metadata{
			Message: "to others",
		},
	}))
	defer n.producer.Stop()
	data := <-subscribe
	assert.Equal(t, websocket.To_ToOthers, decodeMsg(data).Metadata.To)
	assert.Equal(t, "id", decodeMsg(data).Metadata.Id)
	assert.Equal(t, "uid", decodeMsg(data).Metadata.Uid)
	assert.Equal(t, "to others", decodeMsg(data).Metadata.Message)
}

func decodeMsg(data []byte) *websocket.WsMetadataResponse {
	var msg websocket.WsMetadataResponse
	proto.Unmarshal(data, &msg)
	return &msg
}

func newNsqConfig() *gonsq.Config {
	config := gonsq.NewConfig()
	config.MaxInFlight = 1000
	return config
}

func Test_nsq_ToSelf(t *testing.T) {
	if skip {
		t.Skip()
	}
	nss := &nsqSender{
		producer: NewNsqProducer(),
		cfg:      newNsqConfig(),
		addr:     addr,
	}
	n := nss.New("uid", "id").(*nsq)
	defer n.producer.Stop()
	m := &websocket.WsMetadataResponse{
		Metadata: &websocket.Metadata{
			Id:      n.id,
			Uid:     n.uid,
			Slug:    "slug",
			Type:    1,
			End:     true,
			Result:  1,
			Message: "xxx",
			Percent: 0,
		},
	}
	m2 := &websocket.WsMetadataResponse{
		Metadata: &websocket.Metadata{
			Id:      n.id,
			Uid:     n.uid,
			Slug:    "slug",
			Type:    2,
			End:     true,
			Result:  1,
			Message: "xxx",
			Percent: 0,
		},
	}
	subscribe := n.Subscribe()
	assert.Nil(t, n.ToSelf(m))
	var data []byte
	for {
		data = <-subscribe
		var msg websocket.WsMetadataResponse
		proto.Unmarshal(data, &msg)
		if msg.Metadata.To == websocket.To_ToSelf {
			break
		}
		t.Log(msg.Metadata.Id)
	}
	marshal, _ := proto.Marshal(m)
	marshal2, _ := proto.Marshal(m2)
	assert.Equal(t, marshal, data)
	assert.NotEqual(t, marshal2, data)
	assert.Nil(t, n.Close())
	assert.Nil(t, nss.Destroy())
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
	c, _ := gonsq.NewConsumer("test", "duc3", newNsqConfig())
	setLogLevel(c)
	c.Stop()
	p.Stop()
}

func Test_nsq_Publish(t *testing.T) {
	if skip {
		t.Skip()
	}
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, f := testutil.SetGormDB(m, app)
	defer f()
	db.AutoMigrate(&models.Project{}, &models.Namespace{})
	namespace := &models.Namespace{
		Name: "ns",
	}
	db.Create(namespace)
	pmodel := &models.Project{
		Name:         "app",
		PodSelectors: "name=app",
		NamespaceId:  namespace.ID,
	}
	assert.Nil(t, db.Create(pmodel).Error)
	ns := nsqSender{
		producer: NewNsqProducer(),
		cfg:      newNsqConfig(),
		addr:     addr,
	}
	sub := ns.New("a", "a")
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		sub.Run(context.TODO())
	}()
	assert.Nil(t, sub.Join(int64(pmodel.ID)))
	ch := sub.Subscribe()
	assert.Nil(t, sub.Publish(int64(namespace.ID), &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"name": "app",
			},
		},
	}))
	assert.Nil(t, sub.Publish(int64(namespace.ID+1), &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"name": "app",
			},
		},
	}))
	<-ch
	close(sub.(*nsq).eventMsgCh)
	wg.Wait()
	assert.Len(t, sub.(*nsq).eventMsgCh, 0)
	assert.True(t, true)
	ns.Destroy()
}
