package socket

import (
	"errors"
	"sync"
)

var (
	errCancel       = errors.New("取消本次部署，自动回滚到上一个版本！")
	errSignalExists = errors.New("项目已经存在")
)

type CancelSignaler interface {
	Remove(id string)
	Has(id string) bool
	Cancel(id string)
	Add(id string, fn func(error)) error
	CancelAll()
}

var _ CancelSignaler = (*cancelSignal)(nil)

type cancelSignal struct {
	cs map[string]func(error)
	sync.RWMutex
}

func NewCancelSignal() CancelSignaler {
	return &cancelSignal{cs: map[string]func(error){}}
}

func (cs *cancelSignal) Remove(id string) {
	cs.Lock()
	defer cs.Unlock()
	delete(cs.cs, id)
}

func (cs *cancelSignal) Has(id string) bool {
	cs.RLock()
	defer cs.RUnlock()

	_, ok := cs.cs[id]

	return ok
}

func (cs *cancelSignal) Cancel(id string) {
	cs.Lock()
	defer cs.Unlock()
	if fn, ok := cs.cs[id]; ok {
		fn(errCancel)
	}
}

func (cs *cancelSignal) Add(id string, fn func(error)) error {
	cs.Lock()
	defer cs.Unlock()
	if _, ok := cs.cs[id]; ok {
		return errSignalExists
	}
	cs.cs[id] = fn
	return nil
}

func (cs *cancelSignal) CancelAll() {
	cs.Lock()
	defer cs.Unlock()
	for _, f := range cs.cs {
		f(errCancel)
	}
}
