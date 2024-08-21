package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTracingBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := application.NewMockApp(m)
	app.EXPECT().Config().Return(&config.Config{})
	app.EXPECT().Logger().Return(mlog.NewLogger(nil))
	assert.Nil(t, (&TracingBootstrapper{}).Bootstrap(app))
}

func TestTracingBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{"trace"}, (&TracingBootstrapper{}).Tags())
}
