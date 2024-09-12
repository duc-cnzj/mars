package bootstrappers

import (
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPluginBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := application.NewMockApp(m)
	pr := application.NewMockPluginManger(m)
	app.EXPECT().PluginMgr().Return(pr)
	pr.EXPECT().Load(gomock.Any())
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any())
	assert.Nil(t, (&PluginBootstrapper{}).Bootstrap(app))
}

type mockApp struct {
	application.App
	pl application.PluginManger
	cb application.Callback
}

func (a *mockApp) PluginMgr() application.PluginManger {
	return a.pl
}
func (a *mockApp) RegisterAfterShutdownFunc(callback application.Callback) {
	a.cb = callback
}

func TestPluginBootstrapper_Bootstrap2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	manger := application.NewMockPluginManger(m)
	a := &mockApp{pl: manger}
	manger.EXPECT().Load(a).Return(nil)
	assert.Nil(t, (&PluginBootstrapper{}).Bootstrap(a))
	git := application.NewMockGitServer(m)
	git.EXPECT().Destroy()
	manger.EXPECT().Git().Return(git)
	do := application.NewMockDomainManager(m)
	do.EXPECT().Destroy()
	picture := application.NewMockPicture(m)
	picture.EXPECT().Destroy()
	sender := application.NewMockWsSender(m)
	sender.EXPECT().Destroy()
	manger.EXPECT().Picture().Return(picture)
	manger.EXPECT().Domain().Return(do)
	manger.EXPECT().Ws().Return(sender)
	a.cb(a)
}
func TestPluginBootstrapper_Bootstrap3(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	manger := application.NewMockPluginManger(m)
	a := &mockApp{pl: manger}
	manger.EXPECT().Load(a).Return(errors.New("x"))
	assert.Error(t, (&PluginBootstrapper{}).Bootstrap(a))
}

func TestPluginBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&PluginBootstrapper{}).Tags())
}
