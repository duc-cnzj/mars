package services

import (
	"testing"

	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/event/events"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/golang/mock/gomock"
)

func TestVarFunc(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	// 一次 listen，一次 dispatch
	d := mock.NewMockDispatcherInterface(ctrl)
	app.EXPECT().EventDispatcher().Return(d).AnyTimes()
	d.EXPECT().Dispatch(events.EventAuditLog, events.EventAuditLogData{
		Username: "duc",
		Action:   1,
		Msg:      "aa",
	})
	AuditLog("duc", 1, "aa")
	d.EXPECT().Dispatch(events.EventAuditLog, events.EventAuditLogData{
		Username: "abc",
		Action:   types.EventActionType_Upload,
		Msg:      "aa",
		FileId:   1,
	})
	FileAuditLog("abc", "aa", 1)
}
