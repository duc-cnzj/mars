package nsq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/plugins/wssender"
	"github.com/duc-cnzj/mars/v6/internal/plugins/wssender/wssendertest"
	gonsq "github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/labels"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// newApp 构造带真实 ent DB + ProjectRepo 的 PluginApp stub。
func newApp(t *testing.T) (app.PluginApp, biz.ProjectRepo, *ent.Client) {
	t.Helper()
	db := wssendertest.NewDB(t)
	pr := wssendertest.NewTestRepo(t, db)
	return wssendertest.FakeApp{Repo: pr, Log: mlog.NewForConfig(nil)}, pr, db
}

// newTestNSQ 构造仅供纯单元测试的 *nsq（addr/producer 为空，不发起任何连接）。
func newTestNSQ(uid, id string) *nsq {
	return &nsq{
		logger:       mlog.NewForConfig(nil),
		cfg:          gonsq.NewConfig(),
		uid:          uid,
		id:           id,
		msgCh:        make(chan []byte, wssender.MessageChSize),
		eventMsgCh:   make(chan []byte, wssender.MessageChSize),
		consumers:    map[string]nsqConsumer{},
		channelRefs:  map[string]int{},
		pidSelectors: map[int32][]labels.Selector{},
	}
}

// newTestNSQWithDB 构造带真实 ent DB + ProjectRepo 的 *nsq，供 Join/Leave 测试使用。
func newTestNSQWithDB(t *testing.T, uid, id string) (*nsq, *ent.Client) {
	t.Helper()
	n := newTestNSQ(uid, id)
	db := wssendertest.NewDB(t)
	n.projectRepo = wssendertest.NewTestRepo(t, db)
	return n, db
}

// newRealSender 经真实 nsqd Initialize 构造 *nsqSender，注册 Destroy 清理；nsqd 不可达则跳过。
func newRealSender(t *testing.T) *nsqSender {
	t.Helper()
	app, _, _ := newApp(t)
	s := &nsqSender{}
	require.NoError(t, s.Initialize(app, map[string]any{"addr": wssendertest.NSQDAddr(t)}))
	t.Cleanup(func() { _ = s.Destroy() })
	return s
}

// nsqIDSeq 连接 id 自增序列：避免多测试复用 ephemeral 频道/channel 造成串扰。
var nsqIDSeq int64

// testNSQID 返回每次调用唯一的连接 id。
func testNSQID(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("it-%d", atomic.AddInt64(&nsqIDSeq, 1))
}

// nsqDeliver 发布消息并等待投递，超时未达则重试。
// 背景：consumer 连接 nsqd 与 RDY 协商有延迟，ephemeral 频道/主题无排队，订阅生效前发布会丢；
// 连接就绪后后续发布即达，故在 deadline 内重试发布直到收到。
func nsqDeliver(t *testing.T, send func() error, ch <-chan []byte, timeout time.Duration) []byte {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		require.NoError(t, send())
		select {
		case data, ok := <-ch:
			require.True(t, ok, "channel closed unexpectedly")
			return data
		case <-time.After(50 * time.Millisecond):
			if time.Now().After(deadline) {
				t.Fatal("timeout waiting for nsq message delivery")
			}
		}
	}
}

