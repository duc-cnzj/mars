package redis

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/plugins/wssender"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// fakeApp 是 PluginApp 的最小 stub，供 Initialize 测试使用。
type fakeApp struct {
	data   data.Data
	logger mlog.Logger
}

func (f fakeApp) Logger() mlog.Logger { return f.logger }
func (f fakeApp) Data() data.Data     { return f.data }
func (f fakeApp) Cache() data.Cache   { return nil }

func newDB(t *testing.T) *ent.Client {
	t.Helper()
	db, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1&loc=Local")
	require.NoError(t, err)
	require.NoError(t, db.Schema.Create(context.TODO()))
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// seedProject 创建 namespace + project 并返回 nsID、projectID。
func seedProject(t *testing.T, db *ent.Client, selectors []string) (nsID, pid int) {
	t.Helper()
	ns := db.Namespace.Create().SetName("devops-test").SetCreatorEmail("a@b.c").SaveX(context.TODO())
	proj := db.Project.Create().SetName("my-app").SetCreator("tester").
		SetNamespaceID(ns.ID).SetPodSelectors(selectors).SaveX(context.TODO())
	return ns.ID, proj.ID
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

// newPEM 构造连接 miniredis 的 podEventManagers，测试结束自动关闭。
func newPEM(t *testing.T, mr *miniredis.Miniredis, db *ent.Client) *podEventManagers {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr(), DB: 1})
	t.Cleanup(func() { _ = rdb.Close() })
	return &podEventManagers{
		db:           db,
		logger:       mlog.NewForConfig(nil),
		id:           "id-1",
		uid:          "u-1",
		rds:          rdb,
		pubSub:       rdb.Subscribe(context.TODO()),
		ch:           make(chan []byte, wssender.MessageChSize),
		channelRefs:  make(map[string]int),
		pidSelectors: make(map[int32][]labels.Selector),
	}
}

// ---------------------------------------------------------------------------
// lifecycle: Name / Initialize / Destroy
// ---------------------------------------------------------------------------

func TestRedisName(t *testing.T) {
	assert.Equal(t, "ws_sender_redis", (&redisSender{}).Name())
}

