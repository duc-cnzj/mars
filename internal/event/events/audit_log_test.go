package events

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/event"
	"github.com/duc-cnzj/mars/internal/mock"
)

func TestAuditLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	// 一次 listen，一次 dispatch
	app.EXPECT().EventDispatcher().Times(2).Return(event.NewDispatcher(app))
	var called bool
	app.EventDispatcher().Listen(EventAuditLog, func(a any, e contracts.Event) error {
		data := a.(EventAuditLogData)
		assert.Equal(t, "duc", data.Username)
		assert.Equal(t, "hello", data.Msg)
		assert.Equal(t, types.EventActionType_Shell, data.Action)
		assert.Equal(t, "", data.NewS)
		assert.Equal(t, "", data.OldS)
		called = true
		return nil
	})
	AuditLog("duc", types.EventActionType_Shell, "hello", nil, nil)
	assert.True(t, called)
}

func TestFileAuditLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	// 一次 listen，一次 dispatch
	app.EXPECT().EventDispatcher().Times(2).Return(event.NewDispatcher(app))
	var called bool
	app.EventDispatcher().Listen(EventAuditLog, func(a any, e contracts.Event) error {
		data := a.(EventAuditLogData)
		assert.Equal(t, "duc", data.Username)
		assert.Equal(t, "hello", data.Msg)
		assert.Equal(t, types.EventActionType_Upload, data.Action)
		assert.Equal(t, "", data.NewS)
		assert.Equal(t, "", data.OldS)
		assert.Equal(t, 1, data.FileId)
		called = true
		return nil
	})
	FileAuditLog("duc", "hello", 1)
	assert.True(t, called)
}

func TestHandleAuditLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	dbManager := mock.NewMockDBManager(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	// 一次 listen，一次 dispatch
	app.EXPECT().EventDispatcher().Times(2).Return(event.NewDispatcher(app))
	sqlDB, _, _ := sqlmock.New()
	defer sqlDB.Close()
	gormDB, _ := gorm.Open(mysql.New(mysql.Config{SkipInitializeWithVersion: true, Conn: sqlDB}), &gorm.Config{})
	app.EXPECT().DBManager().Times(1).Return(dbManager)
	dbManager.EXPECT().DB().Return(gormDB)
	var called bool
	app.EventDispatcher().Listen(EventAuditLog, func(a any, e contracts.Event) error {
		HandleAuditLog(a, e)
		called = true
		return nil
	})
	FileAuditLog("duc", "hello", 1)
	assert.True(t, called)
}

func TestStringYamlPrettier_PrettyYaml(t *testing.T) {
	assert.Equal(t, (&StringYamlPrettier{Str: "aa"}).PrettyYaml(), "aa")
}

func Test_emptyYamlPrettier_PrettyYaml(t *testing.T) {
	assert.Equal(t, (&emptyYamlPrettier{}).PrettyYaml(), "")
}
