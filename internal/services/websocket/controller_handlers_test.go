package websocket

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/deploy"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

// 本文件覆盖 handler 侧的错误/成功分支（controller.go）与出站消息适配器 messageSender：
// 配合 controller_test.go + controller_ws_test.go 将非 mock 生产函数覆盖补到 100%。

func TestWebsocketManager_HandleJoinRoom_errors(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	wsMock := app.NewMockPubSub(m)
	wsMock.EXPECT().Join(int64(2)).Return(errBoom)
	wsMock.EXPECT().Leave(int64(1), int64(2)).Return(errBoom)

	// Join 先过 RequireProjectAccess（公开命名空间放行）再触达 pubsub，pubsub 出错打日志。
	projBiz := biz.NewMockProjectBiz(m)
	nsBiz := biz.NewMockNamespaceBiz(m)
	projBiz.EXPECT().Show(gomock.Any(), 2).Return(&biz.Project{ID: 2, NamespaceID: 1}, nil)
	nsBiz.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Name: "ns", Private: false}, nil)

	wm := &websocketManager{
		logger:    mlog.NewForConfig(nil),
		accessBiz: biz.NewAccessBiz(nsBiz, projBiz),
	}
	conn := &wsConn{pubSub: wsMock}
	ctx := biz.SetUser(context.TODO(), &biz.UserInfo{ID: "1", Name: "u", Email: "u@mars.com"})

	joinMsg, _ := proto.Marshal(&websocket_pb.ProjectPodEventJoinInput{Type: ProjectPodEvent, Join: true, ProjectId: 2})
	wm.HandleJoinRoom(ctx, conn, ProjectPodEvent, joinMsg)

	leaveMsg, _ := proto.Marshal(&websocket_pb.ProjectPodEventJoinInput{Type: ProjectPodEvent, Join: false, NamespaceId: 1, ProjectId: 2})
	wm.HandleJoinRoom(ctx, conn, ProjectPodEvent, leaveMsg)
}

// 越权回归：普通用户 join 私有命名空间项目被 RequireProjectAccess 拒绝，
// 回错误帧且不触达 PubSub.Join——订阅 Pod 事件流不得泄露私有项目动态。
func TestWebsocketManager_HandleJoinRoom_denied(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	projBiz := biz.NewMockProjectBiz(m)
	nsBiz := biz.NewMockNamespaceBiz(m)
	projBiz.EXPECT().Show(gomock.Any(), 2).Return(&biz.Project{ID: 2, NamespaceID: 1}, nil)
	// 私有命名空间：仅 admin/创建者/成员可见，普通用户 u 无权访问。
	nsBiz.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Name: "ns", Private: true, CreatorEmail: "owner@mars.com"}, nil)

	sub := app.NewMockPubSub(m)
	// 越权拦截 → SendEndError 回错误帧（ToSelf）；不设 Join 期望：若实现误调
	// pubsub.Join，gomock 以"未期望调用"中止——回归防护，与 Clone 的 Get-先于-Clone 同理。
	sub.EXPECT().ToSelf(gomock.Any())

	conn := &wsConn{pubSub: sub, user: &biz.UserInfo{Name: "u"}}
	wm := &websocketManager{
		logger:    mlog.NewForConfig(nil),
		accessBiz: biz.NewAccessBiz(nsBiz, projBiz),
	}
	ctx := biz.SetUser(context.TODO(), &biz.UserInfo{ID: "2", Name: "u", Email: "u@mars.com"})
	joinMsg, _ := proto.Marshal(&websocket_pb.ProjectPodEventJoinInput{Type: ProjectPodEvent, Join: true, ProjectId: 2})
	wm.HandleJoinRoom(ctx, conn, ProjectPodEvent, joinMsg)
}

