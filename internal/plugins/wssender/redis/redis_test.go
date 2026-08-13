package redis

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/plugins/wssender"
	"github.com/duc-cnzj/mars/v6/internal/plugins/wssender/wssendertest"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// testRedisDB 集成测试专用库号：隔离开发默认 DB，启动/清理时 FlushDB 不误伤。
const testRedisDB = 15

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// mustReceiveSubscription 从 PubSub 读取一条应答并断言其为 Subscription 确认。
// go-redis 的 Subscribe/Ping 都只写命令、不等服务端确认，此处显式消费确认应答，
// 使"订阅已在服务端生效"这一时序假设可被证明，避免 Publish 先于订阅生效而丢消息。
func mustReceiveSubscription(tb testing.TB, ps *redis.PubSub) *redis.Subscription {
	tb.Helper()
	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
	defer cancel()
	reply, err := ps.Receive(ctx)
	require.NoError(tb, err)
	sub, ok := reply.(*redis.Subscription)
	require.Truef(tb, ok, "expected *redis.Subscription, got %T", reply)
	return sub
}

// newTestSender creates a redisSender connected to a real Redis (dedicated DB).
// It flushes DB before use and cleans up on test completion. Redis 不可用时整体 t.Skip。
func newTestSender(tb testing.TB) *redisSender {
	tb.Helper()
	return newTestSenderOn(tb, wssendertest.RedisAddr(tb))
}

// newTestSenderOn 在指定真实 Redis 地址上建 sender，供跨实例测试共享同一 Redis。
func newTestSenderOn(tb testing.TB, addr string) *redisSender {
	tb.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: addr, DB: testRedisDB})
	require.NoError(tb, rdb.FlushDB(context.TODO()).Err())

	ctx, cancel := context.WithCancel(context.TODO())

	wsPubSub := rdb.Subscribe(context.TODO())
	require.NoError(tb, wsPubSub.Subscribe(context.TODO(), wssender.BroadcastRoom))
	// go-redis 的 PubSub.Subscribe 只写 SUBSCRIBE 命令、不等服务端确认（_subscribe 仅 writeCmd），
	// PubSub.Ping 同样只写不读应答：二者都无法同步"订阅已生效"。Redis PubSub 无排队，
	// 若 Publish 早于服务端处理 SUBSCRIBE 会直接丢消息（曾致 TestCrossInstanceToAll flake）。
	// 故显式读取 Subscription 确认应答，确保广播频道订阅在服务端生效后 ToAll/ToOthers 首投即达。
	sub := mustReceiveSubscription(tb, wsPubSub)
	require.Equal(tb, wssender.BroadcastRoom, sub.Channel)

	s := &redisSender{
		rds:      rdb,
		logger:   mlog.NewForConfig(nil),
		subs:     make(map[string]*subEntry),
		wsPubSub: wsPubSub,
		msgCh:    wsPubSub.Channel(),
		ctx:      ctx,
		cancel:   cancel,
	}

	go s.dispatcher()

	tb.Cleanup(func() {
		cancel()
		wsPubSub.Close()
		rdb.FlushDB(context.TODO())
		rdb.Close()
	})

	return s
}

func newPubSub(tb testing.TB, s *redisSender, uid, id string) *rdsPubSub {
	tb.Helper()
	return s.New(uid, id).(*rdsPubSub)
}

// mustRead reads from ch within the timeout and returns the data.
// Fails the test if nothing arrives.
func mustRead(t *testing.T, ch <-chan []byte, timeout time.Duration) []byte {
	t.Helper()
	select {
	case data, ok := <-ch:
		require.True(t, ok, "channel closed unexpectedly")
		return data
	case <-time.After(timeout):
		t.Fatal("timeout waiting for message")
		return nil
	}
}

