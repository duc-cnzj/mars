package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v5/websocket"
	"github.com/duc-cnzj/mars/v5/internal/application"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/event"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/util/date"
	"github.com/duc-cnzj/mars/v5/internal/util/yaml"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestEventRepo_List(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	db, _ := data.NewSqliteDB()
	defer db.Close()
	loggerMock := mlog.NewForConfig(nil)
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
	loggerMock := mlog.NewForConfig(nil)
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
	loggerMock := mlog.NewForConfig(nil)
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
	loggerMock := mlog.NewForConfig(nil)
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
	loggerMock := mlog.NewForConfig(nil)
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
	loggerMock := mlog.NewForConfig(nil)
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
	loggerMock := mlog.NewForConfig(nil)
	eventDispatcherMock := event.NewMockDispatcher(m)
	eventDispatcherMock.EXPECT().Listen(gomock.Any(), gomock.Any()).AnyTimes()
	repo := NewEventRepo(nil, nil, nil, nil, loggerMock, dataMock, eventDispatcherMock)

	eventDispatcherMock.EXPECT().Dispatch(AuditLogEvent, NewEventAuditLog("testUser", types.EventActionType(1), "testMessage", AuditWithOldNew(&AnyYamlPrettier{}, &emptyYamlPrettier{})))

	repo.AuditLogWithChange(types.EventActionType(1), "testUser", "testMessage", &AnyYamlPrettier{}, nil)
}

func TestEventRepo_HandleAuditLog(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	db, _ := data.NewSqliteDB()
	defer db.Close()
	loggerMock := mlog.NewForConfig(nil)
	eventDispatcherMock := event.NewMockDispatcher(m)
	eventDispatcherMock.EXPECT().Listen(gomock.Any(), gomock.Any()).AnyTimes()

	repo := NewEventRepo(nil, nil, nil, nil, loggerMock, data.NewDataImpl(&data.NewDataParams{DB: db}), eventDispatcherMock)

	// Test case: AuditLog with no changes
	auditLogData := NewEventAuditLog("testUser", types.EventActionType(1), "testMessage")
	err := repo.HandleAuditLog(auditLogData, AuditLogEvent)
	assert.NoError(t, err)

	// Test case: AuditLog with changes
	auditLogDataWithChanges := NewEventAuditLog("testUser", types.EventActionType(1), "testMessage", AuditWithOldNewStr("old", "new"))
	err = repo.HandleAuditLog(auditLogDataWithChanges, AuditLogEvent)
	assert.NoError(t, err)

	// Test case: AuditLog with file ID
	save := db.File.Create().
		SetUsername("testUser").
		SetPath("/x").
		SaveX(context.TODO())
	auditLogDataWithFileID := NewEventAuditLog("testUser", types.EventActionType(1), "testMessage", AuditWithFileID(save.ID))
	err = repo.HandleAuditLog(auditLogDataWithFileID, AuditLogEvent)
	assert.NoError(t, err)

	// Test case: AuditLog with duration
	auditLogDataWithDuration := NewEventAuditLog("testUser", types.EventActionType(1), "testMessage", AuditWithDuration("2s"))
	err = repo.HandleAuditLog(auditLogDataWithDuration, AuditLogEvent)
	assert.NoError(t, err)
}

func Test_eventRepo_Dispatch(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	db, _ := data.NewSqliteDB()
	defer db.Close()
	loggerMock := mlog.NewForConfig(nil)
	eventDispatcherMock := event.NewMockDispatcher(m)
	eventDispatcherMock.EXPECT().Listen(gomock.Any(), gomock.Any()).AnyTimes()

	repo := NewEventRepo(nil, nil, nil, nil, loggerMock, data.NewDataImpl(&data.NewDataParams{DB: db}), eventDispatcherMock)
	eventDispatcherMock.EXPECT().Dispatch(AuditLogEvent, NewEventAuditLog("testUser", types.EventActionType(1), "testMessage"))
	repo.Dispatch(AuditLogEvent, NewEventAuditLog("testUser", types.EventActionType(1), "testMessage"))
}

