package bootstrappers

import (
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPluginBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mock := app.NewMockApp(m)
	pr := app.NewMockPluginManager(m)
	mock.EXPECT().PluginManager().Return(pr)
	pr.EXPECT().Load(gomock.Any())
	mock.EXPECT().RegisterAfterShutdownFunc(gomock.Any())
	assert.Nil(t, (&PluginBootstrapper{}).Bootstrap(mock))
}

type mockApp struct {
	app.App
	pl app.PluginManager
	cb app.Callback
}

func (a *mockApp) PluginManager() app.PluginManager {
	return a.pl
}
func (a *mockApp) AuthBiz() biz.AuthBiz {
	// PluginBootstrapper 不触达 AuthBiz，仅用于满足 BootstrapDeps 接口。
	return nil
}
func (a *mockApp) ProjectRepo() biz.ProjectRepo {
	// PluginBootstrapper 不触达 ProjectRepo，仅用于满足 BootstrapDeps 接口。
	return nil
}
func (a *mockApp) RegisterAfterShutdownFunc(callback app.Callback) {
	a.cb = callback
}

func TestPluginBootstrapper_Bootstrap_DestroyOnShutdown(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	manager := app.NewMockPluginManager(m)
	a := &mockApp{pl: manager}
	manager.EXPECT().Load(a).Return(nil)
	manager.EXPECT().Destroy()
	assert.Nil(t, (&PluginBootstrapper{}).Bootstrap(a))
	a.cb()
}
func TestPluginBootstrapper_Bootstrap_LoadError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	manager := app.NewMockPluginManager(m)
	a := &mockApp{pl: manager}
	manager.EXPECT().Load(a).Return(errors.New("x"))
	assert.Error(t, (&PluginBootstrapper{}).Bootstrap(a))
}

func TestPluginBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&PluginBootstrapper{}).Tags())
}
