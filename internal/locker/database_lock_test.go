package locker

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"entgo.io/ent/dialect/sql"

	dbpkg "github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/cachelock"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/migrate"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/stretchr/testify/assert"
)

var entClient *ent.Client
var prepared bool

func TestMain(t *testing.M) {
	var (
		user   = os.Getenv("DB_USERNAME")
		port   = os.Getenv("DB_PORT")
		dbname = os.Getenv("DB_DATABASE")
		dbhost = os.Getenv("DB_HOST")
		dbpwd  = os.Getenv("DB_PASSWORD")
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
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, dbpwd, dbhost, port, dbname)
	var err error
	open, _ := sql.Open("mysql", dsn)
	entClient = dbpkg.InitDB(open, mlog.NewForConfig(nil), false, 0, timer.NewReal())
	// InitDB 已无错误返回，这里显式 Ping 确认连接可用：
	// 连不上真实 MySQL 时跳过 DB 集成测试，避免 Schema.Create 直接 Fatal 崩溃。
	if open.DB().Ping() == nil {
		prepared = true
		err = entClient.Schema.Create(
			context.TODO(),
			migrate.WithDropIndex(true),
			migrate.WithDropColumn(true),
		)
		if err != nil {
			log.Fatal(err)
		}
	}
	code := t.Run()
	os.Exit(code)
}

func TestDatabaseLockAcquire(t *testing.T) {
	t.Parallel()
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock, key := setupDatabaseLock(t)
	seconds := int64(60)

	acquired := dbLock.Acquire(key, seconds)
	assert.True(t, acquired, "Expected to acquire lock")

	owner := dbLock.Owner(key)
	assert.Equal(t, dbLock.ID(), owner, "Expected owner to be the same as the ID of the lock")
}

func TestDatabaseLockAcquireWhenLockAlreadyExists(t *testing.T) {
	t.Parallel()
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock, key := setupDatabaseLock(t)
	seconds := int64(60)

	dbLock.Acquire(key, seconds)
	acquired := dbLock.Acquire(key, seconds)
	assert.False(t, acquired, "Expected not to acquire lock when it already exists")
}

func TestDatabaseLockRelease(t *testing.T) {
	t.Parallel()
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock, key := setupDatabaseLock(t)
	seconds := int64(60)

	dbLock.Acquire(key, seconds)
	released := dbLock.Release(key)
	assert.True(t, released, "Expected to release lock")

	owner := dbLock.Owner(key)
	assert.Empty(t, owner, "Expected owner to be empty after release")
}

func TestDatabaseLockForceRelease(t *testing.T) {
	t.Parallel()
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock, key := setupDatabaseLock(t)
	seconds := int64(60)

	dbLock.Acquire(key, seconds)
	dbLock.ForceRelease(key)

	owner := dbLock.Owner(key)
	assert.Empty(t, owner, "Expected owner to be empty after force release")
}

func TestDatabaseLockRenewalAcquire(t *testing.T) {
	t.Parallel()
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock, key := setupDatabaseLock(t)
	seconds := int64(60)
	renewalSeconds := int64(30)

	releaseFunc, acquired := dbLock.RenewalAcquire(key, seconds, renewalSeconds)
	assert.True(t, acquired, "Expected to acquire lock")
	assert.NotNil(t, releaseFunc, "Expected release function to be not nil")

	owner := dbLock.Owner(key)
	assert.Equal(t, dbLock.ID(), owner, "Expected owner to be the same as the ID of the lock")

	releaseFunc()
	ownerAfterRelease := dbLock.Owner(key)
	assert.Empty(t, ownerAfterRelease, "Expected owner to be empty after release")
}

// setupDatabaseLock 返回后端为真实 MySQL 的锁实例及其独占 key（key 取测试名）。
// 使用前只清理该 key 的历史残留：原实现 deleteAllTestKey 清空整张 cache_locks 表，
// 使全部 MySQL 测试天然互斥、无法并行；改为按 key 隔离后各测试互不干扰、可并行。
func setupDatabaseLock(t *testing.T) (*databaseLock, string) {
	t.Helper()
	key := t.Name()
	deleteTestKey(key)
	return NewDatabaseLock(timer.NewReal(), func() *ent.Client { return entClient }, mlog.NewForConfig(nil)).(*databaseLock), key
}

func Test_databaseLock_Type(t *testing.T) {
	t.Parallel()
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock, _ := setupDatabaseLock(t)

	assert.Equal(t, "db", dbLock.Type())
}

func Test_databaseLock_Release(t *testing.T) {
	t.Parallel()
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock, key := setupDatabaseLock(t)
	seconds := int64(60)

	dbLock.Acquire(key, seconds)
	released := dbLock.Release(key)
	assert.True(t, released, "Expected to release lock")

	owner := dbLock.Owner(key)
	assert.Empty(t, owner, "Expected owner to be empty after release")

	released = dbLock.Release(key)
	assert.False(t, released, "Expected not to release lock again")
}

func Test_databaseLock_renewalExistKey(t *testing.T) {
	t.Parallel()
	if !prepared {
		t.Skip("Database not prepared")
	}

	dbLock := NewDatabaseLock(timer.NewReal(), func() *ent.Client { return entClient }, mlog.NewForConfig(nil)).(*databaseLock)

	key := t.Name()
	seconds := int64(60)
	dbLock.ForceRelease(key)
	acquire := dbLock.Acquire(key, seconds)
	assert.True(t, acquire, "Expected to acquire lock")
	exist := dbLock.renewalExistKey(key, seconds)
	assert.Nil(t, exist, "Expected key to exist")

	dbLock.Release(key)
	exist = dbLock.renewalExistKey(key, seconds)
	assert.NotNil(t, exist, "Expected key not to exist")
}

