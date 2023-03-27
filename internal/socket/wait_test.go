package socket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewWaitSocketExit(t *testing.T) {
	assert.IsType(t, (*waitSocketExit)(nil), NewWaitSocketExit())
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
	exit.Inc()
	done := make(chan struct{}, 1)
	go func() {
		exit.Wait()
		done <- struct{}{}
	}()
	time.Sleep(200 * time.Millisecond)
	exit.Dec()
	<-done
	assert.Equal(t, 0, exit.Count())
}