// podEventObj 构造发布到 eventMsgCh 的 pod 事件 JSON。
func podEventObj(channel string, nsID int64) []byte {
	data, _ := json.Marshal(&wssender.ProjectPodEventObj{
		Channel:     channel,
		NamespaceID: nsID,
		Pod:         wssendertest.TestPod(),
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
	s := &nsqSender{}
	err := s.Initialize(app, map[string]any{
		"addr":          wssendertest.NSQDAddr(t),
		"lookupd_addr":  "127.0.0.1:4161",
		"msg_timeout":   10,
		"dial_timeout":  5,
		"read_timeout":  5,
		"write_timeout": 5,
		// heartbeat 必须 < read_timeout（go-nsq config.Validate），取 3s。
		"heartbeat_interval": 3,
	})
	require.NoError(t, err)
	assert.NotNil(t, s.producer)
	assert.Equal(t, wssendertest.NSQDAddr(t), s.addr)
	assert.Equal(t, "127.0.0.1:4161", s.lookupdAddr)
	assert.Same(t, pr, s.projectRepo)
	require.NotNil(t, s.cfg)
	assert.Equal(t, 10*time.Second, s.cfg.MsgTimeout)
	assert.Equal(t, 5*time.Second, s.cfg.DialTimeout)
	assert.Equal(t, 5*time.Second, s.cfg.ReadTimeout)
	assert.Equal(t, 5*time.Second, s.cfg.WriteTimeout)
	assert.Equal(t, 3*time.Second, s.cfg.HeartbeatInterval)
	require.NoError(t, s.Destroy())
}

func TestNsqInitialize_missing_addr(t *testing.T) {
	app, _, _ := newApp(t)
	s := &nsqSender{}
	err := s.Initialize(app, nil)
	assert.ErrorContains(t, err, "add not exits")
}

func TestNsqInitialize_ping_error(t *testing.T) {
	app, _, _ := newApp(t)
	s := &nsqSender{}
	// 死端口：NewProducer 懒连接成功，Ping 失败（真实失败条件）。
	err := s.Initialize(app, map[string]any{"addr": "127.0.0.1:1"})
	assert.Error(t, err)
}

func TestNsqInitialize_invalid_config_error(t *testing.T) {
	app, _, _ := newApp(t)
	s := &nsqSender{}
	// 矛盾超时（heartbeat_interval > read_timeout）触发 config.Validate() 失败，
	// NewProducer 返回 nil producer + err（真实可达分支）。校验先于连接，addr 用死端口即可。
	err := s.Initialize(app, map[string]any{
		"addr":               "127.0.0.1:1",
		"read_timeout":       5,
		"heartbeat_interval": 30,
	})
	assert.Error(t, err)
	assert.Nil(t, s.producer)
}

func TestNsqInitialize_invalid_timeout_args(t *testing.T) {
	app, _, _ := newApp(t)
	s := &nsqSender{}
	// 非法超时参数（<=0）应被忽略，保持默认值。
	require.NoError(t, s.Initialize(app, map[string]any{
		"addr":         wssendertest.NSQDAddr(t),
		"msg_timeout":  -1,
		"dial_timeout": 0,
	}))
	assert.Equal(t, gonsq.NewConfig().MsgTimeout, s.cfg.MsgTimeout)
	assert.Equal(t, gonsq.NewConfig().DialTimeout, s.cfg.DialTimeout)
	require.NoError(t, s.Destroy())
}

func TestNsqDestroy(t *testing.T) {
	app, _, _ := newApp(t)
	s := &nsqSender{}
	require.NoError(t, s.Initialize(app, map[string]any{"addr": wssendertest.NSQDAddr(t)}))
	require.NoError(t, s.Destroy())
}

// ---------------------------------------------------------------------------
// New / Uid / ID / Info
// ---------------------------------------------------------------------------

func TestNsqNew(t *testing.T) {
	s := newRealSender(t)
	n := s.New("uid-1", "id-1").(*nsq)
	assert.Equal(t, "id-1", n.ID())
	assert.Equal(t, "uid-1", n.Uid())
	assert.Equal(t, "id-1#ephemeral", n.ephemeralID())
	assert.Same(t, s.producer, n.producer)
	assert.Same(t, s.logger, n.logger)
	require.NotNil(t, n.msgCh)
	require.NotNil(t, n.eventMsgCh)
	assert.Empty(t, n.consumers)
	assert.Empty(t, n.channelRefs)
	assert.Empty(t, n.pidSelectors)
}

func TestUidIDInfo(t *testing.T) {
	n := newTestNSQ("uid", "id")
	assert.Equal(t, "uid", n.Uid())
	assert.Equal(t, "id", n.ID())
	assert.Nil(t, n.Info())
}

// ---------------------------------------------------------------------------
// ToSelf / ToAll / Publish（真实 nsqd 端到端）
// ---------------------------------------------------------------------------

func TestNsqToSelf_publishes_to_direct_channel(t *testing.T) {
	s := newRealSender(t)
	id := testNSQID(t)
	n := s.New("u1", id).(*nsq)
	ch := n.Subscribe()

	data := nsqDeliver(t, func() error { return n.ToSelf(wssendertest.TestMsg()) }, ch, 3*time.Second)
	var resp websocket_pb.WsProjectPodEventResponse
	require.NoError(t, proto.Unmarshal(data, &resp))
	assert.Equal(t, int32(42), resp.ProjectId)
	require.NoError(t, n.Close())
}

func TestNsqToAll_publishes_to_broadcast_channel(t *testing.T) {
	s := newRealSender(t)
	id1, id2 := testNSQID(t), testNSQID(t)
	n1 := s.New("u1", id1).(*nsq)
	n2 := s.New("u2", id2).(*nsq)
	ch1 := n1.Subscribe()
	ch2 := n2.Subscribe()

	// 广播投递：同一 ToAll 两个连接各收到一份；各自重试直到订阅生效。
	data1 := nsqDeliver(t, func() error { return n1.ToAll(wssendertest.TestMsg()) }, ch1, 3*time.Second)
	data2 := nsqDeliver(t, func() error { return n1.ToAll(wssendertest.TestMsg()) }, ch2, 3*time.Second)
	var resp websocket_pb.WsProjectPodEventResponse
	require.NoError(t, proto.Unmarshal(data1, &resp))
	assert.Equal(t, int32(42), resp.ProjectId)
	require.NoError(t, proto.Unmarshal(data2, &resp))
	assert.Equal(t, int32(42), resp.ProjectId)
	require.NoError(t, n1.Close())
	require.NoError(t, n2.Close())
}

func TestNsqPublish(t *testing.T) {
	_, pr, db := newApp(t)
	s := &nsqSender{}
	require.NoError(t, s.Initialize(wssendertest.FakeApp{Repo: pr, Log: mlog.NewForConfig(nil)}, map[string]any{"addr": wssendertest.NSQDAddr(t)}))
	defer func() { _ = s.Destroy() }()
	id := testNSQID(t)
	n := s.New("u", id).(*nsq)
	nsID, pid := wssendertest.SeedProject(t, db, []string{"app=test"})

	require.NoError(t, n.Join(int64(pid)))

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	go func() { _ = n.Run(ctx) }()
	time.Sleep(200 * time.Millisecond) // 等待 consumer 连接就绪

	data := nsqDeliver(t, func() error { return n.Publish(int64(nsID), wssendertest.TestPod()) }, n.msgCh, 3*time.Second)
	assert.NotEmpty(t, data)
	require.NoError(t, n.Close())
}

func TestTo_publish_error(t *testing.T) {
	app, _, _ := newApp(t)
	s := &nsqSender{}
	require.NoError(t, s.Initialize(app, map[string]any{"addr": wssendertest.NSQDAddr(t)}))
	n := s.New("u", "id").(*nsq)
	s.producer.Stop() // 停止共享 producer → 后续 Publish 失败
	assert.Error(t, n.ToSelf(wssendertest.TestMsg()))
	assert.Error(t, n.ToAll(wssendertest.TestMsg()))
}

func TestPublish_error(t *testing.T) {
	app, _, _ := newApp(t)
	s := &nsqSender{}
	require.NoError(t, s.Initialize(app, map[string]any{"addr": wssendertest.NSQDAddr(t)}))
	n := s.New("u", "id").(*nsq)
	s.producer.Stop()
	assert.Error(t, n.Publish(7, wssendertest.TestPod()))
}

// ---------------------------------------------------------------------------
// Join / Leave
// ---------------------------------------------------------------------------

func TestJoin(t *testing.T) {
	n, db := newTestNSQWithDB(t, "u", "id")
	n.addr = wssendertest.NSQDAddr(t)
	t.Cleanup(func() { _ = n.Close() })
	nsID, pid := wssendertest.SeedProject(t, db, []string{"app=test"})

	require.NoError(t, n.Join(int64(pid)))

	channel := getNsqProjectEventRoom(int64(nsID))
	n.consumersMu.RLock()
	assert.Equal(t, 1, n.channelRefs[channel])
	require.NotNil(t, n.consumers[channel], "Join 应创建真实 consumer")
	n.consumersMu.RUnlock()

	n.pMu.RLock()
	sels := n.pidSelectors[int32(pid)]
	n.pMu.RUnlock()
	require.Len(t, sels, 1)
	assert.True(t, sels[0].Matches(labels.Set(map[string]string{"app": "test"})))
}

func TestJoin_reuses_consumer_for_same_channel(t *testing.T) {
	n, db := newTestNSQWithDB(t, "u", "id")
	n.addr = wssendertest.NSQDAddr(t)
	t.Cleanup(func() { _ = n.Close() })
	nsID, pid := wssendertest.SeedProject(t, db, []string{"app=test"})

	require.NoError(t, n.Join(int64(pid)))
	channel := getNsqProjectEventRoom(int64(nsID))
	n.consumersMu.RLock()
	first := n.consumers[channel]
	n.consumersMu.RUnlock()
	require.NotNil(t, first, "首个 Join 应创建 consumer")

	// 同一 namespace 二次 Join：复用同一 consumer，引用计数 +1。
	require.NoError(t, n.Join(int64(pid)))
	n.consumersMu.RLock()
	assert.Equal(t, 2, n.channelRefs[channel])
	assert.Same(t, first, n.consumers[channel], "同一 namespace 复用同一 consumer")
	n.consumersMu.RUnlock()
}

func TestJoin_unknown_project(t *testing.T) {
	n, _ := newTestNSQWithDB(t, "u", "id")
	assert.Error(t, n.Join(99999))
}

func TestJoin_connect_error(t *testing.T) {
	n, db := newTestNSQWithDB(t, "u", "id")
	n.addr = "127.0.0.1:1" // 死端口 → connect 失败（真实失败条件）
	_, pid := wssendertest.SeedProject(t, db, []string{"app=test"})

	assert.Error(t, n.Join(int64(pid)))
	n.consumersMu.RLock()
	assert.Empty(t, n.consumers)
	n.consumersMu.RUnlock()
}

func TestJoin_invalid_selector_skipped(t *testing.T) {
	n, db := newTestNSQWithDB(t, "u", "id")
	n.addr = wssendertest.NSQDAddr(t)
	t.Cleanup(func() { _ = n.Close() })
	_, pid := wssendertest.SeedProject(t, db, []string{"!!!!bad"})

	require.NoError(t, n.Join(int64(pid)))
	n.pMu.RLock()
	assert.Empty(t, n.pidSelectors[int32(pid)])
	n.pMu.RUnlock()
}

func TestLeave(t *testing.T) {
	n, db := newTestNSQWithDB(t, "u", "id")
	n.addr = wssendertest.NSQDAddr(t)
	t.Cleanup(func() { _ = n.Close() })
	nsID, pid := wssendertest.SeedProject(t, db, []string{"app=test"})

	require.NoError(t, n.Join(int64(pid)))

	channel := getNsqProjectEventRoom(int64(nsID))
	require.NoError(t, n.Leave(int64(nsID), int64(pid)))
	n.consumersMu.RLock()
	_, ok := n.channelRefs[channel]
	assert.False(t, ok, "Leave 后引用计数应移除")
	_, ok = n.consumers[channel]
	assert.False(t, ok, "Leave 后 consumer 应移除")
	n.consumersMu.RUnlock()

	n.pMu.RLock()
	_, ok = n.pidSelectors[int32(pid)]
	assert.False(t, ok)
	n.pMu.RUnlock()
}

func TestLeave_decrements_refcount(t *testing.T) {
	n, db := newTestNSQWithDB(t, "u", "id")
	n.addr = wssendertest.NSQDAddr(t)
	t.Cleanup(func() { _ = n.Close() })
	nsID, pid := wssendertest.SeedProject(t, db, []string{"app=test"})

	require.NoError(t, n.Join(int64(pid)))
	require.NoError(t, n.Join(int64(pid)))

	require.NoError(t, n.Leave(int64(nsID), int64(pid)))
	channel := getNsqProjectEventRoom(int64(nsID))
	n.consumersMu.RLock()
	assert.Equal(t, 1, n.channelRefs[channel], "首次 Leave 仅递减引用计数")
	require.NotNil(t, n.consumers[channel])
	n.consumersMu.RUnlock()
}

// ---------------------------------------------------------------------------
// Run
// ---------------------------------------------------------------------------

func TestRun_dispatches_matching_pod_event(t *testing.T) {
	n, db := newTestNSQWithDB(t, "u", "id")
	nsID, pid := wssendertest.SeedProject(t, db, []string{"app=test"})
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
	n := newTestNSQ("u", "id")

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
	n := newTestNSQ("u", "id")

	cancel := runInBackground(n)
	defer cancel()

	n.eventMsgCh <- []byte("not-json")
	time.Sleep(100 * time.Millisecond)
	// 解码失败仅记日志，不 panic。
}

func TestRun_cancel(t *testing.T) {
	n := newTestNSQ("u", "id")
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
	n := newTestNSQ("u", "id")
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
			Body: wssender.ProtoToMessage(wssendertest.TestMsg(), from, to).Marshal(),
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
	s := newRealSender(t)
	id := testNSQID(t)
	n := s.New("u", id).(*nsq)
	ch := n.Subscribe()
	assert.Equal(t, (<-chan []byte)(n.msgCh), ch)

	n.consumersMu.RLock()
	require.Len(t, n.consumers, 2, "Subscribe 应创建广播 + 直连两个 consumer")
	n.consumersMu.RUnlock()

	// 端到端确认两个 consumer 均已连接：广播投递可收到。
	nsqDeliver(t, func() error { return n.ToAll(wssendertest.TestMsg()) }, ch, 3*time.Second)
	require.NoError(t, n.Close())
}

func TestSubscribe_direct_connect_error(t *testing.T) {
	n := newTestNSQ("u", "id")
	n.addr = "127.0.0.1:1" // 死端口 → 直连 consumer connect 失败
	ch := n.Subscribe()
	_, ok := <-ch
	assert.False(t, ok, "直连 connect 失败应返回已关闭通道")
}

func TestSubscribe_lookupd_connect_error(t *testing.T) {
	n := newTestNSQ("u", "id")
	// ConnectToNSQLookupd 对死端口是 fire-and-forget 轮询（返回 nil 不报错），
	// 唯一同步报错是地址校验：缺 port 触发 buildLookupAddr 的 "missing port"。
	n.lookupdAddr = "127.0.0.1"
	ch := n.Subscribe()
	_, ok := <-ch
	assert.False(t, ok, "lookupd 地址非法应返回已关闭通道")
}

func TestClose(t *testing.T) {
	s := newRealSender(t)
	id := testNSQID(t)
	n := s.New("u", id).(*nsq)
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
		select {
		case <-c.StopChan():
		case <-time.After(2 * time.Second):
			t.Fatal("Close 后 consumer 的 StopChan 应关闭")
		}
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

func TestNsqLoggerAdapter_Output(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := mlog.NewMockLogger(ctrl)
	// lookupd 轮询空主题的 TOPIC_NOT_FOUND 404 是高频噪音，被整体丢弃（不产生日志）。
	logger.EXPECT().Debug(gomock.Any()).Times(0)
	// 其余错误原样转发到 Error。
	logger.EXPECT().Error("some error").Times(1)

	adapter := NewNsqLoggerAdapter(logger)
	require.NoError(t, adapter.Output(2, "TOPIC_NOT_FOUND something"))
	require.NoError(t, adapter.Output(2, "some error"))
}

// ---------------------------------------------------------------------------
// misc
// ---------------------------------------------------------------------------

func TestGetNsqProjectEventRoom(t *testing.T) {
	assert.Equal(t, "project-pod-events-7#ephemeral", getNsqProjectEventRoom[int64](7))
	assert.Equal(t, "project-pod-events-7#ephemeral", getNsqProjectEventRoom(7))
}
