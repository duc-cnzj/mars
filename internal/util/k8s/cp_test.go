package k8s

import (
	"context"
	"os"
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestTarPipe_Read_Failure(t *testing.T) {
	ctx := context.Background()
	src := CopyFileSpec{
		PodName:       "test-pod",
		PodNamespace:  "test-namespace",
		ContainerName: "test-container",
		File:          NewLocalPath("/test/path"),
	}
	copyOptions := NewCopyOptions(mlog.NewForConfig(nil), &rest.Config{}, fake.NewSimpleClientset(), 0, os.Stdout)
	tarPipe := newTarPipe(ctx, src, copyOptions)

	tarPipe.outStream.Write([]byte("test-data"))
	data := make([]byte, 10)
	n, err := tarPipe.Read(data)

	assert.Error(t, err)
	assert.Equal(t, 0, n)
}

func TestLocalPathString(t *testing.T) {
	path := NewLocalPath("/test/path")
	assert.Equal(t, "/test/path", path.String())
}

func TestLocalPathDir(t *testing.T) {
	path := NewLocalPath("/test/path")
	assert.Equal(t, "/test", path.Dir().String())
}

func TestLocalPathBase(t *testing.T) {
	path := NewLocalPath("/test/path")
	assert.Equal(t, "path", path.Base().String())
}

func TestLocalPathClean(t *testing.T) {
	path := NewLocalPath("/test/../path")
	assert.Equal(t, "/path", path.Clean().String())
}

func TestLocalPathJoin(t *testing.T) {
	path := NewLocalPath("/test")
	joinedPath := path.Join(NewLocalPath("path"))
	assert.Equal(t, "/test/path", joinedPath.String())
}

func TestLocalPathStripSlashes(t *testing.T) {
	path := NewLocalPath("/test/path/")
	assert.Equal(t, "test/path", path.StripSlashes().String())
}

func TestRemotePathString(t *testing.T) {
	path := NewRemotePath("/test/path")
	assert.Equal(t, "/test/path", path.String())
}

func TestRemotePathDir(t *testing.T) {
	path := NewRemotePath("/test/path")
	assert.Equal(t, "/test", path.Dir().String())
}

func TestRemotePathBase(t *testing.T) {
	path := NewRemotePath("/test/path")
	assert.Equal(t, "path", path.Base().String())
}

func TestRemotePathClean(t *testing.T) {
	path := NewRemotePath("/test/../path")
	assert.Equal(t, "/path", path.Clean().String())
}

func TestRemotePathJoin(t *testing.T) {
	path := NewRemotePath("/test")
	joinedPath := path.Join(NewRemotePath("path"))
	assert.Equal(t, "/test/path", joinedPath.String())
}

func TestRemotePathStripShortcuts(t *testing.T) {
	path := NewRemotePath("/../test/path")
	assert.Equal(t, "test/path", path.StripShortcuts().String())
}

func TestRemotePathStripSlashes(t *testing.T) {
	path := NewRemotePath("/test/path/")
	assert.Equal(t, "test/path", path.StripSlashes().String())
}

func TestIsRelative(t *testing.T) {
	base := NewLocalPath("/test")
	target := NewLocalPath("/test/path")
	assert.True(t, isRelative(base, target))

	base = NewLocalPath("/test")
	target = NewLocalPath("/another/path")
	assert.False(t, isRelative(base, target))
}

func TestStripTrailingSlash(t *testing.T) {
	assert.Equal(t, "/test/path", stripTrailingSlash("/test/path/"))
	assert.Equal(t, "/test/path", stripTrailingSlash("/test/path"))
}

func TestStripLeadingSlash(t *testing.T) {
	assert.Equal(t, "test/path", stripLeadingSlash("/test/path"))
	assert.Equal(t, "test/path", stripLeadingSlash("test/path"))
}

func TestStripPathShortcuts(t *testing.T) {
	assert.Equal(t, "test/path", stripPathShortcuts("../test/path"))
	assert.Equal(t, "test/path", stripPathShortcuts("test/path"))
}
