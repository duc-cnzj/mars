package event

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
)

func TestEvent_String(t *testing.T) {
	event := Event("testEvent")
	assert.Equal(t, "testEvent", event.String())
}

func TestDispatcher_Listen(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	dispatcher := NewDispatcher(logger)

	eventName := Event("testEvent")
	dispatcher.Listen(eventName, func(any, Event) error { return nil })
	dispatcher.Listen(eventName, func(any, Event) error { return nil })

	assert.Equal(t, 2, len(dispatcher.List()[eventName]))
}

func TestDispatcher_GetListeners(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	dispatcher := NewDispatcher(logger)

	eventName := Event("testEvent")
	dispatcher.Listen(eventName, func(any, Event) error { return nil })

	assert.Equal(t, 1, len(dispatcher.GetListeners(eventName)))
}

// GetListeners must return a copy: appending to it must not grow the
// dispatcher's internal registration.
func TestDispatcher_GetListeners_ReturnsCopy(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	dispatcher := NewDispatcher(logger)

	eventName := Event("testEvent")
	dispatcher.Listen(eventName, func(any, Event) error { return nil })

	got := dispatcher.GetListeners(eventName)
	_ = append(got, func(any, Event) error { return nil })
	assert.Equal(t, 1, len(dispatcher.GetListeners(eventName)))
}

// List must return a copy: mutating the returned map or its slices must not
// affect the dispatcher's internal registration.
func TestDispatcher_List_ReturnsCopy(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	dispatcher := NewDispatcher(logger)

	eventName := Event("testEvent")
	dispatcher.Listen(eventName, func(any, Event) error { return nil })

	got := dispatcher.List()
	got[Event("injected")] = []Listener{func(any, Event) error { return nil }}
	got[eventName] = append(got[eventName], func(any, Event) error { return nil })

	assert.Equal(t, 1, len(dispatcher.List()))
	assert.Equal(t, 1, len(dispatcher.GetListeners(eventName)))
}

// The full loop: an event is dispatched, its listener receives the exact
// payload and event (even when the listener returns an error) and Shutdown
// stops the processing loop.
func TestDispatcher_RunAndDispatch(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	dispatcher := NewDispatcher(logger)

	eventName := Event("testEvent")
	var gotPayload any
	var gotEvent Event
	received := make(chan struct{})
	dispatcher.Listen(eventName, func(payload any, event Event) error {
		gotPayload = payload
		gotEvent = event
		close(received)
		return errors.New("listener failed")
	})

	dispatcher.Run(context.TODO())
	dispatcher.Dispatch(eventName, "payload")

	select {
	case <-received:
	case <-time.After(2 * time.Second):
		t.Fatal("listener was not called")
	}
	assert.Equal(t, eventName, gotEvent)
	assert.Equal(t, "payload", gotPayload)

	dispatcher.Shutdown(context.TODO())
}

// Dispatching an event with no registered listener must be a no-op. The Run
// loop dequeues events in FIFO order, so by the time the handled listener
// fires, the unhandled event has been dequeued and its (empty) listener set
// consulted.
func TestDispatcher_Dispatch_NoListeners(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	dispatcher := NewDispatcher(logger)

	done := make(chan struct{})
	dispatcher.Listen(Event("handled"), func(any, Event) error {
		close(done)
		return nil
	})

	dispatcher.Run(context.TODO())
	dispatcher.Dispatch(Event("unhandled"), "payload")
	dispatcher.Dispatch(Event("handled"), "payload")

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handled listener was not called")
	}

	dispatcher.Shutdown(context.TODO())
}

// Run must stop when the caller's context is cancelled.
func TestDispatcher_Run_ContextDone(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	logger := mlog.NewMockLogger(m)

	ctx, cancel := context.WithCancel(context.TODO())
	cancel()
	dispatcher := &dispatcher{logger: logger, ctx: context.TODO(), ch: make(chan *eventBody)}

	exited := make(chan struct{})
	logger.EXPECT().Info("[Event]: dispatcher running")
	logger.EXPECT().Warning("event dispatcher context done").DoAndReturn(func(...any) {
		close(exited)
	})
	dispatcher.Run(ctx)
	<-exited
}

// Run must stop when the event channel is closed.
func TestDispatcher_Run_ChannelClosed(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	logger := mlog.NewMockLogger(m)

	ch := make(chan *eventBody)
	close(ch)
	dispatcher := &dispatcher{logger: logger, ctx: context.TODO(), ch: ch}

	exited := make(chan struct{})
	logger.EXPECT().Info("[Event]: dispatcher running")
	logger.EXPECT().Warning("event dispatcher channel closed").DoAndReturn(func(...any) {
		close(exited)
	})
	dispatcher.Run(context.TODO())
	<-exited
}

