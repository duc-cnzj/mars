package socket

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v5/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v5/websocket"
	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/locker"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/uploader"
	"github.com/duc-cnzj/mars/v5/internal/util/counter"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func TestWebsocketManager_HandleAuthorize(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	authMock := auth.NewMockAuth(m)
	authMock.EXPECT().VerifyToken("validToken").Return(&auth.JwtClaims{UserInfo: &auth.UserInfo{Name: "testUser"}}, true)
	authMock.EXPECT().VerifyToken("invalidToken").Return(nil, false)

	wm := &WebsocketManager{
		auth: authMock,
	}

	conn := &WsConn{}
	var inputv = websocket_pb.AuthorizeTokenInput{
		Token: "validToken",
	}
	marshalv, _ := proto.Marshal(&inputv)
	wm.HandleAuthorize(context.TODO(), conn, WsAuthorize, marshalv)
	assert.Equal(t, "testUser", conn.GetUser().Name)

	conn = &WsConn{}
	var input = websocket_pb.AuthorizeTokenInput{
		Token: "invalidToken",
	}
	marshal, _ := proto.Marshal(&input)
	wm.HandleAuthorize(context.TODO(), conn, WsAuthorize, marshal)
	assert.Nil(t, conn.GetUser())
}

func TestUpgrader(t *testing.T) {
	assert.True(t, upgrader.CheckOrigin(nil))
}

func TestWebsocketManager_HandleJoinRoom(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	wsMock := application.NewMockPubSub(m)
	wsMock.EXPECT().Join(int64(2))
	wsMock.EXPECT().Leave(int64(1), int64(2))

	conn := &WsConn{
		pubSub: wsMock,
	}
	var input = websocket_pb.ProjectPodEventJoinInput{
		Type:        ProjectPodEvent,
		Join:        true,
		NamespaceId: 1,
		ProjectId:   2,
	}
	marshal, _ := proto.Marshal(&input)
	var linput = websocket_pb.ProjectPodEventJoinInput{
		Type:        ProjectPodEvent,
		Join:        false,
		NamespaceId: 1,
		ProjectId:   2,
	}
	marshal2, _ := proto.Marshal(&linput)

	wm := &WebsocketManager{
		logger: mlog.NewForConfig(nil),
	}
	wm.HandleJoinRoom(context.TODO(), conn, ProjectPodEvent, marshal)

	wm.HandleJoinRoom(context.TODO(), conn, ProjectPodEvent, marshal2)
}

func TestWebsocketManager_HandleStartShell(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	wm := &WebsocketManager{logger: mlog.NewForConfig(nil)}
	sub := application.NewMockPubSub(m)
	conn := &WsConn{pubSub: sub, id: "testConnID", uid: "testConnUID"}

	input := &websocket_pb.WsHandleExecShellInput{
		SessionId: "testSession",
		Container: &websocket_pb.Container{
			Namespace: "testNamespace",
			Pod:       "testPod",
			Container: "testContainer",
		},
	}
	marshal, _ := proto.Marshal(input)
	sub.EXPECT().ToSelf(&WsResponse{
		Metadata: &websocket_pb.Metadata{
			Type:    WsHandleExecShell,
			Result:  ResultError,
			End:     true,
			Uid:     "testConnUID",
			Id:      "testConnID",
			Message: "invalid session sessionID, must format: '<namespace>-<pod>-<container>:<randomID>', input: 'testSession'",
		},
	})
	wm.HandleStartShell(context.TODO(), conn, WsHandleExecShell, marshal)
}

func TestWebsocketManager_HandleShellMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	wm := &WebsocketManager{
		logger: mlog.NewForConfig(nil),
	}
	conn := &WsConn{
		sm: NewSessionMap(wm.logger),
	}

	input := &websocket_pb.TerminalMessageInput{
		Message: &websocket_pb.TerminalMessage{
			SessionId: "testSession",
			Data:      []byte("testData"),
		},
	}
	handler := NewMockPtyHandler(m)
	handler.EXPECT().Send(gomock.Not(nil), &websocket_pb.TerminalMessage{
		SessionId: "testSession",
		Data:      []byte("testData"),
	}).Return(errors.New("x"))
	conn.sm.Set("testSession", handler)
	marshal, _ := proto.Marshal(input)
	wm.HandleShellMessage(context.TODO(), conn, WsHandleExecShellMsg, marshal)
}

