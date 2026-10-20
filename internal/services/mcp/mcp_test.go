package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	changelogpb "github.com/duc-cnzj/mars/api/v6/proto/changelog"
	containerpb "github.com/duc-cnzj/mars/api/v6/proto/container"
	metricspb "github.com/duc-cnzj/mars/api/v6/proto/metrics"
	namespacepb "github.com/duc-cnzj/mars/api/v6/proto/namespace"
	projectpb "github.com/duc-cnzj/mars/api/v6/proto/project"
	apitypes "github.com/duc-cnzj/mars/api/v6/proto/types"
	websocketpb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// fakeProjectSvc 是 projectpb.ProjectServer 的测试替身：嵌入接口即自动满足全部方法
// （包括未导出的 mustEmbedUnimplementedProjectServer——仓库无该 gRPC server 接口的现成 mock），
// 只覆写被测方法，从而能精确断言「MCP 入参怎么搬运成 proto 请求」。
type fakeProjectSvc struct {
	projectpb.ProjectServer
	applyFn         func(context.Context, *projectpb.WebApplyByNameRequest) (*projectpb.WebApplyResponse, error)
	showFn          func(context.Context, *projectpb.ShowRequest) (*projectpb.ShowResponse, error)
	showByNameFn    func(context.Context, *projectpb.ShowByNameRequest) (*projectpb.ShowResponse, error)
	allContainersFn func(context.Context, *projectpb.AllContainersRequest) (*projectpb.AllContainersResponse, error)
}

// WebApplyByName 委派给测试注入的 applyFn。
func (f *fakeProjectSvc) WebApplyByName(ctx context.Context, req *projectpb.WebApplyByNameRequest) (*projectpb.WebApplyResponse, error) {
	return f.applyFn(ctx, req)
}

// Show 委派给测试注入的 showFn。
func (f *fakeProjectSvc) Show(ctx context.Context, req *projectpb.ShowRequest) (*projectpb.ShowResponse, error) {
	return f.showFn(ctx, req)
}

// ShowByName 委派给测试注入的 showByNameFn。
func (f *fakeProjectSvc) ShowByName(ctx context.Context, req *projectpb.ShowByNameRequest) (*projectpb.ShowResponse, error) {
	return f.showByNameFn(ctx, req)
}

// AllContainers 委派给测试注入的 allContainersFn。
func (f *fakeProjectSvc) AllContainers(ctx context.Context, req *projectpb.AllContainersRequest) (*projectpb.AllContainersResponse, error) {
	return f.allContainersFn(ctx, req)
}

// showByNameItem 造一个只带「增量 apply 需要继承的字段」的项目模型（分支/提交/配置/版本/仓库绑定、
// 已落库的自定义元素值）。
func showByNameItem(config, branch, commit string, version, repoID int32, extras []*websocketpb.ExtraValue) *projectpb.ShowResponse {
	return &projectpb.ShowResponse{Item: &apitypes.ProjectModel{
		Id: 7, Name: "svc", Config: config, GitBranch: branch, GitCommit: commit,
		Version: version, RepoId: repoID, FinalExtraValues: extras,
	}}
}

// fakeContainerSvc 是 containerpb.ContainerServer 的测试替身，覆写三个只读/写工具用到的
// unary 方法（ForceDeletePod/IsPodRunning/ContainerLog）。
type fakeContainerSvc struct {
	containerpb.ContainerServer
	forceDeleteFn  func(context.Context, *containerpb.ForceDeletePodRequest) (*containerpb.ForceDeletePodResponse, error)
	isPodRunningFn func(context.Context, *containerpb.IsPodRunningRequest) (*containerpb.IsPodRunningResponse, error)
	containerLogFn func(context.Context, *containerpb.LogRequest) (*containerpb.LogResponse, error)
}

// ForceDeletePod 委派给测试注入的 forceDeleteFn。
func (f *fakeContainerSvc) ForceDeletePod(ctx context.Context, req *containerpb.ForceDeletePodRequest) (*containerpb.ForceDeletePodResponse, error) {
	return f.forceDeleteFn(ctx, req)
}