// Dispatch must never block: when the buffer is full, the event is dropped
// and a warning is logged.
func TestDispatcher_Dispatch_ChannelFull(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	logger := mlog.NewMockLogger(m)

	dispatcher := &dispatcher{
		logger: logger,
		ch:     make(chan *eventBody, 1),
	}

	// First call fills the buffer (no consumer drains it).
	dispatcher.Dispatch(Event("first"), "payload1")
	// Second call hits the non-blocking drop branch.
	logger.EXPECT().Warningf(gomock.Any(), gomock.Any(), gomock.Any())
	dispatcher.Dispatch(Event("second"), "payload2")
}

// TestDispatcher_Run_SemFull_CallCtxDone 回归：信号量满（全部 handler 被慢监听器拖死）时，
// Run 循环阻塞在 sem 发送也必须能响应调用方 ctx 取消退出，否则 Shutdown 后该 goroutine 泄漏。
// 覆盖 Run 内层 select 的 <-ctx.Done() 分支。
func TestDispatcher_Run_SemFull_CallCtxDone(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	callCtx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	block := make(chan struct{})
	defer close(block)
	held := make(chan struct{})
	d := &dispatcher{
		ctx:       context.TODO(),
		ch:        make(chan *eventBody, 2),
		sem:       make(chan struct{}, 1),
		logger:    logger,
		listeners: map[Event][]Listener{},
	}
	// 监听器取得唯一 sem 槽位后永久阻塞，模拟慢监听器拖死信号量。
	d.Listen(Event("e"), func(any, Event) error {
		close(held)
		<-block
		return nil
	})

	done := make(chan struct{})
	go func() {
		_ = d.Run(callCtx)
		close(done)
	}()

	d.Dispatch(Event("e"), nil)
	<-held                      // 首个 handler 已持有 sem 槽位（sem 已满）
	d.Dispatch(Event("e"), nil) // 第二个事件：Run 出队后阻塞在 sem 发送

	cancel() // 调用方 ctx 取消，须穿过内层 sem select 退出
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run 未在信号量满时响应 ctx 取消，goroutine 泄漏")
	}
}

// TestDispatcher_Run_ListenerPanic 验证监听器 panic 被 HandlePanic 兜底：监听器抛 panic
// 不会击穿 Run 循环，后续分发的其它事件仍能被正常处理。注意：panic 会沿着当前 goroutine
// 的分发 for 循环向上展开，同事件的后续监听器会被跳过，但整个 dispatcher 存活。
func TestDispatcher_Run_ListenerPanic(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	dispatcher := NewDispatcher(logger)

	panicked := make(chan struct{})
	dispatcher.Listen(Event("boom"), func(any, Event) error {
		close(panicked)
		panic("listener exploded")
	})
	// 后续分发的新事件：若 panic 击穿 Run 循环，此事件将永远不会被处理。
	after := make(chan struct{})
	dispatcher.Listen(Event("after"), func(any, Event) error {
		close(after)
		return nil
	})

	dispatcher.Run(context.TODO())
	dispatcher.Dispatch(Event("boom"), nil)

	select {
	case <-panicked:
	case <-time.After(2 * time.Second):
		t.Fatal("panicking listener was not called")
	}
	// panic 被 HandlePanic 恢复后，dispatcher 必须继续处理新事件。
	dispatcher.Dispatch(Event("after"), nil)
	select {
	case <-after:
	case <-time.After(2 * time.Second):
		t.Fatal("event after a panic was not handled; Run loop was killed by panic")
	}

	dispatcher.Shutdown(context.TODO())
}

// TestDispatcher_Run_SemFull_Shutdown 回归：与 CallCtxDone 对称，覆盖 Shutdown（内部 ctx）
// 取消时穿过内层 sem select 退出，即 <-d.ctx.Done() 分支。
func TestDispatcher_Run_SemFull_Shutdown(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	dCtx, dCancel := context.WithCancel(context.TODO())

	block := make(chan struct{})
	defer close(block)
	held := make(chan struct{})
	d := &dispatcher{
		ctx:       dCtx,
		ch:        make(chan *eventBody, 2),
		sem:       make(chan struct{}, 1),
		logger:    logger,
		listeners: map[Event][]Listener{},
	}
	d.Listen(Event("e"), func(any, Event) error {
		close(held)
		<-block
		return nil
	})

	done := make(chan struct{})
	go func() {
		_ = d.Run(context.TODO())
		close(done)
	}()

	d.Dispatch(Event("e"), nil)
	<-held
	d.Dispatch(Event("e"), nil)

	dCancel() // Shutdown 取消内部 ctx
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run 未在 Shutdown 时退出，goroutine 泄漏")
	}
}
