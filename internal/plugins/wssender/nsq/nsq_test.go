package nsq

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/plugins/wssender"
	gonsq "github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// ---------------------------------------------------------------------------
// fakes
// ---------------------------------------------------------------------------

// fakeProducer 实现 nsqProducer：记录发布的 topic/body，可注入 Ping/Publish 错误。
type fakeProducer struct {
	mu      sync.Mutex
	topics  []string
	bodies  [][]byte
	pingErr error
	pubErr  error
	stopped bool
}

func newFakeProducer() *fakeProducer { return &fakeProducer{} }

func (f *fakeProducer) Ping() error { return f.pingErr }

func (f *fakeProducer) Stop() { f.stopped = true }

// Publish 记录本次发布并返回注入的 pubErr（nil 表示成功）。
func (f *fakeProducer) Publish(topic string, body []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.pubErr != nil {
		return f.pubErr
	}
	f.topics = append(f.topics, topic)
	f.bodies = append(f.bodies, body)
	return nil
}

// publishedTopics 返回已发布 topic 的拷贝，避免与测试读取竞态。
func (f *fakeProducer) publishedTopics() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.topics...)
}

// publishedBodies 返回已发布 body 的拷贝，避免与测试读取竞态。
func (f *fakeProducer) publishedBodies() [][]byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([][]byte(nil), f.bodies...)
}

// fakeConsumer 实现 nsqConsumer：记录 handler/停止状态，可注入连接错误。
type fakeConsumer struct {
	mu         sync.Mutex
	handler    gonsq.Handler
	stopped    bool
	usedLookup bool
	connectErr error
	stopCh     chan int
}

func newFakeConsumer() *fakeConsumer {
	return &fakeConsumer{stopCh: make(chan int, 1)}
}

// AddHandler 记录注册的 handler。
func (f *fakeConsumer) AddHandler(h gonsq.Handler) {
	f.mu.Lock()
	f.handler = h
	f.mu.Unlock()
}

// ConnectToNSQD 直连 nsqd，返回注入的连接错误。
func (f *fakeConsumer) ConnectToNSQD(addr string) error { return f.connectErr }

// ConnectToNSQLookupd 走 nsqlookupd，返回注入的连接错误。
func (f *fakeConsumer) ConnectToNSQLookupd(addr string) error {
	f.mu.Lock()
	f.usedLookup = true
	f.mu.Unlock()
	return f.connectErr
}

// Stop 标记停止并唤醒 StopChan（多次调用幂等）。
func (f *fakeConsumer) Stop() {
	f.mu.Lock()
	f.stopped = true
	f.mu.Unlock()
	select {
	case f.stopCh <- 1:
	default:
	}
}

// StopChan 返回停止通知通道。
func (f *fakeConsumer) StopChan() <-chan int { return f.stopCh }

func (f *fakeConsumer) isStopped() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.stopped
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// fakeApp 是 PluginApp 的最小 stub，供 Initialize 测试使用。
type fakeApp struct {
	projectRepo biz.ProjectRepo
	logger      mlog.Logger
}

func (f fakeApp) Logger() mlog.Logger          { return f.logger }
func (f fakeApp) ProjectRepo() biz.ProjectRepo { return f.projectRepo }

