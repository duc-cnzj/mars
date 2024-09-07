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

func TestRemotePathString(t *testing.T) {
	path := NewRemotePath("/test/path")
	assert.Equal(t, "/test/path", path.String())
}

func TestStripTrailingSlash(t *testing.T) {
	assert.Equal(t, "/test/path", stripTrailingSlash("/test/path/"))
	assert.Equal(t, "/test/path", stripTrailingSlash("/test/path"))
}
