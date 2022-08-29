package testutil

import (
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetGormDB(m *gomock.Controller, app *mock.MockApplicationInterface) (*gorm.DB, func()) {
	manager := mock.NewMockDBManager(m)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.Exec("PRAGMA foreign_keys = ON", nil)
	s, _ := db.DB()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
	return db, func() {
		s.Close()
	}
}

func MockApp(m *gomock.Controller) *mock.MockApplicationInterface {
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	return app
}

func AssertAuditLogFired(m *gomock.Controller, app *mock.MockApplicationInterface) *mock.MockDispatcherInterface {
	e := mock.NewMockDispatcherInterface(m)
	e.EXPECT().Dispatch(contracts.Event("audit_log"), gomock.Any()).Times(1)
	app.EXPECT().EventDispatcher().Return(e).AnyTimes()

	return e
}
