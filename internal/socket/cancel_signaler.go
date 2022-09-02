package socket

import (
	"errors"
	"sync"
)

var ErrCancel = errors.New("取消本次部署，自动回滚到上一个版本！")

type CancelSignals struct {
	cs map[string]func(error)
	sync.RWMutex
}

func (cs *CancelSignals) Remove(id string) {
	cs.Lock()
	defer cs.Unlock()
	delete(cs.cs, id)
}

func (cs *CancelSignals) Has(id string) bool {
	cs.RLock()
	defer cs.RUnlock()

	_, ok := cs.cs[id]

	return ok
}

func (cs *CancelSignals) Cancel(id string) {
	cs.Lock()
	defer cs.Unlock()
	if fn, ok := cs.cs[id]; ok {
		fn(ErrCancel)
	}
}

func (cs *CancelSignals) Add(id string, fn func(error)) error {
	cs.Lock()
	defer cs.Unlock()
	if _, ok := cs.cs[id]; ok {
		return errors.New("项目已经存在")
	}
	cs.cs[id] = fn
	return nil
}

func (cs *CancelSignals) CancelAll() {
	cs.Lock()
	defer cs.Unlock()
	for _, f := range cs.cs {
		f(ErrCancel)
	}
}
