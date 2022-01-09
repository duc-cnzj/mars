package bootstrappers

import (
	"github.com/coocood/freecache"
	"github.com/duc-cnzj/mars/internal/cache"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
)

type CacheBootstrapper struct{}

func (a *CacheBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	mlog.Debug("CacheBootstrapper booted!")
	cacheSize := 100 * 1024 * 1024
	app.SetCache(cache.NewCache(freecache.NewCache(cacheSize), app.Singleflight()))

	return nil
}
