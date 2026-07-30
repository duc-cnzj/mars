package memory

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v5/websocket"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newTestSender() *memorySender {
	return &memorySender{
		conns:   make(map[string]*Conn),
		idRooms: make(map[string]map[int32]struct{}),
		rooms:   make(namespaceRooms),
		logger:  mlog.NewForConfig(nil),
	}
}

func testMsg() *websocket_pb.WsProjectPodEventResponse {
	return &websocket_pb.WsProjectPodEventResponse{
		Metadata: &websocket_pb.Metadata{
			Id:   "",
			Type: websocket_pb.Type_ProjectPodEvent,
			End:  true,
			To:   websocket_pb.To_ToAll,
		},
		ProjectId: 42,
	}
}

// mustRead attempts to read from ch within the timeout and returns the data.
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
	ms := newTestSender()
	pub := ms.New("uid-1", "id-1").(*memoryPubSub)

	assert.Equal(t, "id-1", pub.ID())
	assert.Equal(t, "uid-1", pub.Uid())
	assert.NotNil(t, pub.Subscribe())
}

func TestAddThenSubscribe(t *testing.T) {
	ms := newTestSender()
	pub := ms.New("u", "x").(*memoryPubSub)
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

func TestToSelf(t *testing.T) {
	ms := newTestSender()
	pub := ms.New("u1", "id1").(*memoryPubSub)
	ch := pub.Subscribe()

	err := pub.ToSelf(testMsg())
	require.NoError(t, err)

	data := mustRead(t, ch, time.Second)
	var resp websocket_pb.WsProjectPodEventResponse
	require.NoError(t, proto.Unmarshal(data, &resp))
	assert.Equal(t, int32(42), resp.ProjectId)
}

func TestToAll(t *testing.T) {
	ms := newTestSender()
	pub1 := ms.New("u1", "id1").(*memoryPubSub)
	pub2 := ms.New("u2", "id2").(*memoryPubSub)
	ch1 := pub1.Subscribe()
	ch2 := pub2.Subscribe()

	require.NoError(t, pub1.ToAll(testMsg()))

	mustRead(t, ch1, time.Second)
	mustRead(t, ch2, time.Second)
}

func TestToOthers(t *testing.T) {
	ms := newTestSender()
	pub1 := ms.New("u1", "id1").(*memoryPubSub)
	pub2 := ms.New("u2", "id2").(*memoryPubSub)
	ch1 := pub1.Subscribe()
	ch2 := pub2.Subscribe()

	require.NoError(t, pub1.ToOthers(testMsg()))

	expectNoRead(t, ch1, 200*time.Millisecond)
	mustRead(t, ch2, time.Second)
}

func TestToSelf_on_PubSub_that_was_Closed(t *testing.T) {
	ms := newTestSender()
	pub := ms.New("u", "id").(*memoryPubSub)
	pub.Close()
	// Must not panic
	err := pub.ToSelf(testMsg())
	assert.NoError(t, err) // no-op is fine
}

// ---------------------------------------------------------------------------
// 2. Close / Delete
// ---------------------------------------------------------------------------

func TestClose_cleans_up_connection(t *testing.T) {
	ms := newTestSender()
	pub := ms.New("u", "id").(*memoryPubSub)
	pub.Close()

	ms.connMu.RLock()
	_, ok := ms.conns["id"]
	ms.connMu.RUnlock()
	assert.False(t, ok, "connection should be removed")
}

func TestMultipleClose_no_panic(t *testing.T) {
	ms := newTestSender()
	pub := ms.New("u", "id").(*memoryPubSub)

	pub.Close()
	assert.NotPanics(t, func() {
		pub.Close()
	})
}

func TestClose_closes_channel(t *testing.T) {
	ms := newTestSender()
	pub := ms.New("u", "id").(*memoryPubSub)
	ch := pub.Subscribe()
	pub.Close()

	// After close, the channel should be closed — reading gives !ok
	select {
	case _, ok := <-ch:
		assert.False(t, ok, "channel should be closed")
	default:
		// The close may not have propagated yet; give it a moment
		time.Sleep(50 * time.Millisecond)
		_, ok := <-ch
		assert.False(t, ok, "channel should be closed after short wait")
	}
}

func TestSubscribe_unknown_id_returns_closed_channel(t *testing.T) {
	ms := newTestSender()
	pub := &memoryPubSub{
		manager: ms,
		uid:     "no-such",
		id:      "no-such",
		logger:  mlog.NewForConfig(nil),
	}
	ch := pub.Subscribe()
	_, ok := <-ch
	assert.False(t, ok, "should get closed channel")
}

// ---------------------------------------------------------------------------
// 3. Info safety
// ---------------------------------------------------------------------------

func TestInfo_returns_snapshot_not_reference(t *testing.T) {
	ms := newTestSender()
	ms.New("u1", "id1")

	info := ms.New("u1", "id1").(*memoryPubSub).Info()
	infoMap, ok := info.(map[string]string)
	require.True(t, ok, "Info must return map[string]string")
	assert.Contains(t, infoMap, "id1")

	// Modify the returned map — must NOT affect internal state
	infoMap["id1"] = "hacked"
	info2 := ms.New("u1", "id1").(*memoryPubSub).Info()
	infoMap2 := info2.(map[string]string)
	assert.Equal(t, "u1", infoMap2["id1"], "internal state must not be affected")
}

func TestInfo_contains_all_connections(t *testing.T) {
	ms := newTestSender()
	ms.New("alice", "a")
	ms.New("bob", "b")
	ms.New("carol", "c")

	info := ms.New("alice", "a").(*memoryPubSub).Info().(map[string]string)
	assert.Equal(t, "alice", info["a"])
	assert.Equal(t, "bob", info["b"])
	assert.Equal(t, "carol", info["c"])
	assert.Len(t, info, 3)
}

func TestInfo_after_close(t *testing.T) {
	ms := newTestSender()
	pub := ms.New("u", "id").(*memoryPubSub)
	pub.Close()

	info := pub.Info().(map[string]string)
	assert.NotContains(t, info, "id")
}

// ---------------------------------------------------------------------------
// 4. Duplicate Add
// ---------------------------------------------------------------------------

func TestAdd_duplicate_id_does_not_replace(t *testing.T) {
	ms := newTestSender()
	ms.Add("original-uid", "same-id")

	ms.connMu.Lock()
	origCh := ms.conns["same-id"].ch
	ms.connMu.Unlock()

	// Add again with different uid
	ms.Add("new-uid", "same-id")

	ms.connMu.RLock()
	defer ms.connMu.RUnlock()

	conn, ok := ms.conns["same-id"]
	require.True(t, ok, "conn must still exist")
	assert.Equal(t, "original-uid", conn.uid, "uid must not be overwritten")
	assert.Equal(t, origCh, conn.ch, "channel must not be replaced")
	assert.Len(t, ms.conns, 1, "only one entry")
}

// ---------------------------------------------------------------------------
// 5. Add/Delete edge cases
// ---------------------------------------------------------------------------

func TestAdd_empty_uid_or_id_is_noop(t *testing.T) {
	ms := newTestSender()
	ms.Add("", "id") // empty uid
	assert.Len(t, ms.conns, 0)

	ms.Add("uid", "") // empty id
	assert.Len(t, ms.conns, 0)
}

func TestDelete_non_existent_id_does_not_panic(t *testing.T) {
	ms := newTestSender()
	ms.Delete("u", "non-existent") // must not panic
}

// ---------------------------------------------------------------------------
// 6. Concurrency: data-race tests  (run: go test -race)
// ---------------------------------------------------------------------------

func TestConcurrentToAll(t *testing.T) {
	const n = 5
	ms := newTestSender()
	pubs := make([]*memoryPubSub, n)
	chans := make([]<-chan []byte, n)
	for i := range n {
		pub := ms.New(fmt.Sprintf("uid%d", i), fmt.Sprintf("id%d", i)).(*memoryPubSub)
		pubs[i] = pub
		chans[i] = pub.Subscribe()
	}

	// Drainers must start BEFORE producers, otherwise channel buffers fill up and deadlock.
	var drainWg sync.WaitGroup
	for i, ch := range chans {
		drainWg.Add(1)
		go func(idx int, c <-chan []byte) {
			defer drainWg.Done()
			for range c {
			}
			t.Logf("drainer %d exited", idx)
		}(i, ch)
	}

	var prodWg sync.WaitGroup
	for _, pub := range pubs {
		prodWg.Add(1)
		go func(p *memoryPubSub) {
			defer prodWg.Done()
			for j := 0; j < 20; j++ {
				_ = p.ToAll(testMsg())
			}
		}(pub)
	}

	prodWg.Wait()
	// Close all publishers so drainer for-range loops exit.
	for _, pub := range pubs {
		pub.Close()
	}
	drainWg.Wait()
}

func TestConcurrentToSelf(t *testing.T) {
	ms := newTestSender()
	pub := ms.New("u", "id").(*memoryPubSub)
	ch := pub.Subscribe()

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = pub.ToSelf(testMsg())
			}
		}()
	}

	// Drain concurrently
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range ch {
		}
	}()

	wg.Wait()
	pub.Close()
	<-done
}