func TestWebsocketManager_HandleCloseShell(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	wm := &WebsocketManager{
		logger: mlog.NewForConfig(nil),
	}
	conn := &WsConn{
		sm: NewSessionMap(wm.logger),
	}

	input := &websocket_pb.TerminalMessageInput{
		Message: &websocket_pb.TerminalMessage{
			SessionId: "testSession",
		},
	}
	handler := NewMockPtyHandler(m)
	handler.EXPECT().Close(gomock.Not(nil), gomock.Any())
	conn.sm.Set("testSession", handler)
	marshal, _ := proto.Marshal(input)
	wm.HandleCloseShell(context.TODO(), conn, WsHandleCloseShell, marshal)
	time.Sleep(1 * time.Second)
}

func TestWebsocketManager_HandleWsCancelDeploy(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	nsRepo := repo.NewMockNamespaceRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	wm := &WebsocketManager{nsRepo: nsRepo, logger: mlog.NewForConfig(nil), eventRepo: eventRepo}
	conn := &WsConn{
		taskManager: NewTaskManager(wm.logger),
		user:        &auth.UserInfo{},
	}

	input := &websocket_pb.CancelInput{
		NamespaceId: 1,
		Name:        "testProject",
	}
	called := false
	conn.taskManager.Register(GetSlugName(input.NamespaceId, input.Name), func(err error) {
		called = true
	})
	marshal, _ := proto.Marshal(input)
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Namespace{}, nil)
	eventRepo.EXPECT().AuditLog(types.EventActionType_CancelDeploy, "", gomock.Any())
	wm.HandleWsCancelDeploy(context.TODO(), conn, WsCancel, marshal)
	assert.True(t, called)
}

func TestWebsocketManager_HandleWsCreateProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	nsRepo := repo.NewMockNamespaceRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	repoRepo := repo.NewMockRepoRepo(m)
	jb := NewMockJobManager(m)
	wm := &WebsocketManager{
		nsRepo:     nsRepo,
		logger:     mlog.NewForConfig(nil),
		eventRepo:  eventRepo,
		jobManager: jb,
		repoRepo:   repoRepo,
	}
	job := NewMockJob(m)
	conn := &WsConn{
		taskManager: NewTaskManager(wm.logger),
		user:        &auth.UserInfo{},
	}

	input := &websocket_pb.CreateProjectInput{
		Type:        WsCreateProject,
		NamespaceId: 1,
		RepoId:      1,
		GitBranch:   "master",
		GitCommit:   "testCommit",
		Config:      "testConfig",
		ExtraValues: []*websocket_pb.ExtraValue{},
		Atomic:      lo.ToPtr(true),
	}
	repoRepo.EXPECT().Show(gomock.Any(), 1).Return(&repo.Repo{
		Name: "app",
		ID:   2,
	}, nil)
	marshal, _ := proto.Marshal(input)
	job.EXPECT().OnFinally(gomock.Not(nil), gomock.Not(nil))
	job.EXPECT().GlobalLock().Return(job)
	job.EXPECT().ID().Return("testID")
	job.EXPECT().Validate().Return(job)
	job.EXPECT().LoadConfigs().Return(job)
	job.EXPECT().Run(gomock.Not(nil)).Return(job)
	job.EXPECT().Finish().Return(job)
	job.EXPECT().Error().Return(nil)
	jb.EXPECT().NewJob(gomock.Cond(func(x any) bool {
		jobInput := x.(*JobInput)
		return jobInput.Type == WsCreateProject &&
			jobInput.NamespaceId == input.NamespaceId &&
			jobInput.RepoID == input.RepoId &&
			jobInput.GitBranch == input.GitBranch &&
			jobInput.GitCommit == input.GitCommit &&
			jobInput.Config == input.Config &&
			*jobInput.Atomic == *input.Atomic &&
			slices.Equal(jobInput.ExtraValues, input.ExtraValues) &&
			jobInput.User == conn.GetUser() &&
			jobInput.PubSub == conn.pubSub &&
			jobInput.Messager != nil
	})).Return(job)
	wm.HandleWsCreateProject(context.TODO(), conn, WsCreateProject, marshal)
}