// deliver 发送消息并等待投递，超时未到则重试。
// 背景：go-redis PubSub.Subscribe 只写 SUBSCRIBE 命令、不等待服务端确认（_subscribe
// 仅 writeCmd），而 Publish 走独立连接，可能先于直连频道的订阅生效而丢消息（Redis
// PubSub 无排队）。ToSelf 依赖 New 里异步订阅的用户直连频道，故须在 deadline 内重试
// 发布直到收到。ToAll 走 setup 时就绪的广播频道，首投即达，不会触发重试。
func deliver(t *testing.T, send func() error, ch <-chan []byte, timeout time.Duration) []byte {
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
				t.Fatal("timeout waiting for message delivery")
			}
		}
	}
}

// expectNoRead asserts that nothing arrives on ch within the timeout.
func expectNoRead(t *testing.T, ch <-chan []byte, timeout time.Duration) {
	t.Helper()
	select {
	case data, ok := <-ch:
		if ok {
			t.Fatalf("unexpected message: %v", data)
		}
		// channel closed — also acceptable for "no message" scenarios
	case <-time.After(timeout):
	}
}

// ---------------------------------------------------------------------------
// 1. Basic lifecycle
// ---------------------------------------------------------------------------

func TestNewPubSub(t *testing.T) {
	s := newTestSender(t)
	pub := newPubSub(t, s, "uid-1", "id-1")

	assert.Equal(t, "id-1", pub.ID())
	assert.Equal(t, "uid-1", pub.Uid())
	assert.NotNil(t, pub.Subscribe())
}

func TestSubscribe_returns_open_channel(t *testing.T) {
	s := newTestSender(t)
	pub := newPubSub(t, s, "u", "x")
	ch := pub.Subscribe()
	require.NotNil(t, ch)

	// Non-blocking check: channel is open and empty
	select {
	case _, ok := <-ch:
		if !ok {
			t.Fatal("channel should be open, not closed")
		}
		t.Fatal("unexpected message on empty channel")
	default:
		// expected — channel is open and empty
	}
}

// ---------------------------------------------------------------------------
// 2. Message routing
// ---------------------------------------------------------------------------

func TestToSelf(t *testing.T) {
	s := newTestSender(t)
	pub := newPubSub(t, s, "u1", "id1")
	ch := pub.Subscribe()

	data := deliver(t, func() error { return pub.ToSelf(wssendertest.TestMsg()) }, ch, 3*time.Second)
	var resp websocket_pb.WsProjectPodEventResponse
	require.NoError(t, proto.Unmarshal(data, &resp))
	assert.Equal(t, int32(42), resp.ProjectId)
}

func TestToAll(t *testing.T) {
	s := newTestSender(t)
	pub1 := newPubSub(t, s, "u1", "id1")
	pub2 := newPubSub(t, s, "u2", "id2")
	ch1 := pub1.Subscribe()
	ch2 := pub2.Subscribe()

	require.NoError(t, pub1.ToAll(wssendertest.TestMsg()))

	mustRead(t, ch1, 3*time.Second)
	mustRead(t, ch2, 3*time.Second)
}

// TestDispatcher_ToToOthers_skips_source 直接向广播频道投递 To_ToOthers 消息，
// 断言 dispatcher 跳过来源连接 id1。ToOthers 方法已删除（无生产调用方），
// 该路由分支保留为防御代码，仍须测试覆盖。
func TestDispatcher_ToToOthers_skips_source(t *testing.T) {
	s := newTestSender(t)
	pub1 := newPubSub(t, s, "u1", "id1")
	pub2 := newPubSub(t, s, "u2", "id2")
	ch1 := pub1.Subscribe()
	ch2 := pub2.Subscribe()

	msg := wssender.ProtoToMessage(wssendertest.TestMsg(), "id1", websocket_pb.To_ToOthers).Marshal()
	require.NoError(t, s.rds.Publish(context.TODO(), wssender.BroadcastRoom, msg).Err())

	expectNoRead(t, ch1, 500*time.Millisecond)
	mustRead(t, ch2, 3*time.Second)
}

