package adapter

import (
	"testing"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"github.com/duc-cnzj/mars/internal/cache"
)

func TestNewGoCacheAdapter(t *testing.T) {
	assert.Implements(t, (*cache.Store)(nil), NewGoCacheAdapter(nil))
}

func TestGoCacheAdapter_Get_Set(t *testing.T) {
	adapter := NewGoCacheAdapter(gocache.New(1*time.Minute, 10*time.Minute))
	_, err := adapter.Get("aaa")
	assert.Equal(t, "key aaa not found", err.Error())
	assert.Nil(t, adapter.Set("aaa", []byte("aaa"), 1))
	v, err := adapter.Get("aaa")
	assert.Nil(t, err)
	assert.Equal(t, "aaa", string(v))
	time.Sleep(2 * time.Second)
	_, err = adapter.Get("aaa")
	assert.Equal(t, "key aaa not found", err.Error())
}