func TestWebsocketManager_HandleWsUpdateProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	nsRepo := repo.NewMockNamespaceRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	repoRepo := repo.NewMockRepoRepo(m)
	proj := repo.NewMockProjectRepo(m)
	mockData := data.NewMockData(m)
	jb := NewMockJobManager(m)
	wm := &WebsocketManager{
		nsRepo:     nsRepo,
		logger:     mlog.NewForConfig(nil),
		eventRepo:  eventRepo,
		jobManager: jb,
		repoRepo:   repoRepo,
		projRepo:   proj,
		data:       mockData,
	}
	mockData.EXPECT().Config().Return(&config.Config{})
	job := NewMockJob(m)
	conn := &WsConn{
		taskManager: NewTaskManager(wm.logger),
		user:        &auth.UserInfo{},
	}

	input := &websocket_pb.UpdateProjectInput{
		ProjectId:   1,
		GitBranch:   "master",
		GitCommit:   "testCommit",
		Config:      "testConfig",
		ExtraValues: []*websocket_pb.ExtraValue{},
		Atomic:      lo.ToPtr(true),
	}
	proj.EXPECT().Show(gomock.Any(), 1).Return(&repo.Project{
		Name:        "appa",
		NamespaceID: 1,
		RepoID:      1,
	}, nil)
	marshal, _ := proto.Marshal(input)
	job.EXPECT().OnFinally(gomock.Any(), gomock.Any())
	job.EXPECT().GlobalLock().Return(job)
	job.EXPECT().ID().Return("testID")
	job.EXPECT().Validate().Return(job)
	job.EXPECT().LoadConfigs().Return(job)
	job.EXPECT().Run(gomock.Any()).Return(job)
	job.EXPECT().Finish().Return(job)
	job.EXPECT().Error().Return(nil)
	jb.EXPECT().NewJob(gomock.Cond(func(x any) bool {
		jobInput := x.(*JobInput)
		return jobInput.Type == WsUpdateProject &&
			jobInput.Name == "appa" &&
			jobInput.NamespaceId == 1 &&
			jobInput.ProjectID == 1 &&
			jobInput.GitBranch == input.GitBranch &&
			jobInput.GitCommit == input.GitCommit &&
			jobInput.Config == input.Config &&
			*jobInput.Atomic == *input.Atomic &&
			slices.Equal(jobInput.ExtraValues, input.ExtraValues) &&
			jobInput.User == conn.GetUser() &&
			jobInput.PubSub == conn.pubSub &&
			jobInput.Messager != nil
	})).Return(job)
	wm.HandleWsUpdateProject(context.TODO(), conn, WsUpdateProject, marshal)
}

func TestNewWebsocketManager(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	logger := mlog.NewForConfig(nil)
	counter := counter.NewCounter()
	projRepo := repo.NewMockProjectRepo(m)
	repoRepo := repo.NewMockRepoRepo(m)
	nsRepo := repo.NewMockNamespaceRepo(m)
	jobManager := NewMockJobManager(m)
	data := data.NewMockData(m)
	pl := application.NewMockPluginManger(m)
	auth := auth.NewMockAuth(m)
	uploader := uploader.NewMockUploader(m)
	locker := locker.NewMockLocker(m)
	clusterRepo := repo.NewMockK8sRepo(m)
	eventRepo := repo.NewMockEventRepo(m)
	executor := repo.NewMockExecutorManager(m)
	fileRepo := repo.NewMockFileRepo(m)

	wm := NewWebsocketManager(timer.NewReal(), logger, counter, projRepo, repoRepo, nsRepo, jobManager, data, pl, auth, uploader, locker, clusterRepo, eventRepo, executor, fileRepo).(*WebsocketManager)

	assert.NotNil(t, wm)
	assert.Equal(t, logger.WithModule("socket/websocket"), wm.logger)
	assert.Equal(t, counter, wm.counter)
	assert.Equal(t, projRepo, wm.projRepo)
	assert.Equal(t, repoRepo, wm.repoRepo)
	assert.Equal(t, nsRepo, wm.nsRepo)
	assert.Equal(t, jobManager, wm.jobManager)
	assert.Equal(t, fileRepo, wm.fileRepo)
	assert.Equal(t, data, wm.data)
	assert.Equal(t, pl, wm.pl)
	assert.Equal(t, auth, wm.auth)
	assert.Equal(t, uploader, wm.uploader)
	assert.Equal(t, locker, wm.locker)
	assert.Equal(t, clusterRepo, wm.k8sRepo)
	assert.Equal(t, eventRepo, wm.eventRepo)
	assert.Equal(t, executor, wm.executor)
	assert.NotNil(t, wm.timer)
	assert.Len(t, wm.handlers, 8)
}
func TestWebsocketManager_TickClusterHealth(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	lockerMock := locker.NewMockLocker(m)
	plMock := application.NewMockPluginManger(m)
	wsMock := application.NewMockWsSender(m)
	k8sRepoMock := repo.NewMockK8sRepo(m)
	loggerMock := mlog.NewForConfig(nil)

	lockerMock.EXPECT().Acquire("TickClusterHealth", int64(5)).Return(true)
	lockerMock.EXPECT().Release("TickClusterHealth")
	plMock.EXPECT().Ws().Return(wsMock)
	sub := application.NewMockPubSub(m)
	wsMock.EXPECT().New(gomock.Any(), gomock.Any()).Return(sub)
	sub.EXPECT().Close()
	sub.EXPECT().ToAll(gomock.Any())
	k8sRepoMock.EXPECT().ClusterInfo().Return(&repo.ClusterInfo{})

	wm := &WebsocketManager{
		locker:             lockerMock,
		pl:                 plMock,
		k8sRepo:            k8sRepoMock,
		logger:             loggerMock,
		healthTickDuration: 1 * time.Second,
	}

	done := make(chan struct{})
	go func() {
		time.Sleep(1100 * time.Millisecond)
		close(done)
	}()
	wm.TickClusterHealth(done)
}
func TestWebsocketManager_Info(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	plMock := application.NewMockPluginManger(m)
	wsMock := application.NewMockWsSender(m)
	sub := application.NewMockPubSub(m)

	plMock.EXPECT().Ws().Return(wsMock)
	wsMock.EXPECT().New(gomock.Any(), gomock.Any()).Return(sub)
	sub.EXPECT().Info().Return(nil)
	sub.EXPECT().Close()

	wm := &WebsocketManager{
		pl: plMock,
	}

	writer := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/info", nil)

	wm.Info(writer, request)

	assert.Equal(t, http.StatusOK, writer.Code)
	assert.Equal(t, "application/json", writer.Header().Get("Content-Type"))
}

