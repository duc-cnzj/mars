package counter

import (
	"context"
	"sync"
)

type Counter interface {
	Inc()
	Dec() bool
	Wait(ctx context.Context) error
	Count() int
}

func NewCounter() Counter {
	mu := &sync.Mutex{}
	return &counter{
		count: 0,
		cond:  sync.NewCond(mu),
		mu:    mu,
	}
}

type counter struct {
	count int
	cond  *sync.Cond
	mu    *sync.Mutex
}

func (w *counter) Inc() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.count++
	w.cond.Broadcast()
}

func (w *counter) Wait(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		for w.Dec() {
		}
	}()
	w.mu.Lock()
	defer w.mu.Unlock()
	for w.count != 0 {
		w.cond.Wait()
	}
	return nil
}

func (w *counter) Dec() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.count > 0 {
		w.count--
		w.cond.Broadcast()
		return true
	}
	return false
}

func (w *counter) Count() int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.count
}
