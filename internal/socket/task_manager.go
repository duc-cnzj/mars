package socket

import (
	"errors"
	"sync"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

var (
	errCancel       = errors.New("取消本次部署，自动回滚到上一个版本！")
	errSignalExists = errors.New("项目已经存在")
)

type TaskManager interface {
	Remove(id string)
	Has(id string) bool
	Stop(id string)
	Register(id string, fn func(error)) error
	StopAll()
}

var _ TaskManager = (*taskManagerImpl)(nil)

type taskManagerImpl struct {
	cs map[string]func(error)
	sync.RWMutex
	logger mlog.Logger
}

func NewTaskManager(logger mlog.Logger) TaskManager {
	return &taskManagerImpl{cs: map[string]func(error){}, logger: logger}
}

func (cs *taskManagerImpl) Remove(id string) {
	cs.Lock()
	defer cs.Unlock()
	delete(cs.cs, id)
}

func (cs *taskManagerImpl) Has(id string) bool {
	cs.RLock()
	defer cs.RUnlock()

	_, ok := cs.cs[id]

	return ok
}

func (cs *taskManagerImpl) Stop(id string) {
	cs.Lock()
	defer cs.Unlock()
	if fn, ok := cs.cs[id]; ok {
		cs.logger.Debugf("stop task\t%v\t%v", id, errCancel)
		fn(errCancel)
	}
}

func (cs *taskManagerImpl) Register(id string, fn func(error)) error {
	cs.Lock()
	defer cs.Unlock()
	if _, ok := cs.cs[id]; ok {
		return errSignalExists
	}
	cs.cs[id] = fn
	return nil
}

func (cs *taskManagerImpl) StopAll() {
	cs.Lock()
	defer cs.Unlock()
	for _, f := range cs.cs {
		f(errCancel)
	}
}
