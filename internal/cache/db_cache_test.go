package cache

import (
	"encoding/base64"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/utils/singleflight"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDBCache_Remember(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	manager := mock.NewMockDBManager(ctrl)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s, _ := db.DB()
	defer s.Close()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
	sf := singleflight.Group{}
	app.EXPECT().Singleflight().Return(&sf).AnyTimes()
	db.AutoMigrate(&models.DBCache{})
	fn := func() ([]byte, error) {
		time.Sleep(2 * time.Second)
		return []byte("data"), nil
	}
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			NewDBCache(app).Remember("key", 10, fn)
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
}

func TestNewDBCache(t *testing.T) {
	assert.Implements(t, (*contracts.CacheInterface)(nil), NewDBCache(nil))
}