func TestToSelf_routed_only_to_target(t *testing.T) {
	s := newTestSender(t)
	pub1 := newPubSub(t, s, "u1", "id1")
	pub2 := newPubSub(t, s, "u2", "id2")
	ch1 := pub1.Subscribe()
	ch2 := pub2.Subscribe()

	deliver(t, func() error { return pub1.ToSelf(wssendertest.TestMsg()) }, ch1, 3*time.Second)
	expectNoRead(t, ch2, 500*time.Millisecond)
}

func TestToSelf_on_closed_PubSub_is_noop(t *testing.T) {
	s := newTestSender(t)
	pub := newPubSub(t, s, "u", "id")
	pub.Close()

	// Must not panic
	err := pub.ToSelf(wssendertest.TestMsg())
	assert.NoError(t, err) // no-op is fine
}

// ---------------------------------------------------------------------------
// 3. Info / Close
// ---------------------------------------------------------------------------

func TestInfo_returns_subscriber_count(t *testing.T) {
	s := newTestSender(t)
	pub1 := newPubSub(t, s, "alice", "a")
	pub2 := newPubSub(t, s, "bob", "b")
	_ = pub2

	info := pub1.Info().(map[string]any)
	assert.Equal(t, 2, info["subscribers"])
	assert.Equal(t, "a", info["id"])
}

func TestInfo_after_close(t *testing.T) {
	s := newTestSender(t)
	pub := newPubSub(t, s, "u", "id")
	pub.Close()

	info := pub.Info().(map[string]any)
	assert.Equal(t, 0, info["subscribers"])
}

func TestClose_removes_subscriber_from_map(t *testing.T) {
	s := newTestSender(t)
	pub := newPubSub(t, s, "u", "close-id")
	_ = pub.Subscribe()

	s.mu.RLock()
	_, exists := s.subs["close-id"]
	s.mu.RUnlock()
	assert.True(t, exists, "sub should exist before Close")

	pub.Close()

	s.mu.RLock()
	_, exists = s.subs["close-id"]
	s.mu.RUnlock()
	assert.False(t, exists, "sub should be removed after Close")
}

func TestMultipleClose_no_panic(t *testing.T) {
	s := newTestSender(t)
	pub := newPubSub(t, s, "u", "id")

	pub.Close()
	assert.NotPanics(t, func() {
		pub.Close()
	})
}

// ---------------------------------------------------------------------------
// 4. sendOrDrop behavior (non-blocking send on full channel)
// ---------------------------------------------------------------------------

func TestSendOrDrop_does_not_block_when_channel_full(t *testing.T) {
	s := newTestSender(t)
	pub := newPubSub(t, s, "u", "id")
	ch := pub.Subscribe()

	// Publish more messages than the buffer can hold.
	msgCount := wssender.MessageChSize * 2
	for i := 0; i < msgCount; i++ {
		require.NoError(t, pub.ToAll(wssendertest.TestMsg()))
	}

	// Give the dispatcher time to process all messages.
	time.Sleep(500 * time.Millisecond)

	// Read whatever made it into the channel (bounded by buffer size).
	var received int
	for {
		select {
		case <-ch:
			received++
		default:
			goto done
		}
	}
done:
	assert.LessOrEqual(t, received, wssender.MessageChSize,
		"channel buffer bounds the max receivable messages")
	assert.GreaterOrEqual(t, received, 1,
		"at least some messages should have arrived")
}

// ---------------------------------------------------------------------------
// 5. Cross-instance simulation (multiple redisSenders, same Redis)
// ---------------------------------------------------------------------------

