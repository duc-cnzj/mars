package repo

import (
	"context"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/event"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/util/date"
	"github.com/duc-cnzj/mars/v4/internal/util/yaml"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestEventRepo_List(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	db, _ := data.NewSqliteDB()
	defer db.Close()
	loggerMock := mlog.NewLogger(nil)
	eventDispatcherMock := event.NewMockDispatcher(m)
	eventDispatcherMock.EXPECT().Listen(gomock.Any(), gomock.Any()).AnyTimes()

	repo := NewEventRepo(nil, nil, nil, nil, loggerMock, data.NewDataImpl(&data.NewDataParams{DB: db}), eventDispatcherMock)

	// seed data
	db.Event.Create().
		SetAction(types.EventActionType(1)).
		SetUsername("ducaaaa").
		SetMessage("qwerty").
		//SetFileID(1).
		SetDuration("2s").
		SetHasDiff(true).
		SaveX(context.TODO())
	db.Event.Create().
		SetAction(types.EventActionType(2)).
		SetUsername("aaaa").
		SetMessage("1111").
		SetDuration("22s").
		SetHasDiff(true).
		SaveX(context.TODO())
	db.Event.Create().
		SetAction(types.EventActionType(3)).
		SetUsername("bbbbb").
		SetMessage("22222").
		SetDuration("2s").
		SetHasDiff(false).
		SaveX(context.TODO())

	input := &ListEventInput{
		Page:        1,
		PageSize:    10,
		ActionType:  types.EventActionType(1),
		Search:      "aaaa",
		OrderIDDesc: lo.ToPtr(true),
	}

	events, pag, err := repo.List(context.TODO(), input)

	assert.Nil(t, err)
	assert.NotNil(t, events)
	assert.NotNil(t, pag)

	assert.Equal(t, 1, len(events))
	assert.Equal(t, int32(1), pag.Page)
	assert.Equal(t, "ducaaaa", events[0].Username)
}

func TestEventRepo_Show(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	db, _ := data.NewSqliteDB()
	defer db.Close()
	loggerMock := mlog.NewLogger(nil)
	eventDispatcherMock := event.NewMockDispatcher(m)
	eventDispatcherMock.EXPECT().Listen(gomock.Any(), gomock.Any()).AnyTimes()

	repo := NewEventRepo(nil, nil, nil, nil, loggerMock, data.NewDataImpl(&data.NewDataParams{DB: db}), eventDispatcherMock)

	event, err := repo.Show(context.TODO(), 1)

	assert.Error(t, err)
	assert.Nil(t, event)
}

func TestEventRepo_AuditLog(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	dataMock := data.NewMockData(m)
	loggerMock := mlog.NewLogger(nil)
	eventDispatcherMock := event.NewMockDispatcher(m)
	eventDispatcherMock.EXPECT().Listen(gomock.Any(), gomock.Any()).AnyTimes()

	repo := NewEventRepo(nil, nil, nil, nil, loggerMock, dataMock, eventDispatcherMock)

	eventDispatcherMock.EXPECT().Dispatch(AuditLogEvent, NewEventAuditLog("testUser", types.EventActionType(1), "testMessage", AuditWithOldNew(&emptyYamlPrettier{}, &emptyYamlPrettier{})))
	repo.AuditLog(types.EventActionType(1), "testUser", "testMessage")
}

func TestEventRepo_FileAuditLog(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	dataMock := data.NewMockData(m)
	loggerMock := mlog.NewLogger(nil)
	eventDispatcherMock := event.NewMockDispatcher(m)
	eventDispatcherMock.EXPECT().Listen(gomock.Any(), gomock.Any()).AnyTimes()

	repo := NewEventRepo(nil, nil, nil, nil, loggerMock, dataMock, eventDispatcherMock)

	eventDispatcherMock.EXPECT().Dispatch(AuditLogEvent, NewEventAuditLog("testUser", types.EventActionType(1), "testMessage", AuditWithFileID(1)))
	repo.FileAuditLog(types.EventActionType(1), "testUser", "testMessage", 1)
}

func TestEventRepo_FileAuditLogWithDuration(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	dataMock := data.NewMockData(m)
	loggerMock := mlog.NewLogger(nil)
	eventDispatcherMock := event.NewMockDispatcher(m)
	eventDispatcherMock.EXPECT().Listen(gomock.Any(), gomock.Any()).AnyTimes()

	repo := NewEventRepo(nil, nil, nil, nil, loggerMock, dataMock, eventDispatcherMock)

	eventDispatcherMock.EXPECT().Dispatch(AuditLogEvent, NewEventAuditLog("testUser", types.EventActionType(1), "testMessage", AuditWithFileID(1), AuditWithDuration(date.HumanDuration(time.Second))))
	repo.FileAuditLogWithDuration(types.EventActionType(1), "testUser", "testMessage", 1, time.Second)
}

func TestEventRepo_AuditLogWithRequest(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	dataMock := data.NewMockData(m)
	loggerMock := mlog.NewLogger(nil)
	eventDispatcherMock := event.NewMockDispatcher(m)
	eventDispatcherMock.EXPECT().Listen(gomock.Any(), gomock.Any()).AnyTimes()

	repo := NewEventRepo(nil, nil, nil, nil, loggerMock, dataMock, eventDispatcherMock)

	req := struct {
		Name string
	}{
		Name: "a",
	}
	marshal, _ := yaml.PrettyMarshal(req)
	eventDispatcherMock.EXPECT().Dispatch(AuditLogEvent, NewEventAuditLog("testUser", types.EventActionType(1), "testMessage", AuditWithOldNewStr("", string(marshal))))
	repo.AuditLogWithRequest(types.EventActionType(1), "testUser", "testMessage", req)
}

func TestEventRepo_AuditLogWithChange(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	dataMock := data.NewMockData(m)
	loggerMock := mlog.NewLogger(nil)
	eventDispatcherMock := event.NewMockDispatcher(m)
	eventDispatcherMock.EXPECT().Listen(gomock.Any(), gomock.Any()).AnyTimes()
	repo := NewEventRepo(nil, nil, nil, nil, loggerMock, dataMock, eventDispatcherMock)

	eventDispatcherMock.EXPECT().Dispatch(AuditLogEvent, NewEventAuditLog("testUser", types.EventActionType(1), "testMessage", AuditWithOldNew(&AnyYamlPrettier{}, &emptyYamlPrettier{})))

	repo.AuditLogWithChange(types.EventActionType(1), "testUser", "testMessage", &AnyYamlPrettier{}, nil)
}
