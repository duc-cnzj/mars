package cache

import (
	"encoding/base64"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"gorm.io/gorm/clause"
)

type DBCache struct {
	singleflightFunc func() *singleflight.Group
	dbFunc           func() *gorm.DB
}

func NewDBCache(singleflightFunc func() *singleflight.Group, dbFunc func() *gorm.DB) *DBCache {
	return &DBCache{singleflightFunc: singleflightFunc, dbFunc: dbFunc}
}

func (c *DBCache) Remember(key string, seconds int, fn func() ([]byte, error)) ([]byte, error) {
	do, err, _ := c.singleflightFunc().Do(c.cacheKey(key), func() (any, error) {
		if seconds <= 0 {
			return fn()
		}

		var cache models.DBCache
		c.dbFunc().Where("`key` = ? and `expired_at` >= ?", key, time.Now()).First(&cache)
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
		if err = c.dbFunc().Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "expired_at"}),
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
	return c.dbFunc().Where("`key` = ?", key).Delete(&models.DBCache{}).Error
}

func (c *DBCache) cacheKey(key string) string {
	return fmt.Sprintf("cache-remember-%s", key)
}
