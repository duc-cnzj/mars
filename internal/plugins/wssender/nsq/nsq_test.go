package nsq

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v5/websocket"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/plugins/wssender"
	gonsq "github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

var (
	testNSQDAddr    = "127.0.0.1:4150"
	testLookupdAddr = "127.0.0.1:4161"
)

func init() {
	if addr := os.Getenv("NSQ_NSQD_ADDR"); addr != "" {
		testNSQDAddr = addr
	}
	if addr := os.Getenv("NSQ_LOOKUPD_ADDR"); addr != "" {
		testLookupdAddr = addr
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func nsqAvailable() bool {
	cfg := gonsq.NewConfig()
	p, err := gonsq.NewProducer(testNSQDAddr, cfg)
	if err != nil {
		return false
	}
	defer p.Stop()
	return p.Ping() == nil
}

// newTestSender creates a nsqSender connected to a local NSQD instance.
func newTestSender(tb testing.TB) *nsqSender {
	tb.Helper()
	if !nsqAvailable() {
		tb.Skipf("NSQ not available at %s (set NSQ_NSQD_ADDR/NSQ_LOOKUPD_ADDR)", testNSQDAddr)
	}

	cfg := gonsq.NewConfig()
	cfg.MaxInFlight = 1000
	// Short poll interval for faster test consumer registration.
	cfg.LookupdPollInterval = 200 * time.Millisecond

	producer, err := gonsq.NewProducer(testNSQDAddr, cfg)
	require.NoError(tb, err)
	require.NoError(tb, producer.Ping())

	// Intentionally leave lookupdAddr empty so consumers connect directly to NSQD
	// instead of via lookupd.  When NSQ runs in Docker the lookupd returns container
	// hostnames that are not resolvable from the host.
	s := &nsqSender{
		producer:    producer,
		cfg:         cfg,
		addr:        testNSQDAddr,
		lookupdAddr: "",
		logger:      mlog.NewForConfig(nil),
	}

	tb.Cleanup(func() {
		producer.Stop()
	})

	return s
}

func newNSQ(tb testing.TB, s *nsqSender, uid, id string) *nsq {
	tb.Helper()
	return s.New(uid, id).(*nsq)
}

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

// mustRead reads from ch within the timeout and returns the data.
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

// expectNoRead asserts that nothing arrives on ch within the timeout.
func expectNoRead(t *testing.T, ch <-chan []byte, timeout time.Duration) {
	t.Helper()
	select {
	case data, ok := <-ch:
		if ok {
			t.Fatalf("unexpected message: %v", data)
		}
	case <-time.After(timeout):
	}
}

// waitForSubscription pauses briefly to allow NSQ consumer registration.
// NSQ subscriptions are asynchronous; this gives the consumer time to
// negotiate the connection and register with the topic before we publish.
func waitForSubscription() {
	time.Sleep(300 * time.Millisecond)
}

// ---------------------------------------------------------------------------
// 1. Basic lifecycle
// ---------------------------------------------------------------------------

func TestNewNSQ(t *testing.T) {
	s := newTestSender(t)
	n := newNSQ(t, s, "uid-1", "id-1")

	assert.Equal(t, "id-1", n.ID())
	assert.Equal(t, "uid-1", n.Uid())
}

func TestSubscribe_returns_open_channel(t *testing.T) {
	s := newTestSender(t)
	n := newNSQ(t, s, "u", "x")
	ch := n.Subscribe()
	defer n.Close()
	require.NotNil(t, ch)

	// Non-blocking check: channel is open and empty.
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
	n := newNSQ(t, s, "u1", "id1")
	ch := n.Subscribe()
	defer n.Close()
	waitForSubscription()

	require.NoError(t, n.ToSelf(testMsg()))

	data := mustRead(t, ch, 3*time.Second)
	var resp websocket_pb.WsProjectPodEventResponse
	require.NoError(t, proto.Unmarshal(data, &resp))
	assert.Equal(t, int32(42), resp.ProjectId)
}

func TestToAll(t *testing.T) {
	s := newTestSender(t)
	n1 := newNSQ(t, s, "u1", "id1")
	n2 := newNSQ(t, s, "u2", "id2")
	ch1 := n1.Subscribe()
	ch2 := n2.Subscribe()
	defer n1.Close()
	defer n2.Close()
	waitForSubscription()

	require.NoError(t, n1.ToAll(testMsg()))

	mustRead(t, ch1, 3*time.Second)
	mustRead(t, ch2, 3*time.Second)
}

func TestToOthers(t *testing.T) {
	s := newTestSender(t)
	n1 := newNSQ(t, s, "u1", "id1")
	n2 := newNSQ(t, s, "u2", "id2")
	ch1 := n1.Subscribe()
	ch2 := n2.Subscribe()
	defer n1.Close()
	defer n2.Close()
	waitForSubscription()

	require.NoError(t, n1.ToOthers(testMsg()))

	// n1 should NOT receive its own ToOthers message.
	expectNoRead(t, ch1, 1*time.Second)
	mustRead(t, ch2, 3*time.Second)
}

func TestToSelf_routed_only_to_target(t *testing.T) {
	s := newTestSender(t)
	n1 := newNSQ(t, s, "u1", "id1")
	n2 := newNSQ(t, s, "u2", "id2")
	ch1 := n1.Subscribe()
	ch2 := n2.Subscribe()
	defer n1.Close()
	defer n2.Close()
	waitForSubscription()

	require.NoError(t, n1.ToSelf(testMsg()))

	// Only pub1 (id1) should receive it.
	mustRead(t, ch1, 3*time.Second)
	expectNoRead(t, ch2, 1*time.Second)
}

func TestToSelf_on_closed_PubSub_is_noop(t *testing.T) {
	s := newTestSender(t)
	n := newNSQ(t, s, "u", "id")
	n.Close()

	// Must not panic — publish is still possible on the shared producer,
	// but there's no consumer to receive it.
	err := n.ToSelf(testMsg())
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// 3. Info / Close
// ---------------------------------------------------------------------------

func TestInfo_returns_nil(t *testing.T) {
	s := newTestSender(t)
	n := newNSQ(t, s, "u", "id")
	assert.Nil(t, n.Info())
}

func TestClose_closes_msgCh(t *testing.T) {
	s := newTestSender(t)
	n := newNSQ(t, s, "u", "id")
	ch := n.Subscribe()

	n.Close()

	// Close() closes msgCh, but the channel buffer may contain data
	// that was enqueued by handlers before Close() was called.
	// Drain the buffer; the range loop exits only after the closed
	// channel has been fully drained.
	for range ch {
	}
}

func TestClose_stops_consumers(t *testing.T) {
	s := newTestSender(t)
	n := newNSQ(t, s, "u", "id")
	ch := n.Subscribe()

	n.consumersMu.RLock()
	consumerCount := len(n.consumers)
	n.consumersMu.RUnlock()
	assert.Equal(t, 2, consumerCount, "Subscribe should create 2 consumers (broadcast + direct)")

	n.Close()

	// The channel from Subscribe() should be closed after Close().
	_, ok := <-ch
	assert.False(t, ok, "channel should be closed after Close()")
}

func TestMultipleClose_no_panic(t *testing.T) {
	s := newTestSender(t)
	n := newNSQ(t, s, "u", "id")
	n.Close()

	assert.NotPanics(t, func() {
		n.Close()
	})
}

// ---------------------------------------------------------------------------
// 4. sendOrDrop behavior (non-blocking send on full channel)
// ---------------------------------------------------------------------------

func TestSendOrDrop_does_not_block_when_channel_full(t *testing.T) {
	s := newTestSender(t)
	n := newNSQ(t, s, "u", "id")
	ch := n.Subscribe()
	defer n.Close()
	waitForSubscription()

	// Publish more messages than the buffer can hold.
	msgCount := wssender.MessageChSize * 2
	for i := 0; i < msgCount; i++ {
		require.NoError(t, n.ToAll(testMsg()))
	}

	// Give the consumer time to process all messages.
	time.Sleep(2 * time.Second)

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
// 5. Message ordering (covers the production out-of-order bug)
// ---------------------------------------------------------------------------

func TestMessageOrdering(t *testing.T) {
	s := newTestSender(t)
	n := newNSQ(t, s, "u", "order-id")
	ch := n.Subscribe()
	defer n.Close()
	waitForSubscription()

	// Drain any stale data from the shared "all#ephemeral" topic that may
	// have been left by a previous test.  The 500ms window allows NSQ to
	// deliver in-flight messages before we start the ordering check.
	time.Sleep(500 * time.Millisecond)
drain:
	for {
		select {
		case <-ch:
		default:
			break drain
		}
	}
	const count = 100
	for i := range count {
		msg := &websocket_pb.WsProjectPodEventResponse{
			Metadata: &websocket_pb.Metadata{
				Id:   fmt.Sprintf("order-id"),
				Type: websocket_pb.Type_ProjectPodEvent,
				End:  true,
				To:   websocket_pb.To_ToSelf,
			},
			ProjectId: int32(i),
		}
		require.NoError(t, n.ToSelf(msg))
	}

	time.Sleep(3 * time.Second)

	var ids []int32
	for {
		select {
		case data := <-ch:
			var resp websocket_pb.WsProjectPodEventResponse
			if err := proto.Unmarshal(data, &resp); err == nil {
				ids = append(ids, resp.ProjectId)
			}
			if len(ids) >= count {
				goto checkOrder
			}
		default:
			goto checkOrder
		}
	}
checkOrder:
	assert.Len(t, ids, count, "should receive all %d messages", count)
	for i := 1; i < len(ids); i++ {
		assert.Greater(t, ids[i], ids[i-1],
			"messages should arrive in order at index %d (got %d after %d)", i, ids[i], ids[i-1])
	}
}

// ---------------------------------------------------------------------------
// 6. Concurrency  (run: go test -race)
// ---------------------------------------------------------------------------

func TestConcurrentToAll(t *testing.T) {
	const n = 5
	s := newTestSender(t)
	nsqs := make([]*nsq, n)
	chans := make([]<-chan []byte, n)
	for i := range n {
		nq := newNSQ(t, s, fmt.Sprintf("uid%d", i), fmt.Sprintf("id%d", i))
		nsqs[i] = nq
		chans[i] = nq.Subscribe()
	}
	// Cleanup all after test.
	t.Cleanup(func() {
		for _, nq := range nsqs {
			nq.Close()
		}
	})
	waitForSubscription()

	// Drainers with ctx-based exit.
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	var drainWg sync.WaitGroup
	for _, ch := range chans {
		drainWg.Add(1)
		go func(c <-chan []byte) {
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
		}(ch)
	}

	var prodWg sync.WaitGroup
	for _, nq := range nsqs {
		prodWg.Add(1)
		go func(nq *nsq) {
			defer prodWg.Done()
			for j := 0; j < 20; j++ {
				_ = nq.ToAll(testMsg())
			}
		}(nq)
	}

	prodWg.Wait()
	cancel() // signal drainers to stop
	drainWg.Wait()
}

func TestConcurrentToSelf(t *testing.T) {
	s := newTestSender(t)
	n := newNSQ(t, s, "u", "id")
	ch := n.Subscribe()
	defer n.Close()
	waitForSubscription()

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
				_ = n.ToSelf(testMsg())
			}
		}()
	}

	wg.Wait()
	cancel()
}

func TestConcurrentSendAndClose(t *testing.T) {
	s := newTestSender(t)
	n := newNSQ(t, s, "u", "id")
	_ = n.Subscribe()
	waitForSubscription()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			_ = n.ToAll(testMsg())
		}
	}()
	go func() {
		defer wg.Done()
		n.Close()
	}()

	wg.Wait()
}

