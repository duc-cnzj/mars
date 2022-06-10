package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache_RememberNoCache(t *testing.T) {
	var i int
	cache := &NoCache{}
	fn := func() {
		cache.Remember("duc", 10, func() ([]byte, error) {
			i++
			return []byte("duccc"), nil
		})
	}
	fn()
	fn()
	fn()
	assert.Equal(t, 3, i)
}

func TestNoCache_Clear(t *testing.T) {
	cache := &NoCache{}
	assert.Nil(t, cache.Clear("aaa"))
}
