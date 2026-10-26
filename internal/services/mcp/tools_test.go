package mcp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	changelogpb "github.com/duc-cnzj/mars/api/v6/proto/changelog"
	containerpb "github.com/duc-cnzj/mars/api/v6/proto/container"
	eventpb "github.com/duc-cnzj/mars/api/v6/proto/event"
	gitpb "github.com/duc-cnzj/mars/api/v6/proto/git"
	hibernatepb "github.com/duc-cnzj/mars/api/v6/proto/hibernate"
	metricspb "github.com/duc-cnzj/mars/api/v6/proto/metrics"
	namespacepb "github.com/duc-cnzj/mars/api/v6/proto/namespace"
	projectpb "github.com/duc-cnzj/mars/api/v6/proto/project"
	repopb "github.com/duc-cnzj/mars/api/v6/proto/repo"
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

		res, out, err := e.toolset().listNamespaces(ctxWithUser(adminUser()), nil,
			listNamespacesIn{paginationIn: paginationIn{Page: 2, PageSize: 5}, Name: "team"})
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

// repoShowWithGitProject 造一个已绑定 GitLab 项目的仓库模型。
func repoShowWithGitProject(repoID, gitProjectID int32) *repopb.ShowResponse {
	return &repopb.ShowResponse{Item: &apitypes.RepoModel{Id: repoID, Name: "svc-repo", GitProjectId: gitProjectID}}
}

// commitOptions 造 n 条提交选项（value 为递增的假 SHA）。
func commitOptions(n int) *gitpb.CommitOptionsResponse {
	items := make([]*gitpb.Option, 0, n)
	for i := 0; i < n; i++ {
		items = append(items, &gitpb.Option{
			Value:  fmt.Sprintf("sha-%03d", i),
			Branch: "main",
			Type:   "commit",
			IsLeaf: true,
		})
	}
	return &gitpb.CommitOptionsResponse{Items: items}
}

