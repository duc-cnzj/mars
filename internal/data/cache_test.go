package data

import (
	"context"
	"encoding/base64"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/dbcache"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/sync/singleflight"
)

func TestCache_Remember(t *testing.T) {
	var i int
	cache := newCache(newGoCacheAdapter(gocache.New(5*time.Minute, 10*time.Minute)), mlog.NewForConfig(nil), &singleflight.Group{})
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
	cache := newCache(newGoCacheAdapter(gocache.New(5*time.Minute, 10*time.Minute)), mlog.NewForConfig(nil), &singleflight.Group{})
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

var errStore = errors.New("store error")

func (e *errorstore) Set(key string, value []byte, expireSeconds int) (err error) {
	return errStore
}

func (e *errorstore) Get(key string) (value []byte, err error) {
	return nil, errors.New("errorstore get err")
}

func (e *errorstore) Delete(key string) error {
	return nil
}

func TestCache_RememberErrorStore(t *testing.T) {
	var i int
	cache := newCache(&errorstore{}, mlog.NewForConfig(nil), &singleflight.Group{})
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
	cache := newCache(newGoCacheAdapter(gocache.New(5*time.Minute, 10*time.Minute)), mlog.NewForConfig(nil), &singleflight.Group{})
	called := 0
	fn := func() ([]byte, error) {
		called++
		return []byte("aaa"), nil
	}
	cache.Remember(NewKey("aaa"), 100, fn, false)
	cache.Remember(NewKey("aaa"), 100, fn, false)
	assert.Nil(t, cache.Clear(NewKey("aaa")))
	cache.Remember(NewKey("aaa"), 100, fn, false)
	assert.Equal(t, 2, called)
	cache.Remember(NewKey("aaa"), 100, fn, false)
	assert.Equal(t, 2, called)
	cache.Remember(NewKey("aaa"), 100, fn, true)
	assert.Equal(t, 3, called)
}

func TestCache_SetWithTTL(t *testing.T) {
	cache := newCache(newGoCacheAdapter(gocache.New(5*time.Minute, 10*time.Minute)), mlog.NewForConfig(nil), &singleflight.Group{})
	cache.SetWithTTL(NewKey("aaa"), []byte("aa"), 100)
	get, _ := cache.(*cacheImpl).store.Get(NewKey("aaa").String())
	assert.Equal(t, "aa", string(get))
}

func TestNewCacheImpl_MemoryDriver(t *testing.T) {
	cfg := &config.Config{CacheDriver: "memory"}
	logger := mlog.NewForConfig(nil)
	d := NewData(cfg, logger)

	cache := NewCacheImpl(cfg, d, logger)

	_, ok := cache.(*metricsForCache)
	assert.True(t, ok)
	assert.IsType(t, &cacheImpl{}, cache.(*metricsForCache).cache)
	assert.IsType(t, &goCacheAdapter{}, cache.(*metricsForCache).cache.(*cacheImpl).store)
}

func TestNewCacheImpl_DbDriver(t *testing.T) {
	cfg := &config.Config{CacheDriver: "db"}
	logger := mlog.NewForConfig(nil)
	d := NewData(cfg, logger)

	cache := NewCacheImpl(cfg, d, logger)

	_, ok := cache.(*metricsForCache)
	assert.True(t, ok)
	assert.IsType(t, &cacheImpl{}, cache.(*metricsForCache).cache)
	assert.IsType(t, &cacheDBStore{}, cache.(*metricsForCache).cache.(*cacheImpl).store)
}

func TestNewCacheImpl_UnknownDriver(t *testing.T) {
	cfg := &config.Config{CacheDriver: "unknown"}
	logger := mlog.NewForConfig(nil)
	d := NewData(cfg, logger)

	cache := NewCacheImpl(cfg, d, logger)

	_, ok := cache.(*metricsForCache)
	assert.True(t, ok)
	assert.IsType(t, &noCache{}, cache.(*metricsForCache).cache)
}

// TestCache_RememberNoCache 覆盖 noCache 后端 Remember 每次直调 fn 不透传缓存。
func TestCache_RememberNoCache(t *testing.T) {
	var i int
	cache := &noCache{}
	fn := func() {
		cache.Remember(NewKey("duc"), 10, func() ([]byte, error) {
			i++
			return []byte("duccc"), nil
		}, false)
	}
	fn()
	fn()
	fn()
	assert.Equal(t, 3, i)
}

// TestNoCache_Clear 覆盖 noCache 清空恒 nil。
func TestNoCache_Clear(t *testing.T) {
	cache := &noCache{}
	assert.Nil(t, cache.Clear(NewKey("aaa")))
}

// TestNoCache_SetWithTTL 覆盖 noCache 写入恒 nil。
func TestNoCache_SetWithTTL(t *testing.T) {
	cache := &noCache{}
	assert.Nil(t, cache.SetWithTTL(NewKey("aaa"), []byte("x"), 1))
}

// TestDBStore_Get 覆盖 DB 后端的读：无记录返回底层错误；有未过期记录时 base64 解码返回。
func TestDBStore_Get(t *testing.T) {
	sqliteDB, _ := NewSqliteDB()
	defer sqliteDB.Close()

	s := &cacheDBStore{
		d: NewDataImpl(&NewDataParams{DB: sqliteDB}),
	}
	_, err := s.Get("test")
	assert.Error(t, err)

	sqliteDB.DBCache.Create().SetKey("test").SetValue("test").SetExpiredAt(time.Now().Add(10 * time.Second)).Exec(context.TODO())

	v, err := s.Get("test")
	assert.Nil(t, err)
	decodeString, _ := base64.StdEncoding.DecodeString("test")
	assert.Equal(t, decodeString, v)
}

// TestDBStore_Set 覆盖 DB 后端的写入与 upsert：二次 Set 同键更新 value 且刷新过期时间。
func TestDBStore_Set(t *testing.T) {
	sqliteDB, _ := NewSqliteDB()
	defer sqliteDB.Close()

	s := &cacheDBStore{
		d: NewDataImpl(&NewDataParams{DB: sqliteDB}),
	}
	err := s.Set("test", []byte("test"), 10)
	assert.Nil(t, err)
	only1, _ := sqliteDB.DBCache.Query().Where(dbcache.Key("test")).Only(context.TODO())
	assert.Equal(t, "test", b64ToStr(only1.Value))
	err = s.Set("test", []byte("testxxx"), 10)
	assert.Nil(t, err)
	only2, _ := sqliteDB.DBCache.Query().Where(dbcache.Key("test")).Only(context.TODO())
	assert.Equal(t, "testxxx", b64ToStr(only2.Value))
	assert.NotEqual(t, only1.ExpiredAt, only2.ExpiredAt)
}

// TestDBStore_Delete 覆盖 DB 后端的删除：删除后查询返回错误；不存在时删除也返回 nil。
func TestDBStore_Delete(t *testing.T) {
	sqliteDB, _ := NewSqliteDB()
	defer sqliteDB.Close()

	s := &cacheDBStore{
		d: NewDataImpl(&NewDataParams{DB: sqliteDB}),
	}
	err := s.Delete("test")
	assert.Nil(t, err)
	sqliteDB.DBCache.Create().SetKey("test").SetValue("test").SetExpiredAt(time.Now().Add(10 * time.Second)).Exec(context.TODO())
	err = s.Delete("test")
	assert.Nil(t, err)
	_, err = sqliteDB.DBCache.Query().Where(dbcache.Key("test")).Only(context.TODO())
	assert.Error(t, err)
}

// b64ToStr 测试辅助：把库里的 base64 编码值解码回字符串做断言。
func b64ToStr(v string) string {
	decodeString, _ := base64.StdEncoding.DecodeString(v)
	return string(decodeString)
}

// TestNewGoCacheAdapter 覆盖 memory 后端实现 store 接口。
func TestNewGoCacheAdapter(t *testing.T) {
	assert.Implements(t, (*store)(nil), newGoCacheAdapter(nil))
}

// TestGoCacheAdapter_Get_Set_Delete 覆盖 memory 后端的读写删：TTL 过期后 Get 返回 not found。
func TestGoCacheAdapter_Get_Set_Delete(t *testing.T) {
	adapter := newGoCacheAdapter(gocache.New(1*time.Minute, 10*time.Minute))
	_, err := adapter.Get("aaa")
	assert.Equal(t, "key aaa not found", err.Error())
	assert.Nil(t, adapter.Set("aaa", []byte("aaa"), 1))
	v, err := adapter.Get("aaa")
	assert.Nil(t, err)
	assert.Equal(t, "aaa", string(v))
	// Set 的 TTL 为 1 秒，等待超过 TTL 触发懒过期。
	time.Sleep(1100 * time.Millisecond)
	_, err = adapter.Get("aaa")
	assert.Equal(t, "key aaa not found", err.Error())
	assert.Nil(t, adapter.Set("bbb", []byte("bbb"), 100))
	assert.Nil(t, adapter.Delete("bbb"))
	_, err = adapter.Get("bbb")
	assert.Equal(t, "key bbb not found", err.Error())
}

func TestMetricsForCache_Clear(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	c := NewMockCache(m)
	c.EXPECT().Clear(NewKey("a")).Times(1)
	mc := &metricsForCache{cache: c}
	mc.Clear(NewKey("a"))
}

func TestMetricsForCache_Remember(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	c := NewMockCache(m)
	bytesRet := []byte{'a', 'b'}
	fn := func() ([]byte, error) {
		return bytesRet, nil
	}
	c.EXPECT().Remember(NewKey("a"), int(10), gomock.Any(), false).Times(1).Return(bytesRet, nil)
	mc := &metricsForCache{cache: c}
	remember, err := mc.Remember(NewKey("a"), 10, fn, false)
	assert.Equal(t, bytesRet, remember)
	assert.Nil(t, err)
}

func TestNewMetricsForCache(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	c := NewMockCache(m)
	cache := newMetricsForCache(c)
	assert.Equal(t, c, cache.(*metricsForCache).cache)
	assert.Implements(t, (*Cache)(nil), cache)
}

func TestMetricsForCache_SetWithTTL(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	c := NewMockCache(m)
	c.EXPECT().SetWithTTL(NewKey("a"), []byte("aaa"), int(1)).Times(1)
	mc := &metricsForCache{cache: c}
	mc.SetWithTTL(NewKey("a"), []byte("aaa"), int(1))
}