func TestConcurrentSendAndClose(t *testing.T) {
	ms := newTestSender()
	pub := ms.New("u", "id").(*memoryPubSub)
	_ = pub.Subscribe()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			_ = pub.ToAll(testMsg())
		}
	}()
	go func() {
		defer wg.Done()
		pub.Close()
	}()

	wg.Wait()
}

func TestConcurrentAddDelete(t *testing.T) {
	ms := newTestSender()
	var wg sync.WaitGroup

	for i := range 20 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := fmt.Sprintf("id%d", n)
			ms.Add(fmt.Sprintf("u%d", n), id)
			ms.Delete(fmt.Sprintf("u%d", n), id)
		}(i)
	}
	wg.Wait()

	ms.connMu.RLock()
	// The map may be non-empty if Add+Delete interleaved with other Adds,
	// but every key must have a valid Conn.
	for id, conn := range ms.conns {
		assert.NotNil(t, conn)
		assert.Equal(t, conn.id, id)
	}
	ms.connMu.RUnlock()
}

func TestConcurrentAddAndToAll(t *testing.T) {
	ms := newTestSender()
	pub := ms.New("pub", "pub").(*memoryPubSub)
	_ = pub.Subscribe()

	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ms.Add(fmt.Sprintf("u%d", n), fmt.Sprintf("id%d", n))
		}(i)
	}

	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = pub.ToAll(testMsg())
		}()
	}

	wg.Wait()
}