func Test_databaseLock_renewalExistKey_Concurrent(t *testing.T) {
	t.Parallel()
	if !prepared {
		t.Skip("Database not prepared")
	}

	dbLock := NewDatabaseLock(timer.NewReal(), func() *ent.Client { return entClient }, mlog.NewForConfig(nil)).(*databaseLock)

	key := t.Name()
	seconds := int64(60)
	dbLock.ForceRelease(key)
	acquire := dbLock.Acquire(key, seconds)
	assert.True(t, acquire, "Expected to acquire lock")

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			exist := dbLock.renewalExistKey(key, seconds)
			assert.Nil(t, exist, "Expected key to exist")
		}()
	}
	wg.Wait()

	dbLock.Release(key)
	exist := dbLock.renewalExistKey(key, seconds)
	assert.NotNil(t, exist, "Expected key not to exist")
}

func TestDatabaseLock_ConcurrentRenewalExistKey(t *testing.T) {
	t.Parallel()
	if !prepared {
		t.Skip("Database not prepared")
	}
	lock := NewDatabaseLock(timer.NewReal(), func() *ent.Client { return entClient }, mlog.NewForConfig(nil)).(*databaseLock)
	anotherLock := NewDatabaseLock(timer.NewReal(), func() *ent.Client { return entClient }, mlog.NewForConfig(nil)).(*databaseLock)
	anotherLock2 := NewDatabaseLock(timer.NewReal(), func() *ent.Client { return entClient }, mlog.NewForConfig(nil)).(*databaseLock)
	key := t.Name()
	seconds := int64(10)
	deleteTestKey(key)

	// Acquire the lock
	acquired := lock.Acquire(key, seconds)
	assert.True(t, acquired)

	var wg sync.WaitGroup
	stopChan := make(chan struct{})
	wg.Add(3)

	var renewedByOwner, renewedByAnother, renewedByAnother2 int

	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopChan:
				return
			default:
				if err := lock.renewalExistKey(key, seconds); err == nil {
					renewedByOwner++
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopChan:
				return
			default:
				if err := anotherLock.renewalExistKey(key, seconds); err == nil {
					renewedByAnother++
				}
			}
		}
	}()
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopChan:
				return
			default:
				if err := anotherLock2.renewalExistKey(key, seconds); err == nil {
					renewedByAnother2++
				}
			}
		}
	}()

	// Run the test for a certain duration
	testDuration := 5 * time.Second
	time.Sleep(testDuration)
	close(stopChan)
	wg.Wait()

	// Ensure only the owner was able to renew the lock
	assert.Greater(t, renewedByOwner, 0, "The owner should be able to renew the lock")
	assert.Equal(t, 0, renewedByAnother, "A different owner should not be able to renew the lock")
	assert.Equal(t, 0, renewedByAnother2, "A different owner should not be able to renew the lock")
}

func Test_databaseLock_Acquire(t *testing.T) {
	t.Parallel()
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock, key := setupDatabaseLock(t)

	// 预置一条"已过期但未被清扫"的锁，验证 Acquire 能接管过期锁。
	// 过期时长取 30s 而非 60s：cleanupExpiredLocks 会删除 expired_at < now-60s 的行，
	// 若置为 now-60s 恰好落在清扫边界上，并行跑的其它测试的 Acquire 会在本测试的
	// Create 与接管之间把该行扫掉，导致 updateExpiredLock 匹配 0 行、接管失败（flaky）。
	_, err := entClient.CacheLock.Create().SetOwner("xxx").SetKey(key).SetExpiredAt(time.Now().Add(-time.Second * 30)).Save(context.TODO())
	assert.Nil(t, err)

	acquire := dbLock.Acquire(key, 60)
	assert.True(t, acquire, "Expected to acquire lock")
}

func Test_databaseLock_RenewalAcquire(t *testing.T) {
	t.Parallel()
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock, key := setupDatabaseLock(t)
	_, err := entClient.CacheLock.Create().SetOwner("xxx").SetKey(key).SetExpiredAt(time.Now().Add(time.Second * 60)).Save(context.TODO())
	assert.Nil(t, err)
	renewalAcquire, b := dbLock.RenewalAcquire(key, 60, 100)
	assert.False(t, b, "Expected not to acquire lock")
	assert.Nil(t, renewalAcquire, "Expected renewalAcquire to be nil")
}

// deleteTestKey 只清理指定 key 的锁记录，不影响同表内其它测试，是并行隔离的前提。
func deleteTestKey(key string) {
	entClient.CacheLock.Delete().Where(cachelock.Key(key)).Exec(context.TODO())
}

func Test_databaseLock_renewalRoutine(t *testing.T) {
	t.Parallel()
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock, key := setupDatabaseLock(t)
	assert.True(t, dbLock.Acquire(key, 2), "前置加锁必须成功，否则续期会因 not owner 提前退出")

	timeout, cancelFunc := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancelFunc()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		dbLock.renewalRoutine(timeout, key, 10, 2)
	}()

	wg.Wait()
	defer dbLock.Release(key)
	first, err := entClient.CacheLock.Query().Where(cachelock.Key(key)).First(context.TODO())
	assert.NoError(t, err, "续期协程不应删除锁记录")
	// 每 2s 续期一次、ttl 10s，5s 时锁的有效期应被推到未来。
	// 原断言 |now-ExpiredAt| > 3 恰好卡在 3.0 边界（负载下 flaky），
	// 且续期彻底失效时旧行的偏差同样 > 3、会假阳性通过。
	assert.True(t, first.ExpiredAt.After(time.Now()), "续期协程应持续推后过期时间，锁不应过期")
}
