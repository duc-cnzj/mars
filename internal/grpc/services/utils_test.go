package services

import (
	"testing"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/event/events"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"go.uber.org/mock/gomock"
)

func TestVarFunc(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	// 一次 listen，一次 dispatch
	d := mock.NewMockDispatcherInterface(ctrl)
	app.EXPECT().EventDispatcher().Return(d).AnyTimes()
	d.EXPECT().Dispatch(events.EventAuditLog, events.NewEventAuditLog("duc", 1, "aa"))
	AuditLog("duc", 1, "aa")
	d.EXPECT().Dispatch(events.EventAuditLog, events.NewEventAuditLog(
		"abc",
		types.EventActionType_Upload,
		"aa",
		events.AuditWithFileID(1)))
	FileAuditLog("abc", "aa", 1)
}
