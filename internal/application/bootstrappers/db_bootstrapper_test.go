package bootstrappers

import (
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/config"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDBBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	app := application.NewMockApp(m)
	app.EXPECT().Data().Return(mockData).AnyTimes()
	mockData.EXPECT().InitDB().Return(func() error { return nil }, nil)
	app.EXPECT().Config().Return(&config.Config{DBAutoMigrate: true})
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any())
	app.EXPECT().Logger().Return(mlog.NewForConfig(nil))
	db, _ := data.NewSqliteDB()
	defer db.Close()
	mockData.EXPECT().DB().Return(db)
	a := &DBBootstrapper{}
	assert.Nil(t, a.Bootstrap(app))
}

func TestDBBootstrapper_Bootstrap_Fail(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	app := application.NewMockApp(m)
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
