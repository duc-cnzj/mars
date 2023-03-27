package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/adapter"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/singleflight"
)

func TestNewCache(t *testing.T) {
	assert.Implements(t, (*contracts.CacheInterface)(nil), NewCache(nil, nil))
}

func TestCache_Remember(t *testing.T) {
	var i int
	cache := NewCache(adapter.NewGoCacheAdapter(gocache.New(5*time.Minute, 10*time.Minute)), &singleflight.Group{})
	fn := func() {
		cache.Remember(NewKey("duc"), 10, func() ([]byte, error) {
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
		cache.Remember(NewKey("cache-y"), 1, func() ([]byte, error) {
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
		cache.Remember(NewKey("cache-z"), 1, func() ([]byte, error) {
			z++
			return nil, errors.New("error fn3")
		})
	}
	fn3()
	fn3()
	fn3()
	assert.Equal(t, 3, z)

	nocacheCalled := 0
	cache.Remember(NewKey("cache-nocache"), 10, func() ([]byte, error) {
		nocacheCalled++
		return nil, nil
	})
	cache.Remember(NewKey("cache-nocache"), 0, func() ([]byte, error) {
		nocacheCalled++
		return nil, nil
	})
	assert.Equal(t, 2, nocacheCalled)
}

type errorstore struct{}

var storeerr = errors.New("store error")

func (e *errorstore) Set(key string, value []byte, expireSeconds int) (err error) {
	return storeerr
}

func (e *errorstore) Get(key string) (value []byte, err error) {
	return nil, errors.New("errorstore get err")
}

func (e *errorstore) Delete(key string) error {
	return nil
}

func TestCache_RememberErrorStore(t *testing.T) {
	var i int
	cache := NewCache(&errorstore{}, &singleflight.Group{})
	fn := func() ([]byte, error) {
		return cache.Remember(NewKey("duc"), 10, func() ([]byte, error) {
			i++
			return []byte("duccc"), nil
		})
	}
	data, err := fn()
	assert.Equal(t, []byte("duccc"), data)
	assert.Nil(t, err)
	_, err = fn()
	assert.Nil(t, err)
	assert.Equal(t, 2, i)
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(adapter.NewGoCacheAdapter(gocache.New(5*time.Minute, 10*time.Minute)), &singleflight.Group{})
	called := 0
	fn := func() ([]byte, error) {
		called++
		return []byte("aaa"), nil
	}
	// +1
	cache.Remember(NewKey("aaa"), 100, fn)
	// +0
	cache.Remember(NewKey("aaa"), 100, fn)
	assert.Nil(t, cache.Clear(NewKey("aaa")))
	// +1
	cache.Remember(NewKey("aaa"), 100, fn)
	assert.Equal(t, 2, called)
	cache.Remember(NewKey("aaa"), 100, fn)
	assert.Equal(t, 2, called)
}

func TestCache_SetWithTTL(t *testing.T) {
	cache := NewCache(adapter.NewGoCacheAdapter(gocache.New(5*time.Minute, 10*time.Minute)), &singleflight.Group{})
	cache.SetWithTTL(NewKey("aaa"), []byte("aa"), 100)
	get, _ := cache.(*Cache).store.Get(NewKey("aaa").String())
	assert.Equal(t, "aa", string(get))
}