func TestRedisInitialize_success(t *testing.T) {
	mr := miniredis.RunT(t)
	db := newDB(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	md := data.NewMockData(ctrl)
	md.EXPECT().DB().Return(db)

	s := &redisSender{}
	err := s.Initialize(fakeApp{data: md, logger: mlog.NewForConfig(nil)}, map[string]any{
		"addr":     mr.Addr(),
		"password": "",
		"db":       1,
	})
	require.NoError(t, err)
	assert.NotNil(t, s.rds)
	assert.NotNil(t, s.subs)
	assert.NotNil(t, s.wsPubSub)
	assert.NotNil(t, s.msgCh)
	assert.Same(t, db, s.db)
	require.NoError(t, s.Destroy())
}

func TestRedisInitialize_missing_addr(t *testing.T) {
	s := &redisSender{}
	err := s.Initialize(fakeApp{logger: mlog.NewForConfig(nil)}, nil)
	assert.ErrorContains(t, err, "addr")
}

func TestRedisInitialize_bad_addr(t *testing.T) {
	db := newDB(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	md := data.NewMockData(ctrl)
	md.EXPECT().DB().Return(db)

	s := &redisSender{}
	err := s.Initialize(fakeApp{data: md, logger: mlog.NewForConfig(nil)}, map[string]any{
		"addr": "127.0.0.1:1", // 连接拒绝 → Ping 失败
	})
	assert.Error(t, err)
}

func TestRedisInitialize_subscribe_error(t *testing.T) {
	mr := miniredis.RunT(t)
	db := newDB(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	md := data.NewMockData(ctrl)
	md.EXPECT().DB().Return(db)

	orig := wsSubscribe
	wsSubscribe = func(*redis.PubSub, context.Context, string) error {
		return redis.ErrClosed
	}
	t.Cleanup(func() { wsSubscribe = orig })

	s := &redisSender{}
	err := s.Initialize(fakeApp{data: md, logger: mlog.NewForConfig(nil)}, map[string]any{"addr": mr.Addr()})
	assert.Error(t, err)
}

func TestRedisDestroy(t *testing.T) {
	s := newTestSender(t)
	assert.NoError(t, s.Destroy())
}

// ---------------------------------------------------------------------------
// dispatcher: msgCh 关闭分支
// ---------------------------------------------------------------------------

func TestDispatcher_msgCh_closed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	ch := make(chan *redis.Message)
	s := &redisSender{logger: mlog.NewForConfig(nil), msgCh: ch, ctx: ctx, cancel: cancel}

	done := make(chan struct{})
	go func() {
		s.dispatcher()
		close(done)
	}()
	close(ch)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("dispatcher did not exit after msgCh closed")
	}
}

// ---------------------------------------------------------------------------
// podEventManagers: Publish
// ---------------------------------------------------------------------------

func TestPodEventPublish(t *testing.T) {
	mr := miniredis.RunT(t)
	pem := newPEM(t, mr, nil)
	channel := wssender.GetProjectPodEventRoom(7)

	require.NoError(t, pem.pubSub.Subscribe(context.TODO(), channel))

	require.NoError(t, pem.Publish(7, testPod()))

	select {
	case msg := <-pem.pubSub.Channel():
		var obj wssender.ProjectPodEventObj
		require.NoError(t, json.Unmarshal([]byte(msg.Payload), &obj))
		assert.Equal(t, int64(7), obj.NamespaceID)
		assert.Equal(t, channel, obj.Channel)
		assert.Equal(t, "pod-1", obj.Pod.Name)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for published pod event")
	}
}

func TestPodEventPublish_error(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	_ = rdb.Close() // 关闭客户端 → Publish 失败
	pem := &podEventManagers{rds: rdb, logger: mlog.NewForConfig(nil), ch: make(chan []byte, 8)}

	err := pem.Publish(7, testPod())
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// podEventManagers: Join / Leave
// ---------------------------------------------------------------------------

func TestPodEventJoin(t *testing.T) {
	mr := miniredis.RunT(t)
	db := newDB(t)
	nsID, pid := seedProject(t, db, []string{"app=test"})
	pem := newPEM(t, mr, db)

	require.NoError(t, pem.Join(int64(pid)))

	channel := wssender.GetProjectPodEventRoom(nsID)
	pem.mu.RLock()
	assert.Equal(t, 1, pem.channelRefs[channel])
	pem.mu.RUnlock()

	pem.pmu.RLock()
	sels := pem.pidSelectors[int32(pid)]
	pem.pmu.RUnlock()
	require.Len(t, sels, 1)
	assert.True(t, sels[0].Matches(labels.Set(map[string]string{"app": "test"})))
}

func TestPodEventJoin_invalid_selector_skipped(t *testing.T) {
	mr := miniredis.RunT(t)
	db := newDB(t)
	_, pid := seedProject(t, db, []string{"!!!!bad"})
	pem := newPEM(t, mr, db)

	require.NoError(t, pem.Join(int64(pid)))
	pem.pmu.RLock()
	assert.Empty(t, pem.pidSelectors[int32(pid)])
	pem.pmu.RUnlock()
}

func TestPodEventJoin_unknown_project(t *testing.T) {
	mr := miniredis.RunT(t)
	pem := newPEM(t, mr, newDB(t))
	assert.Error(t, pem.Join(99999))
}

func TestPodEventJoin_subscribe_error(t *testing.T) {
	mr := miniredis.RunT(t)
	db := newDB(t)
	_, pid := seedProject(t, db, []string{"app=test"})
	pem := newPEM(t, mr, db)
	require.NoError(t, pem.pubSub.Close()) // 关闭 pubsub → Subscribe 失败

	assert.Error(t, pem.Join(int64(pid)))
}

func TestPodEventLeave(t *testing.T) {
	mr := miniredis.RunT(t)
	db := newDB(t)
	nsID, pid := seedProject(t, db, []string{"app=test"})
	pem := newPEM(t, mr, db)

	require.NoError(t, pem.Join(int64(pid)))
	require.NoError(t, pem.Join(int64(pid))) // 引用计数 → 2
	channel := wssender.GetProjectPodEventRoom(nsID)
	pem.mu.RLock()
	assert.Equal(t, 2, pem.channelRefs[channel])
	pem.mu.RUnlock()

	// 首次 Leave：计数降到 1，选择器按项目移除。
	require.NoError(t, pem.Leave(int64(nsID), int64(pid)))
	pem.mu.RLock()
	assert.Equal(t, 1, pem.channelRefs[channel])
	pem.mu.RUnlock()
	pem.pmu.RLock()
	_, ok := pem.pidSelectors[int32(pid)]
	pem.pmu.RUnlock()
	assert.False(t, ok)

	// 二次 Leave：计数归零，channel 与 consumer 一并移除。
	require.NoError(t, pem.Leave(int64(nsID), int64(pid)))
	pem.mu.RLock()
	_, ok = pem.channelRefs[channel]
	pem.mu.RUnlock()
	assert.False(t, ok)
}

func TestPodEventLeave_unsubscribe_error(t *testing.T) {
	mr := miniredis.RunT(t)
	db := newDB(t)
	nsID, pid := seedProject(t, db, []string{"app=test"})
	pem := newPEM(t, mr, db)
	require.NoError(t, pem.Join(int64(pid)))
	require.NoError(t, pem.pubSub.Close()) // 关闭 pubsub → Unsubscribe 失败

	assert.Error(t, pem.Leave(int64(nsID), int64(pid)))
}

// ---------------------------------------------------------------------------
// podEventManagers: Run
// ---------------------------------------------------------------------------

func TestPodEventRun_dispatches(t *testing.T) {
	mr := miniredis.RunT(t)
	db := newDB(t)
	nsID, pid := seedProject(t, db, []string{"app=test"})
	pem := newPEM(t, mr, db)
	require.NoError(t, pem.Join(int64(pid)))

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	go func() { _ = pem.Run(ctx) }()
	time.Sleep(100 * time.Millisecond) // 等待订阅就绪

	require.NoError(t, pem.Publish(int64(nsID), testPod()))

	select {
	case data := <-pem.ch:
		assert.NotEmpty(t, data)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for pod event dispatch")
	}
}

func TestPodEventRun_skips_unsubscribed_channel(t *testing.T) {
	mr := miniredis.RunT(t)
	pem := newPEM(t, mr, newDB(t))

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	go func() { _ = pem.Run(ctx) }()
	time.Sleep(100 * time.Millisecond)

	// 订阅一个 channelRefs 之外的频道：Run 应跳过该消息不 panic。
	require.NoError(t, pem.pubSub.Subscribe(context.TODO(), "extra"))
	require.NoError(t, pem.rds.Publish(context.TODO(), "extra", `{"channel":"extra"}`).Err())
	time.Sleep(100 * time.Millisecond)
}

func TestPodEventRun_malformed_json(t *testing.T) {
	mr := miniredis.RunT(t)
	db := newDB(t)
	nsID, pid := seedProject(t, db, []string{"app=test"})
	pem := newPEM(t, mr, db)
	channel := wssender.GetProjectPodEventRoom(nsID)
	require.NoError(t, pem.Join(int64(pid)))

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	go func() { _ = pem.Run(ctx) }()
	time.Sleep(100 * time.Millisecond)

	// 合法频道但 payload 非法 JSON：Run 记日志后继续，不 panic。
	require.NoError(t, pem.rds.Publish(context.TODO(), channel, "not-json").Err())
	time.Sleep(100 * time.Millisecond)
}

func TestPodEventRun_cancel(t *testing.T) {
	mr := miniredis.RunT(t)
	pem := newPEM(t, mr, newDB(t))
	ctx, cancel := context.WithCancel(context.TODO())

	done := make(chan error, 1)
	go func() { done <- pem.Run(ctx) }()
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		assert.Error(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
}

func TestPodEventRun_ch_closed(t *testing.T) {
	mr := miniredis.RunT(t)
	pem := newPEM(t, mr, newDB(t))
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- pem.Run(ctx) }()
	time.Sleep(100 * time.Millisecond)
	require.NoError(t, pem.pubSub.Close())

	select {
	case err := <-done:
		assert.Error(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after pubsub close")
	}
}

// ---------------------------------------------------------------------------
// rdsPubSub: 发布错误分支（客户端已关闭）
// ---------------------------------------------------------------------------

func TestToSelf_closed_client_returns_error(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	require.NoError(t, rdb.Ping(context.TODO()).Err())
	_ = rdb.Close()

	p := &rdsPubSub{
		rds:     rdb,
		id:      "id",
		uid:     "uid",
		manager: &redisSender{},
		ch:      make(chan []byte, wssender.MessageChSize),
	}
	err := p.ToSelf(testMsg())
	assert.Error(t, err)
}
