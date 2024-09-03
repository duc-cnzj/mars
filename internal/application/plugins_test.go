package application

import (
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type testPlugin struct {
	GitServer
	WsSender
	Picture
	DomainManager
}

func (tp *testPlugin) Initialize(app App, args map[string]any) error {
	return nil
}

func (tp *testPlugin) Destroy() error {
	return nil
}

func (tp *testPlugin) Name() string {
	return "test"
}

func TestPluginManagerLoad(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	RegisterPlugin("test", &testPlugin{})

	app := NewMockApp(m)
	cfg := &config.Config{
		DomainManagerPlugin: config.Plugin{Name: "test"},
		WsSenderPlugin:      config.Plugin{Name: "test"},
		GitServerPlugin:     config.Plugin{Name: "test"},
		PicturePlugin:       config.Plugin{Name: "test"},
	}

	logger := mlog.NewForConfig(nil)

	manager, err := NewPluginManager(cfg, logger)
	assert.NoError(t, err)

	err = manager.Load(app)
	assert.NoError(t, err)

	assert.NotNil(t, manager.Domain())
	assert.NotNil(t, manager.Ws())
	assert.NotNil(t, manager.Git())
	assert.NotNil(t, manager.Picture())
}

func TestGetPlugins(t *testing.T) {
	assert.Equal(t, pluginSet, (&manager{}).GetPlugins())
	assert.NotNil(t, (&manager{}).GetPlugins())
}
