package counter

import (
	"context"
	"fmt"
	"sync"
)

type Counter interface {
	Inc()
	Dec()
	Wait(ctx context.Context) error
	Count() int
}

func NewCounter() Counter {
	mu := &sync.Mutex{}
	return &counter{
		count: 0,
		cond:  sync.Cond{L: mu},
		mu:    mu,
	}
}

type counter struct {
	count int
	cond  sync.Cond
	mu    *sync.Mutex
}

func (w *counter) Inc() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.count++
	w.cond.Broadcast()
}

func (w *counter) Wait(ctx context.Context) error {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	for w.count != 0 {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done, remaining count: %d", w.count)
		default:
		}
		w.cond.Wait()
	}
	return nil
}

func (w *counter) Dec() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.count--
	w.cond.Broadcast()
}

func (w *counter) Count() int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.count
}
