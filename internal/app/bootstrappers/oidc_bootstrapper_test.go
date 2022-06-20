package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
)

func TestOidcBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().Config().Times(1).Return(&config.Config{})
	app.EXPECT().SetOidc(gomock.Any()).Times(1)
	(&OidcBootstrapper{}).Bootstrap(app)
}
