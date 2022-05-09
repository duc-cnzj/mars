package socket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWaitSocketExit(t *testing.T) {
	assert.IsType(t, (*WaitSocketExit)(nil), NewWaitSocketExit())
}

func TestWaitSocketExit_Count(t *testing.T) {
	exit := NewWaitSocketExit()
	assert.Equal(t, 0, exit.Count())
}

func TestWaitSocketExit_Dec(t *testing.T) {
	exit := NewWaitSocketExit()
	exit.Dec()
	assert.Equal(t, -1, exit.Count())
}

func TestWaitSocketExit_Inc(t *testing.T) {
	exit := NewWaitSocketExit()
	exit.Inc()
	exit.Inc()
	exit.Inc()
	assert.Equal(t, 3, exit.Count())
}

func TestWaitSocketExit_Wait(t *testing.T) {
	exit := NewWaitSocketExit()

	for i := 0; i < 100; i++ {
		go func() {
			exit.Inc()
		}()
	}
	for i := 0; i < 100; i++ {
		go func() {
			exit.Dec()
		}()
	}
	exit.Wait()
	assert.Equal(t, 0, exit.Count())
}