func newDB(t *testing.T) *ent.Client {
	t.Helper()
	db, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1&loc=Local")
	require.NoError(t, err)
	require.NoError(t, db.Schema.Create(context.TODO()))
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// seedProject 创建 namespace + project 并返回 nsID、pid。
func seedProject(t *testing.T, db *ent.Client, selectors []string) (nsID, pid int) {
	t.Helper()
	ns := db.Namespace.Create().SetName("devops-test").SetCreatorEmail("a@b.c").SaveX(context.TODO())
	proj := db.Project.Create().SetName("my-app").SetCreator("tester").
		SetNamespaceID(ns.ID).SetPodSelectors(selectors).SaveX(context.TODO())
	return ns.ID, proj.ID
}

// newTestRepo 构造绑定指定 ent DB 的真实 ProjectRepo。
func newTestRepo(t *testing.T, db *ent.Client) biz.ProjectRepo {
	t.Helper()
	impl := data.NewDataImpl(&data.NewDataParams{Cfg: &config.Config{}, DB: db})
	return data.NewProjectRepo(mlog.NewForConfig(nil), impl)
}

// newApp 构造带真实 ent DB + ProjectRepo 的 PluginApp stub。
func newApp(t *testing.T) (app.PluginApp, biz.ProjectRepo, *ent.Client) {
	t.Helper()
	db := newDB(t)
	pr := newTestRepo(t, db)
	return fakeApp{projectRepo: pr, logger: mlog.NewForConfig(nil)}, pr, db
}

// newTestNSQWithDB 构造带真实 ent DB + ProjectRepo 的 *nsq，供 Join/Leave 测试使用。
func newTestNSQWithDB(t *testing.T, fp *fakeProducer, uid, id string) (*nsq, *ent.Client) {
	t.Helper()
	n := newTestNSQ(fp, uid, id)
	db := newDB(t)
	n.projectRepo = newTestRepo(t, db)
	return n, db
}

// setProducer 覆盖 newProducer 测试缝，测试结束自动还原。
func setProducer(t *testing.T, fn func(addr string, cfg *gonsq.Config) (nsqProducer, error)) {
	t.Helper()
	orig := newProducer
	newProducer = fn
	t.Cleanup(func() { newProducer = orig })
}

// setConsumer 覆盖 newConsumer 测试缝，测试结束自动还原。
func setConsumer(t *testing.T, fn func(topic, channel string, cfg *gonsq.Config) (nsqConsumer, error)) {
	t.Helper()
	orig := newConsumer
	newConsumer = fn
	t.Cleanup(func() { newConsumer = orig })
}

// newTestNSQ 构造直接注入 fakeProducer 的 *nsq，供单元测试使用。
func newTestNSQ(fp *fakeProducer, uid, id string) *nsq {
	return &nsq{
		logger:       mlog.NewForConfig(nil),
		cfg:          gonsq.NewConfig(),
		uid:          uid,
		id:           id,
		producer:     fp,
		msgCh:        make(chan []byte, wssender.MessageChSize),
		eventMsgCh:   make(chan []byte, wssender.MessageChSize),
		consumers:    map[string]nsqConsumer{},
		channelRefs:  map[string]int{},
		pidSelectors: map[int32][]labels.Selector{},
	}
}

// testMsg 构造一个最小可序列化的 websocket 消息。
func testMsg() *websocket_pb.WsProjectPodEventResponse {
	return &websocket_pb.WsProjectPodEventResponse{
		Metadata: &websocket_pb.Metadata{
			Id:     "",
			Type:   websocket_pb.Type_ProjectPodEvent,
			End:    true,
			Result: websocket_pb.ResultType_Success,
			To:     websocket_pb.To_ToAll,
		},
		ProjectId: 42,
	}
}

// testPod 构造带选择器匹配标签的测试 Pod。
func testPod() *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "devops-test",
			Labels:    map[string]string{"app": "test"},
		},
	}
}

// podEventObj 构造发布到 eventMsgCh 的 pod 事件 JSON。
func podEventObj(channel string, nsID int64) []byte {
	data, _ := json.Marshal(&wssender.ProjectPodEventObj{
		Channel:     channel,
		NamespaceID: nsID,
		Pod:         testPod(),
	})
	return data
}

// runInBackground 启动 Run 并等 50ms 让其就绪，返回取消函数。
func runInBackground(n *nsq) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.TODO())
	go func() { _ = n.Run(ctx) }()
	time.Sleep(50 * time.Millisecond)
	return cancel
}

// ---------------------------------------------------------------------------
// Name / Initialize / Destroy
// ---------------------------------------------------------------------------

func TestNsqName(t *testing.T) {
	assert.Equal(t, "ws_sender_nsq", (&nsqSender{}).Name())
}

func TestNsqInitialize_success(t *testing.T) {
	app, pr, _ := newApp(t)
	fp := newFakeProducer()
	setProducer(t, func(addr string, cfg *gonsq.Config) (nsqProducer, error) {
		return fp, nil
	})

	s := &nsqSender{}
	err := s.Initialize(app, map[string]any{
		"addr":               "127.0.0.1:4150",
		"lookupd_addr":       "127.0.0.1:4161",
		"msg_timeout":        10,
		"dial_timeout":       5,
		"read_timeout":       5,
		"write_timeout":      5,
		"heartbeat_interval": 30,
	})
	require.NoError(t, err)
	assert.Same(t, fp, s.producer)
	assert.Equal(t, "127.0.0.1:4150", s.addr)
	assert.Equal(t, "127.0.0.1:4161", s.lookupdAddr)
	assert.Same(t, pr, s.projectRepo)
	require.NotNil(t, s.cfg)
	assert.Equal(t, 10*time.Second, s.cfg.MsgTimeout)
	assert.Equal(t, 5*time.Second, s.cfg.DialTimeout)
	assert.Equal(t, 5*time.Second, s.cfg.ReadTimeout)
	assert.Equal(t, 5*time.Second, s.cfg.WriteTimeout)
	assert.Equal(t, 30*time.Second, s.cfg.HeartbeatInterval)
	require.NoError(t, s.Destroy())
}

