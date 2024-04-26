package cache

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMetricsForCache_Clear(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	c := mock.NewMockCacheInterface(m)
	c.EXPECT().Clear(NewKey("a")).Times(1)
	mc := &MetricsForCache{Cache: c}
	mc.Clear(NewKey("a"))
}

func TestMetricsForCache_Remember(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	c := mock.NewMockCacheInterface(m)
	bytesRet := []byte{'a', 'b'}
	fn := func() ([]byte, error) {
		return bytesRet, nil
	}
	c.EXPECT().Remember(NewKey("a"), int(10), gomock.Any()).Times(1).Return(bytesRet, nil)
	mc := &MetricsForCache{Cache: c}
	remember, err := mc.Remember(NewKey("a"), 10, fn)
	assert.Equal(t, bytesRet, remember)
	assert.Nil(t, err)
}

func TestNewMetricsForCache(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	c := mock.NewMockCacheInterface(m)
	cache := NewMetricsForCache(c)
	assert.Equal(t, c, cache.(*MetricsForCache).Cache)
	assert.Implements(t, (*contracts.CacheInterface)(nil), cache)
}

func TestMetricsForCache_SetWithTTL(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	c := mock.NewMockCacheInterface(m)
	c.EXPECT().SetWithTTL(NewKey("a"), []byte("aaa"), int(1)).Times(1)
	mc := &MetricsForCache{Cache: c}
	mc.SetWithTTL(NewKey("a"), []byte("aaa"), int(1))
}

type mockStore struct {
	contracts.Store
}

func TestMetricsForCache_Store(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	c := mock.NewMockCacheInterface(m)
	c.EXPECT().Store().Return(&mockStore{}).Times(2)
	mc := &MetricsForCache{Cache: c}
	assert.Same(t, c.Store(), mc.Store())
}
