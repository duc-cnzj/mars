package cache

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
)

type DBCache struct {
	app contracts.ApplicationInterface
}

func NewDBCache(app contracts.ApplicationInterface) *DBCache {
	return &DBCache{app: app}
}

func (c *DBCache) Remember(key string, seconds int, fn func() ([]byte, error)) ([]byte, error) {
	do, err, _ := c.app.Singleflight().Do(fmt.Sprintf("cache-remember-%s", key), func() (any, error) {
		if seconds <= 0 {
			return fn()
		}

		var cache models.DBCache
		c.app.DBManager().DB().Where("`key` = ? and `expired_at` >= ?", key, time.Now()).First(&cache)
		if cache.ID > 0 {
			bs, err := base64.StdEncoding.DecodeString(cache.Value)
			if err == nil {
				return bs, nil
			}
		}
		bytes, err := fn()
		if err != nil {
			return nil, err
		}
		toString := base64.StdEncoding.EncodeToString(bytes)
		cache = models.DBCache{
			Key:       key,
			Value:     toString,
			ExpiredAt: time.Now().Add(time.Duration(seconds) * time.Second),
		}
		if err = c.app.DBManager().DB().Create(&cache).Error; err != nil {
			mlog.Error(err)
		}
		return bytes, nil
	})
	if err != nil {
		return nil, err
	}
	return do.([]byte), nil
}
