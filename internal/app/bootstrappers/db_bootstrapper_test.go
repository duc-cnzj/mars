package bootstrappers

import (
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/config"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDBBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	app := app.NewMockApp(m)
	app.EXPECT().Data().Return(mockData).AnyTimes()
	mockData.EXPECT().InitDB().Return(func() error { return nil }, nil)
	app.EXPECT().Config().Return(&config.Config{DBAutoMigrate: true})
	// 注册后立即执行闭包，覆盖 after-shutdown 钩子体（真实 shutdown 时会触发）。
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).Do(func(f func()) { f() })
	app.EXPECT().Logger().Return(mlog.NewForConfig(nil))
	// 自动迁移收敛到 Data.Migrate()，不再触碰 DB() 客户端。
	mockData.EXPECT().Migrate().Return(nil)
	a := &DBBootstrapper{}
	assert.Nil(t, a.Bootstrap(app))
}

func TestDBBootstrapper_Bootstrap_NoAutoMigrate(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	app := app.NewMockApp(m)
	app.EXPECT().Data().Return(mockData).AnyTimes()
	mockData.EXPECT().InitDB().Return(func() error { return nil }, nil)
	app.EXPECT().Config().Return(&config.Config{DBAutoMigrate: false})
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any())
	app.EXPECT().Logger().Return(mlog.NewForConfig(nil))
	// DBAutoMigrate=false 时不再走 Schema.Create，不触碰 DB()。
	a := &DBBootstrapper{}
	assert.Nil(t, a.Bootstrap(app))
}

func TestDBBootstrapper_Bootstrap_Fail(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	app := app.NewMockApp(m)
	app.EXPECT().Data().Return(mockData).AnyTimes()
	mockData.EXPECT().InitDB().Return(nil, errors.New("x"))
	a := &DBBootstrapper{}
	assert.Error(t, a.Bootstrap(app))
}

func TestDBBootstrapper_Tags(t *testing.T) {
	a := &DBBootstrapper{}
	got := a.Tags()
	want := []string{}
	assert.Equal(t, got, want)
}
