package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestEventBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := application.NewMockApp(m)
	eb := &EventBootstrapper{}
	app.EXPECT().Dispatcher()
	app.EXPECT().Logger().Return(mlog.NewLogger(nil))
	app.EXPECT().AddServer(gomock.Any())
	assert.Nil(t, eb.Bootstrap(app))
}

func TestEventBootstrapper_Tags(t *testing.T) {
	eb := &EventBootstrapper{}
	assert.Equal(t, eb.Tags(), []string{})
}
