package cache

import "github.com/google/wire"

var WireCache = wire.NewSet(NewCacheImpl)
