package mcp

import (
	"context"
	"errors"
	"strings"
	"testing"

	changelogpb "github.com/duc-cnzj/mars/api/v6/proto/changelog"
	containerpb "github.com/duc-cnzj/mars/api/v6/proto/container"
	metricspb "github.com/duc-cnzj/mars/api/v6/proto/metrics"
	namespacepb "github.com/duc-cnzj/mars/api/v6/proto/namespace"
	projectpb "github.com/duc-cnzj/mars/api/v6/proto/project"
	apitypes "github.com/duc-cnzj/mars/api/v6/proto/types"
	websocketpb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
)

// 本文件逐工具直调 toolset 方法，覆盖每个工具的入参映射、出参回吐与错误分支。
// 出参口径（2026-10-16 拍板）：能复用平台 gRPC 服务接口的工具一律复用，出参直接
// 回吐 proto 模型（声明为 any → 不生成 outputSchema），故断言从「手工投影结构体」
// 改为「proto 模型的 GetXxx()」。传输层（HTTP/MCP 协议/鉴权）由 mcp_test.go 覆盖。

// fakeContainerBiz 是 biz.ContainerBiz 的测试替身：只覆写 ExecOnce，让 exec_command 的
// 错误帧码映射（截断/超时/启动失败/退出码）可以由测试逐帧精确投喂，不依赖真实 k8s 执行。
type fakeContainerBiz struct {
	biz.ContainerBiz
	execOnceFn func(context.Context, biz.ExecOnceStream, *biz.UserInfo, *biz.ExecOnceInput) error
}

// ExecOnce 委派给测试注入的 execOnceFn。
func (f *fakeContainerBiz) ExecOnce(ctx context.Context, s biz.ExecOnceStream, u *biz.UserInfo, in *biz.ExecOnceInput) error {
	return f.execOnceFn(ctx, s, u, in)
}

// TestToolset_ListNamespaces 覆盖入参搬运（名称过滤 + 分页指针）与下游报错上抛。
func TestToolset_ListNamespaces(t *testing.T) {
	t.Run("名称过滤与分页搬进 proto 请求", func(t *testing.T) {
		e := newTestEnv(t)
		var got *namespacepb.ListRequest
		e.nsSvc.listFn = func(_ context.Context, req *namespacepb.ListRequest) (*namespacepb.ListResponse, error) {
			got = req
			return &namespacepb.ListResponse{
				Items: []*apitypes.NamespaceModel{{Id: 1, Name: "team-a", Private: true}},
				Count: 1, Page: 1, PageSize: 15,
			}, nil
		}

		res, out, err := e.toolset().listNamespaces(ctxWithUser(adminUser()), nil, listNamespacesIn{Name: "team", Page: 2, PageSize: 5})
		require.NoError(t, err)
		assert.Nil(t, res)
		require.NotNil(t, got.Name)
		assert.Equal(t, "team", got.GetName())
		assert.Equal(t, int32(2), got.GetPage())
		assert.Equal(t, int32(5), got.GetPageSize())

		resp := out.(*namespacepb.ListResponse)
		require.Len(t, resp.GetItems(), 1)
		assert.Equal(t, "team-a", resp.GetItems()[0].GetName())
		assert.Equal(t, int32(1), resp.GetCount())
	})

	t.Run("未传过滤词时 Name 保持 nil", func(t *testing.T) {
		e := newTestEnv(t)
		e.nsSvc.listFn = func(_ context.Context, req *namespacepb.ListRequest) (*namespacepb.ListResponse, error) {
			assert.Nil(t, req.Name)
			return &namespacepb.ListResponse{}, nil
		}

		_, _, err := e.toolset().listNamespaces(ctxWithUser(adminUser()), nil, listNamespacesIn{})
		require.NoError(t, err)
	})

	t.Run("下游报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.nsSvc.listFn = func(context.Context, *namespacepb.ListRequest) (*namespacepb.ListResponse, error) {
			return nil, errors.New("db down")
		}

		_, _, err := e.toolset().listNamespaces(ctxWithUser(adminUser()), nil, listNamespacesIn{})
		require.ErrorContains(t, err, "db down")
	})
}

// TestToolset_GetCurrentUser 返回 ctx 中由鉴权中间件注入的身份与权限属性；
// 无用户上下文（鉴权链路故障）时上抛错误，不返回空身份。
func TestToolset_GetCurrentUser(t *testing.T) {
	t.Run("管理员身份", func(t *testing.T) {
		e := newTestEnv(t)
		_, out, err := e.toolset().getCurrentUser(ctxWithUser(adminUser()), nil, getCurrentUserIn{})
		require.NoError(t, err)
		got := out.(getCurrentUserOut)
		assert.Equal(t, "1", got.ID)
		assert.Equal(t, "admin@mars.com", got.Email)
		assert.Equal(t, "admin", got.Name)
		assert.Equal(t, []string{schematype.MarsAdmin}, got.Roles)
		assert.True(t, got.IsAdmin)
		assert.False(t, got.IsSuperAdmin)
	})

	t.Run("普通用户", func(t *testing.T) {
		e := newTestEnv(t)
		_, out, err := e.toolset().getCurrentUser(ctxWithUser(normalUser()), nil, getCurrentUserIn{})
		require.NoError(t, err)
		got := out.(getCurrentUserOut)
		assert.False(t, got.IsAdmin)
		assert.NotNil(t, got.Roles, "无角色时应回吐空数组而不是 nil（JSON null）")
		assert.Empty(t, got.Roles)
	})

	t.Run("无用户上下文", func(t *testing.T) {
		e := newTestEnv(t)
		_, _, err := e.toolset().getCurrentUser(context.Background(), nil, getCurrentUserIn{})
		require.ErrorContains(t, err, "user not found")
	})
}

