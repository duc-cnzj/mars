package k8s

import (
	"bytes"
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
	restclient "k8s.io/client-go/rest"
)

func TestRemotePathString(t *testing.T) {
	path := NewRemotePath("/test/path")
	assert.Equal(t, "/test/path", path.String())
}

func TestStripTrailingSlash(t *testing.T) {
	assert.Equal(t, "/test/path", stripTrailingSlash("/test/path/"))
	assert.Equal(t, "/test/path", stripTrailingSlash("/test/path"))
	assert.Equal(t, "", stripTrailingSlash(""))
}

func TestNewCopyOptions(t *testing.T) {
	options := NewCopyOptions(mlog.NewForConfig(nil), &restclient.Config{}, fake.NewSimpleClientset(), 10, &bytes.Buffer{})
	assert.NotNil(t, options)
	assert.NotNil(t, options.logger)
	assert.NotNil(t, options.ClientConfig)
	assert.NotNil(t, options.Clientset)
	assert.NotEmpty(t, options.MaxTries)
	assert.NotNil(t, options.errOut)
}
