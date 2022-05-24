package socket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCancelSignals_Add(t *testing.T) {
	cs := &CancelSignals{
		cs: map[string]func(error){},
	}
	var called bool
	cs.Add("a", func(err error) {
		called = true
	})
	cs.Cancel("a")
	assert.True(t, called)
	err := cs.Add("a", func(err error) {
		called = true
	})
	assert.Error(t, err)
}

func TestCancelSignals_Cancel(t *testing.T) {
	cs := &CancelSignals{
		cs: map[string]func(error){},
	}
	var called bool
	cs.Add("a", func(err error) {
		called = true
	})
	cs.Cancel("a")
	assert.True(t, called)
}

func TestCancelSignals_CancelAll(t *testing.T) {
	cs := &CancelSignals{
		cs: map[string]func(error){},
	}
	var acalled bool
	var bcalled bool
	cs.Add("a", func(err error) {
		acalled = true
	})
	cs.Add("b", func(err error) {
		bcalled = true
	})
	cs.CancelAll()
	assert.True(t, acalled)
	assert.True(t, bcalled)
}

func TestCancelSignals_Has(t *testing.T) {
	cs := &CancelSignals{
		cs: map[string]func(error){},
	}
	cs.Add("a", func(err error) {
	})
	assert.True(t, cs.Has("a"))
	assert.False(t, cs.Has("b"))
}

func TestCancelSignals_Remove(t *testing.T) {
	cs := &CancelSignals{
		cs: map[string]func(error){},
	}
	cs.Add("a", func(err error) {
	})
	assert.True(t, cs.Has("a"))
	cs.Remove("a")
	assert.False(t, cs.Has("a"))
}