func TestWebsocketManager_HandleStartShell_success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	// StartShell 前先过命名空间级访问门卫（RequireNamespaceAccessByName）：
	// FindByName 返回公开命名空间放行，测试聚焦 StartShell 成功路径本身。
	nsBiz := biz.NewMockNamespaceBiz(m)
	nsBiz.EXPECT().FindByName(gomock.Any(), "testNamespace").Return(&biz.Namespace{Name: "testNamespace", Private: false}, nil)

	// StartShell 内部把真实 ptyHandler 存入 sessionMap；runTerminal 经 Get 取回。
	// 用 MockSessionMapper 拦截：Get 返回 mock pty，ClosePty 走 mock，避免真实 ptyHandler
	// 的 Close 里 200ms sleep + recorder/eventRepo 副作用。
	sm := NewMockSessionMapper(m)
	sm.EXPECT().Set(gomock.Any(), gomock.Any())
	mockPty := NewMockPtyHandler(m)
	sm.EXPECT().Get(gomock.Any()).Return(mockPty, true)
	closed := make(chan struct{})
	sm.EXPECT().Close(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(context.Context, string, uint32, string) {
		close(closed)
	})

	// shell="" → 尝试 validShells，bash 首轮成功即退出循环。
	mockPty.EXPECT().IsClosed().Return(false).AnyTimes()
	mockPty.EXPECT().SetShell(gomock.Any()).AnyTimes()

	fileRepo := data.NewMockFileRepo(m)
	fileRepo.EXPECT().NewRecorder(gomock.Any(), gomock.Any()).Return(data.NewMockRecorder(m))

	k8sRepo := data.NewMockK8sRepo(m)
	k8sRepo.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	sub := app.NewMockPubSub(m)
	// HandleStartShell 成功帧 WsHandleShellResponse → ToSelf。
	sub.EXPECT().ToSelf(gomock.Any())

	wm := &websocketManager{logger: mlog.NewForConfig(nil), fileRepo: fileRepo, k8sRepo: k8sRepo, accessBiz: biz.NewAccessBiz(nsBiz, nil)}
	conn := &wsConn{
		pubSub:   sub,
		id:       "connID",
		uid:      "connUID",
		sessions: sm,
		user:     &biz.UserInfo{Name: "user"},
	}

	input := &websocket_pb.WsHandleExecShellInput{
		SessionId: "testNamespace-testPod-testContainer:abc",
		Container: &websocket_pb.Container{Namespace: "testNamespace", Pod: "testPod", Container: "testContainer"},
	}
	marshal, _ := proto.Marshal(input)
	ctx := biz.SetUser(context.TODO(), &biz.UserInfo{ID: "1", Name: "user", Email: "user@mars.com"})
	wm.HandleStartShell(ctx, conn, WsHandleExecShell, marshal)

	select {
	case <-closed:
	case <-time.After(3 * time.Second):
		t.Fatal("runTerminal did not close pty")
	}
}

// 越权回归：普通用户对私有命名空间的容器启动 shell 被 RequireNamespaceAccessByName
// 拒绝，回错误帧且不触达 StartShell——交互终端可执行任意命令，不得绕过命名空间访问控制
// 进入私有命名空间容器（RCE 级越权，与 gRPC container.Exec 对齐）。
func TestWebsocketManager_HandleStartShell_denied(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	nsBiz := biz.NewMockNamespaceBiz(m)
	// 私有命名空间：仅 admin/创建者/成员可见，普通用户 u 无权访问。
	nsBiz.EXPECT().FindByName(gomock.Any(), "testNamespace").
		Return(&biz.Namespace{Name: "testNamespace", Private: true, CreatorEmail: "owner@mars.com"}, nil)

	sub := app.NewMockPubSub(m)
	// 越权拦截 → SendEndError 回错误帧（ToSelf）；不设 sessionMap/k8s 期望：
	// 若实现误调 StartShell，gomock 以"未期望调用"中止——回归防护。
	sub.EXPECT().ToSelf(gomock.Any())

	conn := &wsConn{pubSub: sub, id: "connID", uid: "connUID"}
	wm := &websocketManager{
		logger:    mlog.NewForConfig(nil),
		accessBiz: biz.NewAccessBiz(nsBiz, nil),
	}
	ctx := biz.SetUser(context.TODO(), &biz.UserInfo{ID: "2", Name: "u", Email: "u@mars.com"})
	input := &websocket_pb.WsHandleExecShellInput{
		SessionId: "testNamespace-testPod-testContainer:abc",
		Container: &websocket_pb.Container{Namespace: "testNamespace", Pod: "testPod", Container: "testContainer"},
	}
	marshal, _ := proto.Marshal(input)
	wm.HandleStartShell(ctx, conn, WsHandleExecShell, marshal)
}

