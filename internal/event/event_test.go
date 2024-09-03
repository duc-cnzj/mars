package event

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

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
