package services

// aiagent_test.go 覆盖 AI 状态/启用两个 HTTP 端点的全部分支：
// 状态查询（未配置/已禁用/正常/各类错误）、启用（鉴权/仓库校验/命名空间冲突/
// 幂等/仓库不符/首次部署/部署失败/回读失败），以及项目名派生与超管创建者判定。

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/deploy"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/locker"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/hasher"
	"github.com/gorilla/mux"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const aiAgentTestEmail = "alice@mars.local"

// aiAgentTestKit 组装 aiAgentHandler 的 mock 依赖、路由与可注入的部署替身。
type aiAgentTestKit struct {
	ctrl      *gomock.Controller
	router    *mux.Router
	authBiz   *biz.MockAuthBiz
	repoBiz   *biz.MockRepoBiz
	nsBiz     *biz.MockNamespaceBiz
	projBiz   *biz.MockProjectBiz
	k8sRepo   *data.MockK8sRepo
	tokenBiz  *biz.MockAccessTokenBiz
	activity  *biz.MockAgentActivity
	locker    *locker.MockLocker
	eventRepo *data.MockEventRepo
	// pluginMgr 提供 ws 广播通道：部署路径必须拿到真实 PubSub 而非 EmptyPubSub。
	pluginMgr *app.MockPluginManager
	wsSender  *app.MockWsSender
	handler   *aiAgentHandler

	// 部署替身记录：用例可断言调用次数/入参，并注入返回错误。
	applyCalls int
	applyCtx   context.Context
	applyDeps  deploy.ApplyProjectDeps
	applyInput *deploy.ApplyProjectInput
	applyErr   error
}

// newAIAgentTestKit 构造处理器与路由，apply 默认成功且记录调用。
func newAIAgentTestKit(t *testing.T) *aiAgentTestKit {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	k := &aiAgentTestKit{
		ctrl:      ctrl,
		authBiz:   biz.NewMockAuthBiz(ctrl),
		repoBiz:   biz.NewMockRepoBiz(ctrl),
		nsBiz:     biz.NewMockNamespaceBiz(ctrl),
		projBiz:   biz.NewMockProjectBiz(ctrl),
		k8sRepo:   data.NewMockK8sRepo(ctrl),
		tokenBiz:  biz.NewMockAccessTokenBiz(ctrl),
		activity:  biz.NewMockAgentActivity(ctrl),
		locker:    locker.NewMockLocker(ctrl),
		eventRepo: data.NewMockEventRepo(ctrl),
		pluginMgr: app.NewMockPluginManager(ctrl),
		wsSender:  app.NewMockWsSender(ctrl),
	}
	h := newAIAgentHandler(aiAgentDeps{
		AuthBiz:        k.authBiz,
		RepoBiz:        k.repoBiz,
		NamespaceBiz:   k.nsBiz,
		ProjectBiz:     k.projBiz,
		K8sRepo:        k.k8sRepo,
		AccessTokenBiz: k.tokenBiz,
		AgentActivity:  k.activity,
		Locker:         k.locker,
		EventBiz:       biz.NewEventBiz(k.eventRepo),
		PluginMgr:      k.pluginMgr,
		Logger:         mlog.NewForConfig(nil),
	})
	h.apply = func(ctx context.Context, deps deploy.ApplyProjectDeps, input *deploy.ApplyProjectInput) (deploy.Job, error) {
		k.applyCalls++
		k.applyCtx = ctx
		k.applyDeps = deps
		k.applyInput = input
		return nil, k.applyErr
	}
	k.handler = h
	k.router = mux.NewRouter()
	h.Register(k.router)
	return k
}

// expectLogin 预置登录鉴权成功链路（VerifyToken + EffectiveRoles）。
func (k *aiAgentTestKit) expectLogin(email string) {
	k.authBiz.EXPECT().VerifyToken(gomock.Any(), "Bearer tok").Return(&biz.UserInfo{Name: "alice", Email: email}, nil)
	k.authBiz.EXPECT().EffectiveRoles(gomock.Any(), email, gomock.Any()).Return(nil, nil)
}

// expectNamespaceLock 预置空间初始化串行锁获取成功；release 由 handler 在退出时调用。
func (k *aiAgentTestKit) expectNamespaceLock() {
	k.locker.EXPECT().
		RenewalAcquire(aiAgentNamespaceLockKey, deploy.LockTTLSeconds, deploy.LockRenewalSeconds).
		Return(func() {}, true)
}

// expectWakeLock 预置项目级唤醒锁获取成功，键必须与部署全局锁一致（namespace+name 哈希）。
func (k *aiAgentTestKit) expectWakeLock(namespaceID int, name string) {
	k.locker.EXPECT().
		RenewalAcquire(deploy.GetSlugName(namespaceID, name), deploy.LockTTLSeconds, deploy.LockRenewalSeconds).
		Return(func() {}, true)
}

// expectTokenRotationAudit 预置会话令牌轮换的审计记录（不含令牌原文）。
func (k *aiAgentTestKit) expectTokenRotationAudit(projectID int, name string) {
	k.eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Update, "alice", aiAgentTestEmail,
		"唤醒 AI Agent 沙箱并轮换会话令牌",
		map[string]any{"project_id": projectID, "project_name": name},
	)
}

// expectDeployPubSub 预置部署路径的 ws 广播通道：deployAgent 必须从插件管理器取
// 真实 PubSub（runner 据此广播 WsReloadProjects），且用完后关闭。
func (k *aiAgentTestKit) expectDeployPubSub() *app.MockPubSub {
	pubsub := app.NewMockPubSub(k.ctrl)
	k.pluginMgr.EXPECT().Ws().Return(k.wsSender)
	k.wsSender.EXPECT().New("", "").Return(pubsub)
	pubsub.EXPECT().Close().Return(nil)
	return pubsub
}