// ---------------------------------------------------------------------------
// 7. Channel / goroutine leak detection
// ---------------------------------------------------------------------------

func TestNoGoroutineLeakAfterClose(t *testing.T) {
	// Record baseline goroutine count
	baseline := runtime.NumGoroutine()

	// Create and destroy 100 PubSubs
	for i := range 100 {
		ms := newTestSender()
		pub := ms.New(fmt.Sprintf("u%d", i), fmt.Sprintf("id%d", i)).(*memoryPubSub)
		_ = pub.Subscribe()
		pub.Close()
	}

	// Force GC and give goroutines time to settle
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	after := runtime.NumGoroutine()
	// Allow some slack (Go runtime goroutines, GC, etc.)
	assert.LessOrEqual(t, after, baseline+2,
		"goroutine count should not grow significantly after creating/closing PubSubs")
}

func TestChannelNotLeakedAfterClose(t *testing.T) {
	ms := newTestSender()
	const count = 50

	pubs := make([]*memoryPubSub, count)
	for i := range count {
		pub := ms.New(fmt.Sprintf("u%d", i), fmt.Sprintf("id%d", i)).(*memoryPubSub)
		pubs[i] = pub
		_ = pub.Subscribe()
	}

	// All connections should exist
	ms.connMu.RLock()
	assert.Len(t, ms.conns, count)
	ms.connMu.RUnlock()

	// Close all
	for _, pub := range pubs {
		pub.Close()
	}

	// All connections should be cleaned up
	ms.connMu.RLock()
	assert.Len(t, ms.conns, 0, "all connections must be removed after close")
	ms.connMu.RUnlock()
}

func TestToAll_does_not_panic_when_other_conn_is_closed(t *testing.T) {
	ms := newTestSender()
	pubA := ms.New("a", "a").(*memoryPubSub)
	pubB := ms.New("b", "b").(*memoryPubSub)
	_ = pubA.Subscribe()
	_ = pubB.Subscribe()

	// Close B, then A sends ToAll. B's channel is closed but A must not panic
	pubB.Close()
	assert.NotPanics(t, func() {
		_ = pubA.ToAll(testMsg())
	})
}

// ---------------------------------------------------------------------------
// 8. Room operations (Join / Leave / Publish) — DB-dependent, limited scope
//    These tests verify locking safety and data structure integrity without DB.
// ---------------------------------------------------------------------------

