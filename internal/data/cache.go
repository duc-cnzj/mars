package data

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/dbcache"
	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/singleflight"
)

// store 是缓存后端的统一存取接口：cacheImpl 依赖它把"内存/DB"两种后端隔离开。
type store interface {
	// Get 读取键对应的缓存值。
	Get(key string) (value []byte, err error)
	// Set 写入缓存值并指定过期秒数。
	Set(key string, value []byte, expireSeconds int) (err error)
	// Delete 删除缓存键。
	Delete(key string) error
}

// CacheKey 是缓存的键抽象：String 用于存取，Slug 用于指标分组。
type CacheKey interface {
	// String 返回缓存存取用的具体键。
	String() string
	// Slug 返回缓存指标的固定分组名。
	Slug() string
}

// Cache 是缓存门面：Remember 带 singleflight 合并与过期读取，SetWithTTL/Clear 直通 store 后端。
type Cache interface {
	// SetWithTTL 直接写缓存值并指定过期秒数。
	SetWithTTL(key CacheKey, value []byte, seconds int) error
	// Remember 读取缓存，未命中/过期时执行 fn 回填（带 singleflight 合并）。
	Remember(key CacheKey, seconds int, fn func() ([]byte, error), force bool) ([]byte, error)
	// Clear 删除缓存键。
	Clear(key CacheKey) error
}

// Key 是 CacheKey 的默认实现：slug 为指标分组名，key 为 fmt.Sprintf(slug, vals...) 拼出的具体键。
type Key struct {
	slug string
	key  string
}

// NewKey 构造缓存键：slug 是格式串（如 "all_projects" 或 "all_branches_%d"），vals 按序填充。
func NewKey(slug string, vals ...any) *Key {
	return &Key{slug: slug, key: fmt.Sprintf(slug, vals...)}
}

// String 返回缓存存取用的具体键。
func (c *Key) String() string {
	return c.key
}

// Slug 返回缓存指标的固定分组名。
func (c *Key) Slug() string {
	return c.slug
}

// noCache 是 Cache 的空实现（未配置缓存驱动时的兜底）：Remember 直接执行 fn，其余方法空转。
type noCache struct{}

// Remember 不缓存，直接调用 fn。
func (n *noCache) Remember(key CacheKey, seconds int, fn func() ([]byte, error), force bool) ([]byte, error) {
	return fn()
}

// Clear 空实现。
func (n *noCache) Clear(key CacheKey) error {
	return nil
}

// SetWithTTL 空实现。
func (n *noCache) SetWithTTL(key CacheKey, value []byte, seconds int) error {
	return nil
}

// metricsForCache 是 Cache 的指标装饰器：Remember 观测耗时，命中后记录字节数。
type metricsForCache struct {
	cache Cache
}

// newMetricsForCache 用指标装饰器包一层 Cache。
func newMetricsForCache(c Cache) Cache {
	return &metricsForCache{cache: c}
}

// Remember 观测缓存读耗时，成功后记录字节数。
func (m *metricsForCache) Remember(key CacheKey, seconds int, fn func() ([]byte, error), force bool) ([]byte, error) {
	labels := prometheus.Labels{"key": key.Slug()}
	defer func(t time.Time) {
		metrics.CacheRememberDuration.With(labels).Observe(time.Since(t).Seconds())
	}(time.Now())
	bytes, err := m.cache.Remember(key, seconds, fn, force)
	if err == nil {
		metrics.CacheBytesGauge.With(labels).Set(float64(len(bytes)))
	}

	return bytes, err
}

// Clear 委托内部 Cache。
func (m *metricsForCache) Clear(key CacheKey) error {
	return m.cache.Clear(key)
}

// SetWithTTL 委托内部 Cache。
func (m *metricsForCache) SetWithTTL(key CacheKey, value []byte, seconds int) error {
	return m.cache.SetWithTTL(key, value, seconds)
}

// cacheImpl 是 Cache 的生产实现：基于 store 后端 + singleflight 合并重复读。
type cacheImpl struct {
	store  store
	sf     *singleflight.Group
	logger mlog.Logger
}

// NewCacheImpl 按配置的缓存驱动构造 Cache：memory 用 GoCache、db 用 newCacheDBStore、
// 其他值回落 noCache，最后统一套指标装饰器。singleflight 是本包缓存实现的
// 私有合并原语，在此内聚自建，无需 wire 注入。db 驱动只需 DB 取数，故注入
// DBGetter 窄端口而非完整 dataStore（ISP：cache 域不摸 k8s 客户端）。
func NewCacheImpl(cfg *config.Config, d DBGetter, logger mlog.Logger) Cache {
	logger = logger.WithModule("cache/cacheImpl")
	sf := &singleflight.Group{}

	var ca Cache
	switch cfg.CacheDriver {
	case "memory":
		ca = newCache(
			newGoCacheAdapter(
				gocache.New(5*time.Minute, 10*time.Minute),
			),
			logger,
			sf,
		)
	case "db":
		ca = newCache(newCacheDBStore(d), logger, sf)
	default:
		ca = &noCache{}
	}
	return newMetricsForCache(ca)
}