func TestNsqInitialize_missing_addr(t *testing.T) {
	app, _, _ := newApp(t)
	s := &nsqSender{}
	err := s.Initialize(app, nil)
	assert.ErrorContains(t, err, "add not exits")
}

func TestNsqInitialize_producer_error(t *testing.T) {
	app, _, _ := newApp(t)
	setProducer(t, func(addr string, cfg *gonsq.Config) (nsqProducer, error) {
		return nil, errors.New("boom")
	})
	s := &nsqSender{}
	err := s.Initialize(app, map[string]any{"addr": "127.0.0.1:4150"})
	assert.ErrorContains(t, err, "boom")
}

func TestNsqInitialize_ping_error(t *testing.T) {
	app, _, _ := newApp(t)
	fp := newFakeProducer()
	fp.pingErr = errors.New("ping failed")
	setProducer(t, func(addr string, cfg *gonsq.Config) (nsqProducer, error) {
		return fp, nil
	})
	s := &nsqSender{}
	err := s.Initialize(app, map[string]any{"addr": "127.0.0.1:4150"})
	assert.ErrorContains(t, err, "ping failed")
	assert.True(t, fp.stopped, "Ping 失败后必须 Stop 释放连接")
}

func TestNsqInitialize_invalid_timeout_args(t *testing.T) {
	app, _, _ := newApp(t)
	fp := newFakeProducer()
	setProducer(t, func(addr string, cfg *gonsq.Config) (nsqProducer, error) {
		return fp, nil
	})
	s := &nsqSender{}
	// 非法超时参数（<=0）应被忽略，保持默认值。
	err := s.Initialize(app, map[string]any{
		"addr":         "127.0.0.1:4150",
		"msg_timeout":  -1,
		"dial_timeout": 0,
	})
	require.NoError(t, err)
	assert.Equal(t, gonsq.NewConfig().MsgTimeout, s.cfg.MsgTimeout)
	assert.Equal(t, gonsq.NewConfig().DialTimeout, s.cfg.DialTimeout)
}

func TestNsqDestroy(t *testing.T) {
	app, _, _ := newApp(t)
	fp := newFakeProducer()
	setProducer(t, func(addr string, cfg *gonsq.Config) (nsqProducer, error) {
		return fp, nil
	})
	s := &nsqSender{}
	require.NoError(t, s.Initialize(app, map[string]any{"addr": "127.0.0.1:4150"}))
	require.NoError(t, s.Destroy())
	assert.True(t, fp.stopped)
}

// ---------------------------------------------------------------------------
// New / Uid / ID / Info
// ---------------------------------------------------------------------------

func TestNsqNew(t *testing.T) {
	fp := newFakeProducer()
	s := &nsqSender{
		logger:   mlog.NewForConfig(nil),
		cfg:      gonsq.NewConfig(),
		producer: fp,
	}
	n := s.New("uid-1", "id-1").(*nsq)
	assert.Equal(t, "id-1", n.ID())
	assert.Equal(t, "uid-1", n.Uid())
	assert.Equal(t, "id-1#ephemeral", n.ephemeralID())
	assert.Same(t, fp, n.producer)
	assert.Same(t, s.logger, n.logger)
	require.NotNil(t, n.msgCh)
	require.NotNil(t, n.eventMsgCh)
	assert.Empty(t, n.consumers)
	assert.Empty(t, n.channelRefs)
	assert.Empty(t, n.pidSelectors)
}

func TestUidIDInfo(t *testing.T) {
	n := newTestNSQ(newFakeProducer(), "uid", "id")
	assert.Equal(t, "uid", n.Uid())
	assert.Equal(t, "id", n.ID())
	assert.Nil(t, n.Info())
}

// ---------------------------------------------------------------------------
// ToSelf / ToAll / Publish
// ---------------------------------------------------------------------------