// TestToolset_ListProjects 覆盖空间门卫后的项目列举与越权拒绝——它是仍在 biz 路径上的
// 手工投影工具（平台无「按空间名 + 项目名过滤」的现成接口）。
func TestToolset_ListProjects(t *testing.T) {
	t.Run("门卫通过后列举项目", func(t *testing.T) {
		e := newTestEnv(t)
		e.nsRepo.EXPECT().FindByName(gomock.Any(), "team-a").Return(&biz.Namespace{ID: 2, Name: "team-a"}, nil)
		var got *biz.ListProjectInput
		e.projRepo.EXPECT().List(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, in *biz.ListProjectInput) ([]*biz.Project, *pagination.Pagination, error) {
				got = in
				return []*biz.Project{{ID: 7, Name: "svc"}}, &pagination.Pagination{Count: 1, Page: 1, PageSize: 15}, nil
			})

		_, out, err := e.toolset().listProjects(ctxWithUser(adminUser()), nil, listProjectsIn{Namespace: "team-a", Name: "svc"})
		require.NoError(t, err)
		assert.Equal(t, int32(2), got.NamespaceID)
		assert.Equal(t, "svc", got.Name)
		res := out.(listProjectsOut)
		require.Len(t, res.Projects, 1)
		assert.Equal(t, 7, res.Projects[0].ID)
		assert.Equal(t, int32(1), res.Total)
	})

	t.Run("私有空间越权拒绝", func(t *testing.T) {
		e := newTestEnv(t)
		e.nsRepo.EXPECT().FindByName(gomock.Any(), "private-ns").Return(
			&biz.Namespace{ID: 9, Name: "private-ns", Private: true, CreatorEmail: "other@mars.com"}, nil)

		_, _, err := e.toolset().listProjects(ctxWithUser(normalUser()), nil, listProjectsIn{Namespace: "private-ns"})
		require.Error(t, err)
	})

	t.Run("项目列举报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.nsRepo.EXPECT().FindByName(gomock.Any(), "team-a").Return(&biz.Namespace{ID: 2, Name: "team-a"}, nil)
		e.projRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, nil, errors.New("list fail"))

		_, _, err := e.toolset().listProjects(ctxWithUser(adminUser()), nil, listProjectsIn{Namespace: "team-a"})
		require.ErrorContains(t, err, "list fail")
	})
}

// TestToolset_GetProject 覆盖项目详情的原样回吐：出参必须是 ProjectSvc.Show 的项目模型
// （含 config——模型判断「该改哪个键」的唯一依据），MCP 侧不再另造投影结构。
func TestToolset_GetProject(t *testing.T) {
	e := newTestEnv(t)
	var got *projectpb.ShowRequest
	e.projectSvc.showFn = func(_ context.Context, req *projectpb.ShowRequest) (*projectpb.ShowResponse, error) {
		got = req
		return &projectpb.ShowResponse{Item: &apitypes.ProjectModel{
			Id: 7, Name: "svc", Config: "image:\n  tag: v1\nreplicas: 1", Version: 3,
		}}, nil
	}

	_, out, err := e.toolset().getProject(ctxWithUser(adminUser()), nil, getProjectIn{ProjectID: 7})
	require.NoError(t, err)
	assert.Equal(t, int32(7), got.GetId())
	model := out.(*apitypes.ProjectModel)
	assert.Equal(t, int32(7), model.GetId())
	assert.Equal(t, "svc", model.GetName())
	assert.Equal(t, "image:\n  tag: v1\nreplicas: 1", model.GetConfig())
	assert.Equal(t, int32(3), model.GetVersion())
}

// TestToolset_GetProjectNotFound 覆盖项目不存在时错误原样上抛（含 isNotFound 语义不被吞）。
func TestToolset_GetProjectNotFound(t *testing.T) {
	e := newTestEnv(t)
	e.projectSvc.showFn = func(context.Context, *projectpb.ShowRequest) (*projectpb.ShowResponse, error) {
		return nil, errors.New("project not found")
	}

	_, _, err := e.toolset().getProject(ctxWithUser(adminUser()), nil, getProjectIn{ProjectID: 7})
	require.ErrorContains(t, err, "project not found")
}

// TestToolset_ListProjectContainers 覆盖容器视图的原样回吐（出参即 StateContainer 模型）。
func TestToolset_ListProjectContainers(t *testing.T) {
	e := newTestEnv(t)
	var got *projectpb.AllContainersRequest
	e.projectSvc.allContainersFn = func(_ context.Context, req *projectpb.AllContainersRequest) (*projectpb.AllContainersResponse, error) {
		got = req
		return &projectpb.AllContainersResponse{Items: []*apitypes.StateContainer{
			{Namespace: "team-a", Pod: "web-0", Container: "app", Ready: true},
		}}, nil
	}

	_, out, err := e.toolset().listProjectContainers(ctxWithUser(adminUser()), nil, listProjectContainersIn{ProjectID: 7})
	require.NoError(t, err)
	assert.Equal(t, int32(7), got.GetId())
	resp := out.(*projectpb.AllContainersResponse)
	require.Len(t, resp.GetItems(), 1)
	assert.Equal(t, "web-0", resp.GetItems()[0].GetPod())
	assert.True(t, resp.GetItems()[0].GetReady())
}

// TestToolset_CheckPodStatus 覆盖非运行态原因回传（出参即 IsPodRunningResponse）与下游报错。
func TestToolset_CheckPodStatus(t *testing.T) {
	t.Run("回吐运行态与原因", func(t *testing.T) {
		e := newTestEnv(t)
		var got *containerpb.IsPodRunningRequest
		e.contSvc.isPodRunningFn = func(_ context.Context, req *containerpb.IsPodRunningRequest) (*containerpb.IsPodRunningResponse, error) {
			got = req
			return &containerpb.IsPodRunningResponse{Running: false, Reason: "CrashLoopBackOff"}, nil
		}

		_, out, err := e.toolset().checkPodStatus(ctxWithUser(adminUser()), nil, checkPodStatusIn{Namespace: "team-a", Pod: "web-0"})
		require.NoError(t, err)
		assert.Equal(t, "team-a", got.GetNamespace())
		assert.Equal(t, "web-0", got.GetPod())
		resp := out.(*containerpb.IsPodRunningResponse)
		assert.False(t, resp.GetRunning())
		assert.Equal(t, "CrashLoopBackOff", resp.GetReason())
	})

	t.Run("下游报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.contSvc.isPodRunningFn = func(context.Context, *containerpb.IsPodRunningRequest) (*containerpb.IsPodRunningResponse, error) {
			return nil, errors.New("pod lookup failed")
		}

		_, _, err := e.toolset().checkPodStatus(ctxWithUser(adminUser()), nil, checkPodStatusIn{Namespace: "team-a", Pod: "web-0"})
		require.ErrorContains(t, err, "pod lookup failed")
	})
}