// TestToolset_ListGitBranches 覆盖分支查询的定位链（项目 → 仓库 → GitLab 项目）、
// 白名单过滤开关与配置缺失/下游报错分支。
func TestToolset_ListGitBranches(t *testing.T) {
	t.Run("默认按仓库白名单过滤", func(t *testing.T) {
		e := newTestEnv(t)
		e.expectProjectAccess(7, 3)
		repoID, gitProjectID := int32(3), int32(9)
		e.repoSvc.showFn = func(_ context.Context, req *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			assert.Equal(t, repoID, req.GetId())
			return repoShowWithGitProject(repoID, gitProjectID), nil
		}
		var got *gitpb.BranchOptionsRequest
		e.gitSvc.branchOptionsFn = func(_ context.Context, req *gitpb.BranchOptionsRequest) (*gitpb.BranchOptionsResponse, error) {
			got = req
			return &gitpb.BranchOptionsResponse{Items: []*gitpb.Option{{Value: "main", Branch: "main", Type: "branch"}}}, nil
		}

		_, out, err := e.toolset().listGitBranches(ctxWithUser(adminUser()), nil, listGitBranchesIn{ProjectID: 7})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, gitProjectID, got.GetGitProjectId())
		assert.Equal(t, repoID, got.GetRepoId(), "默认必须下发 repo_id 让服务按白名单过滤")
		items := out.(*gitpb.BranchOptionsResponse).GetItems()
		require.Len(t, items, 1)
		assert.Equal(t, "main", items[0].GetValue())
	})

	t.Run("include_all 时不下发 repo_id", func(t *testing.T) {
		e := newTestEnv(t)
		e.expectProjectAccess(7, 3)
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return repoShowWithGitProject(3, 9), nil
		}
		e.gitSvc.branchOptionsFn = func(_ context.Context, req *gitpb.BranchOptionsRequest) (*gitpb.BranchOptionsResponse, error) {
			assert.Zero(t, req.GetRepoId(), "include_all 时不得按仓库白名单过滤")
			return &gitpb.BranchOptionsResponse{}, nil
		}

		_, _, err := e.toolset().listGitBranches(ctxWithUser(adminUser()), nil, listGitBranchesIn{ProjectID: 7, IncludeAll: true})
		require.NoError(t, err)
	})

	t.Run("项目未绑定仓库返回 InvalidArgument", func(t *testing.T) {
		e := newTestEnv(t)
		e.expectProjectAccess(7, 0)

		_, _, err := e.toolset().listGitBranches(ctxWithUser(adminUser()), nil, listGitBranchesIn{ProjectID: 7})
		require.ErrorContains(t, err, "未绑定仓库")
	})

	t.Run("仓库未配置 GitLab 项目返回 InvalidArgument", func(t *testing.T) {
		e := newTestEnv(t)
		e.expectProjectAccess(7, 3)
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return repoShowWithGitProject(3, 0), nil
		}

		_, _, err := e.toolset().listGitBranches(ctxWithUser(adminUser()), nil, listGitBranchesIn{ProjectID: 7})
		require.ErrorContains(t, err, "未配置 GitLab 项目")
	})

	t.Run("仓库查询报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.expectProjectAccess(7, 3)
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return nil, errors.New("repo query failed")
		}

		_, _, err := e.toolset().listGitBranches(ctxWithUser(adminUser()), nil, listGitBranchesIn{ProjectID: 7})
		require.ErrorContains(t, err, "repo query failed")
	})

	t.Run("下游报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.expectProjectAccess(7, 3)
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return repoShowWithGitProject(3, 9), nil
		}
		e.gitSvc.branchOptionsFn = func(context.Context, *gitpb.BranchOptionsRequest) (*gitpb.BranchOptionsResponse, error) {
			return nil, errors.New("gitlab 不可达")
		}

		_, _, err := e.toolset().listGitBranches(ctxWithUser(adminUser()), nil, listGitBranchesIn{ProjectID: 7})
		require.ErrorContains(t, err, "gitlab 不可达")
	})

	t.Run("门卫拒绝时不触达下游", func(t *testing.T) {
		e := newTestEnv(t)
		e.nsRepo.EXPECT().Show(gomock.Any(), 2).Return(
			&biz.Namespace{ID: 2, Name: "private-ns", Private: true, CreatorEmail: "other@mars.com"}, nil)
		e.projRepo.EXPECT().Show(gomock.Any(), 7).Return(
			&biz.Project{ID: 7, RepoID: 3, NamespaceID: 2}, nil)
		// 命名空间拒绝后门卫回查 agent 命名空间；NotFound 表示非 agent 空间。
		e.nsRepo.EXPECT().FindByName(gomock.Any(), biz.AgentNamespace).Return(nil, errs.WrapNotFound(errors.New("no ns"), "find agent ns"))
		// 非当前用户的 agent 项目：门卫回 PermissionDenied，下游 fake 未设函数即证明未触达。
		_, _, err := e.toolset().listGitBranches(ctxWithUser(normalUser()), nil, listGitBranchesIn{ProjectID: 7})
		require.ErrorContains(t, err, "没有权限执行该操作")
	})
}