func TestToSelf_publishes_to_direct_channel(t *testing.T) {
	fp := newFakeProducer()
	n := newTestNSQ(fp, "u1", "id1")

	require.NoError(t, n.ToSelf(testMsg()))

	topics := fp.publishedTopics()
	require.Len(t, topics, 1)
	assert.Equal(t, "id1#ephemeral", topics[0])

	bodies := fp.publishedBodies()
	msg, err := wssender.DecodeMessage(bodies[0])
	require.NoError(t, err)
	assert.Equal(t, websocket_pb.To_ToSelf, msg.To)
	assert.Equal(t, "id1", msg.ID)
	assert.NotEmpty(t, msg.Data)
}

func TestToAll_publishes_to_broadcast_channel(t *testing.T) {
	fp := newFakeProducer()
	n := newTestNSQ(fp, "u1", "id1")
	require.NoError(t, n.ToAll(testMsg()))
	assert.Equal(t, []string{ephemeralBroadcastRoom}, fp.publishedTopics())
}

func TestTo_publish_error(t *testing.T) {
	fp := newFakeProducer()
	fp.pubErr = errors.New("publish failed")
	n := newTestNSQ(fp, "u", "id")
	assert.Error(t, n.ToSelf(testMsg()))
	assert.Error(t, n.ToAll(testMsg()))
}

func TestPublish(t *testing.T) {
	fp := newFakeProducer()
	n := newTestNSQ(fp, "u", "id")

	require.NoError(t, n.Publish(7, testPod()))

	topics := fp.publishedTopics()
	require.Len(t, topics, 1)
	assert.Equal(t, getNsqProjectEventRoom(int64(7)), topics[0])

	bodies := fp.publishedBodies()
	var obj wssender.ProjectPodEventObj
	require.NoError(t, json.Unmarshal(bodies[0], &obj))
	assert.Equal(t, int64(7), obj.NamespaceID)
	assert.Equal(t, topics[0], obj.Channel)
	require.NotNil(t, obj.Pod)
	assert.Equal(t, "pod-1", obj.Pod.Name)
}

func TestPublish_error(t *testing.T) {
	fp := newFakeProducer()
	fp.pubErr = errors.New("boom")
	n := newTestNSQ(fp, "u", "id")
	assert.Error(t, n.Publish(7, testPod()))
}

// ---------------------------------------------------------------------------
// Join / Leave
// ---------------------------------------------------------------------------

func TestJoin(t *testing.T) {
	fp := newFakeProducer()
	n, db := newTestNSQWithDB(t, fp, "u", "id")
	nsID, pid := seedProject(t, db, []string{"app=test"})

	var topics []string
	fc := newFakeConsumer()
	setConsumer(t, func(topic, channel string, cfg *gonsq.Config) (nsqConsumer, error) {
		topics = append(topics, topic)
		return fc, nil
	})

	require.NoError(t, n.Join(int64(pid)))

	channel := getNsqProjectEventRoom(int64(nsID))
	assert.Equal(t, []string{channel}, topics)
	n.consumersMu.RLock()
	assert.Equal(t, 1, n.channelRefs[channel])
	assert.Same(t, fc, n.consumers[channel])
	n.consumersMu.RUnlock()

	n.pMu.RLock()
	sels := n.pidSelectors[int32(pid)]
	n.pMu.RUnlock()
	require.Len(t, sels, 1)
	assert.True(t, sels[0].Matches(labels.Set(map[string]string{"app": "test"})))
}

func TestJoin_reuses_consumer_for_same_channel(t *testing.T) {
	fp := newFakeProducer()
	n, db := newTestNSQWithDB(t, fp, "u", "id")
	nsID, pid := seedProject(t, db, []string{"app=test"})

	var created int
	fc := newFakeConsumer()
	setConsumer(t, func(topic, channel string, cfg *gonsq.Config) (nsqConsumer, error) {
		created++
		return fc, nil
	})

	require.NoError(t, n.Join(int64(pid)))
	require.NoError(t, n.Join(int64(pid)))
	assert.Equal(t, 1, created, "同一 namespace 复用同一 consumer")

	channel := getNsqProjectEventRoom(int64(nsID))
	n.consumersMu.RLock()
	assert.Equal(t, 2, n.channelRefs[channel])
	n.consumersMu.RUnlock()
}

func TestJoin_unknown_project(t *testing.T) {
	fp := newFakeProducer()
	n, _ := newTestNSQWithDB(t, fp, "u", "id")
	assert.Error(t, n.Join(99999))
}

func TestJoin_consumer_create_error(t *testing.T) {
	fp := newFakeProducer()
	n, db := newTestNSQWithDB(t, fp, "u", "id")
	_, pid := seedProject(t, db, []string{"app=test"})

	setConsumer(t, func(topic, channel string, cfg *gonsq.Config) (nsqConsumer, error) {
		return nil, errors.New("create failed")
	})
	assert.Error(t, n.Join(int64(pid)))

	n.consumersMu.RLock()
	assert.Empty(t, n.channelRefs, "失败时不应登记引用计数")
	n.consumersMu.RUnlock()
}