func TestConcurrentNewAndClose(t *testing.T) {
	s := newTestSender(t)
	var wg sync.WaitGroup

	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := fmt.Sprintf("id%d", n)
			nq := newNSQ(t, s, fmt.Sprintf("u%d", n), id)
			nq.Close()
		}(i)
	}
	wg.Wait()
}

func TestConcurrentPublishWithoutDrainers(t *testing.T) {
	// Many publishers concurrently, no drainers — sendOrDrop prevents blocking.
	s := newTestSender(t)
	var wg sync.WaitGroup

	for i := range 5 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := fmt.Sprintf("cid%d", n)
			nq := newNSQ(t, s, fmt.Sprintf("cu%d", n), id)
			defer nq.Close()
			_ = nq.Subscribe()
			for j := 0; j < 10; j++ {
				_ = nq.ToAll(testMsg())
			}
		}(i)
	}
	wg.Wait()
}

func TestConcurrentAddDeleteSubs(t *testing.T) {
	s := newTestSender(t)
	var wg sync.WaitGroup

	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := fmt.Sprintf("cid%d", n)
			nq := newNSQ(t, s, fmt.Sprintf("cu%d", n), id)
			_ = nq.Subscribe()
			nq.Close()
		}(i)
	}
	wg.Wait()
}

