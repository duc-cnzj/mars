package plugins

import (
	"sync"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/stretchr/testify/assert"
)

type mockApp struct {
	cached   bool
	cache    contracts.CacheInterface
	p        map[string]contracts.PluginInterface
	callback int
	contracts.ApplicationInterface
}

func (receiver *mockApp) Cache() contracts.CacheInterface {
	return receiver.cache
}

func (receiver *mockApp) Config() *config.Config {
	return &config.Config{
		GitServerCached:     receiver.cached,
		DomainManagerPlugin: config.Plugin{Name: "test"},
		PicturePlugin:       config.Plugin{Name: "picture"},
		WsSenderPlugin:      config.Plugin{Name: "sender"},
		GitServerPlugin:     config.Plugin{Name: "git"},
	}
}

func (receiver *mockApp) GetPluginByName(n string) contracts.PluginInterface {
	return receiver.p[n]
}

func (receiver *mockApp) RegisterAfterShutdownFunc(callback contracts.Callback) {
	receiver.callback++
}

type testDm struct {
	DomainManager
	initialized bool
}

func (t *testDm) Initialize(args map[string]any) error {
	t.initialized = true
	return nil
}

func TestGetDomainManager(t *testing.T) {
	dm := &testDm{}
	ma := &mockApp{
		p: map[string]contracts.PluginInterface{"test": dm},
	}
	instance.SetInstance(ma)
	domainManagerOnce = sync.Once{}
	GetDomainManager()
	assert.Equal(t, 1, ma.callback)
	assert.True(t, dm.initialized)
}
