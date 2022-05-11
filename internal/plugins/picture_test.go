package plugins

import (
	"sync"
	"testing"

	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/stretchr/testify/assert"
)

type testPicture struct {
	PictureInterface
	called bool
}

func (t *testPicture) Initialize(args map[string]any) error {
	t.called = true
	return nil
}

func TestGetPicture(t *testing.T) {
	p := &testPicture{}
	ma := &mockApp{
		p: map[string]contracts.PluginInterface{"picture": p},
	}
	instance.SetInstance(ma)
	pictureOnce = sync.Once{}
	GetPicture()
	assert.Equal(t, 1, ma.callback)
	assert.True(t, p.called)
}
