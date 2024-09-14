package timer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRealTimerNow(t *testing.T) {
	realTimer := NewReal()
	now := realTimer.Now()
	assert.WithinDuration(t, time.Now(), now, time.Second, "The time returned by realTimer.Now() should be within a second of the current time")
}

func TestRealTimerType(t *testing.T) {
	assert.Implements(t, (*Timer)(nil), new(realTimer))
}
