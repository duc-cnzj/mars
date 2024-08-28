package locker

import (
	"context"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/util/rand"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
)

type MemStore struct {
	m map[string]*MemItem
	sync.RWMutex
}

func (s *MemStore) Add(key string, item *MemItem) {
	s.Lock()
	defer s.Unlock()
	s.m[key] = item
}

func (s *MemStore) Delete(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, key)
}

func (s *MemStore) Get(key string) *MemItem {
	s.RLock()
	defer s.RUnlock()
	return s.m[key]
}

func (s *MemStore) Update(key string, item *MemItem) {
	s.Lock()
	defer s.Unlock()
	s.m[key] = item
}

func (s *MemStore) Range(fn func(string, *MemItem, map[string]*MemItem)) {
	s.Lock()
	defer s.Unlock()
	for k, item := range s.m {
		fn(k, item, s.m)
	}
}

func (s *MemStore) Count() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.m)
}

func NewMemStore() *MemStore {
	return &MemStore{
		m: make(map[string]*MemItem),
	}
}

type MemItem struct {
	owner     string
	expiresAt int64
}

type memoryLock struct {
	sync.Mutex

	owner   string
	lottery [2]int
	timer   timer.Timer
	locks   *MemStore
	logger  mlog.Logger
}

func NewMemoryLock(timer timer.Timer, lottery [2]int, store *MemStore, logger mlog.Logger) Locker {
	return &memoryLock{
		owner:   rand.String(40),
		lottery: lottery,
		timer:   timer,
		locks:   store,
		logger:  logger,
	}
}

func (m *memoryLock) ID() string {
	return m.owner
}

func (m *memoryLock) Type() string {
	return "memory"
}

func (m *memoryLock) acquireInternal(key string, seconds int64) (bool, *MemItem) {
	unix := m.timer.Now().Unix()
	expiration := unix + seconds

	item := m.locks.Get(key)
	if item == nil {
		m.locks.Add(key, &MemItem{owner: m.owner, expiresAt: expiration})
		return true, nil
	}
	if item.expiresAt <= unix {
		m.locks.Update(key, &MemItem{owner: m.owner, expiresAt: expiration})
		return true, item
	}
	return false, item
}

func (m *memoryLock) Acquire(key string, seconds int64) bool {
	m.Lock()
	defer m.Unlock()

	acquired, _ := m.acquireInternal(key, seconds)

	if rand.Intn(m.lottery[1]) < m.lottery[0] {
		m.locks.Range(func(k string, item *MemItem, store map[string]*MemItem) {
			if item.expiresAt < m.timer.Now().Unix()-60 {
				delete(store, k)
			}
		})
	}

	return acquired
}

func (m *memoryLock) Count() int {
	return m.locks.Count()
}

func (m *memoryLock) RenewalAcquire(key string, seconds int64, renewalSeconds int64) (func(), bool) {
	if m.Acquire(key, seconds) {
		ctx, cancelFunc := context.WithCancel(context.TODO())
		go func() {
			defer m.logger.HandlePanic("[lock]: key: " + key)

			ticker := time.NewTicker(time.Second * time.Duration(renewalSeconds))
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					m.logger.Debug("[lock]: canceled: " + key)
					return
				case <-ticker.C:
					if !m.renewalExistKey(key, seconds) {
						m.logger.Warning("[lock]: err renewal lock: " + key)
						return
					}
				}
			}
		}()
		return func() {
			cancelFunc()
			m.Release(key)
		}, true
	}
	return nil, false
}

func (m *memoryLock) Release(key string) bool {
	m.Lock()
	defer m.Unlock()
	if item := m.locks.Get(key); item != nil && item.owner == m.owner {
		m.locks.Delete(key)
		return true
	}
	return false
}

func (m *memoryLock) ForceRelease(key string) bool {
	m.locks.Delete(key)
	return true
}

func (m *memoryLock) Owner(key string) string {
	if item := m.locks.Get(key); item != nil {
		return item.owner
	}
	return ""
}

func (m *memoryLock) renewalExistKey(key string, seconds int64) bool {
	m.Lock()
	defer m.Unlock()

	unix := m.timer.Now().Unix()
	expiration := unix + seconds

	item := m.locks.Get(key)
	if item == nil || item.owner != m.owner {
		return false
	}

	m.locks.Update(key, &MemItem{owner: m.owner, expiresAt: expiration})
	return true
}
