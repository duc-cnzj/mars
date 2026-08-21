package websocket

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/deploy"
	"github.com/duc-cnzj/mars/v6/internal/locker"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/counter"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func TestWebsocketManager_HandleAuthorize(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	authMock := biz.NewMockAuthBiz(m)
	authMock.EXPECT().VerifyToken(gomock.Any(), "validToken").Return(&biz.UserInfo{Name: "testUser"}, nil)
	authMock.EXPECT().VerifyToken(gomock.Any(), "invalidToken").Return(nil, errors.New("invalid"))

	wm := &websocketManager{
		authBiz: authMock,
		logger:  mlog.NewForConfig(nil),
	}

	conn := &wsConn{}
	var inputv = websocket_pb.AuthorizeTokenInput{
		Token: "validToken",
	}
	marshalv, _ := proto.Marshal(&inputv)
	wm.HandleAuthorize(context.TODO(), conn, WsAuthorize, marshalv)
	assert.Equal(t, "testUser", conn.GetUser().Name)

	// 无效 token：与 master 一致静默失败，不发送任何帧，GetUser 保持 nil。
	conn = &wsConn{}
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

	wsMock := app.NewMockPubSub(m)
	wsMock.EXPECT().Join(int64(2))
	wsMock.EXPECT().Leave(int64(1), int64(2))

	// Join 走 RequireProjectAccess（projBiz.Show + nsRepo.Show + CanAccessNamespace）：
	// 公开命名空间项目任意登录用户放行；Leave 无鉴权直接退订。
	projBiz := biz.NewMockProjectBiz(m)
	nsBiz := biz.NewMockNamespaceBiz(m)
	projBiz.EXPECT().Show(gomock.Any(), 2).Return(&biz.Project{ID: 2, NamespaceID: 1}, nil)
	nsBiz.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Name: "ns", Private: false}, nil)

	conn := &wsConn{
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

	wm := &websocketManager{
		logger:    mlog.NewForConfig(nil),
		accessBiz: biz.NewAccessBiz(nsBiz, projBiz),
	}
	ctx := biz.SetUser(context.TODO(), &biz.UserInfo{ID: "1", Name: "u", Email: "u@mars.com"})
	wm.HandleJoinRoom(ctx, conn, ProjectPodEvent, marshal)

	wm.HandleJoinRoom(ctx, conn, ProjectPodEvent, marshal2)
}

func TestWebsocketManager_HandleStartShell(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	// StartShell 前先过命名空间级访问门卫（RequireNamespaceAccessByName）：
	// FindByName 返回公开命名空间放行，随后 StartShell 才对非规范 sessionID 报错。
	nsBiz := biz.NewMockNamespaceBiz(m)
	nsBiz.EXPECT().FindByName(gomock.Any(), "testNamespace").Return(&biz.Namespace{Name: "testNamespace", Private: false}, nil)

	wm := &websocketManager{
		logger:    mlog.NewForConfig(nil),
		accessBiz: biz.NewAccessBiz(nsBiz, nil),
	}
	sub := app.NewMockPubSub(m)
	conn := &wsConn{pubSub: sub, id: "testConnID", uid: "testConnUID"}

	input := &websocket_pb.WsHandleExecShellInput{
		SessionId: "testSession",
		Container: &websocket_pb.Container{
			Namespace: "testNamespace",
			Pod:       "testPod",
			Container: "testContainer",
		},
	}
	marshal, _ := proto.Marshal(input)
	sub.EXPECT().ToSelf(&wsResponse{
		Metadata: &websocket_pb.Metadata{
			Type:    WsHandleExecShell,
			Result:  deploy.ResultError,
			End:     true,
			Uid:     "testConnUID",
			Id:      "testConnID",
			Message: "invalid session sessionID, must format: '<namespace>-<pod>-<container>:<randomID>', input: 'testSession'",
		},
	})
	ctx := biz.SetUser(context.TODO(), &biz.UserInfo{ID: "1", Name: "u", Email: "u@mars.com"})
	wm.HandleStartShell(ctx, conn, WsHandleExecShell, marshal)
}

func TestWebsocketManager_HandleShellMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	wm := &websocketManager{
		logger: mlog.NewForConfig(nil),
	}
	conn := &wsConn{
		sessions: NewSessionMap(wm.logger),
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
	conn.sessions.Set("testSession", handler)
	marshal, _ := proto.Marshal(input)
	wm.HandleShellMessage(context.TODO(), conn, WsHandleExecShellMsg, marshal)
}

