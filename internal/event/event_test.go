package event

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/stretchr/testify/assert"
)

func TestDispatcher_ListenAndHasListeners(t *testing.T) {
	logger := mlog.NewEmptyLogger()
	dispatcher := NewDispatcher(logger)

	eventName := Event("testEvent")
	listener := func(any any, e Event) error {
		return nil
	}

	dispatcher.Listen(eventName, listener)
	assert.True(t, dispatcher.HasListeners(eventName))
}

func TestDispatcher_Forget(t *testing.T) {
	logger := mlog.NewEmptyLogger()
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
	logger := mlog.NewEmptyLogger()
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
	logger := mlog.NewEmptyLogger()
	dispatcher := NewDispatcher(logger)

	eventName := Event("testEvent")
	called := 0
	listener := func(any any, e Event) error {
		called++
		return errors.New("error")
	}

	dispatcher.Listen(eventName, listener)
	go dispatcher.Run(context.Background())
	dispatcher.Dispatch(eventName, nil)
	time.Sleep(1 * time.Second)
	dispatcher.Shutdown(context.Background())

	assert.Equal(t, 1, called)
}
