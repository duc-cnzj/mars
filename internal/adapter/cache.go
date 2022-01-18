package adapter

import (
	"errors"
	"time"

	"github.com/patrickmn/go-cache"
)

type GoCacheAdapter struct {
	c *cache.Cache
}

func NewGoCacheAdapter(c *cache.Cache) *GoCacheAdapter {
	return &GoCacheAdapter{c: c}
}

func (g *GoCacheAdapter) Get(key string) (value []byte, err error) {
	v, b := g.c.Get(key)
	if !b {
		return nil, errors.New("not found")
	}
	return v.([]byte), nil
}

func (g *GoCacheAdapter) Set(key string, value []byte, expireSeconds int) (err error) {
	g.c.Set(key, value, time.Second*time.Duration(expireSeconds))

	return nil
}
