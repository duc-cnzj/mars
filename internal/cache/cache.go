package cache

import (
	"time"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	gocache "github.com/patrickmn/go-cache"

	"golang.org/x/sync/singleflight"
)

type Store interface {
	Get(key string) (value []byte, err error)
	Set(key string, value []byte, expireSeconds int) (err error)
	Delete(key string) error
}

type CacheKey interface {
	String() string
	Slug() string
}

type Cache interface {
	SetWithTTL(key CacheKey, value []byte, seconds int) error
	Remember(key CacheKey, seconds int, fn func() ([]byte, error)) ([]byte, error)
	Clear(key CacheKey) error
	Store() Store
}

type cacheImpl struct {
	store  Store
	sf     *singleflight.Group
	logger mlog.Logger
}

func NewCacheImpl(cfg *config.Config, data data.Data, logger mlog.Logger, sf *singleflight.Group) (ca Cache) {
	switch cfg.CacheDriver {
	case "memory":
		ca = newCache(
			NewGoCacheAdapter(
				gocache.New(5*time.Minute, 10*time.Minute),
			),
			logger,
			sf,
		)
	case "db":
		ca = newCache(NewDBStore(data), logger, sf)
	default:
		ca = &NoCache{}
	}
	return newMetricsForCache(ca)
}

func newCache(store Store, logger mlog.Logger, sf *singleflight.Group) Cache {
	return &cacheImpl{store: store, sf: sf, logger: logger}
}

// Remember TODO
func (c *cacheImpl) Remember(key CacheKey, seconds int, fn func() ([]byte, error)) ([]byte, error) {
	do, err, _ := c.sf.Do("CacheRemember:"+key.String(), func() (any, error) {
		if seconds <= 0 {
			return fn()
		}

		res, err := c.store.Get(key.String())
		c.logger.Debugf("CacheRemember: %s, from cacheImpl: %t", key, err == nil)
		if err == nil {
			return res, nil
		}
		res, err = fn()
		if err != nil {
			return nil, err
		}
		// 设置缓存阶段不管它有没有成功，我 fn() 都是成功的，所以需要返回
		err = c.SetWithTTL(key, res, seconds)
		if err != nil {
			c.logger.Errorf("[CACHE MISSING]: key %s err %v", key, err)
		}
		return res, nil
	})

	if err != nil {
		return nil, err
	}

	return do.([]byte), err
}

// SetWithTTL TODO
func (c *cacheImpl) SetWithTTL(key CacheKey, value []byte, seconds int) error {
	return c.store.Set(key.String(), value, seconds)
}

// Clear TODO
func (c *cacheImpl) Clear(key CacheKey) error {
	return c.store.Delete(key.String())
}

// Store TODO
func (c *cacheImpl) Store() Store {
	return c.store
}
