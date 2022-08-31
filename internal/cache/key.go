package cache

import (
	"fmt"

	"github.com/duc-cnzj/mars/internal/contracts"
)

var _ contracts.CacheKeyInterface = (*Key)(nil)

type Key struct {
	slug string
	key  string
}

func NewKey(slug string, vals ...any) *Key {
	return &Key{slug: slug, key: fmt.Sprintf(slug, vals...)}
}

func (c *Key) String() string {
	return c.key
}

func (c *Key) Slug() string {
	return c.slug
}
