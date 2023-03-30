package cache

import (
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"

	"golang.org/x/sync/singleflight"
)

type Cache struct {
	store contracts.Store
	sf    *singleflight.Group
}

func NewCache(store contracts.Store, sf *singleflight.Group) contracts.CacheInterface {
	return &Cache{store: store, sf: sf}
}

// Remember TODO
func (c *Cache) Remember(key contracts.CacheKeyInterface, seconds int, fn func() ([]byte, error)) ([]byte, error) {
	do, err, _ := c.sf.Do("CacheRemember:"+key.String(), func() (any, error) {
		if seconds <= 0 {
			return fn()
		}

		res, err := c.store.Get(key.String())
		mlog.Debugf("CacheRemember: %s, from cache: %t", key, err == nil)
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
			mlog.Errorf("[CACHE MISSING]: key %s err %v", key, err)
		}
		return res, nil
	})

	if err != nil {
		return nil, err
	}

	return do.([]byte), err
}

// SetWithTTL TODO
func (c *Cache) SetWithTTL(key contracts.CacheKeyInterface, value []byte, seconds int) error {
	return c.store.Set(key.String(), value, seconds)
}

// Clear TODO
func (c *Cache) Clear(key contracts.CacheKeyInterface) error {
	return c.store.Delete(key.String())
}

// Store TODO
func (c *Cache) Store() contracts.Store {
	return c.store
}