// TestToolset_ListNamespaceEvents 覆盖事件精简投影（含对象拼接与时间格式化）——它是
// 仍在 biz 路径上的手工投影工具（EventSvc.List 查的是平台审计日志，不是 k8s 事件）。
func TestToolset_ListNamespaceEvents(t *testing.T) {
	e := newTestEnv(t)
	e.nsRepo.EXPECT().FindByName(gomock.Any(), "team-a").Return(&biz.Namespace{ID: 2, Name: "team-a"}, nil)
	e.k8sRepo.EXPECT().ListEvents("team-a").Return([]*eventsv1.Event{{
		Type: "Warning", Reason: "BackOff", Note: "Back-off restarting failed container",
		Regarding:               corev1.ObjectReference{Kind: "Pod", Name: "web-0", FieldPath: "spec.containers{app}"},
		DeprecatedCount:         5,
		DeprecatedLastTimestamp: metav1Time(2026, 1, 2, 15, 4, 5),
	}}, nil)

	_, out, err := e.toolset().listNamespaceEvents(ctxWithUser(adminUser()), nil, listNamespaceEventsIn{Namespace: "team-a"})
	require.NoError(t, err)
	res := out.(listNamespaceEventsOut)
	require.Len(t, res.Events, 1)
	assert.Equal(t, "Pod/web-0", res.Events[0].Object)
	assert.Equal(t, "spec.containers{app}", res.Events[0].Container)
	assert.Equal(t, int32(5), res.Events[0].Count)
	assert.Equal(t, "2026-01-02T15:04:05Z", res.Events[0].LastSeen)
}

// TestToolset_GetContainerLogs 覆盖下游报错上抛、正常回吐与超限截断三条路径。
// 截断后必须把提示语追加进响应模型（出参形状不变，仍由 ContainerSvc 定）。
func TestToolset_GetContainerLogs(t *testing.T) {
	t.Run("下游报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.contSvc.containerLogFn = func(context.Context, *containerpb.LogRequest) (*containerpb.LogResponse, error) {
			return nil, errors.New("pod not running")
		}

		_, _, err := e.toolset().getContainerLogs(ctxWithUser(adminUser()), nil,
			getContainerLogsIn{Namespace: "team-a", Pod: "gone", Container: "app"})
		require.ErrorContains(t, err, "pod not running")
	})

	t.Run("未超限原样回吐日志", func(t *testing.T) {
		e := newTestEnv(t)
		var got *containerpb.LogRequest
		e.contSvc.containerLogFn = func(_ context.Context, req *containerpb.LogRequest) (*containerpb.LogResponse, error) {
			got = req
			return &containerpb.LogResponse{Namespace: "team-a", PodName: "web-0", ContainerName: "app", Log: "hello log"}, nil
		}

		_, out, err := e.toolset().getContainerLogs(ctxWithUser(adminUser()), nil,
			getContainerLogsIn{Namespace: "team-a", Pod: "web-0", Container: "app", ShowEvents: true, TailLines: 42})
		require.NoError(t, err)
		assert.Equal(t, "team-a", got.GetNamespace())
		assert.Equal(t, "web-0", got.GetPod())
		assert.Equal(t, "app", got.GetContainer())
		assert.True(t, got.GetShowEvents())
		assert.Equal(t, int64(42), got.GetTailLines())
		resp := out.(*containerpb.LogResponse)
		assert.Equal(t, "hello log", resp.GetLog())
	})

	t.Run("超限截断并追加提示语", func(t *testing.T) {
		e := newTestEnv(t)
		e.contSvc.containerLogFn = func(context.Context, *containerpb.LogRequest) (*containerpb.LogResponse, error) {
			return &containerpb.LogResponse{Log: strings.Repeat("x", maxLogBytes+1024)}, nil
		}

		_, out, err := e.toolset().getContainerLogs(ctxWithUser(adminUser()), nil,
			getContainerLogsIn{Namespace: "team-a", Pod: "web-0", Container: "app"})
		require.NoError(t, err)
		resp := out.(*containerpb.LogResponse)
		assert.Contains(t, resp.GetLog(), logTruncatedNotice)
		assert.Len(t, resp.GetLog(), maxLogBytes+len(logTruncatedNotice))
	})

	t.Run("tail_lines 负数直接拒绝且不触碰下游", func(t *testing.T) {
		e := newTestEnv(t)
		// contSvc 的 containerLogFn 未注入：若校验没拦住而调用下游会 panic。
		_, _, err := e.toolset().getContainerLogs(ctxWithUser(adminUser()), nil,
			getContainerLogsIn{Namespace: "team-a", Pod: "web-0", Container: "app", TailLines: -1})
		require.ErrorContains(t, err, "tail_lines 不能为负数")
	})

	t.Run("省略 container 且多容器 Pod 时要求显式指定", func(t *testing.T) {
		e := newTestEnv(t)
		e.nsRepo.EXPECT().FindByName(gomock.Any(), "team-a").Return(&biz.Namespace{ID: 1, Name: "team-a"}, nil)
		e.k8sRepo.EXPECT().GetPod("team-a", "web-0").Return(&corev1.Pod{
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}, {Name: "sidecar"}}},
		}, nil)

		_, _, err := e.toolset().getContainerLogs(ctxWithUser(adminUser()), nil,
			getContainerLogsIn{Namespace: "team-a", Pod: "web-0"})
		require.ErrorContains(t, err, "必须显式指定 container")
		require.ErrorContains(t, err, "app, sidecar")
	})

	t.Run("省略 container 且单容器 Pod 时正常取日志", func(t *testing.T) {
		e := newTestEnv(t)
		e.nsRepo.EXPECT().FindByName(gomock.Any(), "team-a").Return(&biz.Namespace{ID: 1, Name: "team-a"}, nil)
		e.k8sRepo.EXPECT().GetPod("team-a", "web-0").Return(&corev1.Pod{
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}}},
		}, nil)
		e.contSvc.containerLogFn = func(_ context.Context, req *containerpb.LogRequest) (*containerpb.LogResponse, error) {
			assert.Empty(t, req.GetContainer())
			return &containerpb.LogResponse{Log: "hello log"}, nil
		}

		_, out, err := e.toolset().getContainerLogs(ctxWithUser(adminUser()), nil,
			getContainerLogsIn{Namespace: "team-a", Pod: "web-0"})
		require.NoError(t, err)
		assert.Equal(t, "hello log", out.(*containerpb.LogResponse).GetLog())
	})

	t.Run("省略 container 且 Pod 读取失败时原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.nsRepo.EXPECT().FindByName(gomock.Any(), "team-a").Return(&biz.Namespace{ID: 1, Name: "team-a"}, nil)
		e.k8sRepo.EXPECT().GetPod("team-a", "gone").Return(nil, errors.New("pod gone"))
		// contSvc 未注入：GetPod 失败必须在调用下游之前返回。

		_, _, err := e.toolset().getContainerLogs(ctxWithUser(adminUser()), nil,
			getContainerLogsIn{Namespace: "team-a", Pod: "gone"})
		require.ErrorContains(t, err, "pod gone")
	})

	t.Run("省略 container 时先过命名空间门卫再读 Pod", func(t *testing.T) {
		e := newTestEnv(t)
		e.nsRepo.EXPECT().FindByName(gomock.Any(), "private-ns").Return(
			&biz.Namespace{ID: 9, Name: "private-ns", Private: true, CreatorEmail: "other@mars.com"}, nil)
		// Pod 级门卫在命名空间拒绝后会回查 agent 命名空间；NotFound 表示非 agent 空间。
		e.nsRepo.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).Return(nil, errs.WrapNotFound(errors.New("no ns"), "find agent ns"))
		// 未设置 k8sRepo.GetPod 期望：越权场景下不得读 Pod（否则 gomock 直接失败）。

		_, _, err := e.toolset().getContainerLogs(ctxWithUser(normalUser()), nil,
			getContainerLogsIn{Namespace: "private-ns", Pod: "web-0"})
		require.Error(t, err)
	})
}

