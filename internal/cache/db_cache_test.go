package cache

import (
	"encoding/base64"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"
	"golang.org/x/sync/singleflight"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func newCacheByApp(app contracts.ApplicationInterface) *DBCache {
	return NewDBCache(app.Singleflight(), func() *gorm.DB {
		return app.DB()
	})
}

func TestDBCache_Remember(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	db, closeFn := testutil.SetGormDB(ctrl, app)
	defer closeFn()
	sf := singleflight.Group{}
	app.EXPECT().Singleflight().Return(&sf).AnyTimes()
	db.AutoMigrate(&models.DBCache{})
	var called int64
	fn := func() ([]byte, error) {
		atomic.AddInt64(&called, 1)
		time.Sleep(2 * time.Second)
		return []byte("data"), nil
	}
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			newCacheByApp(app).Remember(NewKey("key"), 1000, fn)
		}()
	}
	wg.Wait()
	c := models.DBCache{}
	db.First(&c)
	var count int64
	db.Model(&models.DBCache{}).Count(&count)
	assert.Equal(t, int64(1), count)
	assert.Equal(t, "key", c.Key)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("data")), c.Value)
	// 手动修改 value 让 b64 报错，会走 fn 的让 called + 1
	db.Model(c).Update("value", "xxx")
	newCacheByApp(app).Remember(NewKey("key"), 1000, fn)

	assert.Equal(t, int64(2), atomic.LoadInt64(&called))
	newCacheByApp(app).Remember(NewKey("key"), 1000, fn)
	assert.Equal(t, int64(2), atomic.LoadInt64(&called))
	newCacheByApp(app).Remember(NewKey("key"), 1000, fn)
	assert.Equal(t, int64(2), atomic.LoadInt64(&called))

	_, err := newCacheByApp(app).Remember(NewKey("key-err"), 1000, func() ([]byte, error) {
		return nil, errors.New("aaa")
	})
	assert.Equal(t, "aaa", err.Error())

	nocacheCalled := 0
	_, err = newCacheByApp(app).Remember(NewKey("no-cache-0-second"), 10, func() ([]byte, error) {
		nocacheCalled++
		return nil, nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, nocacheCalled)

	_, err = newCacheByApp(app).Remember(NewKey("no-cache-0-second"), 0, func() ([]byte, error) {
		nocacheCalled++
		return nil, nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, nocacheCalled)
}

func TestName(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	db, closeFn := testutil.SetGormDB(ctrl, app)
	defer closeFn()
	sf := singleflight.Group{}
	app.EXPECT().Singleflight().Return(&sf).AnyTimes()
	db.AutoMigrate(&models.DBCache{})
	called := 0
	NewDBCache(app.Singleflight(), func() *gorm.DB {
		if called > 0 {
			db.Migrator().DropTable(&models.DBCache{})
			assert.False(t, db.Migrator().HasTable(&models.DBCache{}))
			return db
		}
		called++
		return db
	}).Remember(NewKey("aaaa-xx"), 1, func() ([]byte, error) {
		return []byte("xxx"), nil
	})
}

func TestNewDBCache(t *testing.T) {
	assert.Implements(t, (*contracts.CacheInterface)(nil), NewDBCache(nil, nil))
}

func TestDBCache_Clear(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	db, closeFn := testutil.SetGormDB(ctrl, app)
	defer closeFn()
	sf := singleflight.Group{}
	app.EXPECT().Singleflight().Return(&sf).AnyTimes()
	db.AutoMigrate(&models.DBCache{})

	var count int64
	db.Model(&models.DBCache{}).Count(&count)
	assert.Equal(t, int64(0), count)

	newCacheByApp(app).Remember(NewKey("key-1"), 100, func() ([]byte, error) {
		return []byte("a-1"), nil
	})
	newCacheByApp(app).Remember(NewKey("key-2"), 100, func() ([]byte, error) {
		return []byte("a-2"), nil
	})
	db.Model(&models.DBCache{}).Count(&count)
	assert.Equal(t, int64(2), count)

	assert.Nil(t, newCacheByApp(app).Clear(NewKey("key-1")))

	var caches []*models.DBCache
	db.Model(&models.DBCache{}).Find(&caches)
	assert.Len(t, caches, 1)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("a-2")), caches[0].Value)
}

func TestDBCache_SetWithTTL(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	db, closeFn := testutil.SetGormDB(ctrl, app)
	defer closeFn()
	sf := singleflight.Group{}
	app.EXPECT().Singleflight().Return(&sf).AnyTimes()
	db.AutoMigrate(&models.DBCache{})

	var (
		count  int64
		keyOne models.DBCache
	)

	db.Model(&models.DBCache{}).Count(&count)
	assert.Equal(t, int64(0), count)

	assert.Nil(t, newCacheByApp(app).SetWithTTL(NewKey("key-1"), []byte("aa"), 100))
	db.Model(&models.DBCache{}).Count(&count)
	assert.Nil(t, db.First(&keyOne).Error)
	decodeString, _ := base64.StdEncoding.DecodeString(keyOne.Value)
	assert.Equal(t, "aa", string(decodeString))
	assert.Equal(t, int64(1), count)

	assert.Nil(t, newCacheByApp(app).SetWithTTL(NewKey("key-1"), []byte("bb"), 100))
	db.Model(&models.DBCache{}).Count(&count)
	assert.Equal(t, int64(1), count)
	assert.Nil(t, db.First(&keyOne).Error)
	decodeString, _ = base64.StdEncoding.DecodeString(keyOne.Value)
	assert.Equal(t, "bb", string(decodeString))
}
