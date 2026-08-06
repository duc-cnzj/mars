package counter

import (
	"context"
	"sync"
)

// Counter 是一个可等待归零的并发计数器，常用于等待一组 goroutine 收尾。
// 使用 NewCounter 构造（mu 需初始化），零值不可直接使用。
type Counter struct {
	count int
	cond  *sync.Cond
	mu    *sync.Mutex
}

// NewCounter 构造一个计数从 0 开始的 Counter。
func NewCounter() *Counter {
	mu := &sync.Mutex{}
	return &Counter{
		count: 0,
		cond:  sync.NewCond(mu),
		mu:    mu,
	}
}

// Inc 计数加一并唤醒所有 Wait 等待者。
func (w *Counter) Inc() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.count++
	w.cond.Broadcast()
}

// Wait 阻塞直到计数归零；ctx 取消时递减计数并返回 ctx.Err()。
// 计数已为零时提前返回，避免为"无待回收工作"的场景白白挂起一个 ctx 监听 goroutine。
func (w *Counter) Wait(ctx context.Context) error {
	w.mu.Lock()
	if w.count == 0 {
		w.mu.Unlock()
		return nil
	}
	w.mu.Unlock()

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
	return ctx.Err()
}

// Dec 计数减一，仅在计数 > 0 时生效，返回是否实际发生了递减。
func (w *Counter) Dec() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.count > 0 {
		w.count--
		w.cond.Broadcast()
		return true
	}
	return false
}

// Count 返回当前计数。
func (w *Counter) Count() int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.count
}
