package cache

import (
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/utils/singleflight"
)

type store interface {
	Get(key []byte) (value []byte, err error)
	Set(key, value []byte, expireSeconds int) (err error)
}

type Cache struct {
	fc store
	sf *singleflight.Group
}

func NewCache(fc store, sf *singleflight.Group) *Cache {
	return &Cache{fc: fc, sf: sf}
}

func (c *Cache) Remember(key string, seconds int, fn func() ([]byte, error)) ([]byte, error) {
	k := []byte(key)
	do, err, _ := c.sf.Do("CacheRemember:"+key, func() (interface{}, error) {
		mlog.Debug("CacheRemember:" + key + " Do.....")
		res, err := c.fc.Get(k)
		if err == nil {
			mlog.Debug("from cache")
			return res, nil
		}
		res, err = fn()
		if err != nil {
			return nil, err
		}
		if err = c.fc.Set([]byte(key), res, seconds); err != nil {
			return nil, err
		}
		return res, nil
	})

	if err != nil {
		return nil, err
	}

	return do.([]byte), err
}
