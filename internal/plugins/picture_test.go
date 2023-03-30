package plugins

import (
	"errors"
	"sync"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/stretchr/testify/assert"
)

type testPicture struct {
	PictureInterface
	called bool
	err    error
}

func (t *testPicture) Initialize(args map[string]any) error {
	t.called = true
	return t.err
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
func TestGetPicture2(t *testing.T) {
	p := &testPicture{
		err: errors.New("xxx"),
	}
	ma := &mockApp{
		p: map[string]contracts.PluginInterface{"picture": p},
	}
	instance.SetInstance(ma)
	pictureOnce = sync.Once{}
	assert.Panics(t, func() {
		GetPicture()
	})
	assert.Equal(t, 0, ma.callback)
	assert.True(t, p.called)

}