func TestJoin_connect_error(t *testing.T) {
	fp := newFakeProducer()
	n, db := newTestNSQWithDB(t, fp, "u", "id")
	_, pid := seedProject(t, db, []string{"app=test"})

	fc := newFakeConsumer()
	fc.connectErr = errors.New("connect failed")
	setConsumer(t, func(topic, channel string, cfg *gonsq.Config) (nsqConsumer, error) {
		return fc, nil
	})
	assert.Error(t, n.Join(int64(pid)))
	assert.True(t, fc.isStopped(), "连接失败必须 Stop 释放")

	n.consumersMu.RLock()
	assert.Empty(t, n.consumers)
	n.consumersMu.RUnlock()
}

func TestJoin_invalid_selector_skipped(t *testing.T) {
	fp := newFakeProducer()
	n, db := newTestNSQWithDB(t, fp, "u", "id")
	_, pid := seedProject(t, db, []string{"!!!!bad"})

	fc := newFakeConsumer()
	setConsumer(t, func(topic, channel string, cfg *gonsq.Config) (nsqConsumer, error) {
		return fc, nil
	})
	require.NoError(t, n.Join(int64(pid)))
	n.pMu.RLock()
	assert.Empty(t, n.pidSelectors[int32(pid)])
	n.pMu.RUnlock()
}

func TestLeave(t *testing.T) {
	fp := newFakeProducer()
	n, db := newTestNSQWithDB(t, fp, "u", "id")
	nsID, pid := seedProject(t, db, []string{"app=test"})

	fc := newFakeConsumer()
	setConsumer(t, func(topic, channel string, cfg *gonsq.Config) (nsqConsumer, error) {
		return fc, nil
	})
	require.NoError(t, n.Join(int64(pid)))

	channel := getNsqProjectEventRoom(int64(nsID))
	require.NoError(t, n.Leave(int64(nsID), int64(pid)))
	n.consumersMu.RLock()
	_, ok := n.channelRefs[channel]
	assert.False(t, ok)
	_, ok = n.consumers[channel]
	assert.False(t, ok)
	n.consumersMu.RUnlock()
	assert.True(t, fc.isStopped())

	n.pMu.RLock()
	_, ok = n.pidSelectors[int32(pid)]
	assert.False(t, ok)
	n.pMu.RUnlock()
}

func TestLeave_decrements_refcount(t *testing.T) {
	fp := newFakeProducer()
	n, db := newTestNSQWithDB(t, fp, "u", "id")
	nsID, pid := seedProject(t, db, []string{"app=test"})

	fc := newFakeConsumer()
	setConsumer(t, func(topic, channel string, cfg *gonsq.Config) (nsqConsumer, error) {
		return fc, nil
	})
	require.NoError(t, n.Join(int64(pid)))
	require.NoError(t, n.Join(int64(pid)))

	require.NoError(t, n.Leave(int64(nsID), int64(pid)))
	channel := getNsqProjectEventRoom(int64(nsID))
	n.consumersMu.RLock()
	assert.Equal(t, 1, n.channelRefs[channel])
	assert.Same(t, fc, n.consumers[channel])
	n.consumersMu.RUnlock()
	assert.False(t, fc.isStopped())
}

// ---------------------------------------------------------------------------
// Run
// ---------------------------------------------------------------------------

func TestRun_dispatches_matching_pod_event(t *testing.T) {
	fp := newFakeProducer()
	n, db := newTestNSQWithDB(t, fp, "u", "id")
	nsID, pid := seedProject(t, db, []string{"app=test"})
	sel, err := labels.Parse("app=test")
	require.NoError(t, err)
	channel := getNsqProjectEventRoom(int64(nsID))
	n.channelRefs[channel] = 1
	n.pidSelectors[int32(pid)] = []labels.Selector{sel}

	cancel := runInBackground(n)
	defer cancel()

	n.eventMsgCh <- podEventObj(channel, int64(nsID))

	select {
	case out := <-n.msgCh:
		assert.NotEmpty(t, out)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for pod event dispatch")
	}
}