// TestToolset_GetPodMetrics 覆盖用量原样回吐与下游报错上抛。
func TestToolset_GetPodMetrics(t *testing.T) {
	t.Run("原样回吐 CPU/内存用量", func(t *testing.T) {
		e := newTestEnv(t)
		var got *metricspb.TopPodRequest
		e.metricSvc.topPodFn = func(_ context.Context, req *metricspb.TopPodRequest) (*metricspb.TopPodResponse, error) {
			got = req
			return &metricspb.TopPodResponse{Cpu: 0.1, Memory: 128, HumanizeCpu: "100m", HumanizeMemory: "128Mi"}, nil
		}

		_, out, err := e.toolset().getPodMetrics(ctxWithUser(adminUser()), nil, getPodMetricsIn{Namespace: "team-a", Pod: "web-0"})
		require.NoError(t, err)
		assert.Equal(t, "team-a", got.GetNamespace())
		assert.Equal(t, "web-0", got.GetPod())
		resp := out.(*metricspb.TopPodResponse)
		assert.Equal(t, "100m", resp.GetHumanizeCpu())
		assert.Equal(t, "128Mi", resp.GetHumanizeMemory())
	})

	t.Run("采样失败原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.metricSvc.topPodFn = func(context.Context, *metricspb.TopPodRequest) (*metricspb.TopPodResponse, error) {
			return nil, errors.New("metrics unavailable")
		}

		_, _, err := e.toolset().getPodMetrics(ctxWithUser(adminUser()), nil, getPodMetricsIn{Namespace: "team-a", Pod: "web-0"})
		require.ErrorContains(t, err, "metrics unavailable")
	})
}

// TestToolset_ListProjectChangelogs 覆盖缺省 limit 兜底、显式 limit 就地收窄与
// only_changed 透传（服务固定取最近 15 条，limit 只是在此之上再截）。
func TestToolset_ListProjectChangelogs(t *testing.T) {
	t.Run("缺省 limit 时全部回吐（不足 10 条）", func(t *testing.T) {
		e := newTestEnv(t)
		var got *changelogpb.FindLastChangelogsByProjectIDRequest
		e.changeSvc.findFn = func(_ context.Context, req *changelogpb.FindLastChangelogsByProjectIDRequest) (*changelogpb.FindLastChangelogsByProjectIDResponse, error) {
			got = req
			return &changelogpb.FindLastChangelogsByProjectIDResponse{Items: []*apitypes.ChangelogModel{
				{Version: 3, Username: "admin", GitBranch: "main", ConfigChanged: true},
			}}, nil
		}

		_, out, err := e.toolset().listProjectChangelogs(ctxWithUser(adminUser()), nil, listProjectChangelogsIn{ProjectID: 7})
		require.NoError(t, err)
		assert.Equal(t, int32(7), got.GetProjectId())
		assert.False(t, got.GetOnlyChanged())
		resp := out.(*changelogpb.FindLastChangelogsByProjectIDResponse)
		require.Len(t, resp.GetItems(), 1)
		assert.Equal(t, int32(3), resp.GetItems()[0].GetVersion())
		assert.True(t, resp.GetItems()[0].GetConfigChanged())
	})

	t.Run("显式 limit 就地收窄 + only_changed 透传", func(t *testing.T) {
		e := newTestEnv(t)
		var got *changelogpb.FindLastChangelogsByProjectIDRequest
		e.changeSvc.findFn = func(_ context.Context, req *changelogpb.FindLastChangelogsByProjectIDRequest) (*changelogpb.FindLastChangelogsByProjectIDResponse, error) {
			got = req
			// 服务固定回吐 15 条上限内的结果，MCP 侧负责按 limit 再截。
			items := make([]*apitypes.ChangelogModel, 3)
			for i := range items {
				items[i] = &apitypes.ChangelogModel{Version: int32(i + 1)}
			}
			return &changelogpb.FindLastChangelogsByProjectIDResponse{Items: items}, nil
		}

		_, out, err := e.toolset().listProjectChangelogs(ctxWithUser(adminUser()), nil,
			listProjectChangelogsIn{ProjectID: 7, Limit: 2, OnlyChanged: true})
		require.NoError(t, err)
		assert.True(t, got.GetOnlyChanged())
		resp := out.(*changelogpb.FindLastChangelogsByProjectIDResponse)
		require.Len(t, resp.GetItems(), 2, "limit=2 时只回吐前两条")
		assert.Equal(t, int32(1), resp.GetItems()[0].GetVersion())
	})

	t.Run("下游报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.changeSvc.findFn = func(context.Context, *changelogpb.FindLastChangelogsByProjectIDRequest) (*changelogpb.FindLastChangelogsByProjectIDResponse, error) {
			return nil, errors.New("changelog fail")
		}

		_, _, err := e.toolset().listProjectChangelogs(ctxWithUser(adminUser()), nil, listProjectChangelogsIn{ProjectID: 7})
		require.ErrorContains(t, err, "changelog fail")
	})
}

