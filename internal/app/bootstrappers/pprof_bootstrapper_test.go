package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPprofBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := app.NewMockApp(m)
	app.EXPECT().AddServer(gomock.Any())
	app.EXPECT().Logger().Return(mlog.NewForConfig(nil))
	assert.Nil(t, (&PprofBootstrapper{}).Bootstrap(app))
}

func TestPprofBootstrapper_Tags(t *testing.T) {
	// tag 用 "profile"，与 --exclude_server 文档/README 一致（而非类型名 "pprof"）。
	assert.Equal(t, []string{"profile"}, (&PprofBootstrapper{}).Tags())
}