// TestToolset_ListGitCommits 覆盖提交查询的入参搬运、limit 缺省/上限截断与参数校验。
func TestToolset_ListGitCommits(t *testing.T) {
	t.Run("入参搬运与 limit 截断", func(t *testing.T) {
		e := newTestEnv(t)
		e.expectProjectAccess(7, 3)
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return repoShowWithGitProject(3, 9), nil
		}
		var got *gitpb.CommitOptionsRequest
		e.gitSvc.commitOptionsFn = func(_ context.Context, req *gitpb.CommitOptionsRequest) (*gitpb.CommitOptionsResponse, error) {
			got = req
			return commitOptions(25), nil
		}

		_, out, err := e.toolset().listGitCommits(ctxWithUser(adminUser()), nil,
			listGitCommitsIn{ProjectID: 7, Branch: "main", Limit: 3})
		require.NoError(t, err)
		assert.Equal(t, int32(9), got.GetGitProjectId())
		assert.Equal(t, "main", got.GetBranch())
		items := out.(*gitpb.CommitOptionsResponse).GetItems()
		require.Len(t, items, 3)
		assert.Equal(t, "sha-000", items[0].GetValue(), "value 必须是完整 commit SHA")
	})

	t.Run("limit 缺省 20", func(t *testing.T) {
		e := newTestEnv(t)
		e.expectProjectAccess(7, 3)
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return repoShowWithGitProject(3, 9), nil
		}
		e.gitSvc.commitOptionsFn = func(context.Context, *gitpb.CommitOptionsRequest) (*gitpb.CommitOptionsResponse, error) {
			return commitOptions(25), nil
		}

		_, out, err := e.toolset().listGitCommits(ctxWithUser(adminUser()), nil,
			listGitCommitsIn{ProjectID: 7, Branch: "main"})
		require.NoError(t, err)
		assert.Len(t, out.(*gitpb.CommitOptionsResponse).GetItems(), defaultCommitLimit)
	})

	t.Run("limit 超上限截到 100", func(t *testing.T) {
		e := newTestEnv(t)
		e.expectProjectAccess(7, 3)
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return repoShowWithGitProject(3, 9), nil
		}
		e.gitSvc.commitOptionsFn = func(context.Context, *gitpb.CommitOptionsRequest) (*gitpb.CommitOptionsResponse, error) {
			return commitOptions(120), nil
		}

		_, out, err := e.toolset().listGitCommits(ctxWithUser(adminUser()), nil,
			listGitCommitsIn{ProjectID: 7, Branch: "main", Limit: 500})
		require.NoError(t, err)
		assert.Len(t, out.(*gitpb.CommitOptionsResponse).GetItems(), maxCommitLimit)
	})

	t.Run("branch 为空直接拒绝且不触达门卫", func(t *testing.T) {
		e := newTestEnv(t)
		// 未设任何门卫/下游 mock：一旦继续执行即因缺期望而失败。
		_, _, err := e.toolset().listGitCommits(ctxWithUser(adminUser()), nil, listGitCommitsIn{ProjectID: 7})
		require.ErrorContains(t, err, "branch 不能为空")
	})

	t.Run("项目未绑定仓库返回 InvalidArgument", func(t *testing.T) {
		e := newTestEnv(t)
		e.expectProjectAccess(7, 0)

		_, _, err := e.toolset().listGitCommits(ctxWithUser(adminUser()), nil,
			listGitCommitsIn{ProjectID: 7, Branch: "main"})
		require.ErrorContains(t, err, "未绑定仓库")
	})

	t.Run("下游报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.expectProjectAccess(7, 3)
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return repoShowWithGitProject(3, 9), nil
		}
		e.gitSvc.commitOptionsFn = func(context.Context, *gitpb.CommitOptionsRequest) (*gitpb.CommitOptionsResponse, error) {
			return nil, errors.New("branch not found")
		}

		_, _, err := e.toolset().listGitCommits(ctxWithUser(adminUser()), nil,
			listGitCommitsIn{ProjectID: 7, Branch: "nope"})
		require.ErrorContains(t, err, "branch not found")
	})
}

// TestToolset_CheckApplyStatus 覆盖部署判定工具的入参搬运与失败诊断出参。
func TestToolset_CheckApplyStatus(t *testing.T) {
	t.Run("入参搬运与失败诊断回吐", func(t *testing.T) {
		e := newTestEnv(t)
		var got *projectpb.CheckApplyStatusRequest
		e.projectSvc.checkApplyFn = func(_ context.Context, req *projectpb.CheckApplyStatusRequest) (*projectpb.CheckApplyStatusResponse, error) {
			got = req
			return &projectpb.CheckApplyStatusResponse{
				Status: 3, Reason: "容器异常",
				Failures: []*projectpb.ContainerFailure{{Reason: "CrashLoopBackOff", Logs: "boom"}},
			}, nil
		}

		_, out, err := e.toolset().checkApplyStatus(ctxWithUser(adminUser()), nil, checkApplyStatusIn{ProjectID: 7})
		require.NoError(t, err)
		assert.Equal(t, int32(7), got.GetId())
		resp := out.(*projectpb.CheckApplyStatusResponse)
		assert.Equal(t, int32(3), int32(resp.GetStatus()))
		assert.Equal(t, "容器异常", resp.GetReason())
		require.Len(t, resp.GetFailures(), 1)
		assert.Equal(t, "boom", resp.GetFailures()[0].GetLogs())
	})

	t.Run("下游报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.projectSvc.checkApplyFn = func(context.Context, *projectpb.CheckApplyStatusRequest) (*projectpb.CheckApplyStatusResponse, error) {
			return nil, errors.New("k8s unavailable")
		}

		_, _, err := e.toolset().checkApplyStatus(ctxWithUser(adminUser()), nil, checkApplyStatusIn{ProjectID: 7})
		require.ErrorContains(t, err, "k8s unavailable")
	})
}

