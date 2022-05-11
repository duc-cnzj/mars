package adapter

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/cache"
	"github.com/stretchr/testify/assert"
)

func TestNewGoCacheAdapter(t *testing.T) {
	assert.Implements(t, (*cache.Store)(nil), NewGoCacheAdapter(nil))
}