// do 发起请求并返回响应记录器；token 为空表示不携带鉴权头。
func (k *aiAgentTestKit) do(t *testing.T, method, path, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	rec := httptest.NewRecorder()
	k.router.ServeHTTP(rec, req)
	return rec
}

// decodeStatus 解析 200 响应为 aiStatusResponse。
func decodeStatus(t *testing.T, rec *httptest.ResponseRecorder) *aiStatusResponse {
	t.Helper()
	var out aiStatusResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	return &out
}

// agentNotFound 返回 biz 语义的 NotFound（数据层未配置 AI 仓库时的真实形态）。
func agentNotFound() error {
	return errs.WrapNotFound(errors.New("not found"), "get ai agent repo")
}

// aiAgentTestRepo 构造带 <.UserToken> 模板的 AI Agent 仓库 fixture。
// serveEnable 会做前置校验：values 未引用该变量时直接 400，故所有走启用流程的
// 用例都必须带上该模板。
func aiAgentTestRepo(id int, name string, enabled bool) *biz.Repo {
	return &biz.Repo{
		ID:      id,
		Name:    name,
		Enabled: enabled,
		MarsConfig: &mars.Config{
			ValuesYaml: "mars:\n  token: \"<.UserToken>\"",
		},
	}
}

// ---- GET /api/ai/status ----

