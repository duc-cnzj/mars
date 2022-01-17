package bootstrappers

import (
	"time"

	"github.com/duc-cnzj/mars/internal/adapter"
	cachein "github.com/duc-cnzj/mars/internal/cache"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/patrickmn/go-cache"
)

type CacheBootstrapper struct{}

func (a *CacheBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	mlog.Debug("CacheBootstrapper booted!")
	c := cache.New(5*time.Minute, 10*time.Minute)
	app.SetCache(cachein.NewCache(adapter.NewGoCacheAdapter(c), app.Singleflight()))

	return nil
}