func TestCrossInstanceToSelf(t *testing.T) {
	s1 := newTestSenderOn(t, wssendertest.RedisAddr(t))
	s2 := newTestSenderOn(t, wssendertest.RedisAddr(t))

	pub1 := newPubSub(t, s1, "u1", "x-id1")
	pub2 := newPubSub(t, s2, "u2", "x-id2")
	ch1 := pub1.Subscribe()
	ch2 := pub2.Subscribe()

	// Only pub1 should receive it.
	deliver(t, func() error { return pub1.ToSelf(wssendertest.TestMsg()) }, ch1, 3*time.Second)
	expectNoRead(t, ch2, 500*time.Millisecond)

	pub1.Close()
	pub2.Close()
}

func TestCrossInstanceToAll(t *testing.T) {
	s1 := newTestSenderOn(t, wssendertest.RedisAddr(t))
	s2 := newTestSenderOn(t, wssendertest.RedisAddr(t))

	pub1 := newPubSub(t, s1, "u1", "x-id1")
	pub2 := newPubSub(t, s2, "u2", "x-id2")
	ch1 := pub1.Subscribe()
	ch2 := pub2.Subscribe()

	require.NoError(t, pub1.ToAll(wssendertest.TestMsg()))

	mustRead(t, ch1, 3*time.Second)
	mustRead(t, ch2, 3*time.Second)

	pub1.Close()
	pub2.Close()
}

// ---------------------------------------------------------------------------
// 6. Concurrency  (run: go test -race)
// ---------------------------------------------------------------------------

func TestConcurrentToAll(t *testing.T) {
	const n = 5
	s := newTestSender(t)
	pubs := make([]*rdsPubSub, n)
	chans := make([]<-chan []byte, n)
	for i := range n {
		pub := newPubSub(t, s, fmt.Sprintf("uid%d", i), fmt.Sprintf("id%d", i))
		pubs[i] = pub
		chans[i] = pub.Subscribe()
	}

	// Drainers with ctx-based exit (channels are NOT closed by Close()).
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	var drainWg sync.WaitGroup
	for i, ch := range chans {
		drainWg.Add(1)
		go func(idx int, c <-chan []byte) {
			defer drainWg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case _, ok := <-c:
					if !ok {
						return
					}
				}
			}
		}(i, ch)
	}

	var prodWg sync.WaitGroup
	for _, pub := range pubs {
		prodWg.Add(1)
		go func(p *rdsPubSub) {
			defer prodWg.Done()
			for j := 0; j < 20; j++ {
				_ = p.ToAll(wssendertest.TestMsg())
			}
		}(pub)
	}

	prodWg.Wait()
	cancel() // signal drainers to stop
	drainWg.Wait()
}

func TestConcurrentToSelf(t *testing.T) {
	s := newTestSender(t)
	pub := newPubSub(t, s, "u", "id")
	ch := pub.Subscribe()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-ch:
				if !ok {
					return
				}
			}
		}
	}()

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = pub.ToSelf(wssendertest.TestMsg())
			}
		}()
	}

	wg.Wait()
	cancel()
}

func TestConcurrentSendAndClose(t *testing.T) {
	s := newTestSender(t)
	pub := newPubSub(t, s, "u", "id")
	_ = pub.Subscribe()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			_ = pub.ToAll(wssendertest.TestMsg())
		}
	}()
	go func() {
		defer wg.Done()
		pub.Close()
	}()

	wg.Wait()
}

func TestConcurrentNewAndClose(t *testing.T) {
	s := newTestSender(t)
	var wg sync.WaitGroup

	for i := range 20 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := fmt.Sprintf("id%d", n)
			pub := newPubSub(t, s, fmt.Sprintf("u%d", n), id)
			pub.Close()
		}(i)
	}
	wg.Wait()

	s.mu.RLock()
	assert.Len(t, s.subs, 0, "all subs should be cleaned up")
	s.mu.RUnlock()
}

