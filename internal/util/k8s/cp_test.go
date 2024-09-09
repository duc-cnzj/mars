package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemotePathString(t *testing.T) {
	path := NewRemotePath("/test/path")
	assert.Equal(t, "/test/path", path.String())
}

func TestStripTrailingSlash(t *testing.T) {
	assert.Equal(t, "/test/path", stripTrailingSlash("/test/path/"))
	assert.Equal(t, "/test/path", stripTrailingSlash("/test/path"))
}
