package uploader

import (
	"sync"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewCacheUploader(t *testing.T) {
	uploader := NewCacheUploader(nil, nil)
	assert.IsType(t, &cacheUploader{}, uploader)
}

func Test_byteNum(t *testing.T) {
	assert.Equal(t, int64(10), byteNum([]byte("10")))
	assert.Equal(t, int64(20), byteNum([]byte("20")))
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
	assert.Equal(t, 60, ca.seconds)
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

func Test_toByteNum(t *testing.T) {
	assert.Equal(t, []byte("10"), toByteNum(10))
}
