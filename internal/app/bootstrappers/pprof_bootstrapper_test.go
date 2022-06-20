package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
)

func TestPprofBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().AddServer(gomock.Any()).Times(0)
	app.EXPECT().Config().Times(1).Return(&config.Config{
		ProfileEnabled: false,
	})
	(&PprofBootstrapper{}).Bootstrap(app)
	app.EXPECT().AddServer(gomock.Any()).Times(1)
	app.EXPECT().Config().Times(1).Return(&config.Config{
		ProfileEnabled: true,
	})
	(&PprofBootstrapper{}).Bootstrap(app)
}
