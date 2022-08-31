package cache

import (
	"time"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsForCache struct {
	Cache contracts.CacheInterface
}

func NewMetricsForCache(c contracts.CacheInterface) *MetricsForCache {
	return &MetricsForCache{Cache: c}
}

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

func (m *MetricsForCache) Clear(key contracts.CacheKeyInterface) error {
	return m.Cache.Clear(key)
}