func TestRun_skips_unsubscribed_channel(t *testing.T) {
	n := newTestNSQ(newFakeProducer(), "u", "id")

	cancel := runInBackground(n)
	defer cancel()

	n.eventMsgCh <- podEventObj("some-other-channel", 7)
	time.Sleep(100 * time.Millisecond)
	select {
	case <-n.msgCh:
		t.Fatal("should not dispatch unsubscribed channel")
	default:
	}
}

func TestRun_malformed_json(t *testing.T) {
	n := newTestNSQ(newFakeProducer(), "u", "id")

	cancel := runInBackground(n)
	defer cancel()

	n.eventMsgCh <- []byte("not-json")
	time.Sleep(100 * time.Millisecond)
	// 解码失败仅记日志，不 panic。
}

func TestRun_cancel(t *testing.T) {
	n := newTestNSQ(newFakeProducer(), "u", "id")
	ctx, cancel := context.WithCancel(context.TODO())

	done := make(chan error, 1)
	go func() { done <- n.Run(ctx) }()
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		assert.Error(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
}

func TestRun_ch_closed(t *testing.T) {
	n := newTestNSQ(newFakeProducer(), "u", "id")
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- n.Run(ctx) }()
	time.Sleep(50 * time.Millisecond)
	close(n.eventMsgCh)

	select {
	case err := <-done:
		assert.ErrorContains(t, err, "closed")
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after eventMsgCh closed")
	}
}

// ---------------------------------------------------------------------------
// handler / directHandler
// ---------------------------------------------------------------------------

func TestHandler_HandleMessage(t *testing.T) {
	ch := make(chan []byte, 8)
	h := &handler{id: "id1", msgCh: ch, logger: mlog.NewForConfig(nil)}
	send := func(to websocket_pb.To, from string) {
		require.NoError(t, h.HandleMessage(&gonsq.Message{
			Body: wssender.ProtoToMessage(testMsg(), from, to).Marshal(),
		}))
	}

	// ToSelf → 投递。
	send(websocket_pb.To_ToSelf, "id1")
	assert.Len(t, ch, 1)

	// ToAll → 投递。
	send(websocket_pb.To_ToAll, "any")
	assert.Len(t, ch, 2)

	// ToOthers + 来源非本连接 → 投递。
	send(websocket_pb.To_ToOthers, "other")
	assert.Len(t, ch, 3)

	// ToOthers + 来源为本连接 → 跳过。
	send(websocket_pb.To_ToOthers, "id1")
	assert.Len(t, ch, 3)

	// nil / 空 body → 直接确认，不投递。
	require.NoError(t, h.HandleMessage(nil))
	require.NoError(t, h.HandleMessage(&gonsq.Message{}))
	assert.Len(t, ch, 3)

	// 解码失败 → 记日志并确认，不投递。
	require.NoError(t, h.HandleMessage(&gonsq.Message{Body: []byte("not-json")}))
	assert.Len(t, ch, 3)
}

func TestDirectHandler_HandleMessage(t *testing.T) {
	ch := make(chan []byte, 8)
	d := &directHandler{ch: ch, log: mlog.NewForConfig(nil)}

	require.NoError(t, d.HandleMessage(&gonsq.Message{Body: []byte("payload")}))
	assert.Equal(t, []byte("payload"), <-ch)

	// nil / 空 body → 直接确认，不投递。
	require.NoError(t, d.HandleMessage(nil))
	require.NoError(t, d.HandleMessage(&gonsq.Message{}))
	assert.Empty(t, ch)
}

// ---------------------------------------------------------------------------
// Subscribe / connect / Close
// ---------------------------------------------------------------------------

func TestSubscribe_success(t *testing.T) {
	fp := newFakeProducer()
	n := newTestNSQ(fp, "u", "id")

	setConsumer(t, func(topic, channel string, cfg *gonsq.Config) (nsqConsumer, error) {
		return newFakeConsumer(), nil
	})

	ch := n.Subscribe()
	assert.Equal(t, (<-chan []byte)(n.msgCh), ch)

	n.consumersMu.RLock()
	require.Len(t, n.consumers, 2, "Subscribe 应创建广播 + 直连两个 consumer")
	for _, c := range n.consumers {
		f, ok := c.(*fakeConsumer)
		require.True(t, ok)
		_, ok = f.handler.(*handler)
		assert.True(t, ok, "consumer 必须注册 *handler")
	}
	n.consumersMu.RUnlock()
}