// TestToolset_ErrorPaths 覆盖仍在 biz 路径上的两个工具（list_projects/list_namespace_events）
// 的门卫与下游失败分支——它们没有 services 接口可复用，错误语义仍需原样透传。
func TestToolset_ErrorPaths(t *testing.T) {
	t.Run("listProjects 空间不存在", func(t *testing.T) {
		e := newTestEnv(t)
		e.nsRepo.EXPECT().FindByName(gomock.Any(), "nope").Return(nil, errors.New("namespace not found"))

		_, _, err := e.toolset().listProjects(ctxWithUser(adminUser()), nil, listProjectsIn{Namespace: "nope"})
		require.ErrorContains(t, err, "namespace not found")
	})

	t.Run("listNamespaceEvents 门卫拒绝", func(t *testing.T) {
		e := newTestEnv(t)
		e.nsRepo.EXPECT().FindByName(gomock.Any(), "nope").Return(nil, errors.New("namespace not found"))

		_, _, err := e.toolset().listNamespaceEvents(ctxWithUser(adminUser()), nil, listNamespaceEventsIn{Namespace: "nope"})
		require.ErrorContains(t, err, "namespace not found")
	})

	t.Run("listNamespaceEvents 事件读取报错", func(t *testing.T) {
		e := newTestEnv(t)
		e.nsRepo.EXPECT().FindByName(gomock.Any(), "team-a").Return(&biz.Namespace{ID: 2, Name: "team-a"}, nil)
		e.k8sRepo.EXPECT().ListEvents("team-a").Return(nil, errors.New("events fail"))

		_, _, err := e.toolset().listNamespaceEvents(ctxWithUser(adminUser()), nil, listNamespaceEventsIn{Namespace: "team-a"})
		require.ErrorContains(t, err, "events fail")
	})

	t.Run("listProjectContainers 服务报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.projectSvc.allContainersFn = func(context.Context, *projectpb.AllContainersRequest) (*projectpb.AllContainersResponse, error) {
			return nil, errors.New("list pods fail")
		}

		_, _, err := e.toolset().listProjectContainers(ctxWithUser(adminUser()), nil, listProjectContainersIn{ProjectID: 7})
		require.ErrorContains(t, err, "list pods fail")
	})
}