func TestWebsocketManager_Shutdown(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	c := counter.NewCounter()
	ctx := context.TODO()

	wm := &WebsocketManager{
		counter: c,
	}
	err := wm.Shutdown(ctx)
	assert.NoError(t, err)
}
func TestWebsocketManager_dispatchEvent(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	wm := &WebsocketManager{
		logger: mlog.NewForConfig(nil),
	}

	conn := &WsConn{
		user: &auth.UserInfo{},
	}

	// Test case: valid event type
	wsRequest := &websocket_pb.WsRequestMetadata{
		Type: websocket_pb.Type_HandleAuthorize,
	}
	message := []byte("testMessage")
	wm.handlers = map[websocket_pb.Type]HandleRequestFunc{
		websocket_pb.Type_HandleAuthorize: func(ctx context.Context, c Conn, ty websocket_pb.Type, message []byte) {
			assert.Equal(t, conn, c)
			assert.Equal(t, websocket_pb.Type_HandleAuthorize, ty)
			assert.Equal(t, message, message)
		},
	}
	wm.dispatchEvent(context.TODO(), conn, wsRequest, message)

	// Test case: invalid event type
	wsRequest = &websocket_pb.WsRequestMetadata{
		Type: websocket_pb.Type(-1),
	}
	wm.dispatchEvent(context.TODO(), conn, wsRequest, message)

	// Test case: event type exists but user is not authorized
	wsRequest = &websocket_pb.WsRequestMetadata{
		Type: websocket_pb.Type_HandleAuthorize,
	}
	conn.user = nil
	wm.dispatchEvent(context.TODO(), conn, wsRequest, message)
}

func TestWebsocketManager_Input_error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	sub := application.NewMockPubSub(m)

	wm := &WebsocketManager{
		logger: mlog.NewForConfig(nil),
	}

	conn := &WsConn{
		pubSub: sub,
		user:   &auth.UserInfo{},
	}

	sub.EXPECT().ToSelf(gomock.Any()).MinTimes(1)
	wm.HandleAuthorize(context.TODO(), conn, WsAuthorize, []byte("invalid"))
	wm.HandleJoinRoom(context.TODO(), conn, ProjectPodEvent, []byte("invalid"))
	wm.HandleStartShell(context.TODO(), conn, websocket_pb.Type(0), []byte("invalid"))
	wm.HandleShellMessage(context.TODO(), conn, websocket_pb.Type(0), []byte("invalid"))
	wm.HandleCloseShell(context.TODO(), conn, websocket_pb.Type(0), []byte("invalid"))
	wm.HandleWsCancelDeploy(context.TODO(), conn, websocket_pb.Type(0), []byte("invalid"))
	wm.HandleWsCreateProject(context.TODO(), conn, websocket_pb.Type(0), []byte("invalid"))
	wm.HandleWsUpdateProject(context.TODO(), conn, websocket_pb.Type(0), []byte("invalid"))
}
