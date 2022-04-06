package cache

import (
	"time"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"gorm.io/gorm"
)

type DBCache struct {
	app contracts.ApplicationInterface
}

func NewDBCache(app contracts.ApplicationInterface) *DBCache {
	return &DBCache{app: app}
}

func (c *DBCache) Remember(key string, seconds int, fn func() ([]byte, error)) ([]byte, error) {
	var cache models.DBCache
	if err := c.app.DBManager().DB().Transaction(func(db *gorm.DB) error {
		db.Raw(
			"SELECT `id`, `key`, `value`, `expired_at` from `db_cache` where `key` = ? and `expired_at` >= ? and `deleted_at` IS NULL for update",
			key,
			time.Now()).Scan(&cache)
		mlog.Debugf("CacheRemember: %s, from cache: %t", key, cache.ID > 0)
		if cache.ID > 0 {
			return nil
		}
		bytes, err := fn()
		if err != nil {
			return err
		}
		cache = models.DBCache{
			Key:       key,
			Value:     string(bytes),
			ExpiredAt: time.Now().Add(time.Duration(seconds) * time.Second),
		}
		if err = db.Create(&cache).Error; err != nil {
			mlog.Error(err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return []byte(cache.Value), nil
}