// TestToolset_ApplyProject 覆盖入参搬运（显式值优先）、请求级增量补齐（未传字段与自定义元素值
// 沿用项目当前值）、repo_id 恒取自当前绑定、项目不存在必须报错（只更新不创建）与下游报错。
func TestToolset_ApplyProject(t *testing.T) {
	t.Run("全字段显式：显式值优先，自定义元素值照旧继承", func(t *testing.T) {
		e := newTestEnv(t)
		e.projectSvc.showByNameFn = func(context.Context, *projectpb.ShowByNameRequest) (*projectpb.ShowResponse, error) {
			return showByNameItem("image:\n  tag: v1", "release", "deadbeef", 4, 11,
				[]*websocketpb.ExtraValue{{Path: "web->replicas", Value: "3"}}), nil
		}
		var got *projectpb.WebApplyByNameRequest
		e.projectSvc.applyFn = func(_ context.Context, req *projectpb.WebApplyByNameRequest) (*projectpb.WebApplyResponse, error) {
			got = req
			return &projectpb.WebApplyResponse{
				YamlFiles: []string{"kind: Deployment"},
				Project:   &apitypes.ProjectModel{Id: 7, Name: "svc", Version: 1},
			}, nil
		}

		res, out, err := e.toolset().applyProject(ctxWithUser(adminUser()), nil, applyProjectIn{
			Namespace: "team-a", Name: "svc", GitBranch: "main", GitCommit: "abc",
			Config: "conf", Version: 3, DryRun: true,
		})
		require.NoError(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "team-a", got.GetNamespace())
		assert.Equal(t, "svc", got.GetName())
		assert.Equal(t, "main", got.GetGitBranch(), "显式分支不被继承覆盖")
		assert.Equal(t, "abc", got.GetGitCommit())
		assert.Equal(t, "conf", got.GetConfig())
		// repo_id 不是入参：恒为项目当前绑定，本工具不支持改绑定。
		require.NotNil(t, got.RepoId)
		assert.Equal(t, int32(11), got.GetRepoId())
		require.NotNil(t, got.Version)
		assert.Equal(t, int32(3), got.GetVersion())
		assert.True(t, got.GetDryRun())
		// 自定义元素值无入参可传，恒为「继承现状」——本测试反证它不会被清空。
		require.Len(t, got.GetExtraValues(), 1, "自定义元素值必须原样继承，否则会被重置回仓库默认")
		assert.Equal(t, "web->replicas", got.GetExtraValues()[0].GetPath())
		assert.Equal(t, "3", got.GetExtraValues()[0].GetValue())

		resp := out.(*projectpb.WebApplyResponse)
		assert.Equal(t, []string{"kind: Deployment"}, resp.GetYamlFiles())
		assert.Equal(t, int32(7), resp.GetProject().GetId())
		assert.Equal(t, int32(1), resp.GetProject().GetVersion())
	})

	t.Run("增量：未传字段与自定义元素值从项目当前值补齐，显式字段优先", func(t *testing.T) {
		e := newTestEnv(t)
		var showReq *projectpb.ShowByNameRequest
		e.projectSvc.showByNameFn = func(_ context.Context, req *projectpb.ShowByNameRequest) (*projectpb.ShowResponse, error) {
			showReq = req
			return showByNameItem("image:\n  tag: v1", "release", "deadbeef", 4, 11,
				[]*websocketpb.ExtraValue{{Path: "web->replicas", Value: "3"}}), nil
		}
		var got *projectpb.WebApplyByNameRequest
		e.projectSvc.applyFn = func(_ context.Context, req *projectpb.WebApplyByNameRequest) (*projectpb.WebApplyResponse, error) {
			got = req
			return &projectpb.WebApplyResponse{Project: &apitypes.ProjectModel{Id: 7, Name: "svc", Version: 5}}, nil
		}

		// 只传 config（人改配置的典型场景）：分支/提交/版本/自定义元素值都应从当前项目继承。
		_, out, err := e.toolset().applyProject(ctxWithUser(adminUser()), nil,
			applyProjectIn{Namespace: "team-a", Name: "svc", Config: "image:\n  tag: v2"})
		require.NoError(t, err)
		assert.Equal(t, "team-a", showReq.GetNamespace())
		assert.Equal(t, "svc", showReq.GetName())

		assert.Equal(t, "image:\n  tag: v2", got.GetConfig(), "显式传入的 config 覆盖当前配置")
		assert.Equal(t, "release", got.GetGitBranch(), "未传分支应沿用项目当前分支")
		assert.Equal(t, "deadbeef", got.GetGitCommit(), "未传提交应沿用项目当前提交")
		require.NotNil(t, got.RepoId, "repo_id 恒下发项目当前绑定")
		assert.Equal(t, int32(11), got.GetRepoId())
		require.NotNil(t, got.Version, "未传版本应补上刚读到的当前版本")
		assert.Equal(t, int32(4), got.GetVersion())
		require.Len(t, got.GetExtraValues(), 1, "自定义元素值应原样继承，避免 replicas 被重置回仓库默认")
		assert.Equal(t, "web->replicas", got.GetExtraValues()[0].GetPath())
		assert.Equal(t, "3", got.GetExtraValues()[0].GetValue())
		assert.Equal(t, int32(5), out.(*projectpb.WebApplyResponse).GetProject().GetVersion())
	})

	t.Run("extra_values 只改指定元素，其余沿用现状（key 用 final_extra_values 的 path）", func(t *testing.T) {
		e := newTestEnv(t)
		base := []*websocketpb.ExtraValue{
			{Path: "web->replicas", Value: "3"},
			{Path: "web->image", Value: "redis"},
		}
		e.projectSvc.showByNameFn = func(context.Context, *projectpb.ShowByNameRequest) (*projectpb.ShowResponse, error) {
			return showByNameItem("conf", "release", "deadbeef", 4, 11, base), nil
		}
		var got *projectpb.WebApplyByNameRequest
		e.projectSvc.applyFn = func(_ context.Context, req *projectpb.WebApplyByNameRequest) (*projectpb.WebApplyResponse, error) {
			got = req
			return &projectpb.WebApplyResponse{}, nil
		}

		// 只改 replicas：image 必须原样留着，否则 ElementsLoader 会把它落回仓库默认值。
		_, _, err := e.toolset().applyProject(ctxWithUser(adminUser()), nil,
			applyProjectIn{Namespace: "team-a", Name: "svc", ExtraValues: map[string]string{"web->replicas": "5"}})
		require.NoError(t, err)
		require.Len(t, got.GetExtraValues(), 2)
		assert.Equal(t, "web->replicas", got.GetExtraValues()[0].GetPath())
		assert.Equal(t, "5", got.GetExtraValues()[0].GetValue(), "指定元素被覆盖")
		assert.Equal(t, "web->image", got.GetExtraValues()[1].GetPath())
		assert.Equal(t, "redis", got.GetExtraValues()[1].GetValue(), "未指定的元素沿用现状，不得被重置")
		assert.Equal(t, "3", base[0].GetValue(), "不得就地改写 ShowByName 响应里的切片")
	})

	t.Run("项目不存在（NotFound）：直接报错且不触发部署（只更新、不创建）", func(t *testing.T) {
		e := newTestEnv(t)
		e.projectSvc.showByNameFn = func(context.Context, *projectpb.ShowByNameRequest) (*projectpb.ShowResponse, error) {
			return nil, errs.NotFound("项目 svc 在空间 team-a 下不存在或已被删除")
		}
		var applied bool
		e.projectSvc.applyFn = func(context.Context, *projectpb.WebApplyByNameRequest) (*projectpb.WebApplyResponse, error) {
			applied = true
			return &projectpb.WebApplyResponse{}, nil
		}

		_, _, err := e.toolset().applyProject(ctxWithUser(adminUser()), nil,
			applyProjectIn{Namespace: "team-a", Name: "svc", ExtraValues: map[string]string{"web->replicas": "2"}})
		require.Error(t, err)
		assert.False(t, applied, "项目不存在时绝不能把请求落到部署链路，否则会凭空创建项目")
		assert.Contains(t, err.Error(), "只支持更新已有项目")
	})

	t.Run("显式传入的 version 优先于读到的当前版本", func(t *testing.T) {
		e := newTestEnv(t)
		e.projectSvc.showByNameFn = func(context.Context, *projectpb.ShowByNameRequest) (*projectpb.ShowResponse, error) {
			return showByNameItem("replicas: 1", "main", "abc", 4, 11, nil), nil
		}
		var got *projectpb.WebApplyByNameRequest
		e.projectSvc.applyFn = func(_ context.Context, req *projectpb.WebApplyByNameRequest) (*projectpb.WebApplyResponse, error) {
			got = req
			return &projectpb.WebApplyResponse{}, nil
		}

		_, _, err := e.toolset().applyProject(ctxWithUser(adminUser()), nil,
			applyProjectIn{Namespace: "team-a", Name: "svc", Version: 99})
		require.NoError(t, err)
		assert.Equal(t, int32(11), got.GetRepoId(), "repo_id 恒取自当前绑定")
		assert.Equal(t, int32(99), got.GetVersion())
		assert.Equal(t, "main", got.GetGitBranch(), "其余未传字段仍从当前项目继承")
	})

	t.Run("读到的版本为 0 时 version 指针保持 nil（不破坏未指定语义）", func(t *testing.T) {
		e := newTestEnv(t)
		e.projectSvc.showByNameFn = func(context.Context, *projectpb.ShowByNameRequest) (*projectpb.ShowResponse, error) {
			return showByNameItem("replicas: 1", "", "", 0, 0, nil), nil
		}
		var got *projectpb.WebApplyByNameRequest
		e.projectSvc.applyFn = func(_ context.Context, req *projectpb.WebApplyByNameRequest) (*projectpb.WebApplyResponse, error) {
			got = req
			return &projectpb.WebApplyResponse{}, nil
		}

		_, _, err := e.toolset().applyProject(ctxWithUser(adminUser()), nil,
			applyProjectIn{Namespace: "team-a", Name: "svc"})
		require.NoError(t, err)
		// 版本 0 落 0 会被乐观锁判成并发冲突，故保持指针 nil；repo_id 则恒下发（0 在 proto 里与未传等价）。
		assert.Nil(t, got.Version)
		require.NotNil(t, got.RepoId)
		assert.Equal(t, int32(0), got.GetRepoId())
	})

	t.Run("ShowByName 非 NotFound 错误原样上抛（不伪装成项目不存在）", func(t *testing.T) {
		e := newTestEnv(t)
		e.projectSvc.showByNameFn = func(context.Context, *projectpb.ShowByNameRequest) (*projectpb.ShowResponse, error) {
			return nil, errors.New("db down")
		}
		_, _, err := e.toolset().applyProject(ctxWithUser(adminUser()), nil,
			applyProjectIn{Namespace: "team-a", Name: "svc"})
		require.ErrorContains(t, err, "db down")
	})

	t.Run("下游报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.projectSvc.showByNameFn = func(context.Context, *projectpb.ShowByNameRequest) (*projectpb.ShowResponse, error) {
			return showByNameItem("conf", "main", "abc", 4, 11, nil), nil
		}
		e.projectSvc.applyFn = func(context.Context, *projectpb.WebApplyByNameRequest) (*projectpb.WebApplyResponse, error) {
			return nil, errors.New("render failed")
		}

		_, _, err := e.toolset().applyProject(ctxWithUser(adminUser()), nil, applyProjectIn{Namespace: "team-a", Name: "svc"})
		require.ErrorContains(t, err, "render failed")
	})
}

