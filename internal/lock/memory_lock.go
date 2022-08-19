package lock

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/internal/utils/recovery"
)

var defaultStore = newStore()

type memStore struct {
	m map[string]*memItem
	sync.RWMutex
}

func (s *memStore) Add(key string, i *memItem) {
	s.Lock()
	defer s.Unlock()
	s.m[key] = i
}

func (s *memStore) Delete(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, key)
}

func (s *memStore) Get(key string) *memItem {
	s.RLock()
	defer s.RUnlock()
	return s.m[key]
}

func (s *memStore) Update(key string, m *memItem) {
	s.Lock()
	defer s.Unlock()
	s.m[key] = m
}

func (s *memStore) Range(fn func(string, *memItem, map[string]*memItem)) {
	s.Lock()
	defer s.Unlock()
	for k, item := range s.m {
		fn(k, item, s.m)
	}
}

func (s *memStore) Count() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.m)
}

func newStore() *memStore {
	return &memStore{
		m: make(map[string]*memItem),
	}
}

type memItem struct {
	owner     string
	expiresAt int64
}

type memoryLock struct {
	sync.Mutex

	owner   string
	lottery [2]int
	timer   timer
	locks   *memStore
}

func NewMemoryLock(lottery [2]int, s *memStore) contracts.Locker {
	if s == nil {
		s = defaultStore
	}
	return &memoryLock{lottery: lottery, owner: utils.RandomString(40), timer: &realTimers{}, locks: s}
}

func (m *memoryLock) Acquire(key string, seconds int64) bool {
	var (
		acquired bool
		item     *memItem

		unix       = m.timer.Unix()
		expiration = unix + seconds
	)

	m.Lock()
	defer m.Unlock()

	if item = m.locks.Get(key); item == nil {
		m.locks.Add(key, &memItem{
			owner:     m.owner,
			expiresAt: expiration,
		})
		acquired = true
	}
	if !acquired {
		if item.expiresAt <= unix {
			m.locks.Update(key, &memItem{
				owner:     m.owner,
				expiresAt: expiration,
			})
			acquired = true
		}
	}

	if rand.Intn(m.lottery[1]) < m.lottery[0] {
		m.locks.Range(func(k string, l *memItem, m map[string]*memItem) {
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

func (m *memoryLock) RenewalAcquire(key string, seconds int64, renewalSeconds int) (releaseFn func(), acquired bool) {
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
		item     *memItem

		unix       = m.timer.Unix()
		expiration = unix + seconds
	)

	if item = m.locks.Get(key); item == nil {
		return false
	}

	if item.owner == m.owner {
		m.locks.Update(key, &memItem{
			owner:     m.owner,
			expiresAt: expiration,
		})
		acquired = true
	}

	return acquired
}
