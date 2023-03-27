package socket

import (
	"sync"
)

var Wait WaitSocketExitInterface = NewWaitSocketExit()

type WaitSocketExitInterface interface {
	Inc()
	Dec()
	Wait()
	Count() int
}

func NewWaitSocketExit() WaitSocketExitInterface {
	mu := &sync.Mutex{}
	return &waitSocketExit{
		count: 0,
		cond:  sync.Cond{L: mu},
		mu:    mu,
	}
}

type waitSocketExit struct {
	count int
	cond  sync.Cond
	mu    *sync.Mutex
}

func (w *waitSocketExit) Inc() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.count++
	w.cond.Broadcast()
}

func (w *waitSocketExit) Wait() {
	w.cond.L.Lock()
	for w.count != 0 {
		w.cond.Wait()
	}
	w.cond.L.Unlock()
}

func (w *waitSocketExit) Dec() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.count--
	w.cond.Broadcast()
}

func (w *waitSocketExit) Count() int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.count
}
