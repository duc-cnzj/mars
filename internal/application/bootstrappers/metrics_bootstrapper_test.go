package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMetricsBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := application.NewMockApp(m)
	app.EXPECT().AddServer(gomock.Any())
	app.EXPECT().Config().Return(&config.Config{})
	app.EXPECT().Logger().Return(mlog.NewForConfig(nil))
	app.EXPECT().PrometheusRegistry()
	assert.Nil(t, (&MetricsBootstrapper{}).Bootstrap(app))
}

func TestMetricsBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{"metrics"}, (&MetricsBootstrapper{}).Tags())
}
