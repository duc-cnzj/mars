package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
)

func TestMetricsBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().SetMetrics(gomock.Any()).Times(1)
	app.EXPECT().AddServer(gomock.Any()).Times(1)
	(&MetricsBootstrapper{}).Bootstrap(app)
}