// newCache 构造 cacheImpl。
func newCache(store store, logger mlog.Logger, sf *singleflight.Group) Cache {
	return &cacheImpl{store: store, sf: sf, logger: logger}
}

// Remember 带 singleflight 合并读缓存；seconds<=0 或未命中时执行 fn 并回写。
func (c *cacheImpl) Remember(key CacheKey, seconds int, fn func() ([]byte, error), force bool) ([]byte, error) {
	sfKey := fmt.Sprintf("CacheRemember:%v-%v", key.String(), force)
	do, err, _ := c.sf.Do(sfKey, func() (any, error) {
		if seconds <= 0 {
			return fn()
		}

		res, err := c.store.Get(key.String())
		c.logger.Debugf("CacheRemember: %s, from cacheImpl: %t", key, err == nil)
		if err == nil && !force {
			return res, nil
		}
		res, err = fn()
		if err != nil {
			return nil, err
		}
		// 缓存写回是尽力而为：fn() 已算出值，写缓存失败不应让整个请求失败，
		// 故刻意吞错不冒泡。该错误无法上抛给调用方，按日志规范"必须保留的
		// 例外"原地打印一条，否则缓存写失败将完全无痕。
		err = c.SetWithTTL(key, res, seconds)
		if err != nil {
			c.logger.Errorf("[CACHE WRITE FAILED]: key %s err %v", key, err)
		}
		return res, nil
	})

	if err != nil {
		return nil, err
	}

	return do.([]byte), err
}

// SetWithTTL 写缓存并指定过期秒数。
func (c *cacheImpl) SetWithTTL(key CacheKey, value []byte, seconds int) error {
	return c.store.Set(key.String(), value, seconds)
}

// Clear 删除缓存键。
func (c *cacheImpl) Clear(key CacheKey) error {
	return c.store.Delete(key.String())
}

// cacheDBStore 是 store 的 DB 后端实现：缓存落 dbcache 表，value 经 base64 编码，
// 过期时间落库查询时过滤。命名带 cache 前缀以与 dataStore（repo 存储端口）区分。
type cacheDBStore struct {
	d DBGetter
}

// newCacheDBStore 构造基于数据库的缓存后端（NewCacheImpl 的 "db" 驱动使用）。
// 只注入 DBGetter 窄端口：DB 后端只需 DB()，不持有 dataStore 全貌。
func newCacheDBStore(d DBGetter) store {
	return &cacheDBStore{d: d}
}

// Get 读取未过期（ExpiredAt >= 当前时间）的缓存值并 base64 解码返回；无记录时返回底层错误。
func (d *cacheDBStore) Get(key string) (value []byte, err error) {
	first, err := d.d.DB().DBCache.Query().Where(dbcache.Key(key), dbcache.ExpiredAtGTE(time.Now())).Only(context.TODO())
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(first.Value)
}

// Set 写入缓存：value 先 base64 编码，已存在时按 key 冲突更新 value 与过期时间（upsert）。
func (d *cacheDBStore) Set(key string, value []byte, seconds int) (err error) {
	toString := base64.StdEncoding.EncodeToString(value)

	return d.d.DB().DBCache.Create().
		SetKey(key).
		SetValue(toString).
		SetExpiredAt(time.Now().Add(time.Duration(seconds) * time.Second)).
		OnConflict().
		UpdateValue().
		UpdateExpiredAt().
		Exec(context.TODO())
}

// Delete 删除指定缓存键，键不存在时也返回 nil（ent Delete 对空结果不报错）。
func (d *cacheDBStore) Delete(key string) error {
	_, err := d.d.DB().DBCache.Delete().Where(dbcache.Key(key)).Exec(context.TODO())
	return err
}

// goCacheAdapter 是 store 的 memory 后端实现：包装 patrickmn/go-cache 内存缓存。
type goCacheAdapter struct {
	c *gocache.Cache
}

// newGoCacheAdapter 构造基于内存缓存的后端（NewCacheImpl 的 "memory" 驱动使用）。
func newGoCacheAdapter(c *gocache.Cache) store {
	return &goCacheAdapter{c: c}
}

// Get 读取缓存值，未命中返回带键的错误。
func (g *goCacheAdapter) Get(key string) (value []byte, err error) {
	v, b := g.c.Get(key)
	if !b {
		return nil, fmt.Errorf("key %s not found", key)
	}
	return v.([]byte), nil
}

// Set 写入缓存并指定过期秒数。
func (g *goCacheAdapter) Set(key string, value []byte, expireSeconds int) (err error) {
	g.c.Set(key, value, time.Second*time.Duration(expireSeconds))
	return nil
}

// Delete 删除缓存键。
func (g *goCacheAdapter) Delete(key string) error {
	g.c.Delete(key)
	return nil
}