func TestWebsocketManager_HandleCreateProject_installError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	nsBiz := biz.NewMockNamespaceBiz(m)
	repoBiz := biz.NewMockRepoBiz(m)
	gitBiz := biz.NewMockGitBiz(m)
	projBiz := biz.NewMockProjectBiz(m)
	jb := deploy.NewMockJobManager(m)
	wm := &websocketManager{logger: mlog.NewForConfig(nil), accessBiz: biz.NewAccessBiz(nsBiz, nil), jobManager: jb, repoBiz: repoBiz, gitBiz: gitBiz, projBiz: projBiz}

	job := deploy.NewMockJob(m)
	conn := &wsConn{taskManager: NewTaskManager(wm.logger), user: &biz.UserInfo{}}

	input := &websocket_pb.CreateProjectInput{Type: WsCreateProject, NamespaceId: 1, RepoId: 1}
	// 名缺省由 deploy.ApplyProject 收敛，ws 层只透传原始 name，repoBiz.Get 仅 ApplyProject 调用一次。
	repoBiz.EXPECT().Get(gomock.Any(), 1).Return(&biz.Repo{Name: "app", ID: 2}, nil).Times(1)
	nsBiz.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Name: "ns-1", Private: false}, nil)
	jb.EXPECT().NewJob(gomock.Any()).Return(job)
	job.EXPECT().ID().Return("jobID")
	job.EXPECT().OnFinally(gomock.Any(), gomock.Any())
	job.EXPECT().GlobalLock().Return(job)
	job.EXPECT().Validate().Return(job)
	job.EXPECT().LoadConfigs().Return(job)
	job.EXPECT().Run(gomock.Any()).Return(job)
	job.EXPECT().Finish().Return(job)
	// 流水线失败 → installProject 返回 err → HandleCreateProject 打日志（不吞错）。
	job.EXPECT().Error().Return(errBoom)

	marshal, _ := proto.Marshal(input)
	wm.HandleCreateProject(context.TODO(), conn, WsCreateProject, marshal)
}

func TestWebsocketManager_HandleUpdateProject_showError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	projBiz := biz.NewMockProjectBiz(m)
	projBiz.EXPECT().Show(gomock.Any(), 1).Return(nil, errBoom)

	sub := app.NewMockPubSub(m)
	sub.EXPECT().ToSelf(gomock.Any())

	// RequireProjectAccess 先经 projBiz.Show 加载项目，失败即回错误帧（未触达 nsRepo）。
	wm := &websocketManager{logger: mlog.NewForConfig(nil), accessBiz: biz.NewAccessBiz(nil, projBiz), projBiz: projBiz}
	conn := &wsConn{pubSub: sub, user: &biz.UserInfo{}}

	input := &websocket_pb.UpdateProjectInput{ProjectId: 1}
	marshal, _ := proto.Marshal(input)
	wm.HandleUpdateProject(context.TODO(), conn, WsUpdateProject, marshal)
}