// ---------------------------------------------------------------------------
// 7. Goroutine leak detection
// ---------------------------------------------------------------------------

func TestNoGoroutineLeakAfterDestroy(t *testing.T) {
	s := newTestSender(t)

	// Take baseline AFTER sender creation (producer goroutines are stable).
	baseline := runtime.NumGoroutine()

	nsqs := make([]*nsq, 5)
	for i := range 5 {
		nq := newNSQ(t, s, fmt.Sprintf("u%d", i), fmt.Sprintf("id%d", i))
		nsqs[i] = nq
		_ = nq.Subscribe()
	}
	waitForSubscription()

	for _, nq := range nsqs {
		nq.Close()
	}

	runtime.GC()
	time.Sleep(500 * time.Millisecond)

	after := runtime.NumGoroutine()
	// NSQ consumers spin up internal I/O goroutines that should
	// be fully cleaned up by Stop().  Allow a small headroom for
	// goroutines that may still be winding down.
	assert.LessOrEqual(t, after, baseline+10,
		"goroutine count should not grow significantly after creating/closing PubSubs")
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkToAll(b *testing.B) {
	s := newTestSender(b)
	n := newNSQ(b, s, "pub", "pub")
	_ = n.Subscribe()
	defer n.Close()
	waitForSubscription()

	// Add slow subscribers that never read — sendOrDrop prevents blocking.
	for i := range 5 {
		nq := newNSQ(b, s, fmt.Sprintf("slow%d", i), fmt.Sprintf("slow%d", i))
		_ = nq.Subscribe()
		defer nq.Close()
	}

	msg := testMsg()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = n.ToAll(msg)
	}
}