func TestSubscribe_with_lookupd_uses_lookupd_addr(t *testing.T) {
	fp := newFakeProducer()
	n := newTestNSQ(fp, "u", "id")
	n.lookupdAddr = "127.0.0.1:4161"

	fc := newFakeConsumer()
	setConsumer(t, func(topic, channel string, cfg *gonsq.Config) (nsqConsumer, error) {
		return fc, nil
	})
	_ = n.Subscribe()
	assert.True(t, fc.usedLookup)
}

func TestSubscribe_consumer_create_error(t *testing.T) {
	fp := newFakeProducer()
	n := newTestNSQ(fp, "u", "id")

	setConsumer(t, func(topic, channel string, cfg *gonsq.Config) (nsqConsumer, error) {
		return nil, errors.New("create failed")
	})
	ch := n.Subscribe()
	_, ok := <-ch
	assert.False(t, ok, "广播 consumer 创建失败应返回已关闭通道")
}

func TestSubscribe_direct_consumer_create_error(t *testing.T) {
	fp := newFakeProducer()
	n := newTestNSQ(fp, "u", "id")

	var calls int
	fc := newFakeConsumer()
	setConsumer(t, func(topic, channel string, cfg *gonsq.Config) (nsqConsumer, error) {
		calls++
		if calls == 1 {
			return fc, nil
		}
		return nil, errors.New("create failed")
	})
	ch := n.Subscribe()
	_, ok := <-ch
	assert.False(t, ok)
	assert.True(t, fc.isStopped(), "直连 consumer 创建失败时广播 consumer 必须 Stop")
}

func TestSubscribe_direct_connect_error(t *testing.T) {
	fp := newFakeProducer()
	n := newTestNSQ(fp, "u", "id")

	fc := newFakeConsumer()
	fc.connectErr = errors.New("connect failed")
	setConsumer(t, func(topic, channel string, cfg *gonsq.Config) (nsqConsumer, error) {
		return fc, nil
	})
	ch := n.Subscribe()
	_, ok := <-ch
	assert.False(t, ok)
	assert.True(t, fc.isStopped())
}

func TestSubscribe_broadcast_connect_error(t *testing.T) {
	fp := newFakeProducer()
	n := newTestNSQ(fp, "u", "id")

	fcDirect := newFakeConsumer()
	fcBroadcast := newFakeConsumer()
	fcBroadcast.connectErr = errors.New("connect failed")
	calls := 0
	setConsumer(t, func(topic, channel string, cfg *gonsq.Config) (nsqConsumer, error) {
		calls++
		if calls == 1 {
			return fcBroadcast, nil
		}
		return fcDirect, nil
	})
	ch := n.Subscribe()
	_, ok := <-ch
	assert.False(t, ok)
	assert.True(t, fcDirect.isStopped())
	assert.True(t, fcBroadcast.isStopped())
}

func TestClose(t *testing.T) {
	fp := newFakeProducer()
	n := newTestNSQ(fp, "u", "id")

	setConsumer(t, func(topic, channel string, cfg *gonsq.Config) (nsqConsumer, error) {
		return newFakeConsumer(), nil
	})
	_ = n.Subscribe()

	n.consumersMu.RLock()
	cs := make([]nsqConsumer, 0, len(n.consumers))
	for _, c := range n.consumers {
		cs = append(cs, c)
	}
	n.consumersMu.RUnlock()
	require.Len(t, cs, 2)

	require.NoError(t, n.Close())
	for _, c := range cs {
		assert.True(t, c.(*fakeConsumer).isStopped(), "Close 必须停止全部 consumer")
	}

	// 重复 Close 幂等，不 panic。
	require.NoError(t, n.Close())
}

func TestConsumerWrapper_StopChan(t *testing.T) {
	c, err := gonsq.NewConsumer("topic", "channel", gonsq.NewConfig())
	require.NoError(t, err)
	// 必须注册 handler：go-nsq 仅在 handlerLoop 退出时触发 exit() 关闭 StopChan，
	// 无 handler 时 Stop() 后 StopChan 永闭（生产代码 connect 中总是先 AddHandler）。
	c.AddHandler(gonsq.HandlerFunc(func(m *gonsq.Message) error { return nil }))
	w := &consumerWrapper{Consumer: c}

	c.Stop()
	select {
	case <-w.StopChan():
	case <-time.After(2 * time.Second):
		t.Fatal("StopChan 应在 Stop 后关闭")
	}
}

// ---------------------------------------------------------------------------
// setLogLevel / NsqLoggerAdapter
// ---------------------------------------------------------------------------

