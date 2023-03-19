package bootstrappers

import (
	"errors"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"gorm.io/gorm"

	"github.com/duc-cnzj/mars/v4/internal/adapter"
	cachein "github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/cachelock"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

type CacheBootstrapper struct{}

func (a *CacheBootstrapper) Tags() []string {
	return []string{}
}

func (a *CacheBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	cfg := app.Config()
	if cfg.DBDriver == "sqlite" && cfg.CacheDriver == "db" {
		cfg.CacheDriver = "memory"
		mlog.Warning(`使用 DBDriver 为 "sqlite" 时，CacheDriver,CacheLock 只能使用 "memory"!`)
	}
	driver := cfg.CacheDriver
	mlog.Infof("CacheBootstrapper booted! driver: %s", driver)

	switch driver {
	case "db":
		app.SetCache(cachein.NewDBCache(app.Singleflight(), func() *gorm.DB {
			return app.DB()
		}))
		app.SetCacheLock(cachelock.NewDatabaseLock([2]int{2, 100}, func() *gorm.DB {
			return app.DB()
		}))
	case "memory":
		c := gocache.New(5*time.Minute, 10*time.Minute)
		app.SetCacheLock(cachelock.NewMemoryLock([2]int{2, 100}, nil))
		app.SetCache(cachein.NewCache(adapter.NewGoCacheAdapter(c), app.Singleflight()))
	default:
		return errors.New("unknown cache driver: " + driver)
	}

	return nil
}