// TestAIAgentHandler_Status_Unconfigured 未配置 AI Agent 仓库：200 + enabled=false。
func TestAIAgentHandler_Status_Unconfigured(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(nil, agentNotFound())
	// respond 仍会查一次 agent 状态：空间不存在时按「尚未部署」处理。
	k.nsBiz.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).Return(nil, errs.WrapNotFound(errors.New("no ns"), "find ns"))

	rec := k.do(t, http.MethodGet, aiStatusPath, "Bearer tok")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"enabled":false,"agent":null}`, rec.Body.String())
}

// TestAIAgentHandler_Status_DisabledWithoutAgent 仓库已禁用且尚未部署：enabled=false、agent=null。
func TestAIAgentHandler_Status_DisabledWithoutAgent(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "agent-repo", false), nil)
	k.nsBiz.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).Return(nil, errs.WrapNotFound(errors.New("no ns"), "find ns"))

	rec := k.do(t, http.MethodGet, aiStatusPath, "Bearer tok")
	assert.Equal(t, http.StatusOK, rec.Code)
	got := decodeStatus(t, rec)
	assert.False(t, got.Enabled)
	assert.Nil(t, got.Agent)
}

// TestAIAgentHandler_Status_WithAgent 正常路径：返回仓库信息与当前用户 agent 项目状态。
func TestAIAgentHandler_Status_WithAgent(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "agent-repo", true), nil)
	k.nsBiz.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).Return(&biz.Namespace{ID: 3, Name: biz.AgentNamespace}, nil)
	name := biz.AgentProjectName(aiAgentTestEmail)
	k.projBiz.EXPECT().FindByName(gomock.Any(), name, 3).
		Return(&biz.Project{ID: 11, Name: name, RepoID: 7}, nil)
	k.k8sRepo.EXPECT().GetDeployment(biz.AgentNamespace, biz.AgentDeploymentName(name)).
		Return(&appsv1.Deployment{
			Spec:   appsv1.DeploymentSpec{Replicas: lo.ToPtr(int32(1))},
			Status: appsv1.DeploymentStatus{ReadyReplicas: 1},
		}, nil)

	rec := k.do(t, http.MethodGet, aiStatusPath, "Bearer tok")
	assert.Equal(t, http.StatusOK, rec.Code)
	got := decodeStatus(t, rec)
	assert.True(t, got.Enabled)
	require.NotNil(t, got.Agent)
	assert.Equal(t, 11, got.Agent.ProjectID)
	assert.Equal(t, agentSandboxRunning, got.Agent.Sandbox)
}

// TestAIAgentHandler_Status_SandboxStates 覆盖沙箱运行态的全部派生分支。
func TestAIAgentHandler_Status_SandboxStates(t *testing.T) {
	tests := []struct {
		name string
		dep  *appsv1.Deployment
		err  error
		want string
	}{
		{
			name: "副本为 0 视为休眠",
			dep:  &appsv1.Deployment{Spec: appsv1.DeploymentSpec{Replicas: lo.ToPtr(int32(0))}},
			want: agentSandboxSleeping,
		},
		{
			name: "已扩容但无就绪副本视为启动中",
			dep:  &appsv1.Deployment{Spec: appsv1.DeploymentSpec{Replicas: lo.ToPtr(int32(1))}},
			want: agentSandboxStarting,
		},
		{
			name: "Replicas 为 nil 且就绪视为运行中",
			dep:  &appsv1.Deployment{Status: appsv1.DeploymentStatus{ReadyReplicas: 1}},
			want: agentSandboxRunning,
		},
		{
			name: "Deployment 不存在视为尚未部署",
			err:  errs.WrapNotFound(errors.New("not found"), "get deployment"),
			want: agentSandboxAbsent,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := newAIAgentTestKit(t)
			k.expectLogin(aiAgentTestEmail)
			k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
			k.nsBiz.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).
				Return(&biz.Namespace{ID: 3, Name: biz.AgentNamespace}, nil)
			name := biz.AgentProjectName(aiAgentTestEmail)
			k.projBiz.EXPECT().FindByName(gomock.Any(), name, 3).
				Return(&biz.Project{ID: 11, Name: name, RepoID: 7}, nil)
			k.k8sRepo.EXPECT().GetDeployment(biz.AgentNamespace, biz.AgentDeploymentName(name)).
				Return(tt.dep, tt.err)

			rec := k.do(t, http.MethodGet, aiStatusPath, "Bearer tok")
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tt.want, decodeStatus(t, rec).Agent.Sandbox)
		})
	}
}

// TestAIAgentHandler_Status_SandboxError Deployment 查询非 NotFound 失败 → 500。
func TestAIAgentHandler_Status_SandboxError(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	k.nsBiz.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).
		Return(&biz.Namespace{ID: 3, Name: biz.AgentNamespace}, nil)
	name := biz.AgentProjectName(aiAgentTestEmail)
	k.projBiz.EXPECT().FindByName(gomock.Any(), name, 3).
		Return(&biz.Project{ID: 11, Name: name, RepoID: 7}, nil)
	k.k8sRepo.EXPECT().GetDeployment(biz.AgentNamespace, biz.AgentDeploymentName(name)).
		Return(nil, errors.New("k8s down"))

	rec := k.do(t, http.MethodGet, aiStatusPath, "Bearer tok")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestAIAgentHandler_Status_NoAgentProject 空间存在但用户尚未部署 agent：agent=null。
func TestAIAgentHandler_Status_NoAgentProject(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	k.nsBiz.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).Return(&biz.Namespace{ID: 3}, nil)
	k.projBiz.EXPECT().FindByName(gomock.Any(), biz.AgentProjectName(aiAgentTestEmail), 3).
		Return(nil, errs.WrapNotFound(errors.New("not found"), "find project"))

	rec := k.do(t, http.MethodGet, aiStatusPath, "Bearer tok")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Nil(t, decodeStatus(t, rec).Agent)
}

// TestAIAgentHandler_Respond_EncodeError 覆盖响应写出失败：只记日志不 panic。
func TestAIAgentHandler_Respond_EncodeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	nsBiz := biz.NewMockNamespaceBiz(ctrl)
	projBiz := biz.NewMockProjectBiz(ctrl)
	k8sRepo := data.NewMockK8sRepo(ctrl)
	h := &aiAgentHandler{deps: aiAgentDeps{
		NamespaceBiz: nsBiz,
		ProjectBiz:   projBiz,
		K8sRepo:      k8sRepo,
		Logger:       mlog.NewForConfig(nil),
	}}
	nsBiz.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).Return(&biz.Namespace{ID: 3, Name: biz.AgentNamespace}, nil)
	projBiz.EXPECT().FindByName(gomock.Any(), gomock.Any(), 3).Return(&biz.Project{ID: 11, Name: "p"}, nil)
	k8sRepo.EXPECT().GetDeployment(gomock.Any(), gomock.Any()).Return(&appsv1.Deployment{}, nil)

	assert.NotPanics(t, func() {
		h.respond(context.Background(), &errResponseWriter{header: http.Header{}}, aiAgentTestEmail, nil)
	})
}

// TestAIAgentHandler_Status_RepoError 非 NotFound 的仓库查询失败 → 500。
func TestAIAgentHandler_Status_RepoError(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(nil, errors.New("db down"))

	rec := k.do(t, http.MethodGet, aiStatusPath, "Bearer tok")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestAIAgentHandler_Status_NamespaceError 空间查询非 NotFound 失败 → 500。
func TestAIAgentHandler_Status_NamespaceError(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	k.nsBiz.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).Return(nil, errors.New("db down"))

	rec := k.do(t, http.MethodGet, aiStatusPath, "Bearer tok")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestAIAgentHandler_Status_ProjectError 项目查询非 NotFound 失败 → 500。
func TestAIAgentHandler_Status_ProjectError(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	k.nsBiz.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).Return(&biz.Namespace{ID: 3}, nil)
	k.projBiz.EXPECT().FindByName(gomock.Any(), gomock.Any(), 3).Return(nil, errors.New("db down"))

	rec := k.do(t, http.MethodGet, aiStatusPath, "Bearer tok")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestAIAgentHandler_Unauthorized 未携带 token → 401，且不触碰业务依赖。
func TestAIAgentHandler_Unauthorized(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.authBiz.EXPECT().VerifyToken(gomock.Any(), "").Return(nil, errors.New("invalid token"))

	rec := k.do(t, http.MethodGet, aiStatusPath, "")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ---- POST /api/ai/agent/enable ----

// TestAIAgentHandler_Enable_EmptyEmail 用户缺少 email：400，无法派生项目名。
func TestAIAgentHandler_Enable_EmptyEmail(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin("")

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "email")
}

// TestAIAgentHandler_Enable_RepoUnconfigured 未配置 AI 仓库 → 400。
func TestAIAgentHandler_Enable_RepoUnconfigured(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(nil, agentNotFound())

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "尚未配置")
}

// TestAIAgentHandler_Enable_RepoQueryError 非 NotFound 的仓库查询失败 → 500。
func TestAIAgentHandler_Enable_RepoQueryError(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(nil, errors.New("db down"))

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestAIAgentHandler_Enable_RepoDisabled 仓库已禁用 → 400。
func TestAIAgentHandler_Enable_RepoDisabled(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", false), nil)

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "禁用")
}

// TestAIAgentHandler_Enable_MissingUserTokenTemplate values 未引用 <.UserToken> → 400。
// 该配置下部署不会签发令牌，沙箱能起来但回连 mars 全部 401，必须在入口拦住。
func TestAIAgentHandler_Enable_MissingUserTokenTemplate(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(&biz.Repo{ID: 7, Enabled: true}, nil)

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "<.UserToken>")
	assert.Zero(t, k.applyCalls)
}

// TestAIAgentHandler_Enable_NamespaceCreateError 空间创建失败 → 500。
func TestAIAgentHandler_Enable_NamespaceCreateError(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	k.expectNamespaceLock()
	k.nsBiz.EXPECT().Create(gomock.Any(), biz.AgentNamespace, aiAgentNamespaceDesc, schematype.SuperAdminEmail).
		Return(nil, false, errors.New("k8s boom"))

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "准备 AI Agent 命名空间失败")
}

// TestAIAgentHandler_Enable_NamespaceOccupied 同名空间被非系统用户占用：409，绝不接管。
func TestAIAgentHandler_Enable_NamespaceOccupied(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	k.expectNamespaceLock()
	k.nsBiz.EXPECT().Create(gomock.Any(), biz.AgentNamespace, aiAgentNamespaceDesc, schematype.SuperAdminEmail).
		Return(&biz.Namespace{ID: 3, CreatorEmail: "owner@mars.local"}, true, nil)

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "不会接管")
	assert.Zero(t, k.applyCalls)
}

// TestAIAgentHandler_Enable_NamespaceConfigError 强制私有失败 → 500。
func TestAIAgentHandler_Enable_NamespaceConfigError(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	k.expectNamespaceLock()
	k.nsBiz.EXPECT().Create(gomock.Any(), biz.AgentNamespace, aiAgentNamespaceDesc, schematype.SuperAdminEmail).
		Return(&biz.Namespace{ID: 3, CreatorEmail: schematype.SuperAdminEmail}, true, nil)
	k.nsBiz.EXPECT().UpdateConfig(gomock.Any(), &biz.UpdateConfigInput{
		ID:      3,
		Private: lo.ToPtr(true),
		Emails:  []string{},
	}).Return(nil, errors.New("db down"))

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestAIAgentHandler_Enable_NamespaceInitializing 并发首次启用时抢不到空间初始化锁：
// 409 提示稍后重试，不继续走创建流程。
func TestAIAgentHandler_Enable_NamespaceInitializing(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	k.locker.EXPECT().
		RenewalAcquire(aiAgentNamespaceLockKey, deploy.LockTTLSeconds, deploy.LockRenewalSeconds).
		Return(nil, false)

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "正在初始化")
	assert.Zero(t, k.applyCalls)
}

// TestAIAgentHandler_EnsureAgentNamespace_CreatedAudited 平台首次自动创建 mars-agent
// 空间时写入审计（谁触发、创建了哪个空间）；已存在的空间不重复记录。
func TestAIAgentHandler_EnsureAgentNamespace_CreatedAudited(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectNamespaceLock()
	k.nsBiz.EXPECT().Create(gomock.Any(), biz.AgentNamespace, aiAgentNamespaceDesc, schematype.SuperAdminEmail).
		Return(&biz.Namespace{ID: 8, Name: biz.AgentNamespace, CreatorEmail: schematype.SuperAdminEmail}, false, nil)
	k.eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Create, "alice", aiAgentTestEmail,
		"自动创建 AI Agent 命名空间 "+biz.AgentNamespace,
		map[string]any{"namespace_id": 8, "namespace": biz.AgentNamespace},
	)
	k.nsBiz.EXPECT().UpdateConfig(gomock.Any(), &biz.UpdateConfigInput{
		ID:      8,
		Private: lo.ToPtr(true),
		Emails:  []string{},
	}).Return(&biz.Namespace{ID: 8, Name: biz.AgentNamespace}, nil)

	ns, err := k.handler.ensureAgentNamespace(context.Background(), &biz.UserInfo{Name: "alice", Email: aiAgentTestEmail})
	require.NoError(t, err)
	assert.Equal(t, 8, ns.ID)
}

// TestAIAgentHandler_Enable_WakeLockBusy 唤醒换令牌与并发部署共用项目级锁：
// 抢锁失败返回 409（可重试），不签发新令牌。
func TestAIAgentHandler_Enable_WakeLockBusy(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	name := expectExistingAgent(k, 7, false)
	// 活跃度在进入唤醒临界区之前刷新，抢锁失败也要先 Touch。
	k.activity.EXPECT().Touch(gomock.Any(), name).Return(nil)
	// 抢锁失败时不做任何 Deployment 读取：运行态判断必须在锁内完成，
	// 锁外快照不能用来决策（可能已被并发缩容改写）。
	k.locker.EXPECT().
		RenewalAcquire(deploy.GetSlugName(3, name), deploy.LockTTLSeconds, deploy.LockRenewalSeconds).
		Return(nil, false)

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "正在部署中")
}

// expectAgentNamespace 预置「空间已存在且由超管创建」的链路（展示名也算超管）。
func expectAgentNamespace(k *aiAgentTestKit) {
	k.expectNamespaceLock()
	k.nsBiz.EXPECT().Create(gomock.Any(), biz.AgentNamespace, aiAgentNamespaceDesc, schematype.SuperAdminEmail).
		Return(&biz.Namespace{ID: 3, CreatorEmail: biz.SuperAdminName}, true, nil)
	k.nsBiz.EXPECT().UpdateConfig(gomock.Any(), &biz.UpdateConfigInput{
		ID:      3,
		Private: lo.ToPtr(true),
		Emails:  []string{},
	}).Return(&biz.Namespace{ID: 3, Name: biz.AgentNamespace}, nil)
}

// expectExistingAgent 预置「空间就绪 + 项目已存在」的幂等链路，返回项目名。
// respond 表示请求会走到成功响应：项目查两次（幂等检查 + respond），并补空间查询；
// 中途失败的错误路径只查一次项目、不进入 respond。
func expectExistingAgent(k *aiAgentTestKit, repoID int, respond bool) string {
	expectAgentNamespace(k)
	name := biz.AgentProjectName(aiAgentTestEmail)
	k.projBiz.EXPECT().FindByName(gomock.Any(), name, 3).
		Return(&biz.Project{ID: 11, Name: name, RepoID: repoID}, nil).Times(lo.Ternary(respond, 2, 1))
	if !respond {
		return name
	}
	k.nsBiz.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).
		Return(&biz.Namespace{ID: 3, Name: biz.AgentNamespace}, nil)
	return name
}

// expectSandboxRunning 预置 respond 阶段的沙箱运行态查询（Deployment 正在运行）。
// 用例需在唤醒/重建路径的 Deployment 查询之后声明，保证 gomock 按序消费。
func expectSandboxRunning(k *aiAgentTestKit, name string) {
	k.k8sRepo.EXPECT().GetDeployment(biz.AgentNamespace, biz.AgentDeploymentName(name)).
		Return(&appsv1.Deployment{Spec: appsv1.DeploymentSpec{Replicas: lo.ToPtr(int32(1))}}, nil)
}

// TestAIAgentHandler_Enable_ExistingMatchingRepo 幂等：项目已存在且绑定同一仓库，直接返回状态不重部署。
func TestAIAgentHandler_Enable_ExistingMatchingRepo(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "agent-repo", true), nil)
	expectAgentNamespace(k)
	name := biz.AgentProjectName(aiAgentTestEmail)
	// 前置幂等检查 + respond 回吐状态各查一次项目；respond 还会再查一次空间。
	k.projBiz.EXPECT().FindByName(gomock.Any(), name, 3).Return(&biz.Project{ID: 11, Name: name, RepoID: 7}, nil).Times(2)
	k.nsBiz.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).
		Return(&biz.Namespace{ID: 3, Name: biz.AgentNamespace}, nil)
	// 沙箱仍在运行（replicas=1）：唤醒是空操作，但也要进锁与空闲缩容互斥。
	k.k8sRepo.EXPECT().GetDeployment(biz.AgentNamespace, biz.AgentDeploymentName(name)).
		Return(&appsv1.Deployment{Spec: appsv1.DeploymentSpec{Replicas: lo.ToPtr(int32(1))}}, nil)
	k.expectWakeLock(3, name)
	expectSandboxRunning(k, name)
	k.activity.EXPECT().Touch(gomock.Any(), name).Return(nil)

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusOK, rec.Code)
	got := decodeStatus(t, rec)
	require.NotNil(t, got.Agent)
	assert.Equal(t, 11, got.Agent.ProjectID)
	assert.Zero(t, k.applyCalls, "已存在的 agent 项目不应重复部署")
}

// TestAIAgentHandler_Enable_WakeScaledDownSandbox 沙箱已缩容：换发会话令牌写入 Secret、
// 撤销旧令牌并扩到 1；项目本身不重复部署。
func TestAIAgentHandler_Enable_WakeScaledDownSandbox(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	name := expectExistingAgent(k, 7, true)
	depName := biz.AgentDeploymentName(name)
	// 记录调用顺序：运行态必须在项目锁内读取，锁外快照可能已被并发缩容改写。
	var callOrder []string
	k.locker.EXPECT().
		RenewalAcquire(deploy.GetSlugName(3, name), deploy.LockTTLSeconds, deploy.LockRenewalSeconds).
		DoAndReturn(func(string, int64, int64) (func(), bool) {
			callOrder = append(callOrder, "lock")
			return func() {}, true
		})
	k.k8sRepo.EXPECT().GetDeployment(biz.AgentNamespace, depName).
		DoAndReturn(func(string, string) (*appsv1.Deployment, error) {
			callOrder = append(callOrder, "get")
			return &appsv1.Deployment{Spec: appsv1.DeploymentSpec{Replicas: lo.ToPtr(int32(0))}}, nil
		})
	k.tokenBiz.EXPECT().Grant(gomock.Any(), &biz.GrantAccessTokenInput{
		ExpireSeconds: int32(deploy.UserTokenTTL / time.Second),
		Usage:         deploy.UserTokenUsage(11),
		User:          &biz.UserInfo{Name: "alice", Email: aiAgentTestEmail},
	}).Return(&biz.AccessToken{Token: "new-token"}, nil)
	k.k8sRepo.EXPECT().GetSecret(gomock.Any(), biz.AgentNamespace, depName).Return(&corev1.Secret{}, nil)
	k.k8sRepo.EXPECT().UpdateSecret(gomock.Any(), biz.AgentNamespace, depName, gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string, secret *corev1.Secret) (*corev1.Secret, error) {
			assert.Equal(t, "new-token", string(secret.Data[biz.AgentSessionTokenSecretKey]))
			return secret, nil
		})
	k.tokenBiz.EXPECT().RevokeByUsageExcept(gomock.Any(), deploy.UserTokenUsage(11), "new-token").Return(nil)
	k.expectTokenRotationAudit(11, name)
	k.k8sRepo.EXPECT().ScaleDeployment(gomock.Any(), biz.AgentNamespace, depName, int32(1)).Return(nil)
	expectSandboxRunning(k, name)
	k.activity.EXPECT().Touch(gomock.Any(), name).Return(nil)

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Zero(t, k.applyCalls, "唤醒已存在的 agent 项目不应重复部署")
	require.GreaterOrEqual(t, len(callOrder), 2)
	assert.Equal(t, []string{"lock", "get"}, callOrder[:2], "Deployment 运行态必须在项目锁内读取")
}

// TestAIAgentHandler_Enable_WakeNoDeployment Deployment 已丢失：项目行还在，走更新部署重建沙箱。
func TestAIAgentHandler_Enable_WakeNoDeployment(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	name := expectExistingAgent(k, 7, true)
	// Deployment 运行态在项目锁内读取：即使最终判定为丢失，也必须先拿到锁。
	k.expectWakeLock(3, name)
	k.k8sRepo.EXPECT().GetDeployment(biz.AgentNamespace, biz.AgentDeploymentName(name)).
		Return(nil, errs.WrapNotFound(errors.New("not found"), "get deployment"))
	expectSandboxRunning(k, name)
	k.activity.EXPECT().Touch(gomock.Any(), name).Return(nil)
	pubsub := k.expectDeployPubSub()

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, k.applyCalls, "Deployment 丢失必须触发一次更新部署")
	require.NotNil(t, k.applyInput)
	assert.Equal(t, websocket_pb.Type_UpdateProject, k.applyInput.JobInput.Type)
	assert.Equal(t, int32(11), k.applyInput.JobInput.ProjectID)
	assert.Same(t, pubsub, k.applyInput.JobInput.PubSub, "重建部署必须带真实 ws 广播通道")
}

// TestAIAgentHandler_Enable_WakeDeployError Deployment 丢失且重建失败 → 500。
func TestAIAgentHandler_Enable_WakeDeployError(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	name := expectExistingAgent(k, 7, false)
	k.expectWakeLock(3, name)
	k.k8sRepo.EXPECT().GetDeployment(biz.AgentNamespace, biz.AgentDeploymentName(name)).
		Return(nil, errs.WrapNotFound(errors.New("not found"), "get deployment"))
	k.activity.EXPECT().Touch(gomock.Any(), name).Return(nil)
	k.applyErr = errors.New("helm boom")
	k.expectDeployPubSub()

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	// 内部错误只回吐通用文案，不泄漏部署链路细节。
	assert.Contains(t, rec.Body.String(), "重建 AI Agent 沙箱失败")
	assert.NotContains(t, rec.Body.String(), "helm boom")
	assert.Equal(t, 1, k.applyCalls)
}

// TestAIAgentHandler_Enable_WakeNilReplicas Replicas 为 nil 等价于 k8s 默认 1：视为运行中不唤醒。
func TestAIAgentHandler_Enable_WakeNilReplicas(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	name := expectExistingAgent(k, 7, true)
	k.k8sRepo.EXPECT().GetDeployment(biz.AgentNamespace, biz.AgentDeploymentName(name)).
		Return(&appsv1.Deployment{}, nil)
	k.expectWakeLock(3, name)
	expectSandboxRunning(k, name)
	k.activity.EXPECT().Touch(gomock.Any(), name).Return(nil)

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestAIAgentHandler_Enable_WakeDeploymentError 查询 Deployment 非 NotFound 失败：唤醒失败 → 500。
func TestAIAgentHandler_Enable_WakeDeploymentError(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	name := expectExistingAgent(k, 7, false)
	k.activity.EXPECT().Touch(gomock.Any(), name).Return(nil)
	k.expectWakeLock(3, name)
	k.k8sRepo.EXPECT().GetDeployment(biz.AgentNamespace, biz.AgentDeploymentName(name)).
		Return(nil, errors.New("k8s down"))

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "唤醒")
}

// expectScaledDown 预置「沙箱已缩容」的 Deployment 查询。
func expectScaledDown(k *aiAgentTestKit, name string) {
	k.k8sRepo.EXPECT().GetDeployment(biz.AgentNamespace, biz.AgentDeploymentName(name)).
		Return(&appsv1.Deployment{Spec: appsv1.DeploymentSpec{Replicas: lo.ToPtr(int32(0))}}, nil)
}

// expectWakeGrant 预置签发新会话令牌的成功链路。
func expectWakeGrant(k *aiAgentTestKit, token string) {
	k.tokenBiz.EXPECT().Grant(gomock.Any(), gomock.Any()).Return(&biz.AccessToken{Token: token}, nil)
}

// TestAIAgentHandler_Enable_WakeGrantError 签发新令牌失败 → 500。
func TestAIAgentHandler_Enable_WakeGrantError(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	name := expectExistingAgent(k, 7, false)
	expectScaledDown(k, name)
	k.activity.EXPECT().Touch(gomock.Any(), name).Return(nil)
	k.expectWakeLock(3, name)
	k.tokenBiz.EXPECT().Grant(gomock.Any(), gomock.Any()).Return(nil, errors.New("db down"))

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestAIAgentHandler_Enable_WakeSecretError 读取沙箱 Secret 失败 → 500。
func TestAIAgentHandler_Enable_WakeSecretError(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	name := expectExistingAgent(k, 7, false)
	expectScaledDown(k, name)
	k.activity.EXPECT().Touch(gomock.Any(), name).Return(nil)
	k.expectWakeLock(3, name)
	expectWakeGrant(k, "new-token")
	k.k8sRepo.EXPECT().GetSecret(gomock.Any(), biz.AgentNamespace, biz.AgentDeploymentName(name)).
		Return(nil, errors.New("secret not found"))
	// 新令牌未写入沙箱：必须尽力撤销；撤销失败只记 Warning，不掩盖原始错误。
	k.tokenBiz.EXPECT().Revoke(gomock.Any(), "new-token").Return(errors.New("db down"))

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestAIAgentHandler_Enable_WakeUpdateSecretError 写入沙箱 Secret 失败 → 500。
func TestAIAgentHandler_Enable_WakeUpdateSecretError(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	name := expectExistingAgent(k, 7, false)
	expectScaledDown(k, name)
	k.activity.EXPECT().Touch(gomock.Any(), name).Return(nil)
	k.expectWakeLock(3, name)
	expectWakeGrant(k, "new-token")
	k.k8sRepo.EXPECT().GetSecret(gomock.Any(), biz.AgentNamespace, biz.AgentDeploymentName(name)).Return(&corev1.Secret{}, nil)
	k.k8sRepo.EXPECT().UpdateSecret(gomock.Any(), biz.AgentNamespace, biz.AgentDeploymentName(name), gomock.Any()).
		Return(nil, errors.New("update failed"))
	k.tokenBiz.EXPECT().Revoke(gomock.Any(), "new-token").Return(nil)

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestAIAgentHandler_Enable_WakeRevokeError 撤销旧令牌失败只记 Warning，唤醒仍成功。
func TestAIAgentHandler_Enable_WakeRevokeError(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	name := expectExistingAgent(k, 7, true)
	expectScaledDown(k, name)
	k.expectWakeLock(3, name)
	expectWakeGrant(k, "new-token")
	k.k8sRepo.EXPECT().GetSecret(gomock.Any(), biz.AgentNamespace, biz.AgentDeploymentName(name)).Return(&corev1.Secret{}, nil)
	k.k8sRepo.EXPECT().UpdateSecret(gomock.Any(), biz.AgentNamespace, biz.AgentDeploymentName(name), gomock.Any()).
		Return(&corev1.Secret{}, nil)
	k.tokenBiz.EXPECT().RevokeByUsageExcept(gomock.Any(), deploy.UserTokenUsage(11), "new-token").
		Return(errors.New("db down"))
	k.expectTokenRotationAudit(11, name)
	k.k8sRepo.EXPECT().ScaleDeployment(gomock.Any(), biz.AgentNamespace, biz.AgentDeploymentName(name), int32(1)).Return(nil)
	expectSandboxRunning(k, name)
	k.activity.EXPECT().Touch(gomock.Any(), name).Return(nil)

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestAIAgentHandler_Enable_WakeScaleError 扩容失败 → 500。
func TestAIAgentHandler_Enable_WakeScaleError(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	name := expectExistingAgent(k, 7, false)
	expectScaledDown(k, name)
	k.activity.EXPECT().Touch(gomock.Any(), name).Return(nil)
	k.expectWakeLock(3, name)
	expectWakeGrant(k, "new-token")
	k.k8sRepo.EXPECT().GetSecret(gomock.Any(), biz.AgentNamespace, biz.AgentDeploymentName(name)).Return(&corev1.Secret{}, nil)
	k.k8sRepo.EXPECT().UpdateSecret(gomock.Any(), biz.AgentNamespace, biz.AgentDeploymentName(name), gomock.Any()).
		Return(&corev1.Secret{}, nil)
	k.tokenBiz.EXPECT().RevokeByUsageExcept(gomock.Any(), deploy.UserTokenUsage(11), "new-token").Return(nil)
	k.expectTokenRotationAudit(11, name)
	k.k8sRepo.EXPECT().ScaleDeployment(gomock.Any(), biz.AgentNamespace, biz.AgentDeploymentName(name), int32(1)).
		Return(errors.New("scale failed"))

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestAIAgentHandler_Enable_ExistingOtherRepo 同名项目绑定其它仓库：409，拒绝静默接管。
func TestAIAgentHandler_Enable_ExistingOtherRepo(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	expectAgentNamespace(k)
	name := biz.AgentProjectName(aiAgentTestEmail)
	k.projBiz.EXPECT().FindByName(gomock.Any(), name, 3).Return(&biz.Project{ID: 11, Name: name, RepoID: 99}, nil)

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "已被其他仓库占用")
	assert.Zero(t, k.applyCalls)
}

// TestAIAgentHandler_Enable_ProjectQueryError 非 NotFound 的项目查询失败 → 500。
func TestAIAgentHandler_Enable_ProjectQueryError(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	expectAgentNamespace(k)
	k.projBiz.EXPECT().FindByName(gomock.Any(), gomock.Any(), 3).Return(nil, errors.New("db down"))

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestAIAgentHandler_Enable_FirstDeploy 首次启用：走共享部署编排并回读项目状态。
func TestAIAgentHandler_Enable_FirstDeploy(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "agent-repo", true), nil)
	expectAgentNamespace(k)
	name := biz.AgentProjectName(aiAgentTestEmail)
	gomock.InOrder(
		k.projBiz.EXPECT().FindByName(gomock.Any(), name, 3).Return(nil, errs.WrapNotFound(errors.New("not found"), "find project")),
		k.projBiz.EXPECT().FindByName(gomock.Any(), name, 3).Return(&biz.Project{ID: 11, Name: name, RepoID: 7}, nil),
		k.projBiz.EXPECT().FindByName(gomock.Any(), name, 3).Return(&biz.Project{ID: 11, Name: name, RepoID: 7}, nil),
	)
	k.nsBiz.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).
		Return(&biz.Namespace{ID: 3, Name: biz.AgentNamespace}, nil)
	expectSandboxRunning(k, name)
	// 活跃度刷新失败只记 Warning，不影响启用结果。
	k.activity.EXPECT().Touch(gomock.Any(), name).Return(errors.New("cache down"))
	pubsub := k.expectDeployPubSub()

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusOK, rec.Code)
	got := decodeStatus(t, rec)
	require.NotNil(t, got.Agent)
	assert.Equal(t, 11, got.Agent.ProjectID)

	assert.Equal(t, 1, k.applyCalls)
	require.NotNil(t, k.applyInput)
	assert.Equal(t, websocket_pb.Type_CreateProject, k.applyInput.JobInput.Type)
	assert.Equal(t, int32(3), k.applyInput.JobInput.NamespaceId)
	assert.Equal(t, int32(7), k.applyInput.JobInput.RepoID)
	assert.Equal(t, name, k.applyInput.JobInput.Name)
	require.NotNil(t, k.applyInput.JobInput.User)
	assert.Equal(t, aiAgentTestEmail, k.applyInput.JobInput.User.Email)
	assert.Same(t, pubsub, k.applyInput.JobInput.PubSub, "部署必须带真实 ws 广播通道（WsReloadProjects）")
	assert.NotNil(t, k.applyInput.JobInput.Messager, "内部部署需使用 noop messager")
}

// TestAIAgentHandler_Enable_DeployInProgress 并发启用撞上部署全局锁 → 409（可重试）。
func TestAIAgentHandler_Enable_DeployInProgress(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	expectAgentNamespace(k)
	k.projBiz.EXPECT().FindByName(gomock.Any(), gomock.Any(), 3).
		Return(nil, errs.WrapNotFound(errors.New("not found"), "find project"))
	k.applyErr = deploy.ErrDeployInProgress
	k.expectDeployPubSub()

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "正在部署中")
}

// TestAIAgentHandler_Enable_RebuildInProgress 重建沙箱撞上全局锁同样返回 409。
func TestAIAgentHandler_Enable_RebuildInProgress(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	name := expectExistingAgent(k, 7, false)
	k.expectWakeLock(3, name)
	k.k8sRepo.EXPECT().GetDeployment(biz.AgentNamespace, biz.AgentDeploymentName(name)).
		Return(nil, errs.WrapNotFound(errors.New("not found"), "get deployment"))
	k.activity.EXPECT().Touch(gomock.Any(), name).Return(nil)
	k.applyErr = deploy.ErrDeployInProgress
	k.expectDeployPubSub()

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Equal(t, 1, k.applyCalls)
}

// TestAIAgentHandler_Enable_DeployError 部署失败 → 500，不再回读项目。
func TestAIAgentHandler_Enable_DeployError(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	expectAgentNamespace(k)
	k.projBiz.EXPECT().FindByName(gomock.Any(), gomock.Any(), 3).
		Return(nil, errs.WrapNotFound(errors.New("not found"), "find project"))
	k.applyErr = errors.New("helm boom")
	k.expectDeployPubSub()

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "启用 AI Agent 部署失败")
	assert.NotContains(t, rec.Body.String(), "helm boom")
	assert.Equal(t, 1, k.applyCalls)
}

// TestAIAgentHandler_Enable_ReadBackError 部署成功但回读项目失败 → 500。
func TestAIAgentHandler_Enable_ReadBackError(t *testing.T) {
	k := newAIAgentTestKit(t)
	k.expectLogin(aiAgentTestEmail)
	k.repoBiz.EXPECT().GetAIAgent(gomock.Any()).Return(aiAgentTestRepo(7, "", true), nil)
	expectAgentNamespace(k)
	gomock.InOrder(
		k.projBiz.EXPECT().FindByName(gomock.Any(), gomock.Any(), 3).Return(nil, errs.WrapNotFound(errors.New("not found"), "find project")),
		k.projBiz.EXPECT().FindByName(gomock.Any(), gomock.Any(), 3).Return(nil, errors.New("db down")),
	)
	k.expectDeployPubSub()

	rec := k.do(t, http.MethodPost, aiAgentEnablePath, "Bearer tok")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "回读")
}

// ---- 纯函数 ----

// TestAgentProjectName 项目名必须由邮箱可读且唯一地派生，并在重复调用间保持稳定。
func TestAgentProjectName(t *testing.T) {
	name := biz.AgentProjectName(aiAgentTestEmail)
	assert.Equal(t, biz.AgentProjectPrefix+"alice-"+hasher.Hash(aiAgentTestEmail)[:12], name)
	assert.NotEqual(t, name, biz.AgentProjectName("bob@mars.local"))
	assert.Equal(t, name, biz.AgentProjectName(aiAgentTestEmail), "同一邮箱必须派生出同名项目")
}

// TestAgentTokenUser 沙箱会话令牌恒锚定创建者：创建者本人保留完整身份；
// 其他人触发只带创建者邮箱；creator 缺失（历史项目）或非 agent 项目退回触发者。
func TestAgentTokenUser(t *testing.T) {
	owner := "owner@mars.local"
	ownerInfo := &biz.UserInfo{Email: owner, Name: "owner", Roles: []string{"mars_admin"}}
	admin := &biz.UserInfo{Email: "admin@mars.local", Name: "admin"}
	agentProj := &biz.Project{ID: 1, Name: biz.AgentProjectName(owner), Creator: owner}

	assert.Same(t, ownerInfo, agentTokenUser(agentProj, ownerInfo), "创建者本人触发时保留完整身份")
	assert.Equal(t, &biz.UserInfo{Email: owner}, agentTokenUser(agentProj, admin), "他人触发只带创建者邮箱")
	assert.Same(t, admin, agentTokenUser(&biz.Project{Name: agentProj.Name}, admin), "creator 缺失时退回触发者")
	assert.Same(t, admin, agentTokenUser(&biz.Project{Name: "my-app", Creator: owner}, admin), "非 agent 项目不受锚定规则影响")
	assert.Same(t, admin, agentTokenUser(nil, admin), "项目缺失时保持既有语义：用触发者身份")
	assert.Nil(t, agentTokenUser(agentProj, nil))
}

// TestIsSuperAdminNamespaceCreator 超管邮箱与展示名都算系统创建者。
func TestIsSuperAdminNamespaceCreator(t *testing.T) {
	assert.True(t, isSuperAdminNamespaceCreator(schematype.SuperAdminEmail))
	assert.True(t, isSuperAdminNamespaceCreator(biz.SuperAdminName))
	assert.False(t, isSuperAdminNamespaceCreator("owner@mars.local"))
	assert.False(t, isSuperAdminNamespaceCreator(""))
}
