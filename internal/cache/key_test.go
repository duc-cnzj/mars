package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheKey_Slug(t *testing.T) {
	assert.Equal(t, "a-%s", NewKey("a-%s", "b").Slug())
}

func TestCacheKey_String(t *testing.T) {
	assert.Equal(t, "a-b", NewKey("a-%s", "b").String())
}

func TestNewCacheKey(t *testing.T) {
	ck := NewKey("a-%s", "b")
	assert.Equal(t, "a-b", ck.key)
	assert.Equal(t, "a-%s", ck.slug)
}