// TestToolset_GetProjectEndpoints 覆盖入口地址与资源用量汇总的出参回吐。
func TestToolset_GetProjectEndpoints(t *testing.T) {
	t.Run("入参搬运与端点回吐", func(t *testing.T) {
		e := newTestEnv(t)
		var got *projectpb.MemoryCpuAndEndpointsRequest
		e.projectSvc.endpointsFn = func(_ context.Context, req *projectpb.MemoryCpuAndEndpointsRequest) (*projectpb.MemoryCpuAndEndpointsResponse, error) {
			got = req
			return &projectpb.MemoryCpuAndEndpointsResponse{
				Urls: []*apitypes.ServiceEndpoint{{Name: "web", Url: "https://svc.example.com", Type: "Ingress"}},
				Cpu:  "300m", Memory: "300Mi",
			}, nil
		}

		_, out, err := e.toolset().getProjectEndpoints(ctxWithUser(adminUser()), nil, getProjectEndpointsIn{ProjectID: 7})
		require.NoError(t, err)
		assert.Equal(t, int32(7), got.GetId())
		resp := out.(*projectpb.MemoryCpuAndEndpointsResponse)
		require.Len(t, resp.GetUrls(), 1)
		assert.Equal(t, "https://svc.example.com", resp.GetUrls()[0].GetUrl())
		assert.Equal(t, "300m", resp.GetCpu())
		assert.Equal(t, "300Mi", resp.GetMemory())
	})

	t.Run("下游报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.projectSvc.endpointsFn = func(context.Context, *projectpb.MemoryCpuAndEndpointsRequest) (*projectpb.MemoryCpuAndEndpointsResponse, error) {
			return nil, errors.New("ingress query failed")
		}

		_, _, err := e.toolset().getProjectEndpoints(ctxWithUser(adminUser()), nil, getProjectEndpointsIn{ProjectID: 7})
		require.ErrorContains(t, err, "ingress query failed")
	})
}

// TestToolset_GetProjectResourceTree 覆盖拓扑树的入参搬运与节点/边回吐。
func TestToolset_GetProjectResourceTree(t *testing.T) {
	t.Run("入参搬运与拓扑回吐", func(t *testing.T) {
		e := newTestEnv(t)
		var got *projectpb.ResourceTreeRequest
		e.projectSvc.resourceTreeFn = func(_ context.Context, req *projectpb.ResourceTreeRequest) (*projectpb.ResourceTreeResponse, error) {
			got = req
			return &projectpb.ResourceTreeResponse{
				Status: 2,
				Nodes:  []*projectpb.ResourceTreeNode{{Id: "pod-web-1", Kind: "Pod", Name: "web-1", Status: "healthy"}},
				Edges:  []*projectpb.ResourceTreeEdge{{Id: "e1", Type: "owner", Source: "deployment-web", Target: "pod-web-1"}},
			}, nil
		}

		_, out, err := e.toolset().getProjectResourceTree(ctxWithUser(adminUser()), nil, getProjectResourceTreeIn{ProjectID: 7})
		require.NoError(t, err)
		assert.Equal(t, int32(7), got.GetId())
		resp := out.(*projectpb.ResourceTreeResponse)
		require.Len(t, resp.GetNodes(), 1)
		assert.Equal(t, "Pod", resp.GetNodes()[0].GetKind())
		require.Len(t, resp.GetEdges(), 1)
		assert.Equal(t, "owner", resp.GetEdges()[0].GetType())
	})

	t.Run("下游报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.projectSvc.resourceTreeFn = func(context.Context, *projectpb.ResourceTreeRequest) (*projectpb.ResourceTreeResponse, error) {
			return nil, errors.New("list resources failed")
		}

		_, _, err := e.toolset().getProjectResourceTree(ctxWithUser(adminUser()), nil, getProjectResourceTreeIn{ProjectID: 7})
		require.ErrorContains(t, err, "list resources failed")
	})
}

