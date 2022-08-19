package cache

import (
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/cron/commands"
	"github.com/duc-cnzj/mars/internal/mlog"

	"golang.org/x/sync/singleflight"
)

func init() {
	commands.Register(func(manager contracts.CronManager, app contracts.ApplicationInterface) {
		manager.NewCommand("help2", func() {
			mlog.Info("need help!")
		}).At("08:51")
	})
}

type Store interface {
	Get(key string) (value []byte, err error)
	Set(key string, value []byte, expireSeconds int) (err error)
	Delete(key string) error
}

type Cache struct {
	fc Store
	sf *singleflight.Group
}

func NewCache(fc Store, sf *singleflight.Group) *Cache {
	return &Cache{fc: fc, sf: sf}
}

func (c *Cache) Remember(key string, seconds int, fn func() ([]byte, error)) ([]byte, error) {
	do, err, _ := c.sf.Do("CacheRemember:"+key, func() (any, error) {
		if seconds <= 0 {
			return fn()
		}

		res, err := c.fc.Get(key)
		mlog.Debugf("CacheRemember: %s, from cache: %t", key, err == nil)
		if err == nil {
			return res, nil
		}
		res, err = fn()
		if err != nil {
			return nil, err
		}
		// 设置缓存阶段不管它有没有成功，我 fn() 都是成功的，所以需要返回
		err = c.fc.Set(key, res, seconds)
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

func (c *Cache) Clear(key string) error {
	return c.fc.Delete(key)
}
