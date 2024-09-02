package counter

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCounterIncrements(t *testing.T) {
	counter := NewCounter()
	counter.Inc()
	assert.Equal(t, 1, counter.Count())
}

func TestCounterDecrements(t *testing.T) {
	counter := NewCounter()
	counter.Inc()
	counter.Dec()
	assert.Equal(t, 0, counter.Count())
}

func TestCounterDoesNotDecrementBelowZero(t *testing.T) {
	counter := NewCounter()
	decremented := counter.Dec()
	assert.False(t, decremented)
	assert.Equal(t, 0, counter.Count())
}

func TestCounterWaitReturnsWhenCountIsZero(t *testing.T) {
	counter := NewCounter()
	err := counter.Wait(context.Background())
	assert.Nil(t, err)
}

func TestCounterWaitReturnsOnError(t *testing.T) {
	counter := NewCounter()
	counter.Inc()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := counter.Wait(ctx)
	assert.Error(t, err)
}
