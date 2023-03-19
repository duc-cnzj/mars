package cache

import (
	"encoding/base64"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/models"
)

type DBCache struct {
	sf     *singleflight.Group
	dbFunc func() *gorm.DB
}

func NewDBCache(sf *singleflight.Group, dbFunc func() *gorm.DB) *DBCache {
	return &DBCache{sf: sf, dbFunc: dbFunc}
}

func (c *DBCache) Remember(key contracts.CacheKeyInterface, seconds int, fn func() ([]byte, error)) ([]byte, error) {
	do, err, _ := c.sf.Do(c.cacheKey(key.String()), func() (any, error) {
		if seconds <= 0 {
			return fn()
		}

		var cache models.DBCache
		c.dbFunc().Where("`key` = ? and `expired_at` >= ?", key.String(), time.Now()).First(&cache)
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
		if err := c.SetWithTTL(key, bytes, seconds); err != nil {
			mlog.Error(err)
		}
		return bytes, nil
	})
	if err != nil {
		return nil, err
	}
	return do.([]byte), nil
}

func (c *DBCache) Clear(key contracts.CacheKeyInterface) error {
	return c.dbFunc().Where("`key` = ?", key.String()).Delete(&models.DBCache{}).Error
}

func (c *DBCache) SetWithTTL(key contracts.CacheKeyInterface, value []byte, seconds int) error {
	toString := base64.StdEncoding.EncodeToString(value)
	cache := models.DBCache{
		Key:       key.String(),
		Value:     toString,
		ExpiredAt: time.Now().Add(time.Duration(seconds) * time.Second),
	}

	return c.dbFunc().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "expired_at"}),
	}).Create(&cache).Error
}

func (c *DBCache) cacheKey(key string) string {
	return fmt.Sprintf("cache-remember-%s", key)
}
