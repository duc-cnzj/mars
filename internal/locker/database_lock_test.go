package locker

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/cachelock"
	"github.com/duc-cnzj/mars/v4/internal/ent/migrate"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/util/timer"
	"github.com/stretchr/testify/assert"
)

var db *ent.Client
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
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, dbpwd, dbhost, port, dbname)
	var err error
	db, err = data.InitDB(dsn, mlog.NewLogger(nil))
	if err == nil {
		prepared = true
		err = db.Schema.Create(
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
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock := setupDatabaseLock()
	key := "testKey"
	seconds := int64(60)

	acquired := dbLock.Acquire(key, seconds)
	assert.True(t, acquired, "Expected to acquire lock")

	owner := dbLock.Owner(key)
	assert.Equal(t, dbLock.ID(), owner, "Expected owner to be the same as the ID of the lock")
}

func TestDatabaseLockAcquireWhenLockAlreadyExists(t *testing.T) {
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock := setupDatabaseLock()
	key := "testKey"
	seconds := int64(60)

	dbLock.Acquire(key, seconds)
	acquired := dbLock.Acquire(key, seconds)
	assert.False(t, acquired, "Expected not to acquire lock when it already exists")
}

func TestDatabaseLockRelease(t *testing.T) {
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock := setupDatabaseLock()
	key := "testKey"
	seconds := int64(60)

	dbLock.Acquire(key, seconds)
	released := dbLock.Release(key)
	assert.True(t, released, "Expected to release lock")

	owner := dbLock.Owner(key)
	assert.Empty(t, owner, "Expected owner to be empty after release")
}

func TestDatabaseLockForceRelease(t *testing.T) {
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock := setupDatabaseLock()
	key := "testKey"
	seconds := int64(60)

	dbLock.Acquire(key, seconds)
	dbLock.ForceRelease(key)

	owner := dbLock.Owner(key)
	assert.Empty(t, owner, "Expected owner to be empty after force release")
}

func TestDatabaseLockRenewalAcquire(t *testing.T) {
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock := setupDatabaseLock()
	key := "testKey"
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

func setupDatabaseLock() *databaseLock {
	deleteAllTestKey()
	return NewDatabaseLock(timer.NewRealTimer(), [2]int{1, 2}, data.NewDataImpl(&data.NewDataParams{DB: db}), mlog.NewLogger(nil)).(*databaseLock)
}

func Test_databaseLock_Type(t *testing.T) {
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock := setupDatabaseLock()

	assert.Equal(t, "database", dbLock.Type())
}

func Test_databaseLock_Release(t *testing.T) {
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock := setupDatabaseLock()
	key := "testKey"
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
	if !prepared {
		t.Skip("Database not prepared")
	}

	dbLock := NewDatabaseLock(timer.NewRealTimer(), [2]int{1, 2}, data.NewDataImpl(&data.NewDataParams{DB: db}), mlog.NewLogger(nil)).(*databaseLock)

	key := "testKey"
	seconds := int64(60)
	dbLock.ForceRelease(key)
	acquire := dbLock.Acquire(key, seconds)
	assert.True(t, acquire, "Expected to acquire lock")
	exist := dbLock.renewalExistKey(key, seconds)
	assert.True(t, exist, "Expected key to exist")

	dbLock.Release(key)
	exist = dbLock.renewalExistKey(key, seconds)
	assert.False(t, exist, "Expected key not to exist")
}

func Test_databaseLock_renewalExistKey_Concurrent(t *testing.T) {
	if !prepared {
		t.Skip("Database not prepared")
	}

	dbLock := NewDatabaseLock(timer.NewRealTimer(), [2]int{1, 2}, data.NewDataImpl(&data.NewDataParams{DB: db}), mlog.NewLogger(nil)).(*databaseLock)

	key := "testKey"
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
			assert.True(t, exist, "Expected key to exist")
		}()
	}
	wg.Wait()

	dbLock.Release(key)
	exist := dbLock.renewalExistKey(key, seconds)
	assert.False(t, exist, "Expected key not to exist")
}

func TestDatabaseLock_ConcurrentRenewalExistKey(t *testing.T) {
	if !prepared {
		t.Skip("Database not prepared")
	}
	lock := NewDatabaseLock(timer.NewRealTimer(), [2]int{1, 2}, data.NewDataImpl(&data.NewDataParams{DB: db}), mlog.NewLogger(nil)).(*databaseLock)
	anotherLock := NewDatabaseLock(timer.NewRealTimer(), [2]int{1, 2}, data.NewDataImpl(&data.NewDataParams{DB: db}), mlog.NewLogger(nil)).(*databaseLock)
	anotherLock2 := NewDatabaseLock(timer.NewRealTimer(), [2]int{1, 2}, data.NewDataImpl(&data.NewDataParams{DB: db}), mlog.NewLogger(nil)).(*databaseLock)
	key := "test_key"
	seconds := int64(10)

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
				if lock.renewalExistKey(key, seconds) {
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
				if anotherLock.renewalExistKey(key, seconds) {
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
				if anotherLock2.renewalExistKey(key, seconds) {
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
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock := setupDatabaseLock()

	_, err := db.CacheLock.Create().SetOwner("xxx").SetKey("testKey").SetExpiredAt(time.Now().Add(-time.Second * 60)).Save(context.TODO())
	assert.Nil(t, err)

	acquire := dbLock.Acquire("testKey", 60)
	assert.True(t, acquire, "Expected to acquire lock")
}

func Test_databaseLock_RenewalAcquire(t *testing.T) {
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock := setupDatabaseLock()
	_, err := db.CacheLock.Create().SetOwner("xxx").SetKey("testKey").SetExpiredAt(time.Now().Add(time.Second * 60)).Save(context.TODO())
	assert.Nil(t, err)
	renewalAcquire, b := dbLock.RenewalAcquire("testKey", 60, 100)
	assert.False(t, b, "Expected not to acquire lock")
	assert.Nil(t, renewalAcquire, "Expected renewalAcquire to be nil")
}

func deleteAllTestKey() {
	db.CacheLock.Delete().Where(cachelock.IDGT(0)).Exec(context.TODO())
}

func Test_databaseLock_renewalRoutine(t *testing.T) {
	if !prepared {
		t.Skip("Database not prepared")
	}
	dbLock := setupDatabaseLock()
	key := "testKey"
	dbLock.Acquire(key, 2)

	timeout, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		dbLock.renewalRoutine(timeout, key, 10, 2)
	}()

	wg.Wait()
	defer dbLock.Release(key)
	first, _ := db.CacheLock.Query().Where(cachelock.Key(key)).First(context.TODO())
	assert.Greater(t, int(first.ExpiredAt.Sub(time.Now()).Seconds()), 3)
}
