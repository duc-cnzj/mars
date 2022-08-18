package lock

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
)

func TestNewMemoryLock(t *testing.T) {
	t.Parallel()
	lock := NewMemoryLock([2]int{1, 2}, newStore())
	assert.Implements(t, (*contracts.Locker)(nil), lock)
	lock2 := NewMemoryLock([2]int{1, 2}, nil)
	assert.Same(t, defaultStore, lock2.(*memoryLock).locks)
}

func TestMemoryLock_Acquire(t *testing.T) {
	t.Parallel()
	key := "Acquire"
	key2 := "Acquire2"
	lock := NewMemoryLock([2]int{-1, 100}, newStore())

	num := 10
	var count int64
	wg := sync.WaitGroup{}
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			acquire := lock.Acquire(key, 100)
			if acquire {
				atomic.AddInt64(&count, 1)
			}
		}(i)
	}
	wg.Wait()
	defer lock.Release(key)
	assert.Equal(t, int64(1), atomic.LoadInt64(&count))
	assert.True(t, lock.Acquire(key2, 1))
	defer lock.Release(key2)
	time.Sleep(2 * time.Second)
	assert.True(t, lock.Acquire(key2, 1))
}

func TestMemoryLock_AcquireLottery(t *testing.T) {
	t.Parallel()
	key := "AcquireLottery"
	key2 := "AcquireLottery2"
	lock := NewMemoryLock([2]int{5, 1}, newStore()).(*memoryLock)
	lock.timer = &mockTimer{l: []int64{100, 162}}
	acquire := lock.Acquire(key, 1)
	defer lock.Release(key)
	assert.True(t, acquire)
	assert.Equal(t, 1, lock.Count())

	lock.Acquire(key2, 1000)
	defer lock.Release(key2)
	assert.Equal(t, 1, lock.Count())
}

func TestMemoryLock_ForceRelease(t *testing.T) {
	t.Parallel()
	key := "ForceRelease"
	s := newStore()
	lockOne := NewMemoryLock([2]int{-1, 100}, s).(*memoryLock)
	lockTwo := NewMemoryLock([2]int{-1, 100}, s).(*memoryLock)

	lockOne.Acquire(key, 1000)
	defer lockOne.Release(key)
	assert.NotEqual(t, lockOne.owner, lockTwo.owner)
	assert.Equal(t, lockOne.owner, lockOne.Owner(key))
	assert.Equal(t, lockOne.owner, lockTwo.Owner(key))

	assert.False(t, lockTwo.Release(key))
	assert.True(t, lockTwo.ForceRelease(key))
}

func TestMemoryLock_Owner(t *testing.T) {
	t.Parallel()
	key := "Owner"
	key2 := "Owner2"
	s := newStore()
	lockOne := NewMemoryLock([2]int{-1, 100}, s).(*memoryLock)
	lockTwo := NewMemoryLock([2]int{-1, 100}, s).(*memoryLock)
	assert.NotEqual(t, lockOne.owner, lockTwo.owner)

	lockOne.Acquire(key, 1000)
	defer lockOne.Release(key)
	assert.Equal(t, lockOne.owner, lockOne.Owner(key))
	assert.Equal(t, lockOne.owner, lockTwo.Owner(key))

	lockTwo.Acquire(key2, 1000)
	defer lockTwo.Release(key2)
	assert.Equal(t, lockTwo.owner, lockOne.Owner(key2))
	assert.Equal(t, lockTwo.owner, lockTwo.Owner(key2))

	assert.Equal(t, "", lockTwo.Owner("not-exists"))
}

func TestMemoryLock_Release(t *testing.T) {
	t.Parallel()
	key := "Release"
	s := newStore()
	lockOne := NewMemoryLock([2]int{-1, 100}, s).(*memoryLock)
	lockTwo := NewMemoryLock([2]int{-1, 100}, s).(*memoryLock)
	assert.NotEqual(t, lockOne.owner, lockTwo.owner)

	lockOne.Acquire(key, 1000)
	defer lockOne.Release(key)

	assert.False(t, lockTwo.Release(key))
	assert.True(t, lockOne.Release(key))
}

func TestMemoryLock_RenewalAcquire(t *testing.T) {
	t.Parallel()
	key := "RenewalAcquire"
	s := newStore()
	lock := NewMemoryLock([2]int{-1, 100}, s).(*memoryLock)
	lock2 := NewMemoryLock([2]int{-1, 100}, s).(*memoryLock)
	var i int64
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if release, ok := lock.RenewalAcquire(key, 3, 2); ok {
			func() {
				defer release()
				atomic.AddInt64(&i, 1)
				time.Sleep(5 * time.Second)
			}()
		}
	}()
	time.Sleep(4 * time.Second)
	assert.False(t, lock2.Acquire(key, 10))
	assert.False(t, lock.Acquire(key, 10))
	wg.Wait()
	assert.Equal(t, int64(1), atomic.LoadInt64(&i))
}

func TestMemoryLock_RenewalAcquire2(t *testing.T) {
	lock := NewMemoryLock([2]int{-1, 100}, newStore()).(*memoryLock)
	assert.False(t, lock.renewalExistKey("not-exists", 10))
	key := "RenewalAcquire2"
	fn, ok := lock.RenewalAcquire(key, 3, 2)
	defer fn()
	assert.True(t, ok)
	_, ok2 := lock.RenewalAcquire(key, 3, 2)
	assert.False(t, ok2)
}

func TestMemoryLock_RenewalAcquire3(t *testing.T) {
	t.Parallel()
	key := "RenewalAcquire3"
	s := newStore()
	lock := NewMemoryLock([2]int{-1, 100}, s).(*memoryLock)
	lock2 := NewMemoryLock([2]int{-1, 100}, s).(*memoryLock)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if release, ok := lock.RenewalAcquire(key, 3, 2); ok {
			func() {
				defer release()
				time.Sleep(5 * time.Second)
			}()
		}
	}()
	go func() {
		time.Sleep(1 * time.Second)
		lock.Release(key)
	}()
	time.Sleep(4 * time.Second)
	assert.True(t, lock2.Acquire(key, 10))
	wg.Wait()
}

func BenchmarkMemoryLock_RenewalAcquire(b *testing.B) {
	mlog.SetLogger(mlog.NewEmptyLogger())
	lock := NewMemoryLock([2]int{-1, 100}, newStore()).(*memoryLock)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%v", i)
		if release, ok := lock.RenewalAcquire(key, 3, 2); ok {
			_ = func() {
				release()
			}
			//lock.Release(key)
		}
	}
}
