package cache

import (
	"time"

	"github.com/duc-cnzj/mars/v5/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsForCache struct {
	Cache Cache
}

// newMetricsForCache impl contracts.Cache
func newMetricsForCache(c Cache) Cache {
	return &MetricsForCache{Cache: c}
}

// Remember TODO.
func (m *MetricsForCache) Remember(key CacheKey, seconds int, fn func() ([]byte, error), force bool) ([]byte, error) {
	labels := prometheus.Labels{"key": key.Slug()}
	defer func(t time.Time) {
		metrics.CacheRememberDuration.With(labels).Observe(time.Since(t).Seconds())
	}(time.Now())
	bytes, err := m.Cache.Remember(key, seconds, fn, force)
	if err == nil {
		metrics.CacheBytesGauge.With(labels).Set(float64(len(bytes)))
	}

	return bytes, err
}

// Clear TODO.
func (m *MetricsForCache) Clear(key CacheKey) error {
	return m.Cache.Clear(key)
}

// Store TODO.
func (m *MetricsForCache) Store() Store {
	return m.Cache.Store()
}

// SetWithTTL TODO.
func (m *MetricsForCache) SetWithTTL(key CacheKey, value []byte, seconds int) error {
	return m.Cache.SetWithTTL(key, value, seconds)
}