func TestWebsocketManager_HandleCloseShell(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	wm := &websocketManager{
		logger: mlog.NewForConfig(nil),
	}
	conn := &wsConn{
		sessions: NewSessionMap(wm.logger),
	}

	input := &websocket_pb.TerminalMessageInput{
		Message: &websocket_pb.TerminalMessage{
			SessionId: "testSession",
		},
	}
	handler := NewMockPtyHandler(m)
	handler.EXPECT().Close(gomock.Not(nil), gomock.Any())
	conn.sessions.Set("testSession", handler)
	marshal, _ := proto.Marshal(input)
	wm.HandleCloseShell(context.TODO(), conn, WsHandleCloseShell, marshal)
	time.Sleep(1 * time.Second)
}

func TestWebsocketManager_HandleCancelDeploy(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	nsRepo := data.NewMockNamespaceRepo(m)
	eventRepo := data.NewMockEventRepo(m)
	wm := &websocketManager{nsRepo: nsRepo, logger: mlog.NewForConfig(nil), eventRepo: eventRepo}
	conn := &wsConn{
		taskManager: NewTaskManager(wm.logger),
		user:        &biz.UserInfo{},
	}

	input := &websocket_pb.CancelInput{
		NamespaceId: 1,
		Name:        "testProject",
	}
	called := false
	conn.taskManager.Register(deploy.GetSlugName(input.NamespaceId, input.Name), func(err error) {
		called = true
	})
	marshal, _ := proto.Marshal(input)
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{}, nil)
	eventRepo.EXPECT().AuditLog(types.EventActionType_CancelDeploy, "", "", gomock.Any())
	wm.HandleCancelDeploy(context.TODO(), conn, WsCancel, marshal)
	assert.True(t, called)
}

func TestWebsocketManager_HandleCreateProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	eventRepo := data.NewMockEventRepo(m)
	nsBiz := biz.NewMockNamespaceBiz(m)
	repoBiz := biz.NewMockRepoBiz(m)
	gitBiz := biz.NewMockGitBiz(m)
	projBiz := biz.NewMockProjectBiz(m)
	jb := deploy.NewMockJobManager(m)
	wm := &websocketManager{
		logger:     mlog.NewForConfig(nil),
		accessBiz:  biz.NewAccessBiz(nsBiz, nil),
		eventRepo:  eventRepo,
		jobManager: jb,
		repoBiz:    repoBiz,
		gitBiz:     gitBiz,
		projBiz:    projBiz,
	}
	job := deploy.NewMockJob(m)
	conn := &wsConn{
		taskManager: NewTaskManager(wm.logger),
		user:        &biz.UserInfo{},
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
	// 空 name 缺省（仓库名）由 deploy.ApplyProject 收敛，ws 层只透传原始 name，
	// 故 repoBiz.Get 仅 ApplyProject 内部调用一次。
	repoBiz.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{
		Name: "app",
		ID:   2,
	}, nil).Times(1)
	// ApplyProject 命名空间访问校验：公开空间放行。
	nsBiz.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Name: "ns-1", Private: false}, nil)
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
		jobInput := x.(*deploy.JobInput)
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
	wm.HandleCreateProject(context.TODO(), conn, WsCreateProject, marshal)
}

func TestWebsocketManager_HandleUpdateProject(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	eventRepo := data.NewMockEventRepo(m)
	nsBiz := biz.NewMockNamespaceBiz(m)
	repoBiz := biz.NewMockRepoBiz(m)
	gitBiz := biz.NewMockGitBiz(m)
	projBiz := biz.NewMockProjectBiz(m)
	jb := deploy.NewMockJobManager(m)
	wm := &websocketManager{
		logger:     mlog.NewForConfig(nil),
		accessBiz:  biz.NewAccessBiz(nsBiz, projBiz),
		eventRepo:  eventRepo,
		jobManager: jb,
		repoBiz:    repoBiz,
		gitBiz:     gitBiz,
		projBiz:    projBiz,
		config:     &config.Config{},
	}
	job := deploy.NewMockJob(m)
	conn := &wsConn{
		taskManager: NewTaskManager(wm.logger),
		user:        &biz.UserInfo{},
	}

	input := &websocket_pb.UpdateProjectInput{
		ProjectId:   1,
		GitBranch:   "master",
		GitCommit:   "testCommit",
		Config:      "testConfig",
		ExtraValues: []*websocket_pb.ExtraValue{},
		Atomic:      lo.ToPtr(true),
	}
	// 更新前先过 RequireProjectAccess（Show + 所属命名空间校验），ctx 需带用户。
	projBiz.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{
		Name:        "appa",
		NamespaceID: 1,
		RepoID:      1,
	}, nil)
	// 命名空间校验两次：handler 的 RequireProjectAccess + ApplyProject 入口。
	nsBiz.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Name: "ns-1", Private: false}, nil).Times(2)
	repoBiz.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{Name: "appa", ID: 1, NeedGitRepo: false}, nil)
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
		jobInput := x.(*deploy.JobInput)
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
	ctx := biz.SetUser(context.TODO(), &biz.UserInfo{ID: "1", Name: "u", Email: "u@mars.com"})
	wm.HandleUpdateProject(ctx, conn, WsUpdateProject, marshal)
}

