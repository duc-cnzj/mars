package event

import (
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/stretchr/testify/assert"
)

func TestDispatcher_Dispatch(t *testing.T) {
	d := NewDispatcher(nil)
	called := false
	d.Listen("evt", func(a any, event contracts.Event) error {
		called = true
		assert.Equal(t, "xxx", a)
		return nil
	})
	d.Dispatch("evt", "xxx")
	assert.True(t, called)
}

func TestDispatcher_Forget(t *testing.T) {
	d := NewDispatcher(nil)
	d.Listen("evt", func(a any, event contracts.Event) error {
		return nil
	})
	d.Forget("evt")
	listeners := d.GetListeners("evt")
	assert.Len(t, listeners, 0)
}

func TestDispatcher_GetListeners(t *testing.T) {
	d := NewDispatcher(nil)
	d.Listen("evt", func(a any, event contracts.Event) error {
		return nil
	})
	listeners := d.GetListeners("evt")
	assert.Len(t, listeners, 1)
}

func TestDispatcher_HasListeners(t *testing.T) {
	d := NewDispatcher(nil)
	d.Listen("evt", func(a any, event contracts.Event) error {
		return nil
	})
	assert.True(t, d.HasListeners("evt"))
	assert.False(t, d.HasListeners("xxx"))
}

func TestDispatcher_Listen(t *testing.T) {
	d := NewDispatcher(nil)
	called := 0
	d.Listen("evt", func(a any, event contracts.Event) error {
		called++
		return nil
	})
	assert.NotNil(t, d.listeners["evt"])
	d.Listen("evt", func(a any, event contracts.Event) error {
		called++
		return errors.New("err called")
	})
	err := d.Dispatch("evt", "")
	assert.Equal(t, 2, called)
	assert.Equal(t, "err called", err.Error())
}

func TestNewDispatcher(t *testing.T) {
	assert.Implements(t, (*contracts.DispatcherInterface)(nil), NewDispatcher(nil))
}