// 越权回归：普通用户更新私有命名空间项目被 RequireProjectAccess 拒绝，
// 回错误帧且不触达部署流水线——更新帧不得用于探测或操作私有项目（与 HandleJoinRoom 对齐）。
func TestWebsocketManager_HandleUpdateProject_denied(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	projBiz := biz.NewMockProjectBiz(m)
	projBiz.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{ID: 1, NamespaceID: 1, Name: "app", RepoID: 2}, nil)
	// 私有命名空间：仅 admin/创建者/成员可见，普通用户 u 无权访问。
	nsBiz := biz.NewMockNamespaceBiz(m)
	nsBiz.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Name: "ns", Private: true, CreatorEmail: "owner@mars.com"}, nil)

	sub := app.NewMockPubSub(m)
	// 越权拦截 → SendEndError 回错误帧（ToSelf）；jobManager 恒 nil——若实现误触达
	// 部署流水线会 panic，回归防护"未授权绝不下发 Job"。
	sub.EXPECT().ToSelf(gomock.Any())

	conn := &wsConn{pubSub: sub, user: &biz.UserInfo{}}
	wm := &websocketManager{
		logger:    mlog.NewForConfig(nil),
		accessBiz: biz.NewAccessBiz(nsBiz, projBiz),
		projBiz:   projBiz,
	}
	ctx := biz.SetUser(context.TODO(), &biz.UserInfo{ID: "2", Name: "u", Email: "u@mars.com"})
	input := &websocket_pb.UpdateProjectInput{ProjectId: 1}
	marshal, _ := proto.Marshal(input)
	wm.HandleUpdateProject(ctx, conn, WsUpdateProject, marshal)
}

func TestWebsocketManager_HandleUpdateProject_installError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	projBiz := biz.NewMockProjectBiz(m)
	projBiz.EXPECT().Show(gomock.Any(), 1).Return(&biz.Project{ID: 1, NamespaceID: 1, Name: "app", RepoID: 2}, nil)

	nsBiz := biz.NewMockNamespaceBiz(m)
	repoBiz := biz.NewMockRepoBiz(m)
	gitBiz := biz.NewMockGitBiz(m)
	jb := deploy.NewMockJobManager(m)
	job := deploy.NewMockJob(m)
	jb.EXPECT().NewJob(gomock.Any()).Return(job)
	job.EXPECT().ID().Return("jobID")
	job.EXPECT().OnFinally(gomock.Any(), gomock.Any())
	job.EXPECT().GlobalLock().Return(job)
	job.EXPECT().Validate().Return(job)
	job.EXPECT().LoadConfigs().Return(job)
	job.EXPECT().Run(gomock.Any()).Return(job)
	job.EXPECT().Finish().Return(job)
	// 流水线失败 → installProject 返回 err → HandleUpdateProject 打日志（不吞错），与 create 分支一致。
	job.EXPECT().Error().Return(errBoom)
	// RequireProjectAccess（projBiz.Show + 所属命名空间校验）+ ApplyProject 访问校验 + 仓库取回（NeedGitRepo=false 无需 git ensure）。
	nsBiz.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Name: "ns-1", Private: false}, nil).Times(2)
	repoBiz.EXPECT().Get(gomock.Any(), 2).Return(&biz.Repo{Name: "app", ID: 2, NeedGitRepo: false}, nil)

	wm := &websocketManager{logger: mlog.NewForConfig(nil), accessBiz: biz.NewAccessBiz(nsBiz, projBiz), projBiz: projBiz, jobManager: jb, config: &config.Config{InstallTimeout: 30 * time.Second}, repoBiz: repoBiz, gitBiz: gitBiz}
	conn := &wsConn{taskManager: NewTaskManager(wm.logger), user: &biz.UserInfo{}}

	input := &websocket_pb.UpdateProjectInput{ProjectId: 1}
	marshal, _ := proto.Marshal(input)
	ctx := biz.SetUser(context.TODO(), &biz.UserInfo{ID: "1", Name: "u", Email: "u@mars.com"})
	wm.HandleUpdateProject(ctx, conn, WsUpdateProject, marshal)
}

func TestWebsocketManager_HandleCancelDeploy_showError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	nsRepo := data.NewMockNamespaceRepo(m)
	nsRepo.EXPECT().Show(gomock.Any(), 1).Return(nil, errBoom)

	wm := &websocketManager{logger: mlog.NewForConfig(nil), nsRepo: nsRepo}
	conn := &wsConn{taskManager: NewTaskManager(wm.logger), user: &biz.UserInfo{}}

	input := &websocket_pb.CancelInput{NamespaceId: 1, Name: "app"}
	// 预注册任务 → RunCancelDeployTask 成功 → 进入 nsRepo.Show 分支，Show 失败走日志+return。
	assert.NoError(t, conn.taskManager.Register(deploy.GetSlugName(1, "app"), func(error) {}))

	marshal, _ := proto.Marshal(input)
	wm.HandleCancelDeploy(context.TODO(), conn, WsCancel, marshal)
}

