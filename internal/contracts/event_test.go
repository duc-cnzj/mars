package contracts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvent_Is(t *testing.T) {
	e := Event("a")
	assert.True(t, e.Is(Event("a")))
	assert.False(t, e.Is(Event("b")))
}

func TestEvent_String(t *testing.T) {
	e := Event("a")
	assert.Equal(t, "a", e.String())
}
