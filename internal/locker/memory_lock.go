package locker

import (
	"context"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/rand"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
)

// staleWindow 僵尸锁判定窗口（秒）：锁项过期超过该时长仍未续期时，被视为僵尸锁回收。
const staleWindow int64 = 60

// MemStore 是内存锁的后端存储，用 map 存放锁项，并通过 RWMutex 保证并发安全。
type MemStore struct {
	m map[string]*MemItem
	sync.RWMutex
}

// Add 写入一个锁项。
func (s *MemStore) Add(key string, item *MemItem) {
	s.Lock()
	defer s.Unlock()
	s.m[key] = item
}

// Delete 删除一个锁项。
func (s *MemStore) Delete(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, key)
}

// Get 读取一个锁项，不存在时返回 nil。
func (s *MemStore) Get(key string) *MemItem {
	s.RLock()
	defer s.RUnlock()
	return s.m[key]
}

// Update 覆盖写入一个锁项。
func (s *MemStore) Update(key string, item *MemItem) {
	s.Lock()
	defer s.Unlock()
	s.m[key] = item
}

// CleanupExpired 删除所有过期时间早于 now-staleWindow 的锁项，用于回收长期未续期的僵尸锁。
func (s *MemStore) CleanupExpired(now, staleWindow int64) {
	s.Lock()
	defer s.Unlock()
	for k, item := range s.m {
		if item.expiresAt < now-staleWindow {
			delete(s.m, k)
		}
	}
}

// NewMemStore 创建一个空的内存锁存储。
func NewMemStore() *MemStore {
	return &MemStore{
		m: make(map[string]*MemItem),
	}
}

// MemItem 表示一个内存锁项：持有者 owner 与过期时间点 expiresAt（Unix 秒）。
type MemItem struct {
	owner     string
	expiresAt int64
}

// memoryLock 是基于进程内存的 Locker 实现。
type memoryLock struct {
	sync.Mutex

	owner   string
	lottery [2]int
	timer   timer.Timer
	locks   *MemStore
	logger  mlog.Logger
}

// NewMemoryLock 创建一个基于内存的锁。
//
// lottery 是 [分子, 分母] 组合：每次 Acquire 有 lottery[0]/lottery[1] 的概率触发一次僵尸锁清理。
func NewMemoryLock(timer timer.Timer, lottery [2]int, store *MemStore, logger mlog.Logger) Locker {
	return &memoryLock{
		owner:   rand.String(40),
		lottery: lottery,
		timer:   timer,
		locks:   store,
		logger:  logger,
	}
}

// ID 返回当前锁实例的持有者标识。
func (m *memoryLock) ID() string {
	return m.owner
}

// Type 返回锁驱动类型标识。
func (m *memoryLock) Type() string {
	return "memory"
}

// acquireInternal 尝试以当前 owner 获取 key 锁：
// key 不存在则直接写入；已存在但过期则接管；未过期则获取失败。
func (m *memoryLock) acquireInternal(key string, seconds int64) bool {
	unix := m.timer.Now().Unix()
	expiration := unix + seconds

	item := m.locks.Get(key)
	if item == nil {
		m.locks.Add(key, &MemItem{owner: m.owner, expiresAt: expiration})
		return true
	}
	if item.expiresAt <= unix {
		m.locks.Update(key, &MemItem{owner: m.owner, expiresAt: expiration})
		return true
	}
	return false
}

// Acquire 尝试获取 key 锁并返回是否成功。
//
// 每次获取会以 lottery 概率触发一次僵尸锁清理，避免过期项无限堆积。
func (m *memoryLock) Acquire(key string, seconds int64) bool {
	m.Lock()
	defer m.Unlock()

	acquired := m.acquireInternal(key, seconds)

	if rand.Intn(m.lottery[1]) < m.lottery[0] {
		m.locks.CleanupExpired(m.timer.Now().Unix(), staleWindow)
	}

	return acquired
}

// RenewalAcquire 获取 key 锁并启动后台续期协程，每 renewalSeconds 秒续期一次，
// 返回 release 函数与是否获取成功。
func (m *memoryLock) RenewalAcquire(key string, seconds int64, renewalSeconds int64) (func(), bool) {
	if m.Acquire(key, seconds) {
		ctx, cancelFunc := context.WithCancel(context.TODO())
		go m.renewalRoutine(ctx, key, seconds, renewalSeconds)
		return func() {
			cancelFunc()
			m.Release(key)
		}, true
	}
	return nil, false
}

// renewalRoutine 在后台周期续期锁，直到 ctx 取消或续期失败。
func (m *memoryLock) renewalRoutine(ctx context.Context, key string, seconds, renewalSeconds int64) {
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
}

// Release 释放 key 锁。仅持有者能释放，成功返回 true。
func (m *memoryLock) Release(key string) bool {
	m.Lock()
	defer m.Unlock()
	if item := m.locks.Get(key); item != nil && item.owner == m.owner {
		m.locks.Delete(key)
		return true
	}
	return false
}

// ForceRelease 无条件释放 key 锁（不校验持有者）。
func (m *memoryLock) ForceRelease(key string) bool {
	m.locks.Delete(key)
	return true
}

// Owner 返回 key 锁当前的持有者，未加锁时返回空串。
func (m *memoryLock) Owner(key string) string {
	if item := m.locks.Get(key); item != nil {
		return item.owner
	}
	return ""
}

// renewalExistKey 若 key 锁仍由当前 owner 持有，则将其过期时间顺延 seconds 秒。
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