func TestAuditLogImpl_Getters(t *testing.T) {
	auditLog := &auditLogImpl{
		Username: "testUser",
		Action:   types.EventActionType(1),
		Msg:      "testMessage",
		OldS:     "old",
		NewS:     "new",
		FileId:   1,
		Duration: "2s",
	}

	assert.Equal(t, "testUser", auditLog.GetUsername())
	assert.Equal(t, types.EventActionType(1), auditLog.GetAction())
	assert.Equal(t, "testMessage", auditLog.GetMsg())
	assert.Equal(t, "old", auditLog.GetOldStr())
	assert.Equal(t, "new", auditLog.GetNewStr())
	assert.Equal(t, 1, auditLog.GetFileID())
	assert.Equal(t, "2s", auditLog.GetDuration())
}

func TestNewEventAuditLog(t *testing.T) {
	auditLog := NewEventAuditLog("testUser", types.EventActionType(1), "testMessage", AuditWithOldNewStr("old", "new"), AuditWithFileID(1), AuditWithDuration("2s"))

	assert.Equal(t, "testUser", auditLog.GetUsername())
	assert.Equal(t, types.EventActionType(1), auditLog.GetAction())
	assert.Equal(t, "testMessage", auditLog.GetMsg())
	assert.Equal(t, "old", auditLog.GetOldStr())
	assert.Equal(t, "new", auditLog.GetNewStr())
	assert.Equal(t, 1, auditLog.GetFileID())
	assert.Equal(t, "2s", auditLog.GetDuration())
}

func TestAuditWithOldNew(t *testing.T) {
	old := &StringYamlPrettier{Str: "old"}
	new := &StringYamlPrettier{Str: "new"}

	auditLog := &auditLogImpl{}
	AuditWithOldNew(old, new)(auditLog)

	assert.Equal(t, "old", auditLog.GetOldStr())
	assert.Equal(t, "new", auditLog.GetNewStr())
}

func TestAuditWithFileID(t *testing.T) {
	auditLog := &auditLogImpl{}
	AuditWithFileID(1)(auditLog)

	assert.Equal(t, 1, auditLog.GetFileID())
}

func TestAuditWithDuration(t *testing.T) {
	auditLog := &auditLogImpl{}
	AuditWithDuration("2s")(auditLog)

	assert.Equal(t, "2s", auditLog.GetDuration())
}

func Test_eventRepo_HandleProjectDeleted(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	loggerMock := mlog.NewForConfig(nil)
	eventDispatcherMock := event.NewMockDispatcher(m)
	eventDispatcherMock.EXPECT().Listen(gomock.Any(), gomock.Any()).AnyTimes()

	pl := application.NewMockPluginManger(m)
	repo := NewEventRepo(nil, nil, pl, nil, loggerMock, nil, eventDispatcherMock)
	sender := application.NewMockWsSender(m)
	pl.EXPECT().Ws().Return(sender)
	sub := application.NewMockPubSub(m)
	sender.EXPECT().New("", "").Return(sub)
	sub.EXPECT().Close()
	sub.EXPECT().ToAll(gomock.Cond(func(x any) bool {
		v := x.(*websocket_pb.WsReloadProjectsResponse)
		return v.NamespaceId == 1 && v.Metadata.Type == websocket_pb.Type_ReloadProjects
	}))

	repo.(*eventRepo).HandleProjectDeleted(&ProjectDeletedPayload{
		NamespaceID: 1,
	}, "")
}
func Test_eventRepo_HandleProjectChanged(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	loggerMock := mlog.NewForConfig(nil)
	eventDispatcherMock := event.NewMockDispatcher(m)
	eventDispatcherMock.EXPECT().Listen(gomock.Any(), gomock.Any()).AnyTimes()

	projectRepoMock := NewMockProjectRepo(m)
	clRepoMock := NewMockChangelogRepo(m)

	pl := application.NewMockPluginManger(m)
	repo := NewEventRepo(projectRepoMock, nil, pl, clRepoMock, loggerMock, nil, eventDispatcherMock)

	// Test case: ProjectChangedData with valid project ID
	projectRepoMock.EXPECT().Show(context.TODO(), 1).Return(&Project{}, nil)
	clRepoMock.EXPECT().FindLastChangeByProjectID(context.TODO(), 1).Return(&Changelog{}, nil)
	clRepoMock.EXPECT().Create(context.TODO(), gomock.Any()).Return(nil, nil)

	err := repo.(*eventRepo).HandleProjectChanged(&ProjectChangedData{
		ID:       1,
		Username: "testUser",
	}, "")
	assert.NoError(t, err)

	// Test case: ProjectChangedData with invalid project ID
	projectRepoMock.EXPECT().Show(context.TODO(), 2).Return(nil, errors.New("project not found"))
	err = repo.(*eventRepo).HandleProjectChanged(&ProjectChangedData{
		ID:       2,
		Username: "testUser",
	}, "")
	assert.Error(t, err)
}

