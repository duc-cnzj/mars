package cache

//go:generate mockgen -destination ./mock_cache.go -package cache github.com/duc-cnzj/mars/v5/internal/cache Cache

import "github.com/google/wire"

var WireCache = wire.NewSet(NewCacheImpl)