// Test_mergeExtraValues 覆盖「以现状为底、用入参覆盖」的合并语义与其边界：
// 空入参原样返回、覆盖已有 path 不改写入参切片、新增 path 按字典序追加、空串是显式置空。
func Test_mergeExtraValues(t *testing.T) {
	newBase := func() []*websocketpb.ExtraValue {
		return []*websocketpb.ExtraValue{
			{Path: "web->replicas", Value: "3", Description: "副本数"},
			{Path: "web->image", Value: "redis"},
		}
	}

	t.Run("override 为空：原样返回 base（不做无谓拷贝）", func(t *testing.T) {
		base := newBase()
		assert.Equal(t, base, mergeExtraValues(base, nil))
		assert.Equal(t, base, mergeExtraValues(base, map[string]string{}))
	})

	t.Run("覆盖已有 path：值替换、顺序不变、不就地改写 base", func(t *testing.T) {
		base := newBase()
		merged := mergeExtraValues(base, map[string]string{"web->replicas": "5"})

		require.Len(t, merged, 2)
		assert.Equal(t, "web->replicas", merged[0].GetPath())
		assert.Equal(t, "5", merged[0].GetValue())
		assert.Equal(t, "web->image", merged[1].GetPath())
		assert.Equal(t, "redis", merged[1].GetValue())
		// base 是 ShowByName 响应里的切片，就地改写会污染上游响应。
		assert.Equal(t, "3", base[0].GetValue())
		assert.NotSame(t, base[0], merged[0])
		// description 不搬运：后端落库时会按当时的元素定义重新固化。
		assert.Empty(t, merged[0].GetDescription())
	})

	t.Run("新增 path：按字典序追加在末尾（消除 map 遍历的随机序）", func(t *testing.T) {
		merged := mergeExtraValues(newBase(), map[string]string{
			"web->zzz": "1", "web->aaa": "2",
		})
		paths := make([]string, 0, len(merged))
		for _, v := range merged {
			paths = append(paths, v.GetPath())
		}
		assert.Equal(t, []string{"web->replicas", "web->image", "web->aaa", "web->zzz"}, paths)
	})

	t.Run("空串是显式置空（map 键存在即「我要设它」）", func(t *testing.T) {
		merged := mergeExtraValues(newBase(), map[string]string{"web->image": ""})
		assert.Equal(t, "", merged[1].GetValue())
	})

	t.Run("base 为 nil 时也能合并（项目尚未落库任何自定义元素值）", func(t *testing.T) {
		merged := mergeExtraValues(nil, map[string]string{"web->replicas": "2"})
		require.Len(t, merged, 1)
		assert.Equal(t, "web->replicas", merged[0].GetPath())
		assert.Equal(t, "2", merged[0].GetValue())
	})
}

// TestToolset_RestartPod 覆盖 grace=0 强制删除的入参构造与出参原样回吐。
func TestToolset_RestartPod(t *testing.T) {
	t.Run("正常删除：grace_period_seconds 必为 0", func(t *testing.T) {
		e := newTestEnv(t)
		var got *containerpb.ForceDeletePodRequest
		e.contSvc.forceDeleteFn = func(_ context.Context, req *containerpb.ForceDeletePodRequest) (*containerpb.ForceDeletePodResponse, error) {
			got = req
			return &containerpb.ForceDeletePodResponse{Deleted: true, Namespace: "team-a", Pod: "web-0", Message: "ok"}, nil
		}

		res, out, err := e.toolset().restartPod(ctxWithUser(adminUser()), nil, restartPodIn{Namespace: "team-a", Pod: "web-0"})
		require.NoError(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "team-a", got.GetNamespace())
		assert.Equal(t, "web-0", got.GetPod())
		assert.Equal(t, int64(0), got.GetGracePeriodSeconds(), "必须 0 宽限期才等价于立即重启")
		resp := out.(*containerpb.ForceDeletePodResponse)
		assert.True(t, resp.GetDeleted())
		assert.Equal(t, "ok", resp.GetMessage())
	})

	t.Run("下游报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.contSvc.forceDeleteFn = func(context.Context, *containerpb.ForceDeletePodRequest) (*containerpb.ForceDeletePodResponse, error) {
			return nil, errors.New("forbidden")
		}

		_, _, err := e.toolset().restartPod(ctxWithUser(adminUser()), nil, restartPodIn{Namespace: "team-a", Pod: "web-0"})
		require.ErrorContains(t, err, "forbidden")
	})
}

