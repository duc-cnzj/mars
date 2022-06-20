package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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
