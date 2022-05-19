package cache

import (
	"encoding/base64"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"
	"github.com/duc-cnzj/mars/internal/utils/singleflight"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

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
	called := 0
	fn := func() ([]byte, error) {
		called++
		time.Sleep(2 * time.Second)
		return []byte("data"), nil
	}
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			NewDBCache(app).Remember("key", 1000, fn)
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
	NewDBCache(app).Remember("key", 1000, fn)

	assert.Equal(t, 2, called)
	NewDBCache(app).Remember("key", 1000, fn)
	assert.Equal(t, 2, called)
	NewDBCache(app).Remember("key", 1000, fn)
	assert.Equal(t, 2, called)

	_, err := NewDBCache(app).Remember("key-err", 1000, func() ([]byte, error) {
		return nil, errors.New("aaa")
	})
	assert.Equal(t, "aaa", err.Error())
}

func TestNewDBCache(t *testing.T) {
	assert.Implements(t, (*contracts.CacheInterface)(nil), NewDBCache(nil))
}