// IsPodRunning 委派给测试注入的 isPodRunningFn。
func (f *fakeContainerSvc) IsPodRunning(ctx context.Context, req *containerpb.IsPodRunningRequest) (*containerpb.IsPodRunningResponse, error) {
	return f.isPodRunningFn(ctx, req)
}

// ContainerLog 委派给测试注入的 containerLogFn。
func (f *fakeContainerSvc) ContainerLog(ctx context.Context, req *containerpb.LogRequest) (*containerpb.LogResponse, error) {
	return f.containerLogFn(ctx, req)
}

// fakeNamespaceSvc 是 namespacepb.NamespaceServer 的测试替身，覆写 List。
type fakeNamespaceSvc struct {
	namespacepb.NamespaceServer
	listFn func(context.Context, *namespacepb.ListRequest) (*namespacepb.ListResponse, error)
}

// List 委派给测试注入的 listFn。
func (f *fakeNamespaceSvc) List(ctx context.Context, req *namespacepb.ListRequest) (*namespacepb.ListResponse, error) {
	return f.listFn(ctx, req)
}

// fakeMetricsSvc 是 metricspb.MetricsServer 的测试替身，覆写 TopPod。
type fakeMetricsSvc struct {
	metricspb.MetricsServer
	topPodFn func(context.Context, *metricspb.TopPodRequest) (*metricspb.TopPodResponse, error)
}

// TopPod 委派给测试注入的 topPodFn。
func (f *fakeMetricsSvc) TopPod(ctx context.Context, req *metricspb.TopPodRequest) (*metricspb.TopPodResponse, error) {
	return f.topPodFn(ctx, req)
}

// fakeChangelogSvc 是 changelogpb.ChangelogServer 的测试替身，覆写 FindLastChangelogsByProjectID。
type fakeChangelogSvc struct {
	changelogpb.ChangelogServer
	findFn func(context.Context, *changelogpb.FindLastChangelogsByProjectIDRequest) (*changelogpb.FindLastChangelogsByProjectIDResponse, error)
}

// FindLastChangelogsByProjectID 委派给测试注入的 findFn。
func (f *fakeChangelogSvc) FindLastChangelogsByProjectID(ctx context.Context, req *changelogpb.FindLastChangelogsByProjectIDRequest) (*changelogpb.FindLastChangelogsByProjectIDResponse, error) {
	return f.findFn(ctx, req)
}

// 测试用固定令牌：LoginHTTP → biz.Authenticate → MockAuthBiz.VerifyToken 认它。
const testToken = "Bearer test-token"

// testEnv 聚合 MCP 集成测试所需的全部替身：数据层 repo mock + 真实 biz 组装出的
// 被测服务端。用真实 biz（而非 biz mock）是为了让 RBAC 判定、日志截断等真实逻辑
// 一起被覆盖——与 services 包 container_test.go 的组装口径一致。
type testEnv struct {
	ctrl       *gomock.Controller
	nsRepo     *data.MockNamespaceRepo
	projRepo   *data.MockProjectRepo
	k8sRepo    *data.MockK8sRepo
	clRepo     *data.MockChangelogRepo
	eventRepo  *data.MockEventRepo
	fileRepo   *data.MockFileRepo
	helmerRepo *data.MockHelmerRepo
	authBiz    *biz.MockAuthBiz
	projectSvc *fakeProjectSvc
	contSvc    *fakeContainerSvc
	nsSvc      *fakeNamespaceSvc
	metricSvc  *fakeMetricsSvc
	changeSvc  *fakeChangelogSvc
	activity   *biz.MockAgentActivity
	// touchedActivity 记录每次活跃度刷新的项目名，供断言心跳键。
	touchedActivity []string
	deps            Deps
	server          *Server
	httpSrv         *httptest.Server
}