// TestToolset_GetCommitDetail 覆盖提交详情入参校验、项目→仓库→GitLab 定位链与下游报错。
func TestToolset_GetCommitDetail(t *testing.T) {
	t.Run("入参搬运与详情回吐", func(t *testing.T) {
		e := newTestEnv(t)
		e.expectProjectAccess(7, 3)
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return repoShowWithGitProject(3, 9), nil
		}
		var got *gitpb.CommitRequest
		e.gitSvc.commitFn = func(_ context.Context, req *gitpb.CommitRequest) (*gitpb.CommitResponse, error) {
			got = req
			return &gitpb.CommitResponse{
				Id: "abc123def456", Title: "fix crash", AuthorName: "duc",
				AuthorEmail: "duc@mars.com", WebUrl: "https://gitlab/commit/abc123def456",
				Message: "fix crash\n\nbody",
			}, nil
		}

		_, out, err := e.toolset().getCommitDetail(ctxWithUser(adminUser()), nil,
			getCommitDetailIn{ProjectID: 7, Branch: "main", Commit: "abc123def456"})
		require.NoError(t, err)
		assert.Equal(t, int32(9), got.GetGitProjectId())
		assert.Equal(t, "main", got.GetBranch())
		assert.Equal(t, "abc123def456", got.GetCommit())
		resp := out.(*gitpb.CommitResponse)
		assert.Equal(t, "duc@mars.com", resp.GetAuthorEmail())
		assert.Equal(t, "https://gitlab/commit/abc123def456", resp.GetWebUrl())
	})

	t.Run("branch 为空直接拒绝且不触达门卫", func(t *testing.T) {
		e := newTestEnv(t)
		_, _, err := e.toolset().getCommitDetail(ctxWithUser(adminUser()), nil,
			getCommitDetailIn{ProjectID: 7, Commit: "abc"})
		require.ErrorContains(t, err, "branch 不能为空")
	})

	t.Run("commit 为空直接拒绝且不触达门卫", func(t *testing.T) {
		e := newTestEnv(t)
		_, _, err := e.toolset().getCommitDetail(ctxWithUser(adminUser()), nil,
			getCommitDetailIn{ProjectID: 7, Branch: "main"})
		require.ErrorContains(t, err, "commit 不能为空")
	})

	t.Run("项目未绑定仓库返回 InvalidArgument", func(t *testing.T) {
		e := newTestEnv(t)
		e.expectProjectAccess(7, 0)

		_, _, err := e.toolset().getCommitDetail(ctxWithUser(adminUser()), nil,
			getCommitDetailIn{ProjectID: 7, Branch: "main", Commit: "abc"})
		require.ErrorContains(t, err, "未绑定仓库")
	})

	t.Run("下游报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.expectProjectAccess(7, 3)
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return repoShowWithGitProject(3, 9), nil
		}
		e.gitSvc.commitFn = func(context.Context, *gitpb.CommitRequest) (*gitpb.CommitResponse, error) {
			return nil, errors.New("commit not found")
		}

		_, _, err := e.toolset().getCommitDetail(ctxWithUser(adminUser()), nil,
			getCommitDetailIn{ProjectID: 7, Branch: "main", Commit: "nope"})
		require.ErrorContains(t, err, "commit not found")
	})
}

