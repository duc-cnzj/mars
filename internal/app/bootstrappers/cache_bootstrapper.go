package bootstrappers

import (
	"errors"
	"time"

	gocache "github.com/patrickmn/go-cache"

	"github.com/duc-cnzj/mars/internal/adapter"
	cachein "github.com/duc-cnzj/mars/internal/cache"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
)

type CacheBootstrapper struct{}

func (a *CacheBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	cfg := app.Config()
	if cfg.DBDriver == "sqlite" && cfg.CacheDriver == "db" {
		cfg.CacheDriver = "memory"
		mlog.Warning(`使用 DBDriver 为 "sqlite" 时，CacheDriver 只能使用 "memory"!`)
	}
	driver := cfg.CacheDriver
	mlog.Infof("CacheBootstrapper booted! driver: %s", driver)

	switch driver {
	case "db":
		app.SetCache(cachein.NewDBCache(app))
	case "memory":
		c := gocache.New(5*time.Minute, 10*time.Minute)
		app.SetCache(cachein.NewCache(adapter.NewGoCacheAdapter(c), app.Singleflight()))
	default:
		return errors.New("unknown cache driver: " + driver)
	}

	return nil
}