func TestConcurrentPublishWithoutDrainers(t *testing.T) {
	// Many publishers concurrently, no drainers — sendOrDrop prevents blocking.
	s := newTestSender(t)
	var wg sync.WaitGroup

	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			pub := newPubSub(t, s, fmt.Sprintf("u%d", n), fmt.Sprintf("id%d", n))
			defer pub.Close()
			for j := 0; j < 10; j++ {
				_ = pub.ToAll(wssendertest.TestMsg())
			}
		}(i)
	}
	wg.Wait()
}

func TestConcurrentAddDeleteSubs(t *testing.T) {
	s := newTestSender(t)
	var wg sync.WaitGroup

	for i := range 20 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := fmt.Sprintf("cid%d", n)
			pub := newPubSub(t, s, fmt.Sprintf("cu%d", n), id)
			// Subscribe and immediately close
			_ = pub.Subscribe()
			pub.Close()
		}(i)
	}
	wg.Wait()

	s.mu.RLock()
	assert.Len(t, s.subs, 0, "all subs should be cleaned up after concurrent AddDelete")
	s.mu.RUnlock()
}

// ---------------------------------------------------------------------------
// 空 id 订阅者（只发不收）：New("", "") 不应注册进 dispatcher，避免其 channel
// 被 ToAll 广播灌满后成为永久丢消息点（对齐 memory 后端 Add 的空 id 跳过）。
// ---------------------------------------------------------------------------

func TestNew_emptyID_doesNotRegisterSubscriber(t *testing.T) {
	s := newTestSender(t)
	pub := newPubSub(t, s, "", "")
	defer pub.Close()

	s.mu.RLock()
	defer s.mu.RUnlock()
	assert.Empty(t, s.subs, "空 id 订阅者（只发不收）不应注册进 dispatcher subs")
}

func TestClose_emptyID_noPanic(t *testing.T) {
	s := newTestSender(t)
	pub := newPubSub(t, s, "", "")
	pub.Close()
	assert.NotPanics(t, func() { pub.Close() }, "空 id 连接重复 Close 应幂等")
}

func TestToAll_emptyID_doesNotReceive(t *testing.T) {
	s := newTestSender(t)
	realPub := newPubSub(t, s, "u", "real-id")
	emptyPub := newPubSub(t, s, "", "")
	defer realPub.Close()
	defer emptyPub.Close()

	realCh := realPub.Subscribe()
	emptyCh := emptyPub.Subscribe()

	// 空 id pubsub 发布 ToAll：真实订阅者收到，空 id 自身（未注册）不应收到。
	deliver(t, func() error { return emptyPub.ToAll(wssendertest.TestMsg()) }, realCh, 3*time.Second)

	select {
	case msg := <-emptyCh:
		t.Fatalf("空 id 订阅者不应收到广播, got %v", msg)
	default:
	}
}

func TestConcurrentEmptyAndRealSubs(t *testing.T) {
	// 空 id（只发不收）与真实订阅者混用，-race 下验证共享 PubSub 串行化无竞态。
	s := newTestSender(t)
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := fmt.Sprintf("rid%d", n)
			pub := newPubSub(t, s, fmt.Sprintf("u%d", n), id)
			_ = pub.Subscribe()
			pub.Close()
		}(i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			empty := newPubSub(t, s, "", "")
			_ = empty.ToAll(wssendertest.TestMsg())
			empty.Close()
		}()
	}
	wg.Wait()

	s.mu.RLock()
	defer s.mu.RUnlock()
	assert.Len(t, s.subs, 0, "全部连接关闭后 subs 应清空")
}

// ---------------------------------------------------------------------------
// 7. Goroutine / channel leak detection
// ---------------------------------------------------------------------------

