package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestLogBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().IsDebug().Return(false).AnyTimes()
	app.EXPECT().Config().Return(&config.Config{LogChannel: "xxx"}).Times(2)
	err := (&LogBootstrapper{}).Bootstrap(app)
	assert.Error(t, err)

	app.EXPECT().Config().Return(&config.Config{LogChannel: "logrus"}).Times(1)
	err = (&LogBootstrapper{}).Bootstrap(app)
	assert.Nil(t, err)
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).Times(1)
	app.EXPECT().Config().Return(&config.Config{LogChannel: "zap"}).Times(1)
	err = (&LogBootstrapper{}).Bootstrap(app)
	assert.Nil(t, err)
}

func TestLogBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&LogBootstrapper{}).Tags())
}
