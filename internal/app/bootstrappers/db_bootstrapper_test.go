package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDBBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().Config().Return(&config.Config{DBDriver: "sqlite", DBDatabase: "file::memory:?cache=shared"}).Times(1)

	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).Times(1)
	app.EXPECT().IsDebug().Return(false).AnyTimes()
	dbm := mock.NewMockDBManager(controller)
	dbm.EXPECT().SetDB(gomock.Any()).Times(1)
	dbm.EXPECT().AutoMigrate(gomock.Any()).Times(1)
	app.EXPECT().DBManager().Return(dbm).AnyTimes()

	assert.Nil(t, (&DBBootstrapper{}).Bootstrap(app))

	app.EXPECT().Config().Return(&config.Config{DBDriver: "xxx"}).Times(1)
	assert.Equal(t, "db_driver must in ['sqlite', 'mysql']", (&DBBootstrapper{}).Bootstrap(app).Error())
}

func TestDBBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{}, (&DBBootstrapper{}).Tags())
}
