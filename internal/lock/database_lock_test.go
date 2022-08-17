package lock

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/internal/adapter"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func TestMain(t *testing.M) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", "root", "", "127.0.0.1", "3306", "lock_db_test")
	gormDB, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: &adapter.GormLoggerAdapter{}})
	sqlDB, _ := gormDB.DB()
	db = gormDB
	var all []*models.CacheLock
	db.Find(&all)
	for _, lock := range all {
		db.Delete(&lock)
	}
	db.AutoMigrate(&models.CacheLock{})
	code := t.Run()
	sqlDB.Close()
	os.Exit(code)
}

func TestNewDatabaseLock(t *testing.T) {
	t.Parallel()
	lock := NewDatabaseLock([2]int{1, 2}, nil)
	assert.Implements(t, (*contracts.Locker)(nil), lock)
}

func TestDatabaseLock_Acquire(t *testing.T) {
	t.Parallel()
	key := "Acquire"
	key2 := "Acquire2"
	lock := NewDatabaseLock([2]int{-1, 100}, db)

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
	var cl = &models.CacheLock{}
	assert.Nil(t, db.Where("`key` = ?", key).First(cl).Error)

	assert.True(t, lock.Acquire(key2, 1))
	defer lock.Release(key2)
	time.Sleep(2 * time.Second)
	assert.True(t, lock.Acquire(key2, 1))
}

type mockTimer struct {
	i int
	l []int64
}

func (m *mockTimer) Unix() int64 {
	t := m.l[m.i]
	m.i++
	return t
}

func TestDatabaseLock_AcquireLottery(t *testing.T) {
	t.Parallel()
	key := "AcquireLottery"
	key2 := "AcquireLottery2"
	lock := NewDatabaseLock([2]int{5, 1}, db)
	lock.timer = &mockTimer{l: []int64{100, 162}}
	acquire := lock.Acquire(key, 1)
	defer lock.Release(key)
	assert.True(t, acquire)
	var count int64
	db.Model(&models.CacheLock{}).Where("`key` = ?", key).Count(&count)
	assert.Equal(t, int64(1), count)

	lock.Acquire(key2, 1000)
	defer lock.Release(key2)
	db.Model(&models.CacheLock{}).Where("`key` = ?", key).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestDatabaseLock_ForceRelease(t *testing.T) {
	t.Parallel()
	key := "ForceRelease"
	lockOne := NewDatabaseLock([2]int{-1, 100}, db)
	lockTwo := NewDatabaseLock([2]int{-1, 100}, db)

	lockOne.Acquire(key, 1000)
	defer lockOne.Release(key)
	assert.NotEqual(t, lockOne.owner, lockTwo.owner)
	assert.Equal(t, lockOne.owner, lockOne.Owner(key))
	assert.Equal(t, lockOne.owner, lockTwo.Owner(key))

	assert.False(t, lockTwo.Release(key))
	assert.True(t, lockTwo.ForceRelease(key))
}

func TestDatabaseLock_Owner(t *testing.T) {
	t.Parallel()
	key := "Owner"
	key2 := "Owner2"
	lockOne := NewDatabaseLock([2]int{-1, 100}, db)
	lockTwo := NewDatabaseLock([2]int{-1, 100}, db)
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

func TestDatabaseLock_Release(t *testing.T) {
	t.Parallel()
	key := "Release"
	lockOne := NewDatabaseLock([2]int{-1, 100}, db)
	lockTwo := NewDatabaseLock([2]int{-1, 100}, db)
	assert.NotEqual(t, lockOne.owner, lockTwo.owner)

	lockOne.Acquire(key, 1000)
	defer lockOne.Release(key)

	assert.False(t, lockTwo.Release(key))
	assert.True(t, lockOne.Release(key))
}

func Test_realTimers_Now(t *testing.T) {
	t.Parallel()
	assert.Greater(t, time.Now().Unix()+5, (&realTimers{}).Unix())
	assert.Less(t, time.Now().Unix()-5, (&realTimers{}).Unix())
}

func TestDatabaseLock_RenewalAcquire(t *testing.T) {
	t.Parallel()
	key := "RenewalAcquire"
	lock := NewDatabaseLock([2]int{-1, 100}, db)
	lock2 := NewDatabaseLock([2]int{-1, 100}, db)
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

func BenchmarkDatabaseLock_RenewalAcquire(b *testing.B) {
	lock := NewDatabaseLock([2]int{-1, 100}, db)
	for i := 0; i < b.N; i++ {
		if release, ok := lock.RenewalAcquire(fmt.Sprintf("key-%v", i), 3, 2); ok {
			release()
		}
	}
}
