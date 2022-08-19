package bootstrappers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
)

func TestPprofBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().AddServer(gomock.Any()).Times(1)
	app.EXPECT().Config().Times(1).Return(&config.Config{})
	(&PprofBootstrapper{}).Bootstrap(app)
}

func TestPprofRunner_Shutdown(t *testing.T) {
	assert.Nil(t, (&pprofRunner{}).Shutdown(context.TODO()))
}
