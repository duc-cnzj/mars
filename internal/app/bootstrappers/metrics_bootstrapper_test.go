package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMetricsBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := app.NewMockApp(m)
	app.EXPECT().AddServer(gomock.Any())
	app.EXPECT().Config().Return(&config.Config{})
	app.EXPECT().Logger().Return(mlog.NewForConfig(nil))
	app.EXPECT().PrometheusRegistry()
	assert.Nil(t, (&MetricsBootstrapper{}).Bootstrap(app))
}

func TestMetricsBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{"metrics"}, (&MetricsBootstrapper{}).Tags())
}