func Test_eventRepo_HandleNamespaceDeleted(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	loggerMock := mlog.NewForConfig(nil)
	eventDispatcherMock := event.NewMockDispatcher(m)
	eventDispatcherMock.EXPECT().Listen(gomock.Any(), gomock.Any()).AnyTimes()

	pl := application.NewMockPluginManger(m)
	repo := NewEventRepo(nil, nil, pl, nil, loggerMock, nil, eventDispatcherMock)
	sender := application.NewMockWsSender(m)
	pl.EXPECT().Ws().Return(sender)
	sub := application.NewMockPubSub(m)
	sender.EXPECT().New("", "").Return(sub)
	sub.EXPECT().Close()
	sub.EXPECT().ToAll(gomock.Cond(func(x any) bool {
		v := x.(*websocket_pb.WsReloadProjectsResponse)
		return v.NamespaceId == 1 && v.Metadata.Type == websocket_pb.Type_ReloadProjects
	}))

	repo.(*eventRepo).HandleNamespaceDeleted(NamespaceDeletedData{
		ID: 1,
	}, "")
}

func TestEventRepo_HandleInjectTlsSecret(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	loggerMock := mlog.NewForConfig(nil)
	eventDispatcherMock := event.NewMockDispatcher(m)
	eventDispatcherMock.EXPECT().Listen(gomock.Any(), gomock.Any()).AnyTimes()

	k8sRepoMock := NewMockK8sRepo(m)
	pl := application.NewMockPluginManger(m)
	domainMock := application.NewMockDomainManager(m)
	pl.EXPECT().Domain().Return(domainMock)
	domainMock.EXPECT().GetCerts().Return("name", "key", "crt")

	repo := NewEventRepo(nil, k8sRepoMock, pl, nil, loggerMock, nil, eventDispatcherMock)

	k8sRepoMock.EXPECT().AddTlsSecret("namespace", "name", "key", "crt").Return(nil, nil)

	err := repo.(*eventRepo).HandleInjectTlsSecret(NamespaceCreatedData{
		NsK8sObj: &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "namespace",
			},
		},
	}, "")

	assert.NoError(t, err)
}

func TestEventRepo_HandleInjectTlsSecret_NoCerts(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	loggerMock := mlog.NewForConfig(nil)
	eventDispatcherMock := event.NewMockDispatcher(m)
	eventDispatcherMock.EXPECT().Listen(gomock.Any(), gomock.Any()).AnyTimes()

	pl := application.NewMockPluginManger(m)
	domainMock := application.NewMockDomainManager(m)
	pl.EXPECT().Domain().Return(domainMock)
	domainMock.EXPECT().GetCerts().Return("", "", "")

	repo := NewEventRepo(nil, nil, pl, nil, loggerMock, nil, eventDispatcherMock)

	err := repo.(*eventRepo).HandleInjectTlsSecret(NamespaceCreatedData{
		NsK8sObj: &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "namespace",
			},
		},
	}, "")

	assert.NoError(t, err)
}

func TestEventRepo_HandleInjectTlsSecret_AddTlsSecretError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	loggerMock := mlog.NewForConfig(nil)
	eventDispatcherMock := event.NewMockDispatcher(m)
	eventDispatcherMock.EXPECT().Listen(gomock.Any(), gomock.Any()).AnyTimes()

	k8sRepoMock := NewMockK8sRepo(m)
	pl := application.NewMockPluginManger(m)
	domainMock := application.NewMockDomainManager(m)
	pl.EXPECT().Domain().Return(domainMock)
	domainMock.EXPECT().GetCerts().Return("name", "key", "crt")

	repo := NewEventRepo(nil, k8sRepoMock, pl, nil, loggerMock, nil, eventDispatcherMock)

	k8sRepoMock.EXPECT().AddTlsSecret("namespace", "name", "key", "crt").Return(nil, errors.New("error"))

	err := repo.(*eventRepo).HandleInjectTlsSecret(NamespaceCreatedData{
		NsK8sObj: &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "namespace",
			},
		},
	}, "")

	assert.NoError(t, err)
}
