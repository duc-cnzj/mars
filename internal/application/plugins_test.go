package application_test

import (
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type testPlugin struct {
	application.GitServer
	application.WsSender
	application.Picture
	application.DomainManager
}

func (tp *testPlugin) Initialize(app application.App, args map[string]any) error {
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

	application.RegisterPlugin("test", &testPlugin{})

	app := application.NewMockApp(m)
	cfg := &config.Config{
		DomainManagerPlugin: config.Plugin{Name: "test"},
		WsSenderPlugin:      config.Plugin{Name: "test"},
		GitServerPlugin:     config.Plugin{Name: "test"},
		PicturePlugin:       config.Plugin{Name: "test"},
	}

	logger := mlog.NewLogger(nil)

	manager, err := application.NewPluginManager(cfg, logger)
	assert.NoError(t, err)

	err = manager.Load(app)
	assert.NoError(t, err)

	assert.NotNil(t, manager.Domain())
	assert.NotNil(t, manager.Ws())
	assert.NotNil(t, manager.Git())
	assert.NotNil(t, manager.Picture())
}
