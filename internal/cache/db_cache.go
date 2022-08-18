package cache

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"gorm.io/gorm/clause"
)

type DBCache struct {
	app contracts.ApplicationInterface
}

func NewDBCache(app contracts.ApplicationInterface) *DBCache {
	return &DBCache{app: app}
}

func (c *DBCache) Remember(key string, seconds int, fn func() ([]byte, error)) ([]byte, error) {
	do, err, _ := c.app.Singleflight().Do(c.cacheKey(key), func() (any, error) {
		if seconds <= 0 {
			return fn()
		}

		var cache models.DBCache
		c.app.DBManager().DB().Where("`key` = ? and `expired_at` >= ?", key, time.Now()).First(&cache)
		if cache.Key != "" {
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
		if err = c.app.DBManager().DB().Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "expired_at", "updated_at"}),
		}).Create(&cache).Error; err != nil {
			mlog.Error(err)
		}
		return bytes, nil
	})
	if err != nil {
		return nil, err
	}
	return do.([]byte), nil
}

func (c *DBCache) Clear(key string) error {
	return c.app.DBManager().DB().Where("`key` = ?", key).Delete(&models.DBCache{}).Error
}

func (c *DBCache) cacheKey(key string) string {
	return fmt.Sprintf("cache-remember-%s", key)
}
