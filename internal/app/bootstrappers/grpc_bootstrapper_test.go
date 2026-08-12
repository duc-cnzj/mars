package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGrpcBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := app.NewMockApp(m)
	app.EXPECT().Config().Return(&config.Config{})
	app.EXPECT().AddServer(gomock.Any())
	app.EXPECT().GrpcRegistry()
	app.EXPECT().AuthBiz()
	app.EXPECT().Logger().Return(mlog.NewForConfig(nil))
	assert.Nil(t, (&GrpcBootstrapper{}).Bootstrap(app))
}

func TestGrpcBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{"api", "grpc"}, (&GrpcBootstrapper{}).Tags())
}