// TestToolset_GetPipelineStatus 覆盖 pipeline 查询的 repo_id 推导、入参校验与下游报错。
func TestToolset_GetPipelineStatus(t *testing.T) {
	t.Run("按项目绑定仓库查询并回吐 job", func(t *testing.T) {
		e := newTestEnv(t)
		e.expectProjectAccess(7, 3)
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return repoShowWithGitProject(3, 9), nil
		}
		var got *gitpb.PipelineInfoByRepoIdRequest
		e.gitSvc.pipelineByRepoFn = func(_ context.Context, req *gitpb.PipelineInfoByRepoIdRequest) (*gitpb.PipelineInfoResponse, error) {
			got = req
			return &gitpb.PipelineInfoResponse{
				Status: "success", WebUrl: "https://gitlab/pipeline/1",
				Jobs: []*gitpb.PipelineJob{{Name: "build", Status: "success", StageName: "build"}},
			}, nil
		}

		_, out, err := e.toolset().getPipelineStatus(ctxWithUser(adminUser()), nil,
			getPipelineStatusIn{ProjectID: 7, Branch: "main", Commit: "abc123def456"})
		require.NoError(t, err)
		assert.Equal(t, int32(3), got.GetRepoId(), "repo_id 必须来自项目绑定，不能由调用方直传")
		assert.Equal(t, "abc123def456", got.GetCommit())
		resp := out.(*gitpb.PipelineInfoResponse)
		assert.Equal(t, "success", resp.GetStatus())
		require.Len(t, resp.GetJobs(), 1)
		assert.Equal(t, "build", resp.GetJobs()[0].GetName())
	})

	t.Run("branch 为空直接拒绝且不触达门卫", func(t *testing.T) {
		e := newTestEnv(t)
		_, _, err := e.toolset().getPipelineStatus(ctxWithUser(adminUser()), nil,
			getPipelineStatusIn{ProjectID: 7, Commit: "abc"})
		require.ErrorContains(t, err, "branch 不能为空")
	})

	t.Run("commit 为空直接拒绝且不触达门卫", func(t *testing.T) {
		e := newTestEnv(t)
		_, _, err := e.toolset().getPipelineStatus(ctxWithUser(adminUser()), nil,
			getPipelineStatusIn{ProjectID: 7, Branch: "main"})
		require.ErrorContains(t, err, "commit 不能为空")
	})

	t.Run("项目未绑定仓库返回 InvalidArgument", func(t *testing.T) {
		e := newTestEnv(t)
		e.expectProjectAccess(7, 0)

		_, _, err := e.toolset().getPipelineStatus(ctxWithUser(adminUser()), nil,
			getPipelineStatusIn{ProjectID: 7, Branch: "main", Commit: "abc"})
		require.ErrorContains(t, err, "未绑定仓库")
	})

	t.Run("下游报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.expectProjectAccess(7, 3)
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return repoShowWithGitProject(3, 9), nil
		}
		e.gitSvc.pipelineByRepoFn = func(context.Context, *gitpb.PipelineInfoByRepoIdRequest) (*gitpb.PipelineInfoResponse, error) {
			return nil, errors.New("pipeline api failed")
		}

		_, _, err := e.toolset().getPipelineStatus(ctxWithUser(adminUser()), nil,
			getPipelineStatusIn{ProjectID: 7, Branch: "main", Commit: "abc"})
		require.ErrorContains(t, err, "pipeline api failed")
	})
}

// TestToolset_ListOperationEvents 覆盖审计日志的分页/搜索/all 入参搬运与下游报错。
func TestToolset_ListOperationEvents(t *testing.T) {
	t.Run("分页与过滤条件搬进请求", func(t *testing.T) {
		e := newTestEnv(t)
		var got *eventpb.ListRequest
		e.eventSvc.listFn = func(_ context.Context, req *eventpb.ListRequest) (*eventpb.ListResponse, error) {
			got = req
			return &eventpb.ListResponse{
				Page: 2, PageSize: 5,
				Items: []*apitypes.EventModel{{Id: 11, Message: "更新项目 7/svc", Username: "admin"}},
			}, nil
		}

		_, out, err := e.toolset().listOperationEvents(ctxWithUser(adminUser()), nil,
			listOperationEventsIn{paginationIn: paginationIn{Page: 2, PageSize: 5}, Search: "svc", All: true})
		require.NoError(t, err)
		require.NotNil(t, got.GetPage())
		assert.Equal(t, int32(2), got.GetPage())
		assert.Equal(t, int32(5), got.GetPageSize())
		assert.Equal(t, "svc", got.GetSearch())
		assert.True(t, got.GetAll())
		resp := out.(*eventpb.ListResponse)
		require.Len(t, resp.GetItems(), 1)
		assert.Equal(t, "更新项目 7/svc", resp.GetItems()[0].GetMessage())
	})

	t.Run("下游报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.eventSvc.listFn = func(context.Context, *eventpb.ListRequest) (*eventpb.ListResponse, error) {
			return nil, errors.New("db down")
		}

		_, _, err := e.toolset().listOperationEvents(ctxWithUser(adminUser()), nil, listOperationEventsIn{})
		require.ErrorContains(t, err, "db down")
	})
}

