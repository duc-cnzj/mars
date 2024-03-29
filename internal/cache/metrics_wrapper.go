package cache

import (
	"time"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsForCache struct {
	Cache contracts.CacheInterface
}

// NewMetricsForCache impl contracts.CacheInterface
func NewMetricsForCache(c contracts.CacheInterface) contracts.CacheInterface {
	return &MetricsForCache{Cache: c}
}

// Remember TODO.
func (m *MetricsForCache) Remember(key contracts.CacheKeyInterface, seconds int, fn func() ([]byte, error)) ([]byte, error) {
	labels := prometheus.Labels{"key": key.Slug()}
	defer func(t time.Time) {
		metrics.CacheRememberDuration.With(labels).Observe(time.Since(t).Seconds())
	}(time.Now())
	bytes, err := m.Cache.Remember(key, seconds, fn)
	if err == nil {
		metrics.CacheBytesGauge.With(labels).Set(float64(len(bytes)))
	}

	return bytes, err
}

// Clear TODO.
func (m *MetricsForCache) Clear(key contracts.CacheKeyInterface) error {
	return m.Cache.Clear(key)
}

// Store TODO.
func (m *MetricsForCache) Store() contracts.Store {
	return m.Cache.Store()
}

// SetWithTTL TODO.
func (m *MetricsForCache) SetWithTTL(key contracts.CacheKeyInterface, value []byte, seconds int) error {
	return m.Cache.SetWithTTL(key, value, seconds)
}