// newTestEnv 组装被测 MCP 服务端并起一个 httptest server：走真实 http.Handler，
// 因此在传输层就把「第三方客户端经 HTTP 能否用」这条链路一并验证到。
func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	ctrl := gomock.NewController(t)
	e := &testEnv{
		ctrl:       ctrl,
		nsRepo:     data.NewMockNamespaceRepo(ctrl),
		projRepo:   data.NewMockProjectRepo(ctrl),
		k8sRepo:    data.NewMockK8sRepo(ctrl),
		clRepo:     data.NewMockChangelogRepo(ctrl),
		eventRepo:  data.NewMockEventRepo(ctrl),
		fileRepo:   data.NewMockFileRepo(ctrl),
		helmerRepo: data.NewMockHelmerRepo(ctrl),
		authBiz:    biz.NewMockAuthBiz(ctrl),
		// services 层接口用 fake 替身：这些接口是 gRPC server（含未导出方法），
		// gomock 生成物不存在；且 mcp 包无法 import internal/services（后者 wire.go 反向
		// 依赖本包，会构成 import 环），故真实服务实现不可用。精确断言「入参 → proto 请求」
		// 比断言调用次数更有价值。
		projectSvc: &fakeProjectSvc{},
		contSvc:    &fakeContainerSvc{},
		nsSvc:      &fakeNamespaceSvc{},
		metricSvc:  &fakeMetricsSvc{},
		changeSvc:  &fakeChangelogSvc{},
		activity:   biz.NewMockAgentActivity(ctrl),
	}
	// 所有已鉴权请求都会刷新活跃度：默认放行并记录项目名。
	e.activity.EXPECT().Touch(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, projectName string) error {
			e.touchedActivity = append(e.touchedActivity, projectName)
			return nil
		}).AnyTimes()

	logger := mlog.NewForConfig(nil)
	k8sBiz := biz.NewK8sBiz(e.k8sRepo)
	nsBiz := biz.NewNamespaceBiz(logger, e.nsRepo, e.k8sRepo, e.helmerRepo, e.eventRepo)
	projBiz := biz.NewProjectBiz(logger, e.projRepo, e.k8sRepo, e.clRepo)

	e.deps = Deps{
		ProjectBiz:    projBiz,
		ContainerBiz:  biz.NewContainerBiz(logger, k8sBiz, biz.NewFileBiz(e.fileRepo), biz.NewEventBiz(e.eventRepo), timer.NewReal()),
		K8sBiz:        k8sBiz,
		AccessBiz:     biz.NewAccessBiz(nsBiz, projBiz),
		AuthBiz:       e.authBiz,
		AgentActivity: e.activity,
		NamespaceSvc:  e.nsSvc,
		ProjectSvc:    e.projectSvc,
		ContainerSvc:  e.contSvc,
		MetricsSvc:    e.metricSvc,
		ChangelogSvc:  e.changeSvc,
		Logger:        logger,
	}
	e.server = NewServer(e.deps)
	e.httpSrv = httptest.NewServer(e.server.handler)
	t.Cleanup(e.httpSrv.Close)
	return e
}

// toolset 暴露被测工具实现，供逐工具单测直接调用（不经 HTTP/协议层）。
func (e *testEnv) toolset() *toolset {
	return &toolset{
		projectBiz:   e.deps.ProjectBiz,
		containerBiz: e.deps.ContainerBiz,
		k8sBiz:       e.deps.K8sBiz,
		accessBiz:    e.deps.AccessBiz,
		namespaceSvc: e.deps.NamespaceSvc,
		projectSvc:   e.deps.ProjectSvc,
		containerSvc: e.deps.ContainerSvc,
		metricsSvc:   e.deps.MetricsSvc,
		changelogSvc: e.deps.ChangelogSvc,
	}
}

// ctxWithUser 构造携带指定用户的 ctx，模拟鉴权中间件注入用户后的请求上下文。
func ctxWithUser(u *biz.UserInfo) context.Context {
	return biz.SetUser(context.Background(), u)
}

// expectAuth 让每次 HTTP 鉴权都成功并返回指定用户（每次返回新实例，避免跨请求共享可变对象）。
func (e *testEnv) expectAuth(user *biz.UserInfo) {
	e.authBiz.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, _ string) (*biz.UserInfo, error) {
			cp := *user
			cp.Roles = append([]string(nil), user.Roles...)
			return &cp, nil
		}).AnyTimes()
	e.authBiz.EXPECT().EffectiveRoles(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, _ string, sso []string) ([]string, error) {
			return append([]string(nil), sso...), nil
		}).AnyTimes()
}

