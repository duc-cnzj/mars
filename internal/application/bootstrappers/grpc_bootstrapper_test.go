package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGrpcBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := application.NewMockApp(m)
	app.EXPECT().Config().Return(&config.Config{})
	app.EXPECT().AddServer(gomock.Any())
	app.EXPECT().GrpcRegistry()
	app.EXPECT().Auth()
	app.EXPECT().Logger().Return(mlog.NewLogger(nil))
	assert.Nil(t, (&GrpcBootstrapper{}).Bootstrap(app))
}

func TestGrpcBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{"api", "grpc"}, (&GrpcBootstrapper{}).Tags())
}
