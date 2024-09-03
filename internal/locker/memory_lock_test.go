package locker

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
	"github.com/stretchr/testify/assert"
)

func TestNewMemoryLock(t *testing.T) {
	t.Parallel()
	lock := NewMemoryLock(timer.NewRealTimer(), [2]int{1, 2}, NewMemStore(), mlog.NewForConfig(nil))
	assert.Implements(t, (*Locker)(nil), lock)
}

func TestMemoryLock_Acquire(t *testing.T) {
	t.Parallel()
	key := "Acquire"
	key2 := "Acquire2"
	lock := NewMemoryLock(timer.NewRealTimer(), [2]int{1, 2}, NewMemStore(), mlog.NewForConfig(nil))

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

type mockTimer struct {
	n time.Time
}

func (m *mockTimer) Now() time.Time {
	return m.n
}

func TestMemoryLock_AcquireLottery(t *testing.T) {
	t.Parallel()

	key := "AcquireLottery"
	key2 := "AcquireLottery2"

	lock := NewMemoryLock(&mockTimer{n: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)}, [2]int{100, 1}, NewMemStore(), mlog.NewForConfig(nil)).(*memoryLock)
	acquire := lock.Acquire(key, 1)
	defer lock.Release(key)
	assert.True(t, acquire)
	assert.Equal(t, 1, lock.Count())

	lock.timer = &mockTimer{n: time.Date(2029, 1, 1, 0, 0, 0, 0, time.UTC)}
	lock.Acquire(key2, 1000)
	defer lock.Release(key2)
	assert.Equal(t, 1, lock.Count())
}

func TestMemoryLock_ForceRelease(t *testing.T) {
	t.Parallel()
	key := "ForceRelease"
	s := NewMemStore()
	lockOne := NewMemoryLock(timer.NewRealTimer(), [2]int{1, 2}, s, mlog.NewForConfig(nil)).(*memoryLock)
	lockTwo := NewMemoryLock(timer.NewRealTimer(), [2]int{1, 2}, s, mlog.NewForConfig(nil)).(*memoryLock)

	lockOne.Acquire(key, 1000)
	defer lockOne.Release(key)
	assert.NotEqual(t, lockOne.owner, lockTwo.owner)
	assert.Equal(t, lockOne.Owner(key), lockOne.owner)
	assert.Equal(t, lockOne.Owner(key), lockTwo.Owner(key))

	assert.False(t, lockTwo.Release(key))
	assert.True(t, lockTwo.ForceRelease(key))
}

func TestMemoryLock_Owner(t *testing.T) {
	t.Parallel()
	key := "Owner"
	key2 := "Owner2"
	s := NewMemStore()
	lockOne := NewMemoryLock(timer.NewRealTimer(), [2]int{1, 2}, s, mlog.NewForConfig(nil)).(*memoryLock)
	lockTwo := NewMemoryLock(timer.NewRealTimer(), [2]int{1, 2}, s, mlog.NewForConfig(nil)).(*memoryLock)
	assert.NotEqual(t, lockOne.owner, lockTwo.owner)

	lockOne.Acquire(key, 1000)
	defer lockOne.Release(key)
	assert.Equal(t, lockOne.owner, lockOne.Owner(key))
	assert.Equal(t, lockOne.Owner(key), lockTwo.Owner(key))

	lockTwo.Acquire(key2, 1000)
	defer lockTwo.Release(key2)
	assert.Equal(t, lockTwo.Owner(key2), lockOne.Owner(key2))
	assert.Equal(t, lockTwo.Owner(key2), lockTwo.Owner(key2))

	assert.Equal(t, "", lockTwo.Owner("not-exists"))
}

func TestMemoryLock_Release(t *testing.T) {
	t.Parallel()
	key := "Release"
	s := NewMemStore()
	lockOne := NewMemoryLock(timer.NewRealTimer(), [2]int{1, 2}, s, mlog.NewForConfig(nil)).(*memoryLock)
	lockTwo := NewMemoryLock(timer.NewRealTimer(), [2]int{1, 2}, s, mlog.NewForConfig(nil)).(*memoryLock)
	assert.NotEqual(t, lockOne.owner, lockTwo.owner)

	lockOne.Acquire(key, 1000)
	defer lockOne.Release(key)

	assert.False(t, lockTwo.Release(key))
	assert.True(t, lockOne.Release(key))
}

func TestMemoryLock_RenewalAcquire(t *testing.T) {
	t.Parallel()
	key := "RenewalAcquire"
	s := NewMemStore()
	lock := NewMemoryLock(timer.NewRealTimer(), [2]int{1, 2}, s, mlog.NewForConfig(nil))
	lock2 := NewMemoryLock(timer.NewRealTimer(), [2]int{1, 2}, s, mlog.NewForConfig(nil))
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
	lock := NewMemoryLock(timer.NewRealTimer(), [2]int{1, 2}, NewMemStore(), mlog.NewForConfig(nil)).(*memoryLock)
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
	s := NewMemStore()
	lock := NewMemoryLock(timer.NewRealTimer(), [2]int{1, 2}, s, mlog.NewForConfig(nil))
	lock2 := NewMemoryLock(timer.NewRealTimer(), [2]int{1, 2}, s, mlog.NewForConfig(nil))
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

type emptyLogger struct{}

func NewEmptyLogger() *emptyLogger {
	return &emptyLogger{}
}

func (e *emptyLogger) Debug(v ...any) {}

func (e *emptyLogger) Debugf(format string, v ...any) {}

func (e *emptyLogger) Warning(v ...any) {}

func (e *emptyLogger) Warningf(format string, v ...any) {}

func (e *emptyLogger) Info(v ...any) {}

func (e *emptyLogger) Infof(format string, v ...any) {}

func (e *emptyLogger) Error(v ...any) {}

func (e *emptyLogger) Errorf(format string, v ...any) {}

func (e *emptyLogger) Fatal(v ...any) {}

func (e *emptyLogger) Fatalf(format string, v ...any) {}

func BenchmarkMemoryLock_RenewalAcquire(b *testing.B) {
	lock := NewMemoryLock(timer.NewRealTimer(), [2]int{1, 2}, NewMemStore(), mlog.NewForConfig(nil)).(*memoryLock)
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

func Test_memoryLock_ID(t *testing.T) {
	id := NewMemoryLock(timer.NewRealTimer(), [2]int{1, 2}, NewMemStore(), mlog.NewForConfig(nil)).(*memoryLock).ID()
	assert.Len(t, id, 40)
}

func Test_memoryLock_Type(t *testing.T) {
	assert.Equal(t, "memory", NewMemoryLock(timer.NewRealTimer(), [2]int{1, 2}, NewMemStore(), mlog.NewForConfig(nil)).Type())
}
