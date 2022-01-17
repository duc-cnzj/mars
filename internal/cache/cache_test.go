package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/internal/adapter"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/utils/singleflight"
	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
)

func TestNewCache(t *testing.T) {
	assert.Implements(t, (*contracts.CacheInterface)(nil), NewCache(nil, nil))
}

func TestCache_Remember(t *testing.T) {
	var i int
	cache := NewCache(adapter.NewGoCacheAdapter(gocache.New(5*time.Minute, 10*time.Minute)), &singleflight.Group{})
	fn := func() {
		cache.Remember("duc", 10, func() ([]byte, error) {
			i++
			return []byte("duccc"), nil
		})
	}
	fn()
	fn()
	fn()
	assert.Equal(t, 1, i)

	y := 0
	fn2 := func() {
		cache.Remember("cache-y", 1, func() ([]byte, error) {
			y++
			return []byte("duccc"), nil
		})
	}
	fn2()
	time.Sleep(2 * time.Second)
	fn2()
	assert.Equal(t, 2, y)

	z := 0
	fn3 := func() {
		cache.Remember("cache-z", 1, func() ([]byte, error) {
			z++
			return nil, errors.New("error fn3")
		})
	}
	fn3()
	fn3()
	fn3()
	assert.Equal(t, 3, z)
}

type errorstore struct{}

var storeerr = errors.New("store error")

func (e *errorstore) Set(key, value []byte, expireSeconds int) (err error) {
	return storeerr
}

func (e *errorstore) Get(key []byte) (value []byte, err error) {
	return nil, errors.New("errorstore get err")
}

func TestCache_RememberErrorStore(t *testing.T) {
	var i int
	cache := NewCache(&errorstore{}, &singleflight.Group{})
	fn := func() ([]byte, error) {
		return cache.Remember("duc", 10, func() ([]byte, error) {
			i++
			return []byte("duccc"), nil
		})
	}
	_, err := fn()
	assert.ErrorIs(t, storeerr, err)
	_, err = fn()
	assert.ErrorIs(t, storeerr, err)
	assert.Equal(t, 2, i)
}

func TestCache_RememberNoCache(t *testing.T) {
	var i int
	cache := &NoCache{}
	fn := func() {
		cache.Remember("duc", 10, func() ([]byte, error) {
			i++
			return []byte("duccc"), nil
		})
	}
	fn()
	fn()
	fn()
	assert.Equal(t, 3, i)
}