// adminUser 返回一个带 mars_admin 角色的用户。
func adminUser() *biz.UserInfo {
	return &biz.UserInfo{ID: "1", Email: "admin@mars.com", Name: "admin", Roles: []string{schematype.MarsAdmin}}
}

// normalUser 返回一个无管理员角色的普通用户。
func normalUser() *biz.UserInfo {
	return &biz.UserInfo{ID: "2", Email: "user@mars.com", Name: "user"}
}

// post 向被测服务端发一条 MCP JSON-RPC 消息（stdio 之外的唯一入口是 HTTP POST），
// 返回 HTTP 状态码、响应头与完整响应体（可能是 SSE 帧）。
func (e *testEnv) post(t *testing.T, sessionID, auth, body string) (int, http.Header, string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, e.httpSrv.URL+Path, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, resp.Header, string(raw)
}

// sseJSON 从 Streamable HTTP 的 text/event-stream 响应里取出第一条 data 帧的 JSON。
func sseJSON(t *testing.T, body string) map[string]any {
	t.Helper()
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(line, "data:"))), &m); err != nil {
			continue
		}
		return m
	}
	t.Fatalf("响应里没有可解析的 SSE data 帧: %q", body)
	return nil
}

// initSession 完成 MCP initialize 握手并返回会话 ID——这是任何标准 MCP 客户端
// 连接服务端的第一步，本测试据此证明协议层与第三方客户端兼容。
func (e *testEnv) initSession(t *testing.T, auth string) string {
	t.Helper()
	code, hdr, body := e.post(t, "", auth,
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}`)
	require.Equal(t, http.StatusOK, code, body)
	sessionID := hdr.Get("Mcp-Session-Id")
	require.NotEmpty(t, sessionID, "initialize 必须下发 Mcp-Session-Id")
	res := sseJSON(t, body)
	result := res["result"].(map[string]any)
	serverInfo := result["serverInfo"].(map[string]any)
	assert.Equal(t, "mars", serverInfo["name"])
	// 展示约定必须随 instructions 下发给模型：向用户汇报时带 ID 是用户后续操作的前提。
	assert.Contains(t, result["instructions"], "必须同时给出 ID 与名称")
	return sessionID
}

// metav1TimeZero 返回零值 metav1.Time，用于断言时间格式化的零值分支。
func metav1TimeZero() metav1.Time { return metav1.Time{} }

// metav1Time 按给定分量构造 UTC metav1.Time，用于断言非零值格式化。
func metav1Time(y, mo, d, h, mi, s int) metav1.Time {
	return metav1.NewTime(time.Date(y, time.Month(mo), d, h, mi, s, 0, time.UTC))
}

// TestServer_Unauthenticated 验证未携带 token 的请求被 LoginHTTP 拦在门外（401），
// 不会进入 MCP 协议层——这是标准 MCP 端点同时复用 mars 鉴权的关键约束。
func TestServer_Unauthenticated(t *testing.T) {
	e := newTestEnv(t)
	// 无 token 时 verify 必然失败：MockAuthBiz 返回鉴权错误，LoginHTTP 应据此回 401。
	e.authBiz.EXPECT().VerifyToken(gomock.Any(), "").Return(nil, errors.New("invalid token"))
	code, _, _ := e.post(t, "", "", `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	assert.Equal(t, http.StatusUnauthorized, code)
}

