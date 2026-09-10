package eventhandler

import (
	"context"
	"errors"
	"testing"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/event"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newTestCoordinator 构造事件协调器：注入可替换的 getCerts/toAll 与 mock repo，
// 隔离每个 handler 的分支验证。dispatcher 字段留空——handler 不经 dispatcher。
func newTestCoordinator(m *gomock.Controller, getCerts func() (string, string, string), toAll func(proto.Message) error) (*EventCoordinator, *data.MockProjectRepo, *data.MockK8sRepo, *data.MockChangelogRepo, *data.MockEventRepo) {
	pr := data.NewMockProjectRepo(m)
	kr := data.NewMockK8sRepo(m)
	cl := data.NewMockChangelogRepo(m)
	er := data.NewMockEventRepo(m)
	if getCerts == nil {
		getCerts = func() (string, string, string) { return "", "", "" }
	}
	if toAll == nil {
		toAll = func(proto.Message) error { return nil }
	}
	return &EventCoordinator{
		logger:      mlog.NewForConfig(nil),
		getCerts:    getCerts,
		toAll:       toAll,
		projectRepo: pr,
		k8sRepo:     kr,
		clRepo:      cl,
		eventRepo:   er,
	}, pr, kr, cl, er
}

// TestNewEventCoordinator_RegistersListeners 验证构造期在 dispatcher 注册了
// 5 个业务事件监听（4 个跨域生命周期 + 1 个 audit 落库，每个事件恰好一个），
// 且不调用插件闭包（惰性闭包不解析）。
func TestNewEventCoordinator_RegistersListeners(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	disp := event.NewDispatcher(mlog.NewForConfig(nil))

	ec := NewEventCoordinator(
		mlog.NewForConfig(nil),
		disp,
		data.NewMockProjectRepo(m),
		data.NewMockK8sRepo(m),
		data.NewMockChangelogRepo(m),
		data.NewMockEventRepo(m),
		&PluginDeps{GetCerts: func() (string, string, string) { return "", "", "" }, ToAll: func(proto.Message) error { return nil }},
	)

	assert.NotNil(t, ec)
	for _, key := range []biz.EventKey{biz.EventNamespaceCreated, biz.EventNamespaceDeleted, biz.EventProjectChanged, biz.EventProjectDeleted, biz.AuditLogEvent} {
		assert.Len(t, disp.GetListeners(event.Event(key)), 1)
	}
}

// TestEventCoordinator_HandleInjectTlsSecret 覆盖 TLS 注入三态：
// 空证书跳过、证书就绪成功注入、注入失败原地打日志（错误不上抛）。
func TestEventCoordinator_HandleInjectTlsSecret(t *testing.T) {
	t.Run("空证书跳过注入", func(t *testing.T) {
		m := gomock.NewController(t)
		t.Cleanup(m.Finish)
		c, _, _, _, _ := newTestCoordinator(m, func() (string, string, string) { return "", "", "" }, nil)
		// 无 k8sRepo EXPECT：若被调用会 panic，反向证明未注入。
		err := c.HandleInjectTlsSecret(biz.NamespaceCreatedData{NsK8sObj: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}}}, biz.EventNamespaceCreated)
		assert.NoError(t, err)
	})

	t.Run("证书就绪成功注入", func(t *testing.T) {
		m := gomock.NewController(t)
		t.Cleanup(m.Finish)
		c, _, kr, _, _ := newTestCoordinator(m, func() (string, string, string) { return "n", "k", "crt" }, nil)
		kr.EXPECT().AddTlsSecret("default", "n", "k", "crt").Return(&corev1.Secret{}, nil)
		err := c.HandleInjectTlsSecret(biz.NamespaceCreatedData{NsK8sObj: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}}}, biz.EventNamespaceCreated)
		assert.NoError(t, err)
	})

	t.Run("注入失败仅打日志", func(t *testing.T) {
		m := gomock.NewController(t)
		t.Cleanup(m.Finish)
		c, _, kr, _, _ := newTestCoordinator(m, func() (string, string, string) { return "n", "k", "crt" }, nil)
		kr.EXPECT().AddTlsSecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))
		err := c.HandleInjectTlsSecret(biz.NamespaceCreatedData{NsK8sObj: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}}}, biz.EventNamespaceCreated)
		// 异步监听场景错误无法上抛，handler 吞掉仅打日志，始终返回 nil。
		assert.NoError(t, err)
	})

	t.Run("错误负载类型为 no-op", func(t *testing.T) {
		m := gomock.NewController(t)
		t.Cleanup(m.Finish)
		c, _, _, _, _ := newTestCoordinator(m, nil, nil)
		err := c.HandleInjectTlsSecret("not NamespaceCreatedData", biz.EventNamespaceCreated)
		assert.NoError(t, err)
	})
}