func TestWebsocketManager_installProject_addCancelDeployTaskError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	nsBiz := biz.NewMockNamespaceBiz(m)
	repoBiz := biz.NewMockRepoBiz(m)
	gitBiz := biz.NewMockGitBiz(m)
	projBiz := biz.NewMockProjectBiz(m)
	jb := deploy.NewMockJobManager(m)
	job := deploy.NewMockJob(m)
	jb.EXPECT().NewJob(gomock.Any()).Return(job)
	job.EXPECT().ID().Return("dupID").AnyTimes()
	// ApplyProject 前置：匿名守卫放行 + 访问校验 + 仓库取回。
	nsBiz.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Name: "ns-1", Private: false}, nil)
	repoBiz.EXPECT().Get(gomock.Any(), 0).Return(&biz.Repo{Name: "app", NeedGitRepo: false}, nil).AnyTimes()

	sub := app.NewMockPubSub(m)
	sub.EXPECT().ToSelf(gomock.Cond(func(x any) bool {
		resp, ok := x.(*websocket_pb.WsMetadataResponse)
		return ok &&
			resp.Metadata.Type == WsCreateProject &&
			resp.Metadata.Result == deploy.ResultDeployFailed &&
			resp.Metadata.Message == "正在清理中，请稍后再试。"
	}))

	wm := &websocketManager{logger: mlog.NewForConfig(nil), accessBiz: biz.NewAccessBiz(nsBiz, nil), jobManager: jb, repoBiz: repoBiz, gitBiz: gitBiz, projBiz: projBiz}
	conn := &wsConn{taskManager: NewTaskManager(wm.logger), pubSub: sub}
	// 预注册同名任务 → AddCancelDeployTask 返回 errSignalExists → 失败帧 + 提前返回。
	assert.NoError(t, conn.taskManager.Register("dupID", func(error) {}))

	err := wm.installProject(context.TODO(), conn, &deploy.JobInput{
		Type:        WsCreateProject,
		NamespaceId: 1,
		Name:        "app",
		User:        &biz.UserInfo{},
	})
	assert.NoError(t, err)
}

func TestWebsocketManager_installProject_onFinallyCallback(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	nsBiz := biz.NewMockNamespaceBiz(m)
	repoBiz := biz.NewMockRepoBiz(m)
	gitBiz := biz.NewMockGitBiz(m)
	projBiz := biz.NewMockProjectBiz(m)
	jb := deploy.NewMockJobManager(m)
	job := deploy.NewMockJob(m)
	jb.EXPECT().NewJob(gomock.Any()).Return(job)
	job.EXPECT().ID().Return("taskID").AnyTimes()
	// OnFinally Do 触发闭包：RemoveCancelDeployTask(job.ID()) + base()。
	job.EXPECT().OnFinally(gomock.Any(), gomock.Any()).Do(func(p int, fn func(err error, base func())) {
		fn(nil, func() {})
	})
	job.EXPECT().GlobalLock().Return(job)
	job.EXPECT().Validate().Return(job)
	job.EXPECT().LoadConfigs().Return(job)
	job.EXPECT().Run(gomock.Any()).Return(job)
	job.EXPECT().Finish().Return(job)
	job.EXPECT().Error().Return(nil)
	// ApplyProject 前置：匿名守卫放行 + 访问校验 + 仓库取回。
	nsBiz.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Name: "ns-1", Private: false}, nil)
	repoBiz.EXPECT().Get(gomock.Any(), 0).Return(&biz.Repo{Name: "app", NeedGitRepo: false}, nil).AnyTimes()

	wm := &websocketManager{logger: mlog.NewForConfig(nil), accessBiz: biz.NewAccessBiz(nsBiz, nil), jobManager: jb, repoBiz: repoBiz, gitBiz: gitBiz, projBiz: projBiz}
	conn := &wsConn{taskManager: NewTaskManager(wm.logger)}

	err := wm.installProject(context.TODO(), conn, &deploy.JobInput{
		Type:        WsCreateProject,
		NamespaceId: 1,
		Name:        "app",
		User:        &biz.UserInfo{},
	})
	assert.NoError(t, err)
	// AddCancelDeployTask 注册 + OnFinally 闭包 Remove，最终任务已清理。
	assert.False(t, conn.taskManager.Has("taskID"))
}
func TestMessageSenderFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	ms := newMessageSender(conn, "slug", websocket_pb.Type_SetUid)

	assert.NotNil(t, ms)
	assert.Equal(t, conn, ms.(*messageSender).conn)
	assert.Equal(t, "slug", ms.(*messageSender).slugName)
	assert.Equal(t, websocket_pb.Type_SetUid, ms.(*messageSender).wsType)
	assert.NotNil(t, ms.(*messageSender).percent)
}

func TestSendDeployedResultFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := app.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).Times(1)
	conn.EXPECT().UID().Return("uid").Times(1)
	conn.EXPECT().ID().Return("id").Times(1)
	sub.EXPECT().ToSelf(gomock.Any()).Do(func(msg app.WebsocketMessage) {
		res := msg.(*websocket_pb.WsMetadataResponse)
		assert.Equal(t, "slug", res.Metadata.Slug)
		assert.Equal(t, websocket_pb.Type_HandleAuthorize, res.Metadata.Type)
		assert.Equal(t, websocket_pb.ResultType_Deployed, res.Metadata.Result)
		assert.True(t, res.Metadata.End)
		assert.Equal(t, "uid", res.Metadata.Uid)
		assert.Equal(t, "id", res.Metadata.Id)
		assert.Equal(t, "message", res.Metadata.Message)
	})

	ms := newMessageSender(conn, "slug", websocket_pb.Type_HandleAuthorize)
	ms.SendDeployedResult(websocket_pb.ResultType_Deployed, "message", &types.ProjectModel{})
}

func TestSendEndErrorFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := app.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).Times(1)
	sub.EXPECT().ToSelf(gomock.Any()).Do(func(msg app.WebsocketMessage) {
		res := msg.(*websocket_pb.WsMetadataResponse)
		assert.Equal(t, websocket_pb.Type_HandleExecShellMsg, res.Metadata.Type)
		assert.Equal(t, deploy.ResultError, res.Metadata.Result)
		assert.True(t, res.Metadata.End)
		assert.Equal(t, "error", res.Metadata.Message)
	})
	conn.EXPECT().UID().Return("uid").Times(1)
	conn.EXPECT().ID().Return("id").Times(1)
	ms := newMessageSender(conn, "slug", websocket_pb.Type_HandleExecShellMsg)
	ms.SendEndError(errors.New("error"))
}

func TestSendProcessPercentFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := app.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).Times(1)
	sub.EXPECT().ToSelf(gomock.Any()).Do(func(msg app.WebsocketMessage) {
		res := msg.(*websocket_pb.WsMetadataResponse)
		assert.Equal(t, WsProcessPercent, res.Metadata.Type)
		assert.Equal(t, deploy.ResultSuccess, res.Metadata.Result)
		assert.False(t, res.Metadata.End)
		assert.Equal(t, int32(50), res.Metadata.Percent)
	})
	conn.EXPECT().UID().Return("uid").Times(1)
	conn.EXPECT().ID().Return("id").Times(1)

	ms := newMessageSender(conn, "slug", websocket_pb.Type_HandleExecShellMsg)
	ms.SendProcessPercent(50)
}

func TestSendMsgFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := app.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).Times(1)
	sub.EXPECT().ToSelf(gomock.Any()).Do(func(msg app.WebsocketMessage) {
		res := msg.(*websocket_pb.WsMetadataResponse)
		assert.Equal(t, websocket_pb.Type_HandleExecShellMsg, res.Metadata.Type)
		assert.Equal(t, deploy.ResultSuccess, res.Metadata.Result)
		assert.False(t, res.Metadata.End)
		assert.Equal(t, "message", res.Metadata.Message)
	})
	conn.EXPECT().UID().Return("uid").Times(1)
	conn.EXPECT().ID().Return("id").Times(1)

	ms := newMessageSender(conn, "slug", websocket_pb.Type_HandleExecShellMsg)
	ms.SendMsg("message")
}

func TestSendMsgWithContainerLogFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := app.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).Times(1)
	containers := []*websocket_pb.Container{{Namespace: "ns", Pod: "pod-1", Container: "app"}}
	sub.EXPECT().ToSelf(gomock.Any()).Do(func(msg app.WebsocketMessage) {
		res := msg.(*websocket_pb.WsWithContainerMessageResponse)
		assert.Equal(t, websocket_pb.Type_HandleExecShellMsg, res.Metadata.Type)
		assert.Equal(t, deploy.ResultLogWithContainers, res.Metadata.Result)
		assert.False(t, res.Metadata.End)
		assert.Equal(t, "message", res.Metadata.Message)
		assert.Len(t, res.Containers, 1)
		assert.Equal(t, "app", res.Containers[0].Container)
	})
	conn.EXPECT().UID().Return("uid").Times(1)
	conn.EXPECT().ID().Return("id").Times(1)

	ms := newMessageSender(conn, "slug", websocket_pb.Type_HandleExecShellMsg)
	ms.SendMsgWithContainerLog("message", containers)
}

func TestSendProtoMsgFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := app.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).Times(1)
	sub.EXPECT().ToSelf(gomock.Any()).Do(func(msg app.WebsocketMessage) {
		res := msg.(*websocket_pb.WsMetadataResponse)
		// SendProtoMsg 是直通语义：入参即出参，字段原样透传。
		assert.Equal(t, "passthrough", res.Metadata.Message)
	})
	ms := newMessageSender(conn, "slug", websocket_pb.Type_HandleExecShellMsg)
	ms.SendProtoMsg(&wsResponse{Metadata: &websocket_pb.Metadata{Message: "passthrough"}})
}

func TestProcessPercentAddFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := app.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).Times(1)
	sub.EXPECT().ToSelf(gomock.Any()).Do(func(msg app.WebsocketMessage) {
		res := msg.(*websocket_pb.WsMetadataResponse)
		assert.Equal(t, WsProcessPercent, res.Metadata.Type)
		assert.Equal(t, deploy.ResultSuccess, res.Metadata.Result)
		assert.Equal(t, int32(1), res.Metadata.Percent)
	})
	conn.EXPECT().UID().Return("uid").Times(1)
	conn.EXPECT().ID().Return("id").Times(1)

	ms := newMessageSender(conn, "slug", websocket_pb.Type_HandleExecShellMsg)
	ms.Add()
}

func TestProcessPercentToFunctionality(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := app.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).AnyTimes()
	sub.EXPECT().ToSelf(gomock.Any()).Times(2).Do(func(msg app.WebsocketMessage) {
		res := msg.(*websocket_pb.WsMetadataResponse)
		assert.Equal(t, WsProcessPercent, res.Metadata.Type)
		// To(3)：0→2 步进一帧 + 末尾精确对齐一帧，两帧 Percent 均为 2。
		assert.Equal(t, int32(2), res.Metadata.Percent)
	})
	conn.EXPECT().UID().Return("uid").AnyTimes()
	conn.EXPECT().ID().Return("id").AnyTimes()

	ms := newMessageSender(conn, "slug", websocket_pb.Type_HandleExecShellMsg)
	ms.To(3)
	assert.Equal(t, int64(3), ms.Current())
}

func Test_messageSender_Current(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	sub := app.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub).AnyTimes()
	sub.EXPECT().ToSelf(gomock.Any()).AnyTimes()
	conn.EXPECT().UID().Return("uid").AnyTimes()
	conn.EXPECT().ID().Return("id").AnyTimes()

	ms := newMessageSender(conn, "slug", websocket_pb.Type_HandleExecShellMsg)
	assert.Equal(t, int64(0), ms.Current())
	ms.Add()
	assert.Equal(t, int64(1), ms.Current())
}
