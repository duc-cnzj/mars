package cache

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/data"

	"github.com/duc-cnzj/mars/v5/internal/mlog"
	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/singleflight"
)

func TestCache_Remember(t *testing.T) {
	var i int
	cache := newCache(NewGoCacheAdapter(gocache.New(5*time.Minute, 10*time.Minute)), mlog.NewLogger(nil), &singleflight.Group{})
	fn := func() {
		cache.Remember(NewKey("duc"), 10, func() ([]byte, error) {
			i++
			return []byte("duccc"), nil
		}, false)
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
		}, false)
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
		}, false)
	}
	fn3()
	fn3()
	fn3()
	assert.Equal(t, 3, z)

	nocacheCalled := 0
	cache.Remember(NewKey("cache-nocache"), 10, func() ([]byte, error) {
		nocacheCalled++
		return nil, nil
	}, false)
	cache.Remember(NewKey("cache-nocache"), 0, func() ([]byte, error) {
		nocacheCalled++
		return nil, nil
	}, false)
	assert.Equal(t, 2, nocacheCalled)
}

func TestCache_RememberV2(t *testing.T) {
	cache := newCache(NewGoCacheAdapter(gocache.New(5*time.Minute, 10*time.Minute)), mlog.NewLogger(nil), &singleflight.Group{})
	v := atomic.Int64{}
	v2 := atomic.Int64{}

	fn := func() {
		cache.Remember(NewKey("duc"), 10, func() ([]byte, error) {
			v.Add(1)
			time.Sleep(2 * time.Second)
			return []byte("duccc"), nil
		}, false)
	}

	wg := &sync.WaitGroup{}
	wg.Add(5)
	go func() {
		defer wg.Done()
		fn()
	}()
	go func() {
		defer wg.Done()
		fn()
	}()
	go func() {
		defer wg.Done()
		fn()
	}()

	v2Fn := func() ([]byte, error) {
		time.Sleep(2 * time.Second)
		v2.Add(1)
		return []byte("duccc"), nil
	}

	go func() {
		defer wg.Done()
		func() {
			cache.Remember(NewKey("duc"), 10, v2Fn, true)
		}()
	}()
	go func() {
		defer wg.Done()
		func() {
			cache.Remember(NewKey("duc"), 10, v2Fn, true)
		}()
	}()
	wg.Wait()
	assert.Equal(t, int64(1), v.Load())
	assert.Equal(t, int64(1), v2.Load())
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
	cache := newCache(&errorstore{}, mlog.NewLogger(nil), &singleflight.Group{})
	fn := func() ([]byte, error) {
		return cache.Remember(NewKey("duc"), 10, func() ([]byte, error) {
			i++
			return []byte("duccc"), nil
		}, false)
	}
	data, err := fn()
	assert.Equal(t, []byte("duccc"), data)
	assert.Nil(t, err)
	_, err = fn()
	assert.Nil(t, err)
	assert.Equal(t, 2, i)
}

func TestCache_Clear(t *testing.T) {
	cache := newCache(NewGoCacheAdapter(gocache.New(5*time.Minute, 10*time.Minute)), mlog.NewLogger(nil), &singleflight.Group{})
	called := 0
	fn := func() ([]byte, error) {
		called++
		return []byte("aaa"), nil
	}
	// +1
	cache.Remember(NewKey("aaa"), 100, fn, false)
	// +0
	cache.Remember(NewKey("aaa"), 100, fn, false)
	assert.Nil(t, cache.Clear(NewKey("aaa")))
	// +1
	cache.Remember(NewKey("aaa"), 100, fn, false)
	assert.Equal(t, 2, called)
	cache.Remember(NewKey("aaa"), 100, fn, false)
	assert.Equal(t, 2, called)
	cache.Remember(NewKey("aaa"), 100, fn, true)
	assert.Equal(t, 3, called)
}

func TestCache_SetWithTTL(t *testing.T) {
	cache := newCache(NewGoCacheAdapter(gocache.New(5*time.Minute, 10*time.Minute)), mlog.NewLogger(nil), &singleflight.Group{})
	cache.SetWithTTL(NewKey("aaa"), []byte("aa"), 100)
	get, _ := cache.(*cacheImpl).store.Get(NewKey("aaa").String())
	assert.Equal(t, "aa", string(get))
}

func TestNewCacheImpl_MemoryDriver(t *testing.T) {
	cfg := &config.Config{CacheDriver: "memory"}
	logger := mlog.NewLogger(nil)
	data := data.NewData(cfg, logger)
	sf := &singleflight.Group{}

	cache := NewCacheImpl(cfg, data, logger, sf)

	_, ok := cache.(*MetricsForCache)
	assert.True(t, ok)
	assert.IsType(t, &cacheImpl{}, cache.(*MetricsForCache).Cache)
	assert.IsType(t, &goCacheAdapter{}, cache.(*MetricsForCache).Cache.(*cacheImpl).store)
}

func TestNewCacheImpl_DbDriver(t *testing.T) {
	cfg := &config.Config{CacheDriver: "db"}
	logger := mlog.NewLogger(nil)
	data := data.NewData(cfg, logger)
	sf := &singleflight.Group{}

	cache := NewCacheImpl(cfg, data, logger, sf)

	_, ok := cache.(*MetricsForCache)
	assert.True(t, ok)
	assert.IsType(t, &cacheImpl{}, cache.(*MetricsForCache).Cache)
	assert.IsType(t, &dbStore{}, cache.(*MetricsForCache).Cache.(*cacheImpl).store)
}

func TestNewCacheImpl_UnknownDriver(t *testing.T) {
	cfg := &config.Config{CacheDriver: "unknown"}
	logger := mlog.NewLogger(nil)
	data := data.NewData(cfg, logger)
	sf := &singleflight.Group{}

	cache := NewCacheImpl(cfg, data, logger, sf)

	_, ok := cache.(*MetricsForCache)
	assert.True(t, ok)
	assert.IsType(t, &NoCache{}, cache.(*MetricsForCache).Cache)
}