func TestRoomDataStructure_after_Join_then_Leave(t *testing.T) {
	ms := newTestSender()
	pub := ms.New("u", "id").(*memoryPubSub)

	// Simulate Join without DB: manually populate room data structures
	nsID := int32(100)
	projectID := int32(200)

	pub.manager.roomMu.Lock()
	if _, ok := pub.manager.rooms[nsID]; !ok {
		pub.manager.rooms[nsID] = make(projectSubscriptions)
	}
	if _, ok := pub.manager.rooms[nsID][projectID]; !ok {
		pub.manager.rooms[nsID][projectID] = make(socketSubscriptions)
	}
	selectors, _ := labels.Parse("app=test")
	pub.manager.rooms[nsID][projectID][pub.id] = []labels.Selector{selectors}
	if _, ok := pub.manager.idRooms[pub.id]; !ok {
		pub.manager.idRooms[pub.id] = make(map[int32]struct{})
	}
	pub.manager.idRooms[pub.id][nsID] = struct{}{}
	pub.manager.roomMu.Unlock()

	// Verify structure
	pub.manager.roomMu.RLock()
	_, nsOk := pub.manager.rooms[nsID]
	_, pidOk := pub.manager.rooms[nsID][projectID]
	_, sidOk := pub.manager.rooms[nsID][projectID][pub.id]
	_, roomOk := pub.manager.idRooms[pub.id]
	pub.manager.roomMu.RUnlock()

	assert.True(t, nsOk, "nsID must exist")
	assert.True(t, pidOk, "projectID must exist")
	assert.True(t, sidOk, "socketID must exist")
	assert.True(t, roomOk, "idRooms entry must exist")

	// Leave via the actual Leave method
	pub.Leave(int64(nsID), int64(projectID))

	// Verify cleanup
	pub.manager.roomMu.RLock()
	_, nsOk2 := pub.manager.rooms[nsID]
	_, roomOk2 := pub.manager.idRooms[pub.id]
	pub.manager.roomMu.RUnlock()

	assert.False(t, nsOk2, "nsID should be cleaned up")
	assert.False(t, roomOk2, "idRooms should be cleaned up")
}

func TestPublish_empty_room_is_noop(t *testing.T) {
	ms := newTestSender()
	pub := ms.New("u", "id").(*memoryPubSub)
	// No rooms joined → Publish should be a no-op
	err := pub.Publish(1, &corev1.Pod{})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// 9. Registration
// ---------------------------------------------------------------------------

func TestInitRegistration(t *testing.T) {
	ms := &memorySender{}
	assert.Equal(t, "ws_sender_memory", ms.Name())
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkToAll(b *testing.B) {
	ms := newTestSender()
	pub := ms.New("pub", "pub").(*memoryPubSub)
	_ = pub.Subscribe()

	// Add slow subscribers that never read (worst-case for blocking sends)
	for i := range 10 {
		ms.New(fmt.Sprintf("slow%d", i), fmt.Sprintf("slow%d", i))
	}

	msg := testMsg()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = pub.ToAll(msg)
	}
}

func BenchmarkToAll_with_drainers(b *testing.B) {
	ms := newTestSender()
	pub := ms.New("pub", "pub").(*memoryPubSub)
	_ = pub.Subscribe()

	for i := range 10 {
		p := ms.New(fmt.Sprintf("u%d", i), fmt.Sprintf("id%d", i)).(*memoryPubSub)
		// Drain as fast as possible
		ch := p.Subscribe()
		go func() {
			for range ch {
			}
		}()
	}

	msg := testMsg()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = pub.ToAll(msg)
	}
}

func BenchmarkPublish(b *testing.B) {
	ms := newTestSender()
	pub := ms.New("pub", "pub").(*memoryPubSub)

	// Join rooms via direct struct manipulation (no DB)
	nsID := int32(1)
	pid := int32(100)

	pub.manager.roomMu.Lock()
	pub.manager.rooms[nsID] = make(projectSubscriptions)
	pub.manager.rooms[nsID][pid] = make(socketSubscriptions)
	sel, _ := labels.Parse("app=test")
	pub.manager.rooms[nsID][pid]["pub"] = []labels.Selector{sel}
	pub.manager.idRooms["pub"] = map[int32]struct{}{nsID: {}}
	pub.manager.roomMu.Unlock()

	// Drain
	ch := pub.Subscribe()
	go func() { for range ch { } }()

	pod := &corev1.Pod{}
	pod.Labels = map[string]string{"app": "test"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = pub.Publish(int64(nsID), pod)
	}
}