func TestNoGoroutineLeakAfterDestroy(t *testing.T) {
	baseline := runtime.NumGoroutine()

	// Single sender (1 Redis connection, 1 dispatcher). Multiple PubSubs share it.
	s := newTestSender(t)
	pubs := make([]*rdsPubSub, 10)
	for i := range 10 {
		pub := newPubSub(t, s, fmt.Sprintf("u%d", i), fmt.Sprintf("id%d", i))
		pubs[i] = pub
		_ = pub.Subscribe()
	}

	for _, pub := range pubs {
		pub.Close()
	}

	runtime.GC()
	time.Sleep(200 * time.Millisecond)

	after := runtime.NumGoroutine()
	// Allow room for go-redis internal goroutines (pool reaper, pubsub reader, health check)
	// that may still be winding down.
	assert.LessOrEqual(t, after, baseline+15,
		"goroutine count should not grow significantly after creating/closing PubSubs")
}

func TestChannelCleanupAfterMultipleClose(t *testing.T) {
	s := newTestSender(t)
	const count = 20

	pubs := make([]*rdsPubSub, count)
	for i := range count {
		pub := newPubSub(t, s, fmt.Sprintf("u%d", i), fmt.Sprintf("id%d", i))
		pubs[i] = pub
		_ = pub.Subscribe()
	}

	s.mu.RLock()
	assert.Len(t, s.subs, count)
	s.mu.RUnlock()

	for _, pub := range pubs {
		pub.Close()
	}

	s.mu.RLock()
	assert.Len(t, s.subs, 0, "all subs must be removed after close")
	s.mu.RUnlock()
}

// ---------------------------------------------------------------------------
// 8. Malformed / error handling in dispatcher
// ---------------------------------------------------------------------------

func TestDispatcher_handles_malformed_message(t *testing.T) {
	s := newTestSender(t)
	// Publishing garbage to BroadcastRoom should not panic the dispatcher.
	err := s.rds.Publish(context.TODO(), wssender.BroadcastRoom, []byte("not-json")).Err()
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	// If we got here without a panic, the test passes.
}

// ---------------------------------------------------------------------------
// 9. Destroy
// ---------------------------------------------------------------------------

func TestDestroy(t *testing.T) {
	s := newTestSender(t)
	pub := newPubSub(t, s, "u", "id")
	_ = pub.Subscribe()

	s.cancel() // simulate Destroy
	// Must not panic after cancel
	err := pub.ToSelf(wssendertest.TestMsg())
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkToAll(b *testing.B) {
	s := newTestSender(b)
	pub := newPubSub(b, s, "pub", "pub")
	_ = pub.Subscribe()

	// Add slow subscribers that never read — sendOrDrop prevents blocking.
	for i := range 10 {
		newPubSub(b, s, fmt.Sprintf("slow%d", i), fmt.Sprintf("slow%d", i))
	}

	msg := wssendertest.TestMsg()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = pub.ToAll(msg)
	}
}

func BenchmarkToAll_with_drainers(b *testing.B) {
	s := newTestSender(b)
	pub := newPubSub(b, s, "pub", "pub")
	_ = pub.Subscribe()

	// Draining subscribers.
	for i := range 10 {
		p := newPubSub(b, s, fmt.Sprintf("u%d", i), fmt.Sprintf("id%d", i))
		ch := p.Subscribe()
		go func() {
			for range ch {
			}
		}()
	}

	msg := wssendertest.TestMsg()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = pub.ToAll(msg)
	}
}

func BenchmarkToAll_parallel(b *testing.B) {
	s := newTestSender(b)
	pub := newPubSub(b, s, "pub", "pub")
	_ = pub.Subscribe()

	for i := range 10 {
		p := newPubSub(b, s, fmt.Sprintf("u%d", i), fmt.Sprintf("id%d", i))
		ch := p.Subscribe()
		go func() {
			for range ch {
			}
		}()
	}

	msg := wssendertest.TestMsg()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = pub.ToAll(msg)
		}
	})
}

func BenchmarkToSelf(b *testing.B) {
	s := newTestSender(b)
	pub := newPubSub(b, s, "u", "id")

	ch := pub.Subscribe()
	go func() {
		for range ch {
		}
	}()

	msg := wssendertest.TestMsg()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = pub.ToSelf(msg)
	}
}