func BenchmarkToAll_with_drainers(b *testing.B) {
	s := newTestSender(b)
	n := newNSQ(b, s, "pub", "pub")
	_ = n.Subscribe()
	defer n.Close()
	waitForSubscription()

	// Draining subscribers.
	for i := range 5 {
		nq := newNSQ(b, s, fmt.Sprintf("u%d", i), fmt.Sprintf("id%d", i))
		ch := nq.Subscribe()
		defer nq.Close()
		go func() {
			for range ch {
			}
		}()
	}

	msg := testMsg()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = n.ToAll(msg)
	}
}

func BenchmarkToAll_parallel(b *testing.B) {
	s := newTestSender(b)
	n := newNSQ(b, s, "pub", "pub")
	_ = n.Subscribe()
	defer n.Close()
	waitForSubscription()

	for i := range 5 {
		nq := newNSQ(b, s, fmt.Sprintf("u%d", i), fmt.Sprintf("id%d", i))
		ch := nq.Subscribe()
		defer nq.Close()
		go func() {
			for range ch {
			}
		}()
	}

	msg := testMsg()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = n.ToAll(msg)
		}
	})
}

func BenchmarkToSelf(b *testing.B) {
	s := newTestSender(b)
	n := newNSQ(b, s, "u", "id")
	ch := n.Subscribe()
	defer n.Close()
	waitForSubscription()

	go func() {
		for range ch {
		}
	}()

	msg := testMsg()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = n.ToSelf(msg)
	}
}