func TestSetLogLevel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := mlog.NewMockLogger(ctrl)
	// gonsq 后台 goroutine 可能在 setLogLevel 与 Stop 之间写入，允许任意调用；
	// 适配器路由行为由 TestNsqLoggerAdapter_Output 单独严格断言。
	logger.EXPECT().Error(gomock.Any()).AnyTimes()
	logger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any()).AnyTimes()
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

	c, err := gonsq.NewConsumer("topic", "channel", gonsq.NewConfig())
	require.NoError(t, err)

	// *consumerWrapper 分支：解开后递归设置底层真实 consumer。
	w := &consumerWrapper{Consumer: c}
	assert.NotPanics(t, func() { setLogLevel(logger, w) })

	// *gonsq.Consumer 分支。注册 handler 保证 Stop() 能触发 exit() 关闭 StopChan，
	// 避免 rdyLoop goroutine 泄漏（StopChan 的关闭语义见 TestConsumerWrapper_StopChan）。
	c.AddHandler(gonsq.HandlerFunc(func(m *gonsq.Message) error { return nil }))
	assert.NotPanics(t, func() { setLogLevel(logger, c) })
	c.Stop()

	// *gonsq.Producer 分支。
	p, err := gonsq.NewProducer("127.0.0.1:1", gonsq.NewConfig())
	require.NoError(t, err)
	assert.NotPanics(t, func() { setLogLevel(logger, p) })
	p.Stop()

	// 未知类型：no-op，不 panic。
	assert.NotPanics(t, func() { setLogLevel(logger, "unknown") })
}

// ---------------------------------------------------------------------------
// 默认测试缝直测（不覆写 newProducer/newConsumer）：消除 97.5% 覆盖缺口
// ---------------------------------------------------------------------------

// TestDefaultNewProducer_success 直接调用默认 newProducer 缝，验证 go-nsq 真实构造成功。
// 背景：其余测试全部覆写该缝注入 fake，默认闭包从未执行，拉低包覆盖至 97.5%。
func TestDefaultNewProducer_success(t *testing.T) {
	p, err := newProducer("127.0.0.1:1", gonsq.NewConfig())
	require.NoError(t, err)
	assert.NotNil(t, p)
	p.Stop() // NewProducer 仅分配结构体、不发起连接，Stop 安全。
}

// TestDefaultNewProducer_invalid_config 验证默认缝透传 go-nsq 配置校验错误。
func TestDefaultNewProducer_invalid_config(t *testing.T) {
	cfg := gonsq.NewConfig()
	cfg.MaxInFlight = -1 // 小于 min=0 → config.Validate() 失败
	_, err := newProducer("127.0.0.1:1", cfg)
	assert.Error(t, err)
}

// TestDefaultNewConsumer_success 直接调用默认 newConsumer 缝，验证真实 consumer 包装成功。
func TestDefaultNewConsumer_success(t *testing.T) {
	c, err := newConsumer("topic", "channel", gonsq.NewConfig())
	require.NoError(t, err)
	// 必须先注册 handler：go-nsq 仅在 handlerLoop 退出时关闭 StopChan，
	// 无 handler 时 Stop() 后 rdyLoop 泄漏（语义见 TestConsumerWrapper_StopChan）。
	c.AddHandler(gonsq.HandlerFunc(func(m *gonsq.Message) error { return nil }))
	c.Stop()
	select {
	case <-c.StopChan():
	case <-time.After(2 * time.Second):
		t.Fatal("StopChan 应在 Stop 后关闭")
	}
}

// TestDefaultNewConsumer_invalid_topic 验证默认缝透传非法 topic 的错误。
func TestDefaultNewConsumer_invalid_topic(t *testing.T) {
	_, err := newConsumer("", "channel", gonsq.NewConfig()) // 空 topic 不合法
	assert.Error(t, err)
}

func TestNsqLoggerAdapter_Output(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := mlog.NewMockLogger(ctrl)
	logger.EXPECT().Debug("TOPIC_NOT_FOUND something").Times(1)
	logger.EXPECT().Error("some error").Times(1)

	adapter := NewNsqLoggerAdapter(logger)
	require.NoError(t, adapter.Output(2, "TOPIC_NOT_FOUND something"))
	require.NoError(t, adapter.Output(2, "some error"))
}

// ---------------------------------------------------------------------------
// misc
// ---------------------------------------------------------------------------

func TestGetNsqProjectEventRoom(t *testing.T) {
	assert.Equal(t, "project-pod-events:7#ephemeral", getNsqProjectEventRoom[int64](7))
	assert.Equal(t, "project-pod-events:7#ephemeral", getNsqProjectEventRoom(7))
}
