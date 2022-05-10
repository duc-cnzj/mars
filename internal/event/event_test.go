package event

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/contracts"
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
	d.Listen("evt", func(a any, event contracts.Event) error {
		return nil
	})
	assert.NotNil(t, d.listeners["evt"])
}

func TestNewDispatcher(t *testing.T) {
	assert.Implements(t, (*contracts.DispatcherInterface)(nil), NewDispatcher(nil))
}
