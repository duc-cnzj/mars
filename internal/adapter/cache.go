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

func (g *GoCacheAdapter) Get(key []byte) (value []byte, err error) {
	get, b := g.c.Get(string(key))
	if !b {
		return nil, errors.New("not found")
	}
	return get.([]byte), nil
}

func (g *GoCacheAdapter) Set(key, value []byte, expireSeconds int) (err error) {
	g.c.Set(string(key), value, time.Second*time.Duration(expireSeconds))

	return nil
}
