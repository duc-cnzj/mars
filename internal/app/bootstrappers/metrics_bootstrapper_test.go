package bootstrappers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/duc-cnzj/mars/internal/config"

	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
)

func TestMetricsBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().AddServer(gomock.Any()).Times(1)
	app.EXPECT().Config().Return(&config.Config{MetricsPort: "9091"}).Times(1)
	(&MetricsBootstrapper{}).Bootstrap(app)
}

func TestMetricsBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{"metrics"}, (&MetricsBootstrapper{}).Tags())
}

func Test_metricsRunner_Shutdown(t *testing.T) {
	assert.Nil(t, (&metricsRunner{}).Shutdown(context.TODO()))
}
