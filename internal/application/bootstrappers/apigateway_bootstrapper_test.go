package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestApiGatewayBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	app := application.NewMockApp(m)
	app.EXPECT().HttpHandler().Return(nil)
	app.EXPECT().Logger().Return(mlog.NewForConfig(nil))
	app.EXPECT().GrpcRegistry()
	app.EXPECT().Config().Return(&config.Config{}).AnyTimes()
	app.EXPECT().AddServer(gomock.Any())
	a := &ApiGatewayBootstrapper{}
	assert.Nil(t, a.Bootstrap(app))
}

func TestApiGatewayBootstrapper_Tags(t *testing.T) {
	a := &ApiGatewayBootstrapper{}
	got := a.Tags()
	want := []string{"api", "gateway"}
	assert.Equal(t, got, want)
}