// TestEventCoordinator_HandleNamespaceDeleted 验证广播 reload 消息且携带 namespace id。
func TestEventCoordinator_HandleNamespaceDeleted(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	var got proto.Message
	c, _, _, _, _ := newTestCoordinator(m, nil, func(msg proto.Message) error { got = msg; return nil })

	err := c.HandleNamespaceDeleted(biz.NamespaceDeletedData{ID: 42}, biz.EventNamespaceDeleted)
	assert.NoError(t, err)

	wsMsg, ok := got.(*websocket_pb.WsReloadProjectsResponse)
	assert.True(t, ok)
	assert.Equal(t, websocket_pb.Type_ReloadProjects, wsMsg.Metadata.Type)
	assert.Equal(t, int32(42), wsMsg.NamespaceId)
}

// TestEventCoordinator_HandleProjectChanged 覆盖变更日志落库四态：
// 成功落库带 configChanged 判定、项目读取失败上抛、上一条记录缺失仍落库、
// 错误负载为 no-op。
func TestEventCoordinator_HandleProjectChanged(t *testing.T) {
	proj := &biz.Project{
		ID:          1,
		Version:     3,
		Config:      "cfg",
		GitBranch:   "main",
		GitCommit:   "abc",
		DockerImage: []string{"img"},
	}

	t.Run("成功落库并判定 configChanged", func(t *testing.T) {
		m := gomock.NewController(t)
		t.Cleanup(m.Finish)
		c, pr, _, cl, _ := newTestCoordinator(m, nil, nil)
		pr.EXPECT().Show(gomock.Any(), 1).Return(proj, nil)
		cl.EXPECT().FindLastChangeByProjectID(gomock.Any(), 1).Return(&biz.Changelog{Config: "old", GitCommit: "oldsha"}, nil)
		cl.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, in *biz.CreateChangeLogInput) (*biz.Changelog, error) {
			assert.Equal(t, 3, in.Version)
			assert.Equal(t, "u", in.Username)
			assert.Equal(t, "cfg", in.Config)
			assert.Equal(t, "main", in.GitBranch)
			assert.Equal(t, "abc", in.GitCommit)
			// Config "old"→"cfg" 变化：configChanged 应为 true。
			assert.True(t, in.ConfigChanged)
			assert.Equal(t, 1, in.ProjectID)
			return &biz.Changelog{}, nil
		})

		err := c.HandleProjectChanged(&biz.ProjectChangedData{ID: 1, Username: "u"}, biz.EventProjectChanged)
		assert.NoError(t, err)
	})

	t.Run("项目读取失败上抛且不落库", func(t *testing.T) {
		m := gomock.NewController(t)
		t.Cleanup(m.Finish)
		c, pr, _, _, _ := newTestCoordinator(m, nil, nil)
		pr.EXPECT().Show(gomock.Any(), 1).Return(nil, errors.New("boom"))
		// 无 clRepo.Create EXPECT：若被调用会 panic。
		err := c.HandleProjectChanged(&biz.ProjectChangedData{ID: 1}, biz.EventProjectChanged)
		assert.Error(t, err)
	})

	t.Run("上一条记录缺失仍落库且 configChanged 为 false", func(t *testing.T) {
		m := gomock.NewController(t)
		t.Cleanup(m.Finish)
		c, pr, _, cl, _ := newTestCoordinator(m, nil, nil)
		pr.EXPECT().Show(gomock.Any(), 1).Return(proj, nil)
		cl.EXPECT().FindLastChangeByProjectID(gomock.Any(), 1).Return(nil, errors.New("no last change"))
		cl.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, in *biz.CreateChangeLogInput) (*biz.Changelog, error) {
			assert.False(t, in.ConfigChanged)
			return &biz.Changelog{}, nil
		})

		err := c.HandleProjectChanged(&biz.ProjectChangedData{ID: 1}, biz.EventProjectChanged)
		assert.NoError(t, err)
	})

	t.Run("落库失败仅打日志", func(t *testing.T) {
		m := gomock.NewController(t)
		t.Cleanup(m.Finish)
		c, pr, _, cl, _ := newTestCoordinator(m, nil, nil)
		pr.EXPECT().Show(gomock.Any(), 1).Return(proj, nil)
		cl.EXPECT().FindLastChangeByProjectID(gomock.Any(), 1).Return(nil, errors.New("no last change"))
		cl.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("db boom"))
		err := c.HandleProjectChanged(&biz.ProjectChangedData{ID: 1}, biz.EventProjectChanged)
		// 落库失败仅打日志，handler 仍返回 nil。
		assert.NoError(t, err)
	})

	t.Run("错误负载类型为 no-op", func(t *testing.T) {
		m := gomock.NewController(t)
		t.Cleanup(m.Finish)
		c, _, _, _, _ := newTestCoordinator(m, nil, nil)
		err := c.HandleProjectChanged("not ProjectChangedData", biz.EventProjectChanged)
		assert.NoError(t, err)
	})
}

