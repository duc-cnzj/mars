package utils

import (
	"context"
	"sync"
	"time"
)

type CustomErrorContext struct {
	sync.Mutex
	done     chan struct{}
	err      error
	canceled bool
}

func NewCustomErrorContext() (context.Context, func(error)) {
	ctx := &CustomErrorContext{done: make(chan struct{})}
	return ctx, func(err error) {
		ctx.Lock()
		defer ctx.Unlock()
		if ctx.canceled {
			return
		}
		ctx.err = err
		ctx.canceled = true
		close(ctx.done)
	}
}

func (m *CustomErrorContext) Deadline() (deadline time.Time, ok bool) {
	return
}

func (m *CustomErrorContext) Done() <-chan struct{} {
	return m.done
}

func (m *CustomErrorContext) Err() error {
	return m.err
}

func (m *CustomErrorContext) Value(key any) any {
	return nil
}
