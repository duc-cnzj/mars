package bootstrappers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
)

func TestPluginsBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().SetPlugins(gomock.Any()).Times(1)
	(&PluginsBootstrapper{}).Bootstrap(app)
}

func TestPluginsBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&PluginsBootstrapper{}).Tags())
}
