package event

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
)

func TestDispatcher_ListenAndHasListeners(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	dispatcher := NewDispatcher(logger)

	eventName := Event("testEvent")
	listener := func(any any, e Event) error {
		return nil
	}

	dispatcher.Listen(eventName, listener)
	dispatcher.Listen(eventName, func(a any, event Event) error {
		return nil
	})
	assert.True(t, dispatcher.HasListeners(eventName))
	assert.Equal(t, 2, len(dispatcher.List()[eventName]))
}

func TestDispatcher_Forget(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	dispatcher := NewDispatcher(logger)

	eventName := Event("testEvent")
	listener := func(any any, e Event) error {
		return nil
	}

	dispatcher.Listen(eventName, listener)
	dispatcher.Forget(eventName)

	assert.False(t, dispatcher.HasListeners(eventName))
}

func TestDispatcher_GetListeners(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	dispatcher := NewDispatcher(logger)

	eventName := Event("testEvent")
	listener := func(any any, e Event) error {
		return nil
	}

	dispatcher.Listen(eventName, listener)
	listeners := dispatcher.GetListeners(eventName)

	assert.Equal(t, 1, len(listeners))
}

func TestDispatcher_RunAndShutdown(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	dispatcher := NewDispatcher(logger)

	eventName := Event("testEvent")
	mu := sync.RWMutex{}
	called := 0
	listener := func(any any, e Event) error {
		mu.Lock()
		defer mu.Unlock()
		called++
		return errors.New("error")
	}

	dispatcher.Listen(eventName, listener)
	go dispatcher.Run(context.Background())
	dispatcher.Dispatch(eventName, nil)
	time.Sleep(1 * time.Second)
	dispatcher.Shutdown(context.Background())

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 1, called)
}

func TestDispatcher_Run_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	logger := mlog.NewMockLogger(m)
	dispatcher := &dispatcher{logger: logger, ctx: context.Background(), ch: make(chan *eventBody)}

	ctx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()
	logger.EXPECT().Info("[Event]: dispatcher running")
	logger.EXPECT().Warning("event dispatcher context done")
	dispatcher.Run(ctx)
	time.Sleep(1 * time.Second)
}

func TestDispatcher_Run_Error2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	logger := mlog.NewMockLogger(m)
	ch := make(chan *eventBody)
	close(ch)
	dispatcher := &dispatcher{logger: logger, ctx: context.Background(), ch: ch}

	logger.EXPECT().Info("[Event]: dispatcher running")
	logger.EXPECT().Warning("event dispatcher channel closed")
	dispatcher.Run(context.Background())
	time.Sleep(1 * time.Second)
}

func TestEvent_String(t *testing.T) {
	event := Event("testEvent")
	assert.Equal(t, "testEvent", event.String())
}

func TestEvent_Is(t *testing.T) {
	event := Event("testEvent")
	assert.True(t, event.Is("testEvent"))
	assert.False(t, event.Is("testEvent2"))
}

func Test_dispatcher_List(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	dispatcher := NewDispatcher(logger)

	eventName := Event("testEvent")
	listener := func(any any, e Event) error {
		return nil
	}

	dispatcher.Listen(eventName, listener)
	listeners := dispatcher.List()

	assert.Equal(t, 1, len(listeners))
}
