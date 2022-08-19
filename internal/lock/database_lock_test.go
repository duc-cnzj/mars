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
var prepared bool

func TestMain(t *testing.M) {
	var (
		user   string = os.Getenv("DB_USERNAME")
		port   string = os.Getenv("DB_PORT")
		dbname string = os.Getenv("DB_DATABASE")
		dbhost string = os.Getenv("DB_HOST")
		dbpwd  string = os.Getenv("DB_PASSWORD")
	)
	setDefault := func(key *string, value string) {
		if *key == "" {
			*key = value
		}
	}
	setDefault(&user, "root")
	setDefault(&port, "3306")
	setDefault(&dbname, "mars_test_db")
	setDefault(&dbhost, "127.0.0.1")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, dbpwd, dbhost, port, dbname)
	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: &adapter.GormLoggerAdapter{}})
	if err == nil {
		prepared = true
		sqlDB, _ := gormDB.DB()
		defer sqlDB.Close()
		db = gormDB
		db.Migrator().DropTable(&models.CacheLock{})
		db.AutoMigrate(&models.CacheLock{})
	}
	code := t.Run()
	os.Exit(code)
}

func TestNewDatabaseLock(t *testing.T) {
	t.Parallel()
	lock := NewDatabaseLock([2]int{1, 2}, nil)
	assert.Implements(t, (*contracts.Locker)(nil), lock)
}

func TestDatabaseLock_Acquire(t *testing.T) {
	if !prepared {
		t.Skipf("db not installed")
	}
	t.Parallel()
	key := "Acquire"
	key2 := "Acquire2"
	lock := NewDatabaseLock([2]int{-1, 100}, func() *gorm.DB {
		return db
	})

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
	if !prepared {
		t.Skipf("db not installed")
	}
	t.Parallel()
	key := "AcquireLottery"
	key2 := "AcquireLottery2"
	lock := NewDatabaseLock([2]int{5, 1}, func() *gorm.DB {
		return db
	}).(*databaseLock)
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
	if !prepared {
		t.Skipf("db not installed")
	}
	t.Parallel()
	key := "ForceRelease"
	lockOne := NewDatabaseLock([2]int{-1, 100}, func() *gorm.DB {
		return db
	}).(*databaseLock)
	lockTwo := NewDatabaseLock([2]int{-1, 100}, func() *gorm.DB {
		return db
	}).(*databaseLock)

	lockOne.Acquire(key, 1000)
	defer lockOne.Release(key)
	assert.NotEqual(t, lockOne.owner, lockTwo.owner)
	assert.Equal(t, lockOne.owner, lockOne.Owner(key))
	assert.Equal(t, lockOne.owner, lockTwo.Owner(key))

	assert.False(t, lockTwo.Release(key))
	assert.True(t, lockTwo.ForceRelease(key))
}

func TestDatabaseLock_Owner(t *testing.T) {
	if !prepared {
		t.Skipf("db not installed")
	}
	t.Parallel()
	key := "Owner"
	key2 := "Owner2"
	lockOne := NewDatabaseLock([2]int{-1, 100}, func() *gorm.DB {
		return db
	}).(*databaseLock)
	lockTwo := NewDatabaseLock([2]int{-1, 100}, func() *gorm.DB {
		return db
	}).(*databaseLock)
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
	if !prepared {
		t.Skipf("db not installed")
	}
	t.Parallel()
	key := "Release"
	lockOne := NewDatabaseLock([2]int{-1, 100}, func() *gorm.DB {
		return db
	}).(*databaseLock)
	lockTwo := NewDatabaseLock([2]int{-1, 100}, func() *gorm.DB {
		return db
	}).(*databaseLock)
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
	if !prepared {
		t.Skipf("db not installed")
	}
	t.Parallel()
	key := "RenewalAcquire"
	lock := NewDatabaseLock([2]int{-1, 100}, func() *gorm.DB {
		return db
	}).(*databaseLock)
	lock2 := NewDatabaseLock([2]int{-1, 100}, func() *gorm.DB {
		return db
	}).(*databaseLock)
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

func TestDatabaseLock_RenewalAcquire2(t *testing.T) {
	if !prepared {
		t.Skipf("db not installed")
	}
	key := "RenewalAcquire2"
	lock := NewDatabaseLock([2]int{-1, 100}, func() *gorm.DB {
		return db
	}).(*databaseLock)
	assert.False(t, lock.renewalExistKey("not-exists", 10))

	fn, ok := lock.RenewalAcquire(key, 3, 2)
	defer fn()
	assert.True(t, ok)
	_, ok2 := lock.RenewalAcquire(key, 3, 2)
	assert.False(t, ok2)
}

func TestDatabaseLock_RenewalAcquire3(t *testing.T) {
	if !prepared {
		t.Skipf("db not installed")
	}
	t.Parallel()
	key := "RenewalAcquire3"
	lock := NewDatabaseLock([2]int{-1, 100}, func() *gorm.DB {
		return db
	}).(*databaseLock)
	lock2 := NewDatabaseLock([2]int{-1, 100}, func() *gorm.DB {
		return db
	}).(*databaseLock)
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

func BenchmarkDatabaseLock_RenewalAcquire(b *testing.B) {
	if !prepared {
		b.Skipf("db not installed")
	}
	lock := NewDatabaseLock([2]int{-1, 100}, func() *gorm.DB {
		return db
	}).(*databaseLock)
	for i := 0; i < b.N; i++ {
		if release, ok := lock.RenewalAcquire(fmt.Sprintf("key-%v", i), 3, 2); ok {
			release()
		}
	}
}
