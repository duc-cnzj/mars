package adapter

import (
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/contracts"

	"github.com/patrickmn/go-cache"
)

type goCacheAdapter struct {
	c *cache.Cache
}

// NewGoCacheAdapter return Store instance.
func NewGoCacheAdapter(c *cache.Cache) contracts.Store {
	return &goCacheAdapter{c: c}
}

// Get return val from cache.
func (g *goCacheAdapter) Get(key string) (value []byte, err error) {
	v, b := g.c.Get(key)
	if !b {
		return nil, fmt.Errorf("key %s not found", key)
	}
	return v.([]byte), nil
}

// Set val for key.
func (g *goCacheAdapter) Set(key string, value []byte, expireSeconds int) (err error) {
	g.c.Set(key, value, time.Second*time.Duration(expireSeconds))

	return nil
}

// Delete key from cache.
func (g *goCacheAdapter) Delete(key string) error {
	g.c.Delete(key)

	return nil
}
