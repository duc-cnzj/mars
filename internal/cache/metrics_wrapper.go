package cache

import (
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsForCache struct {
	c contracts.CacheInterface
}

func NewMetricsForCache(c contracts.CacheInterface) *MetricsForCache {
	return &MetricsForCache{c: c}
}

func (m *MetricsForCache) Remember(key string, seconds int, fn func() ([]byte, error)) ([]byte, error) {
	bytes, err := m.c.Remember(key, seconds, fn)
	if err == nil {
		metrics.CacheBytesGauge.With(prometheus.Labels{"key": key}).Set(float64(len(bytes)))
	}

	return bytes, err
}

func (m *MetricsForCache) Clear(key string) error {
	return m.c.Clear(key)
}
