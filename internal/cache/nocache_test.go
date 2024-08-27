package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache_RememberNoCache(t *testing.T) {
	var i int
	cache := &NoCache{}
	fn := func() {
		cache.Remember(NewKey("duc"), 10, func() ([]byte, error) {
			i++
			return []byte("duccc"), nil
		}, false)
	}
	fn()
	fn()
	fn()
	assert.Equal(t, 3, i)
}

func TestNoCache_Clear(t *testing.T) {
	cache := &NoCache{}
	assert.Nil(t, cache.Clear(NewKey("aaa")))
}

func TestNoCache_SetWithTTL(t *testing.T) {
	cache := &NoCache{}
	assert.Nil(t, cache.SetWithTTL(NewKey("aaa"), []byte("x"), 1))
}

func TestNoCache_Store(t *testing.T) {
	assert.Nil(t, (&NoCache{}).Store())
}
