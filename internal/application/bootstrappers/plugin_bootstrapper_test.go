package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/application"
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

func TestPluginBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&PluginBootstrapper{}).Tags())
}
