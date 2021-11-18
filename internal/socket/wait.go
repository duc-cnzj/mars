package socket

import (
	"sync"
)

var Wait = NewWaitSocketExit()

type WaitSocketExit struct {
	count int
	cond  sync.Cond
	mu    *sync.Mutex
}

func (w *WaitSocketExit) Inc() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.count++
	w.cond.Broadcast()
}

func (w *WaitSocketExit) Wait() {
	w.cond.L.Lock()
	for w.count != 0 {
		w.cond.Wait()
	}
	w.cond.L.Unlock()
}

func (w *WaitSocketExit) Dec() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.count--
	w.cond.Broadcast()
}

func (w *WaitSocketExit) Count() int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.count
}

func NewWaitSocketExit() *WaitSocketExit {
	mu := &sync.Mutex{}
	return &WaitSocketExit{
		count: 0,
		cond:  sync.Cond{L: mu},
		mu:    mu,
	}
}