func TestNewWebsocketManager(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	logger := mlog.NewForConfig(nil)
	counter := counter.NewCounter()
	projBiz := biz.NewMockProjectBiz(m)
	repoBiz := biz.NewMockRepoBiz(m)
	gitBiz := biz.NewMockGitBiz(m)
	nsBiz := biz.NewMockNamespaceBiz(m)
	accessBiz := biz.NewAccessBiz(nsBiz, nil)
	nsRepo := data.NewMockNamespaceRepo(m)
	jobManager := deploy.NewMockJobManager(m)
	pl := app.NewMockPluginManager(m)
	authBiz := biz.NewMockAuthBiz(m)
	locker := locker.NewMockLocker(m)
	clusterRepo := data.NewMockK8sRepo(m)
	eventRepo := data.NewMockEventRepo(m)
	fileRepo := data.NewMockFileRepo(m)

	cfg := &config.Config{}

	wm := NewWebsocketManager(WebsocketManagerDeps{
		Timer:         timer.NewReal(),
		Logger:        logger,
		Counter:       counter,
		ProjBiz:       projBiz,
		RepoBiz:       repoBiz,
		GitBiz:        gitBiz,
		NsRepo:        nsRepo,
		AccessBiz:     accessBiz,
		JobManager:    jobManager,
		Config:        cfg,
		PluginManager: pl,
		AuthBiz:       authBiz,
		Locker:        locker,
		ClusterRepo:   clusterRepo,
		EventRepo:     eventRepo,
		FileRepo:      fileRepo,
	}).(*websocketManager)

	assert.NotNil(t, wm)
	assert.Equal(t, logger.WithModule("socket/websocket"), wm.logger)
	assert.Equal(t, counter, wm.counter)
	assert.Equal(t, projBiz, wm.projBiz)
	assert.Equal(t, repoBiz, wm.repoBiz)
	assert.Equal(t, gitBiz, wm.gitBiz)
	assert.Equal(t, accessBiz, wm.accessBiz)
	assert.Equal(t, nsRepo, wm.nsRepo)
	assert.Equal(t, jobManager, wm.jobManager)
	assert.Equal(t, fileRepo, wm.fileRepo)
	assert.Equal(t, cfg, wm.config)
	assert.Equal(t, pl, wm.pluginManager)
	assert.Equal(t, authBiz, wm.authBiz)
	assert.Equal(t, locker, wm.locker)
	assert.Equal(t, clusterRepo, wm.k8sRepo)
	assert.Equal(t, eventRepo, wm.eventRepo)
	assert.NotNil(t, wm.timer)
	assert.Len(t, wm.handlers, 8)
}
func TestWebsocketManager_TickClusterHealth(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	lockerMock := locker.NewMockLocker(m)
	plMock := app.NewMockPluginManager(m)
	wsMock := app.NewMockWsSender(m)
	k8sRepoMock := data.NewMockK8sRepo(m)
	loggerMock := mlog.NewForConfig(nil)

	lockerMock.EXPECT().Acquire("TickClusterHealth", int64(5)).Return(true)
	lockerMock.EXPECT().Release("TickClusterHealth")
	plMock.EXPECT().Ws().Return(wsMock)
	sub := app.NewMockPubSub(m)
	wsMock.EXPECT().New(gomock.Any(), gomock.Any()).Return(sub)
	sub.EXPECT().Close()
	sub.EXPECT().ToAll(gomock.Any())
	k8sRepoMock.EXPECT().ClusterInfo().Return(&biz.ClusterInfo{})

	wm := &websocketManager{
		locker:             lockerMock,
		pluginManager:      plMock,
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

	plMock := app.NewMockPluginManager(m)
	wsMock := app.NewMockWsSender(m)
	sub := app.NewMockPubSub(m)

	plMock.EXPECT().Ws().Return(wsMock)
	wsMock.EXPECT().New(gomock.Any(), gomock.Any()).Return(sub)
	sub.EXPECT().Info().Return(nil)
	sub.EXPECT().Close()

	wm := &websocketManager{
		pluginManager: plMock,
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

	wm := &websocketManager{
		counter: c,
	}
	err := wm.Shutdown(ctx)
	assert.NoError(t, err)
}
func TestWebsocketManager_dispatchEvent(t *testing.T) {
	// 必须设 timer：dispatchEvent 的 defer 参数 wc.timer.Now() 对 nil timer 会 panic，被
	// HandlePanicWithCallback 静默吞掉 → handler 不执行 = 测试空转。删不得。
	t.Run("valid type calls handler", func(t *testing.T) {
		wm := &websocketManager{logger: mlog.NewForConfig(nil), timer: timer.NewReal()}
		conn := &wsConn{user: &biz.UserInfo{}}
		called := false
		wm.handlers = map[websocket_pb.Type]HandleRequestFunc{
			websocket_pb.Type_HandleAuthorize: func(ctx context.Context, c Conn, ty websocket_pb.Type, message []byte) {
				called = true
				assert.Equal(t, conn, c)
				assert.Equal(t, websocket_pb.Type_HandleAuthorize, ty)
				assert.Equal(t, []byte("x"), message)
			},
		}
		wsRequest := &websocket_pb.WsRequestMetadata{Type: websocket_pb.Type_HandleAuthorize}
		wm.dispatchEvent(context.TODO(), conn, wsRequest, []byte("x"))
		assert.True(t, called)
	})

	t.Run("unknown type skipped", func(t *testing.T) {
		wm := &websocketManager{logger: mlog.NewForConfig(nil), timer: timer.NewReal()}
		wm.handlers = map[websocket_pb.Type]HandleRequestFunc{}
		conn := &wsConn{user: &biz.UserInfo{}}
		wm.dispatchEvent(context.TODO(), conn, &websocket_pb.WsRequestMetadata{Type: websocket_pb.Type(-1)}, []byte("x"))
	})

	t.Run("user not authorized sends pending frame", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		wm := &websocketManager{logger: mlog.NewForConfig(nil), timer: timer.NewReal()}
		wsMock := app.NewMockPubSub(m)
		// 未认证：发"认证中，请稍等~"帧，handler 不执行
		wsMock.EXPECT().ToSelf(gomock.Any())
		conn := &wsConn{pubSub: wsMock} // 未设置用户
		called := false
		wm.handlers = map[websocket_pb.Type]HandleRequestFunc{
			websocket_pb.Type_CreateProject: func(ctx context.Context, c Conn, ty websocket_pb.Type, message []byte) {
				called = true
			},
		}
		wm.dispatchEvent(context.TODO(), conn, &websocket_pb.WsRequestMetadata{Type: websocket_pb.Type_CreateProject}, []byte("x"))
		assert.False(t, called)
	})

	t.Run("handler panic recovered internally", func(t *testing.T) {
		wm := &websocketManager{logger: mlog.NewForConfig(nil), timer: timer.NewReal()}
		conn := &wsConn{user: &biz.UserInfo{}}
		wm.handlers = map[websocket_pb.Type]HandleRequestFunc{
			websocket_pb.Type_CreateProject: func(ctx context.Context, c Conn, ty websocket_pb.Type, message []byte) {
				panic("boom")
			},
		}
		// 双层 recover（metrics 计数 + HandlePanicWithCallback）吞掉，不向外传播
		assert.NotPanics(t, func() {
			wm.dispatchEvent(context.TODO(), conn, &websocket_pb.WsRequestMetadata{Type: websocket_pb.Type_CreateProject}, []byte("x"))
		})
	})
}

func TestWebsocketManager_Input_error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	sub := app.NewMockPubSub(m)

	wm := &websocketManager{
		logger: mlog.NewForConfig(nil),
	}

	conn := &wsConn{
		pubSub: sub,
		user:   &biz.UserInfo{},
	}

	sub.EXPECT().ToSelf(gomock.Any()).MinTimes(1)
	wm.HandleAuthorize(context.TODO(), conn, WsAuthorize, []byte("invalid"))
	wm.HandleJoinRoom(context.TODO(), conn, ProjectPodEvent, []byte("invalid"))
	wm.HandleStartShell(context.TODO(), conn, websocket_pb.Type(0), []byte("invalid"))
	wm.HandleShellMessage(context.TODO(), conn, websocket_pb.Type(0), []byte("invalid"))
	wm.HandleCloseShell(context.TODO(), conn, websocket_pb.Type(0), []byte("invalid"))
	wm.HandleCancelDeploy(context.TODO(), conn, websocket_pb.Type(0), []byte("invalid"))
	wm.HandleCreateProject(context.TODO(), conn, websocket_pb.Type(0), []byte("invalid"))
	wm.HandleUpdateProject(context.TODO(), conn, websocket_pb.Type(0), []byte("invalid"))
}
