package cache

import (
	"fmt"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
)

var _ contracts.CacheKeyInterface = (*Key)(nil)

type Key struct {
	slug string
	key  string
}

// NewKey impl contracts.CacheKeyInterface
func NewKey(slug string, vals ...any) *Key {
	return &Key{slug: slug, key: fmt.Sprintf(slug, vals...)}
}

// String key string: key-1, key-2
func (c *Key) String() string {
	return c.key
}

// Slug key slug: key-1 => key, key-2 => key
func (c *Key) Slug() string {
	return c.slug
}
