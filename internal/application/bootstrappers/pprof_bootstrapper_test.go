package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPprofBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := application.NewMockApp(m)
	app.EXPECT().AddServer(gomock.Any())
	app.EXPECT().Logger().Return(mlog.NewForConfig(nil))
	assert.Nil(t, (&PprofBootstrapper{}).Bootstrap(app))
}

func TestPprofBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{"pprof"}, (&PprofBootstrapper{}).Tags())
}