// TestEventCoordinator_FiresRegisteredListeners 覆盖 NewEventCoordinator 的惰性
// 插件闭包（GetCerts/ToAll）与 listen 内联回调体：同步取出 dispatcher 上注册的
// 监听并触发，走完整的闭包 → 用例链（wire 期构造、运行期触发的时序）。
func TestEventCoordinator_FiresRegisteredListeners(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)

	// 惰性闭包：GetCerts 提供证书，ToAll 记录广播次数。
	var broadcastCount int
	deps := &PluginDeps{
		GetCerts: func() (string, string, string) { return "n", "k", "crt" },
		ToAll: func(proto.Message) error {
			broadcastCount++
			return nil
		},
	}

	pr := data.NewMockProjectRepo(m)
	pr.EXPECT().Show(gomock.Any(), 5).Return(&biz.Project{ID: 5, Version: 1, Config: "cfg"}, nil)
	kr := data.NewMockK8sRepo(m)
	kr.EXPECT().AddTlsSecret("default", "n", "k", "crt").Return(&corev1.Secret{}, nil)
	cl := data.NewMockChangelogRepo(m)
	cl.EXPECT().FindLastChangeByProjectID(gomock.Any(), 5).Return(nil, errors.New("no last change"))
	cl.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&biz.Changelog{}, nil)
	er := data.NewMockEventRepo(m)
	er.EXPECT().HandleAuditLog("audit-payload", biz.AuditLogEvent).Return(nil)

	disp := event.NewDispatcher(mlog.NewForConfig(nil))
	NewEventCoordinator(mlog.NewForConfig(nil), disp, pr, kr, cl, er, deps)

	// 同步取出监听回调并触发（GetListeners 返回拷贝，事件 key 精确匹配）。
	created := disp.GetListeners(event.Event(biz.EventNamespaceCreated))[0]
	assert.NoError(t, created(biz.NamespaceCreatedData{NsK8sObj: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}}}, event.Event(biz.EventNamespaceCreated)))
	deleted := disp.GetListeners(event.Event(biz.EventNamespaceDeleted))[0]
	assert.NoError(t, deleted(biz.NamespaceDeletedData{ID: 42}, event.Event(biz.EventNamespaceDeleted)))
	changed := disp.GetListeners(event.Event(biz.EventProjectChanged))[0]
	assert.NoError(t, changed(&biz.ProjectChangedData{ID: 5, Username: "u"}, event.Event(biz.EventProjectChanged)))
	projDeleted := disp.GetListeners(event.Event(biz.EventProjectDeleted))[0]
	assert.NoError(t, projDeleted(&biz.ProjectDeletedPayload{NamespaceID: 7, ProjectID: 9}, event.Event(biz.EventProjectDeleted)))
	audit := disp.GetListeners(event.Event(biz.AuditLogEvent))[0]
	assert.NoError(t, audit("audit-payload", event.Event(biz.AuditLogEvent)))

	// HandleNamespaceDeleted 与 HandleProjectDeleted 各触发一次 ToAll（audit 不广播）。
	assert.Equal(t, 2, broadcastCount)
}

// TestEventCoordinator_HandleProjectDeleted 验证广播 reload 消息且携带 namespace id。
func TestEventCoordinator_HandleProjectDeleted(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	var got proto.Message
	c, _, _, _, _ := newTestCoordinator(m, nil, func(msg proto.Message) error { got = msg; return nil })

	err := c.HandleProjectDeleted(&biz.ProjectDeletedPayload{NamespaceID: 7, ProjectID: 9}, biz.EventProjectDeleted)
	assert.NoError(t, err)

	wsMsg, ok := got.(*websocket_pb.WsReloadProjectsResponse)
	assert.True(t, ok)
	assert.Equal(t, websocket_pb.Type_ReloadProjects, wsMsg.Metadata.Type)
	assert.Equal(t, int32(7), wsMsg.NamespaceId)
}

// TestEventCoordinator_HandleAuditLog 覆盖 audit 委托两分支：
// 成功透传事件Repo 的 nil，落库失败错误原样返回（由 dispatcher 统一消费）。
func TestEventCoordinator_HandleAuditLog(t *testing.T) {
	t.Run("成功委托落库", func(t *testing.T) {
		m := gomock.NewController(t)
		t.Cleanup(m.Finish)
		c, _, _, _, er := newTestCoordinator(m, nil, nil)
		er.EXPECT().HandleAuditLog("audit-payload", biz.AuditLogEvent).Return(nil)
		assert.NoError(t, c.HandleAuditLog("audit-payload", biz.AuditLogEvent))
	})

	t.Run("落库失败错误原样返回", func(t *testing.T) {
		m := gomock.NewController(t)
		t.Cleanup(m.Finish)
		c, _, _, _, er := newTestCoordinator(m, nil, nil)
		er.EXPECT().HandleAuditLog("audit-payload", biz.AuditLogEvent).Return(errors.New("db boom"))
		assert.Error(t, c.HandleAuditLog("audit-payload", biz.AuditLogEvent))
	})
}
