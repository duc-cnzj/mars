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