// TestServer_InitializeAndListTools 走完整 initialize + tools/list 流程，断言工具集
// 覆盖只读排查工具与 3 个写工具，且绝不包含 delete_project 等破坏性能力。
//
// 关键断言：写工具的 annotations.readOnlyHint 必须不为 true——这是「客户端弹人工确认」
// 这道审批闸门唯一的生效依据，一旦被打上只读标，客户端会静默放行写操作。
func TestServer_InitializeAndListTools(t *testing.T) {
	e := newTestEnv(t)
	e.expectAuth(adminUser())

	sessionID := e.initSession(t, testToken)
	code, _, body := e.post(t, sessionID, testToken, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	require.Equal(t, http.StatusOK, code, body)

	tools := sseJSON(t, body)["result"].(map[string]any)["tools"].([]any)
	byName := make(map[string]map[string]any, len(tools))
	for _, raw := range tools {
		m := raw.(map[string]any)
		byName[m["name"].(string)] = m
	}

	readOnly := []string{
		"get_current_user",
		"list_namespaces", "list_projects", "get_project", "list_project_containers",
		"check_pod_status", "list_namespace_events", "get_container_logs",
		"get_pod_metrics", "list_project_changelogs",
	}
	write := []string{"apply_project", "restart_pod", "exec_command"}
	for _, want := range append(append([]string{}, readOnly...), write...) {
		assert.Contains(t, byName, want)
	}
	assert.NotContains(t, byName, "delete_project")
	assert.Len(t, byName, 13)

	for _, name := range readOnly {
		ann, _ := byName[name]["annotations"].(map[string]any)
		assert.Equal(t, true, ann["readOnlyHint"], "%s 是只读工具，应可免确认执行", name)
	}
	for _, name := range write {
		ann, _ := byName[name]["annotations"].(map[string]any)
		assert.NotEqual(t, true, ann["readOnlyHint"],
			"%s 是写工具，readOnlyHint 不得为 true，否则客户端不会弹确认", name)
	}
}

// TestServer_TouchesAgentActivity 已鉴权的 MCP 请求按用户邮箱派生的项目刷新活跃度，
// 使工具调用期间空闲关闭任务不会误缩容沙箱。
func TestServer_TouchesAgentActivity(t *testing.T) {
	e := newTestEnv(t)
	e.expectAuth(adminUser())

	e.initSession(t, testToken)

	require.NotEmpty(t, e.touchedActivity)
	assert.Equal(t, biz.AgentProjectName("admin@mars.com"), e.touchedActivity[0])
}

// TestServer_CallTool_ListNamespaces 验证一次真实 tools/call 能打通「HTTP → 鉴权 →
// 工具 → NamespaceSvc → 回吐服务模型 JSON」全链路，并把入参如实搬运成 proto 请求。
func TestServer_CallTool_ListNamespaces(t *testing.T) {
	e := newTestEnv(t)
	e.expectAuth(adminUser())
	e.nsSvc.listFn = func(_ context.Context, req *namespacepb.ListRequest) (*namespacepb.ListResponse, error) {
		// 未传分页时指针指向 0，交由 NamespaceSvc 侧按默认值兜底；此处只断言搬运不带过滤词。
		assert.Zero(t, req.GetPage())
		assert.Nil(t, req.Name)
		return &namespacepb.ListResponse{
			Items: []*apitypes.NamespaceModel{{Id: 1, Name: "team-a", Description: "团队 A", Private: true}},
			Count: 1, Page: 1, PageSize: 15,
		}, nil
	}

	sessionID := e.initSession(t, testToken)
	code, _, body := e.post(t, sessionID, testToken,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_namespaces","arguments":{}}}`)
	require.Equal(t, http.StatusOK, code, body)

	res := sseJSON(t, body)["result"].(map[string]any)
	content := res["content"].([]any)[0].(map[string]any)["text"].(string)
	assert.Contains(t, content, "team-a")
	assert.Contains(t, content, `"count":1`)
}

// TestServer_CallTool_GetProjectReturnsConfig 端到端证明「模型真能看见当前配置」：
// get_project 回吐 ProjectSvc.Show 的项目模型，其中必须带上 config 字段——这是模型
// 判断「该改哪个键」的唯一依据，只断言单测结构体是不够的（字段名/序列化在传输层
// 出错同样会让模型抓瞎）。
func TestServer_CallTool_GetProjectReturnsConfig(t *testing.T) {
	e := newTestEnv(t)
	e.expectAuth(adminUser())
	e.projectSvc.showFn = func(_ context.Context, req *projectpb.ShowRequest) (*projectpb.ShowResponse, error) {
		assert.Equal(t, int32(7), req.GetId())
		return &projectpb.ShowResponse{Item: &apitypes.ProjectModel{
			Id: 7, Name: "svc", Config: "image:\n  tag: v1\nreplicas: 1", Version: 3,
		}}, nil
	}

	sessionID := e.initSession(t, testToken)
	code, _, body := e.post(t, sessionID, testToken,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_project","arguments":{"project_id":7}}}`)
	require.Equal(t, http.StatusOK, code, body)

	res := sseJSON(t, body)["result"].(map[string]any)
	require.NotEqual(t, true, res["isError"], "get_project 不应报错: %v", body)
	content := res["content"].([]any)[0].(map[string]any)["text"].(string)
	assert.Contains(t, content, `"config":"image:\n  tag: v1\nreplicas: 1"`)
}

// TestServer_CallTool_PermissionDenied 验证越权访问被 RBAC 拦下：普通用户访问私有
// 空间时工具返回错误结果（isError=true），而不是泄漏空间内容。用仍走真实 biz 门卫的
// list_namespace_events 承载本断言（其余读工具的门卫已下沉到 services fake 内部）。
func TestServer_CallTool_PermissionDenied(t *testing.T) {
	e := newTestEnv(t)
	e.expectAuth(normalUser())
	e.nsRepo.EXPECT().FindByName(gomock.Any(), "private-ns").Return(
		&biz.Namespace{ID: 9, Name: "private-ns", Private: true, CreatorEmail: "someone-else@mars.com"}, nil)

	sessionID := e.initSession(t, testToken)
	code, _, body := e.post(t, sessionID, testToken,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_namespace_events","arguments":{"namespace":"private-ns"}}}`)
	require.Equal(t, http.StatusOK, code, body)

	res := sseJSON(t, body)["result"].(map[string]any)
	assert.Equal(t, true, res["isError"], "越权访问必须返回 isError=true")
}

// TestServer_CallTool_UnknownNamespace 验证 biz 层返回的 NotFound 原样透传给客户端，
// 让模型能把「空间不存在」和「无权限」区分开。
func TestServer_CallTool_UnknownNamespace(t *testing.T) {
	e := newTestEnv(t)
	e.expectAuth(adminUser())
	e.nsRepo.EXPECT().FindByName(gomock.Any(), "nope").Return(nil, errors.New("namespace not found"))

	sessionID := e.initSession(t, testToken)
	code, _, body := e.post(t, sessionID, testToken,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_namespace_events","arguments":{"namespace":"nope"}}}`)
	require.Equal(t, http.StatusOK, code, body)

	res := sseJSON(t, body)["result"].(map[string]any)
	assert.Equal(t, true, res["isError"])
	content := res["content"].([]any)[0].(map[string]any)["text"].(string)
	assert.Contains(t, content, "namespace not found")
}

// TestServer_CallTool_GetContainerLogsTruncated 验证超大容器日志在 MCP 侧被截断，
// 且被截断的日志串尾部带上提示语（出参形状仍是 ContainerSvc 的原始模型），
// 避免撑爆模型上下文。
func TestServer_CallTool_GetContainerLogsTruncated(t *testing.T) {
	e := newTestEnv(t)
	e.expectAuth(adminUser())
	e.contSvc.containerLogFn = func(_ context.Context, req *containerpb.LogRequest) (*containerpb.LogResponse, error) {
		assert.Equal(t, "team-a", req.GetNamespace())
		return &containerpb.LogResponse{
			Namespace: "team-a", PodName: "p1", ContainerName: "app",
			Log: strings.Repeat("x", maxLogBytes+1024),
		}, nil
	}

	sessionID := e.initSession(t, testToken)
	code, _, body := e.post(t, sessionID, testToken,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_container_logs","arguments":{"namespace":"team-a","pod":"p1","container":"app"}}}`)
	require.Equal(t, http.StatusOK, code, body)

	res := sseJSON(t, body)["result"].(map[string]any)
	require.NotEqual(t, true, res["isError"], "日志读取不应报错: %v", body)
	content := res["content"].([]any)[0].(map[string]any)["text"].(string)
	assert.Contains(t, content, "已截断")
	assert.LessOrEqual(t, len(content), maxLogBytes+512, "文本内容必须被截断到上限附近")
}

// Test_truncateLog 覆盖日志截断的纯函数边界：未超限原样返回、超限截断且不切断
// 多字节 UTF-8 字符。
func Test_truncateLog(t *testing.T) {
	short := strings.Repeat("a", 10)
	got, truncated := truncateLog(short)
	assert.Equal(t, short, got)
	assert.False(t, truncated)

	long := strings.Repeat("a", maxLogBytes+1)
	got, truncated = truncateLog(long)
	assert.True(t, truncated)
	assert.Len(t, got, maxLogBytes)

	// 末尾多字节字符被截半时必须回退到完整边界。
	multi := strings.Repeat("中", maxLogBytes/3+10) // 每个「中」占 3 字节
	got, truncated = truncateLog(multi)
	assert.True(t, truncated)
	assert.LessOrEqual(t, len(got), maxLogBytes)
	assert.True(t, isValidUTF8Boundary(got))
}

// Test_formatTime 覆盖事件时间格式化的零值分支。
func Test_formatTime(t *testing.T) {
	assert.Equal(t, "", formatTime(metav1TimeZero()))
	assert.Equal(t, "2026-01-02T15:04:05Z", formatTime(metav1Time(2026, 1, 2, 15, 4, 5)))
}

// Test_toProjectSummary 覆盖项目概要投影：字段逐项搬运，零值枚举由 proto 渲染。
func Test_toProjectSummary(t *testing.T) {
	got := toProjectSummary(&biz.Project{
		ID: 7, Name: "svc", Version: 3, GitBranch: "main", GitCommit: "abc", UpdatedBy: "admin",
	})
	assert.Equal(t, 7, got.ID)
	assert.Equal(t, "svc", got.Name)
	assert.Equal(t, 3, got.Version)
	assert.Equal(t, "main", got.GitBranch)
	assert.Equal(t, "abc", got.GitCommit)
	assert.Equal(t, "admin", got.UpdatedBy)
	// 零值枚举的字符串形式由 proto 生成，断言非空即可（枚举名随 proto 演进）。
	assert.NotEmpty(t, got.DeployStatus)
}

// TestNewServer_TouchActivityError 覆盖活跃度刷新失败分支：只记 Warning，不得阻断
// MCP 请求——活跃度是空闲缩容的优化依据，不能变成请求的故障点。
func TestNewServer_TouchActivityError(t *testing.T) {
	ctrl := gomock.NewController(t)
	authBiz := biz.NewMockAuthBiz(ctrl)
	authBiz.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(adminUser(), nil)
	authBiz.EXPECT().EffectiveRoles(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]string{schematype.MarsAdmin}, nil)
	activity := biz.NewMockAgentActivity(ctrl)
	activity.EXPECT().Touch(gomock.Any(), biz.AgentProjectName("admin@mars.com")).
		Return(errors.New("cache down")).Times(1)

	server := NewServer(Deps{
		AuthBiz:       authBiz,
		AgentActivity: activity,
		Logger:        mlog.NewForConfig(nil),
	})
	httpSrv := httptest.NewServer(server.handler)
	t.Cleanup(httpSrv.Close)

	req, err := http.NewRequest(http.MethodPost, httpSrv.URL+Path, strings.NewReader(
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", testToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "活跃度刷新失败不得阻断 MCP 请求")
}

// TestNewServer_RegistersMcpRoute 验证服务端门面装配完成且路径契约固定为 /mcp。
func TestNewServer_RegistersMcpRoute(t *testing.T) {
	e := newTestEnv(t)
	assert.NotNil(t, e.server)
	// Path 契约：沙箱侧 agent/config.go 按 "/mcp" 拼接 MCP 地址，改动需同步。
	assert.Equal(t, "/mcp", Path)
	assert.NotNil(t, e.server.handler)
}
