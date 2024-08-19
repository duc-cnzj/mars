package cache

import (
	"testing"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
)

func TestNewGoCacheAdapter(t *testing.T) {
	assert.Implements(t, (*Store)(nil), NewGoCacheAdapter(nil))
}

func TestGoCacheAdapter_Get_Set_Delete(t *testing.T) {
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
	assert.Nil(t, adapter.Set("bbb", []byte("bbb"), 100))
	assert.Nil(t, adapter.Delete("bbb"))
	_, err = adapter.Get("bbb")
	assert.Equal(t, "key bbb not found", err.Error())
}
