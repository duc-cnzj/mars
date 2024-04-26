package uploader

import (
	"sync"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewCacheUploader(t *testing.T) {
	uploader := NewCacheUploader(nil, nil)
	assert.IsType(t, &cacheUploader{}, uploader)
}

func Test_byteToInt64(t *testing.T) {
	assert.Equal(t, int64(10), byteToInt64([]byte("10")))
	assert.Equal(t, int64(20), byteToInt64([]byte("20")))
}

type cacheMock struct {
	key     contracts.CacheKeyInterface
	seconds int
	contracts.CacheInterface
	sync.Once
}

func (c *cacheMock) Remember(key contracts.CacheKeyInterface, seconds int, fn func() ([]byte, error)) ([]byte, error) {
	c.key = key
	c.seconds = seconds
	return fn()
}

func Test_cacheUploader_DirSize(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	up := mock.NewMockUploader(m)
	up.EXPECT().DirSize().Return(int64(10), nil)
	ca := &cacheMock{}
	size, err := (&cacheUploader{cacheFn: func() contracts.CacheInterface {
		return ca
	}, Uploader: up}).DirSize()
	assert.Equal(t, int64(10), size)
	assert.Nil(t, err)
	assert.Equal(t, 60*15, ca.seconds)
	assert.Equal(t, "dir-size", ca.key.String())
}

func Test_cacheUploader_UnWrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	c := &cacheUploader{}
	assert.Nil(t, c.UnWrap())
	up := mock.NewMockUploader(m)
	c.Uploader = up
	assert.Same(t, up, c.UnWrap())
}

func Test_int64ToByte(t *testing.T) {
	assert.Equal(t, []byte("10"), int64ToByte(10))
}