// TestToolset_ListHibernateWindows 覆盖全局休眠时间窗列表回吐与下游报错。
func TestToolset_ListHibernateWindows(t *testing.T) {
	t.Run("回吐时间窗定义", func(t *testing.T) {
		e := newTestEnv(t)
		e.hibSvc.listWindowsFn = func(_ context.Context, _ *hibernatepb.ListWindowsRequest) (*hibernatepb.ListWindowsResponse, error) {
			return &hibernatepb.ListWindowsResponse{Items: []*hibernatepb.HibernateWindow{
				{Id: 1, Name: "night", Start: "22:00", End: "08:00", RepoNames: []string{"svc-repo"}},
			}}, nil
		}

		_, out, err := e.toolset().listHibernateWindows(ctxWithUser(adminUser()), nil, listHibernateWindowsIn{})
		require.NoError(t, err)
		resp := out.(*hibernatepb.ListWindowsResponse)
		require.Len(t, resp.GetItems(), 1)
		assert.Equal(t, "night", resp.GetItems()[0].GetName())
		assert.Equal(t, "22:00", resp.GetItems()[0].GetStart())
		assert.Equal(t, "08:00", resp.GetItems()[0].GetEnd())
	})

	t.Run("下游报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.hibSvc.listWindowsFn = func(context.Context, *hibernatepb.ListWindowsRequest) (*hibernatepb.ListWindowsResponse, error) {
			return nil, errors.New("window query failed")
		}

		_, _, err := e.toolset().listHibernateWindows(ctxWithUser(adminUser()), nil, listHibernateWindowsIn{})
		require.ErrorContains(t, err, "window query failed")
	})
}

// TestToolset_ToggleHibernate 覆盖休眠开关的入参搬运与下游报错。
func TestToolset_ToggleHibernate(t *testing.T) {
	t.Run("开关与时间窗搬运", func(t *testing.T) {
		e := newTestEnv(t)
		var got *projectpb.ToggleHibernateRequest
		e.projectSvc.toggleHibernateFn = func(_ context.Context, req *projectpb.ToggleHibernateRequest) (*projectpb.ToggleHibernateResponse, error) {
			got = req
			return &projectpb.ToggleHibernateResponse{Enabled: true, State: "awake", WindowNames: []string{"night"}}, nil
		}

		_, out, err := e.toolset().toggleHibernate(ctxWithUser(adminUser()), nil,
			toggleHibernateIn{ProjectID: 7, Enabled: true, WindowNames: []string{"night"}})
		require.NoError(t, err)
		assert.Equal(t, int32(7), got.GetProjectId())
		assert.True(t, got.GetEnabled())
		assert.Equal(t, []string{"night"}, got.GetWindowNames())
		resp := out.(*projectpb.ToggleHibernateResponse)
		assert.True(t, resp.GetEnabled())
		assert.Equal(t, "awake", resp.GetState())
	})

	t.Run("下游报错原样上抛", func(t *testing.T) {
		e := newTestEnv(t)
		e.projectSvc.toggleHibernateFn = func(context.Context, *projectpb.ToggleHibernateRequest) (*projectpb.ToggleHibernateResponse, error) {
			return nil, errors.New("window not allowed")
		}

		_, _, err := e.toolset().toggleHibernate(ctxWithUser(adminUser()), nil,
			toggleHibernateIn{ProjectID: 7, Enabled: true, WindowNames: []string{"night"}})
		require.ErrorContains(t, err, "window not allowed")
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
