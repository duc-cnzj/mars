package cache

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMetricsForCache_Clear(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	c := mock.NewMockCacheInterface(m)
	c.EXPECT().Clear("a").Times(1)
	mc := &MetricsForCache{Cache: c}
	mc.Clear("a")
}

func TestMetricsForCache_Remember(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	c := mock.NewMockCacheInterface(m)
	bytesRet := []byte{'a', 'b'}
	fn := func() ([]byte, error) {
		return bytesRet, nil
	}
	c.EXPECT().Remember("a", int(10), gomock.Any()).Times(1).Return(bytesRet, nil)
	mc := &MetricsForCache{Cache: c}
	remember, err := mc.Remember("a", 10, fn)
	assert.Equal(t, bytesRet, remember)
	assert.Nil(t, err)
}

func TestNewMetricsForCache(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	c := mock.NewMockCacheInterface(m)
	cache := NewMetricsForCache(c)
	assert.Equal(t, c, cache.Cache)
	assert.Implements(t, (*contracts.CacheInterface)(nil), cache)
}