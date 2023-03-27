package cachelock

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/utils"
	"github.com/duc-cnzj/mars/v4/internal/utils/recovery"
)

var defaultStore = NewMemStore()

type MemStore struct {
	m map[string]*MemItem
	sync.RWMutex
}

func (s *MemStore) Add(key string, i *MemItem) {
	s.Lock()
	defer s.Unlock()
	s.m[key] = i
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

func (s *MemStore) Update(key string, m *MemItem) {
	s.Lock()
	defer s.Unlock()
	s.m[key] = m
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
	timer   timer
	locks   *MemStore
}

func NewMemoryLock(lottery [2]int, s *MemStore) contracts.Locker {
	if s == nil {
		s = defaultStore
	}
	return &memoryLock{lottery: lottery, owner: utils.RandomString(40), timer: &realTimers{}, locks: s}
}

func (m *memoryLock) ID() string {
	return m.owner
}

func (m *memoryLock) Type() string {
	return "memory"
}

func (m *memoryLock) Acquire(key string, seconds int64) bool {
	var (
		acquired bool
		item     *MemItem

		unix       = m.timer.Unix()
		expiration = unix + seconds
	)

	m.Lock()
	defer m.Unlock()

	if item = m.locks.Get(key); item == nil {
		m.locks.Add(key, &MemItem{
			owner:     m.owner,
			expiresAt: expiration,
		})
		acquired = true
	}
	if !acquired {
		if item.expiresAt <= unix {
			m.locks.Update(key, &MemItem{
				owner:     m.owner,
				expiresAt: expiration,
			})
			acquired = true
		}
	}

	if rand.Intn(m.lottery[1]) < m.lottery[0] {
		m.locks.Range(func(k string, l *MemItem, m map[string]*MemItem) {
			if l.expiresAt < unix-60 {
				delete(m, k)
			}
		})
	}

	return acquired
}

func (m *memoryLock) Count() int {
	return m.locks.Count()
}

func (m *memoryLock) RenewalAcquire(key string, seconds int64, renewalSeconds int64) (releaseFn func(), acquired bool) {
	if m.Acquire(key, seconds) {
		ctx, cancelFunc := context.WithCancel(context.TODO())
		go func() {
			defer recovery.HandlePanic("[lock]: key: " + key)

			ticker := time.NewTicker(time.Second * time.Duration(renewalSeconds))
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					mlog.Debug("[lock]: canceled: " + key)
					return
				case <-ticker.C:
					if !m.renewalExistKey(key, seconds) {
						mlog.Warning("[lock]: err renewal lock: " + key)
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

	var (
		acquired bool
		item     *MemItem

		unix       = m.timer.Unix()
		expiration = unix + seconds
	)

	if item = m.locks.Get(key); item == nil {
		return false
	}

	if item.owner == m.owner {
		m.locks.Update(key, &MemItem{
			owner:     m.owner,
			expiresAt: expiration,
		})
		acquired = true
	}

	return acquired
}
