package events

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	registry = make(map[contracts.Event][]contracts.Listener)
	Register("e1", func(a any, event contracts.Event) error {
		return nil
	})
	Register("e1", func(a any, event contracts.Event) error {
		return nil
	})
	assert.Len(t, registry["e1"], 2)
}

func TestRegisteredEvents(t *testing.T) {
	registry = make(map[contracts.Event][]contracts.Listener)
	Register("e1", func(a any, event contracts.Event) error {
		return nil
	})
	Register("e1", func(a any, event contracts.Event) error {
		return nil
	})
	assert.Len(t, registry["e1"], 2)
	assert.Len(t, RegisteredEvents(), 1)
	assert.Len(t, RegisteredEvents()["e1"], 2)
}