// TestToolset_ExecCommand 覆盖一次性命令的输出收集与错误帧码到结构化出参的映射。
// 用 fakeContainerBiz 逐帧投喂，使每条码路径都被确定性覆盖（真实 biz 分支已在其自身测试覆盖）。
func TestToolset_ExecCommand(t *testing.T) {
	// execEnv 造一个把 ContainerBiz 换成 fake 的 toolset，并预置命名空间门卫。execOnceFn 为
	// fake 注入的「被调时向流里投喂什么帧」的行为。
	execEnv := func(t *testing.T, execOnceFn func(context.Context, biz.ExecOnceStream, *biz.UserInfo, *biz.ExecOnceInput) error) *toolset {
		t.Helper()
		e := newTestEnv(t)
		e.nsRepo.EXPECT().FindByName(gomock.Any(), "team-a").Return(&biz.Namespace{ID: 2, Name: "team-a"}, nil)
		ts := e.toolset()
		ts.containerBiz = &fakeContainerBiz{execOnceFn: execOnceFn}
		return ts
	}

	t.Run("正常输出与入参搬运", func(t *testing.T) {
		userCtx := ctxWithUser(adminUser())
		var got *biz.ExecOnceInput
		ts := execEnv(t, func(_ context.Context, s biz.ExecOnceStream, u *biz.UserInfo, in *biz.ExecOnceInput) error {
			got = in
			assert.Equal(t, "admin", u.Name, "审计需要真实用户身份")
			// 适配器必须原样回吐请求 ctx（biz 用它做取消传播与日志归属）。
			assert.Equal(t, userCtx, s.Context())
			_ = s.Send(&containerpb.ExecResponse{Message: []byte("hi")})
			_ = s.Send(&containerpb.ExecResponse{Message: []byte(" there")})
			return nil
		})

		res, out, err := ts.execCommand(userCtx, nil, execCommandIn{
			Namespace: "team-a", Pod: "web-0", Container: "app", Command: []string{"echo", "hi"}, TimeoutSeconds: 5,
		})
		require.NoError(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "team-a", got.Namespace)
		assert.Equal(t, "web-0", got.Pod)
		assert.Equal(t, "app", got.Container)
		assert.Equal(t, []string{"echo", "hi"}, got.Command)
		assert.Equal(t, int64(5), got.TimeoutSeconds)
		assert.Equal(t, "hi there", out.Output)
		assert.False(t, out.Truncated)
		assert.False(t, out.TimedOut)
		assert.False(t, out.StartFailed)
		assert.Nil(t, out.ExitCode)
	})

	t.Run("真实退出码映射为 exit_code", func(t *testing.T) {
		ts := execEnv(t, func(_ context.Context, s biz.ExecOnceStream, _ *biz.UserInfo, _ *biz.ExecOnceInput) error {
			_ = s.Send(&containerpb.ExecResponse{Message: []byte("boom")})
			_ = s.Send(&containerpb.ExecResponse{Error: &containerpb.ExecError{Code: 1, Message: "exit status 1"}})
			return nil
		})

		_, out, err := ts.execCommand(ctxWithUser(adminUser()), nil, execCommandIn{Namespace: "team-a", Pod: "web-0", Command: []string{"false"}})
		require.NoError(t, err)
		require.NotNil(t, out.ExitCode)
		assert.Equal(t, int32(1), *out.ExitCode)
		assert.False(t, out.Truncated)
		assert.False(t, out.TimedOut)
		assert.False(t, out.StartFailed)
		assert.Empty(t, out.Error, "退出码非 0 属正常执行结果，不应填 error 文案")
	})

	t.Run("截断错误帧 -> truncated + error", func(t *testing.T) {
		ts := execEnv(t, func(_ context.Context, s biz.ExecOnceStream, _ *biz.UserInfo, _ *biz.ExecOnceInput) error {
			_ = s.Send(&containerpb.ExecResponse{Error: &containerpb.ExecError{Code: biz.ExecOnceTruncatedCode, Message: "输出超限"}})
			return nil
		})

		_, out, err := ts.execCommand(ctxWithUser(adminUser()), nil, execCommandIn{Namespace: "team-a", Pod: "web-0", Command: []string{"yes"}})
		require.NoError(t, err)
		assert.True(t, out.Truncated)
		assert.Equal(t, "输出超限", out.Error)
		assert.Nil(t, out.ExitCode)
	})

	t.Run("超时错误帧 -> timed_out + error", func(t *testing.T) {
		ts := execEnv(t, func(_ context.Context, s biz.ExecOnceStream, _ *biz.UserInfo, _ *biz.ExecOnceInput) error {
			_ = s.Send(&containerpb.ExecResponse{Error: &containerpb.ExecError{Code: biz.ExecOnceTimeoutCode, Message: "执行超时"}})
			return nil
		})

		_, out, err := ts.execCommand(ctxWithUser(adminUser()), nil, execCommandIn{Namespace: "team-a", Pod: "web-0", Command: []string{"sleep", "999"}})
		require.NoError(t, err)
		assert.True(t, out.TimedOut)
		assert.Equal(t, "执行超时", out.Error)
	})

	t.Run("启动失败错误帧 -> start_failed + error", func(t *testing.T) {
		ts := execEnv(t, func(_ context.Context, s biz.ExecOnceStream, _ *biz.UserInfo, _ *biz.ExecOnceInput) error {
			_ = s.Send(&containerpb.ExecResponse{Error: &containerpb.ExecError{Code: biz.ExecOnceExecFailedCode, Message: "exec 启动失败"}})
			return nil
		})

		_, out, err := ts.execCommand(ctxWithUser(adminUser()), nil, execCommandIn{Namespace: "team-a", Pod: "web-0", Command: []string{"no-such-bin"}})
		require.NoError(t, err)
		assert.True(t, out.StartFailed)
		assert.Equal(t, "exec 启动失败", out.Error)
	})

	t.Run("未知错误码 -> 不落任何标志位", func(t *testing.T) {
		ts := execEnv(t, func(_ context.Context, s biz.ExecOnceStream, _ *biz.UserInfo, _ *biz.ExecOnceInput) error {
			_ = s.Send(&containerpb.ExecResponse{Error: &containerpb.ExecError{Code: 9999, Message: "未知"}})
			return nil
		})

		_, out, err := ts.execCommand(ctxWithUser(adminUser()), nil, execCommandIn{Namespace: "team-a", Pod: "web-0", Command: []string{"x"}})
		require.NoError(t, err)
		assert.False(t, out.Truncated)
		assert.False(t, out.TimedOut)
		assert.False(t, out.StartFailed)
		assert.Nil(t, out.ExitCode)
	})

	t.Run("超大输出被二次截断到 maxLogBytes", func(t *testing.T) {
		ts := execEnv(t, func(_ context.Context, s biz.ExecOnceStream, _ *biz.UserInfo, _ *biz.ExecOnceInput) error {
			_ = s.Send(&containerpb.ExecResponse{Message: []byte(strings.Repeat("a", maxLogBytes+1024))})
			return nil
		})

		_, out, err := ts.execCommand(ctxWithUser(adminUser()), nil, execCommandIn{Namespace: "team-a", Pod: "web-0", Command: []string{"cat", "/big"}})
		require.NoError(t, err)
		assert.True(t, out.Truncated)
		assert.Len(t, out.Output, maxLogBytes)
	})

	t.Run("门卫拒绝时不触达下游", func(t *testing.T) {
		e := newTestEnv(t)
		e.nsRepo.EXPECT().FindByName(gomock.Any(), "private-ns").Return(
			&biz.Namespace{ID: 9, Name: "private-ns", Private: true, CreatorEmail: "other@mars.com"}, nil)
		// Pod 级门卫在命名空间拒绝后会回查 agent 命名空间；NotFound 表示非 agent 空间。
		e.nsRepo.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).Return(nil, errs.WrapNotFound(errors.New("no ns"), "find agent ns"))
		ts := e.toolset()
		// 未设 execOnceFn：一旦门卫漏过，fake 解引用 nil 会 panic，测试即失败。
		ts.containerBiz = &fakeContainerBiz{}

		_, _, err := ts.execCommand(ctxWithUser(normalUser()), nil, execCommandIn{Namespace: "private-ns", Pod: "p", Command: []string{"id"}})
		require.Error(t, err)
	})

	t.Run("下游执行报错原样上抛", func(t *testing.T) {
		ts := execEnv(t, func(context.Context, biz.ExecOnceStream, *biz.UserInfo, *biz.ExecOnceInput) error {
			return errors.New("spdy 建连失败")
		})

		_, _, err := ts.execCommand(ctxWithUser(adminUser()), nil, execCommandIn{Namespace: "team-a", Pod: "web-0", Command: []string{"id"}})
		require.ErrorContains(t, err, "spdy 建连失败")
	})
}

// TestServer_Register 断言 /mcp 路由确实挂到 mux 上（否则会被前端 SPA 兜底吞掉）。
func TestServer_Register(t *testing.T) {
	e := newTestEnv(t)
	r := mux.NewRouter()
	e.server.Register(r)

	route := r.GetRoute("mcp")
	require.NotNil(t, route)
	tmpl, err := route.GetPathTemplate()
	require.NoError(t, err)
	assert.Equal(t, Path, tmpl)
}
