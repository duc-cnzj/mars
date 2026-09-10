package services

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/cluster"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewClusterSvc(t *testing.T) {
	svc, _ := newClusterSvcWithMocks(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.k8sBiz)
	assert.NotNil(t, svc.logger)
}

func Test_clusterSvc_ClusterInfo(t *testing.T) {
	svc, mocks := newClusterSvcWithMocks(t)

	mocks.k8sBiz.EXPECT().ClusterInfo().Return(&biz.ClusterInfo{
		Status:            "ok",
		FreeMemory:        "100Mi",
		FreeCpu:           "2",
		FreeRequestMemory: "50Mi",
		FreeRequestCpu:    "1",
		TotalMemory:       "200Mi",
		TotalCpu:          "4",
		UsageMemoryRate:   "0.5",
		UsageCpuRate:      "0.25",
		RequestMemoryRate: "0.3",
		RequestCpuRate:    "0.4",
	})

	resp, err := svc.ClusterInfo(context.TODO(), nil)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	if assert.NotNil(t, resp.Item) {
		assert.Equal(t, "ok", resp.Item.Status)
		assert.Equal(t, "100Mi", resp.Item.FreeMemory)
		assert.Equal(t, "2", resp.Item.FreeCpu)
		assert.Equal(t, "50Mi", resp.Item.FreeRequestMemory)
		assert.Equal(t, "1", resp.Item.FreeRequestCpu)
		assert.Equal(t, "200Mi", resp.Item.TotalMemory)
		assert.Equal(t, "4", resp.Item.TotalCpu)
		assert.Equal(t, "0.5", resp.Item.UsageMemoryRate)
		assert.Equal(t, "0.25", resp.Item.UsageCpuRate)
		assert.Equal(t, "0.3", resp.Item.RequestMemoryRate)
		assert.Equal(t, "0.4", resp.Item.RequestCpuRate)
	}
}

type clusterSvcMocks struct {
	ctrl        *gomock.Controller
	k8sBiz      *biz.MockK8sBiz
	nsBiz       *biz.MockNamespaceBiz
	projectBiz  *biz.MockProjectBiz
	projectRepo *data.MockProjectRepo
	nsRepo      *data.MockNamespaceRepo
	clRepo      *data.MockChangelogRepo
}

func newClusterSvcWithMocks(t *testing.T) (*clusterSvc, *clusterSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &clusterSvcMocks{
		ctrl:        ctrl,
		k8sBiz:      biz.NewMockK8sBiz(ctrl),
		nsBiz:       biz.NewMockNamespaceBiz(ctrl),
		projectBiz:  biz.NewMockProjectBiz(ctrl),
		projectRepo: data.NewMockProjectRepo(ctrl),
		nsRepo:      data.NewMockNamespaceRepo(ctrl),
		clRepo:      data.NewMockChangelogRepo(ctrl),
	}
	logger := mlog.NewForConfig(nil)
	s, ok := NewClusterSvc(ClusterSvcDeps{
		K8sBiz:       mocks.k8sBiz,
		NamespaceBiz: mocks.nsBiz,
		ProjectBiz:   mocks.projectBiz,
		AccessBiz:    biz.NewAccessBiz(biz.NewNamespaceBiz(logger, mocks.nsRepo, nil, nil, nil), biz.NewProjectBiz(logger, mocks.projectRepo, nil, nil)),
		ChangelogBiz: biz.NewChangelogBiz(mocks.clRepo),
		Logger:       logger,
	}).(*clusterSvc)
	if !ok {
		panic("NewClusterSvc returned unexpected type")
	}
	return s, mocks
}

// Test_clusterSvc_ClusterBoard 成功路径：biz 看板聚合成 BoardResponse 全字段落位。
func Test_clusterSvc_ClusterBoard(t *testing.T) {
	svc, mocks := newClusterSvcWithMocks(t)

	mocks.nsBiz.EXPECT().ListAllNames(gomock.Any()).Return([]string{"ns-a"}, nil)
	mocks.k8sBiz.EXPECT().ClusterBoard(gomock.Any(), gomock.Any(), gomock.Any()).Return(&biz.ClusterBoard{
		Overview: &biz.ClusterInfo{Status: "health"},
		Nodes: []*biz.ClusterBoardNode{{
			Name: "node01", Role: "master", Status: "Ready",
			CpuCapacity: 4000, CpuUsage: 1000, CpuRequest: 500,
			MemCapacity: 8589934592, MemUsage: 1073741824, MemRequest: 268435456,
		}},
		Namespaces: []*biz.ClusterBoardNamespace{{
			Name: "ns-a", CpuMilli: 500, MemoryBytes: 268435456, PodCount: 2,
		}},
		Pods: []*biz.ClusterBoardPod{{
			Namespace: "ns-a", Pod: "p1", CpuMilli: 500, MemoryBytes: 268435456,
		}},
	}, nil)

	resp, err := svc.ClusterBoard(newAdminUserCtx(), &cluster.BoardRequest{})
	assert.NoError(t, err)
	if assert.NotNil(t, resp.Overview) {
		assert.Equal(t, "health", resp.Overview.Status)
	}
	if assert.Len(t, resp.Nodes, 1) {
		assert.Equal(t, "node01", resp.Nodes[0].Name)
		assert.Equal(t, "master", resp.Nodes[0].Role)
		assert.Equal(t, int64(4000), resp.Nodes[0].CpuCapacity)
		assert.Equal(t, int64(8589934592), resp.Nodes[0].MemCapacity)
	}
	if assert.Len(t, resp.Namespaces, 1) {
		assert.Equal(t, "ns-a", resp.Namespaces[0].Name)
		assert.Equal(t, int32(2), resp.Namespaces[0].PodCount)
	}
	if assert.Len(t, resp.Pods, 1) {
		assert.Equal(t, "p1", resp.Pods[0].Pod)
		assert.Equal(t, int64(500), resp.Pods[0].CpuMilli)
	}
}

// Test_clusterSvc_ClusterBoard_Error 失败路径：管理集合 / biz 聚合任一环节失败均上抛错误。
func Test_clusterSvc_ClusterBoard_Error(t *testing.T) {
	t.Run("管理集合失败", func(t *testing.T) {
		svc, mocks := newClusterSvcWithMocks(t)
		mocks.nsBiz.EXPECT().ListAllNames(gomock.Any()).Return(nil, errors.New("boom"))

		resp, err := svc.ClusterBoard(newAdminUserCtx(), &cluster.BoardRequest{})
		assert.Nil(t, resp)
		assert.Error(t, err)
	})

	t.Run("biz 聚合失败", func(t *testing.T) {
		svc, mocks := newClusterSvcWithMocks(t)
		mocks.nsBiz.EXPECT().ListAllNames(gomock.Any()).Return([]string{"ns-a"}, nil)
		mocks.k8sBiz.EXPECT().ClusterBoard(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))

		resp, err := svc.ClusterBoard(newAdminUserCtx(), &cluster.BoardRequest{})
		assert.Nil(t, resp)
		assert.Error(t, err)
	})
}

// Test_clusterSvc_ClusterBoard_TopSort top_sort 透传：biz 收到与请求一致的排行维度。
func Test_clusterSvc_ClusterBoard_TopSort(t *testing.T) {
	svc, mocks := newClusterSvcWithMocks(t)
	mocks.nsBiz.EXPECT().ListAllNames(gomock.Any()).Return([]string{"ns-a"}, nil)
	mocks.k8sBiz.EXPECT().ClusterBoard(gomock.Any(), gomock.Any(), gomock.Eq("mem")).Return(&biz.ClusterBoard{Overview: &biz.ClusterInfo{}}, nil)

	resp, err := svc.ClusterBoard(newAdminUserCtx(), &cluster.BoardRequest{TopSort: "mem"})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// Test_clusterSvc_ResourceBoard 成功路径：管理集合 + 项目归属 + biz 空间板
// 逐字段映射为 ResourceBoardResponse。
func Test_clusterSvc_ResourceBoard(t *testing.T) {
	svc, mocks := newClusterSvcWithMocks(t)

	mocks.nsBiz.EXPECT().ListAllNames(gomock.Any()).Return([]string{"ns-a"}, nil)
	mocks.projectBiz.EXPECT().ListAllProjectBriefs(gomock.Any()).Return([]*biz.Project{{Name: "web"}}, nil)
	mocks.k8sBiz.EXPECT().ResourceBoard(gomock.Any(), []string{"ns-a"}, gomock.Any()).Return(&biz.ResourceBoard{
		Namespaces: []*biz.ResourceNamespace{{
			Name:            "ns-a",
			PodCount:        2,
			CpuRequestMilli: 1500,
			CpuUsageMilli:   500,
			MemRequestBytes: 268435456,
			MemUsageBytes:   134217728,
			Projects: []*biz.ResourceProject{{
				Name:            "web",
				PodCount:        1,
				CpuRequestMilli: 500,
				CpuUsageMilli:   200,
				MemRequestBytes: 134217728,
				MemUsageBytes:   67108864,
				Workloads: []*biz.ResourceProjectWorkload{{
					Kind:            "Deployment",
					Name:            "web-api",
					PodCount:        1,
					CpuRequestMilli: 500,
					CpuUsageMilli:   200,
					MemRequestBytes: 134217728,
					MemUsageBytes:   67108864,
				}},
			}},
		}},
	}, nil)

	resp, err := svc.ResourceBoard(newAdminUserCtx(), &cluster.InfoRequest{})
	assert.NoError(t, err)
	if assert.Len(t, resp.Namespaces, 1) {
		ns := resp.Namespaces[0]
		assert.Equal(t, "ns-a", ns.Name)
		assert.Equal(t, int32(2), ns.PodCount)
		assert.Equal(t, int64(1500), ns.CpuRequestMilli)
		assert.Equal(t, int64(500), ns.CpuUsageMilli)
		assert.Equal(t, int64(268435456), ns.MemRequestBytes)
		assert.Equal(t, int64(134217728), ns.MemUsageBytes)
		if assert.Len(t, ns.Projects, 1) {
			p := ns.Projects[0]
			assert.Equal(t, "web", p.Name)
			assert.Equal(t, int32(1), p.PodCount)
			assert.Equal(t, int64(500), p.CpuRequestMilli)
			assert.Equal(t, int64(200), p.CpuUsageMilli)
			assert.Equal(t, int64(134217728), p.MemRequestBytes)
			assert.Equal(t, int64(67108864), p.MemUsageBytes)
			if assert.Len(t, p.Workloads, 1) {
				w := p.Workloads[0]
				assert.Equal(t, "Deployment", w.Kind)
				assert.Equal(t, "web-api", w.Name)
				assert.Equal(t, int32(1), w.PodCount)
				assert.Equal(t, int64(500), w.CpuRequestMilli)
				assert.Equal(t, int64(200), w.CpuUsageMilli)
				assert.Equal(t, int64(134217728), w.MemRequestBytes)
				assert.Equal(t, int64(67108864), w.MemUsageBytes)
			}
		}
	}
}

// TestToWorkloadProtos 映射边界：nil/空列表返回空表，多元素逐字段透传。
func TestToWorkloadProtos(t *testing.T) {
	assert.Empty(t, toWorkloadProtos(nil))
	assert.Empty(t, toWorkloadProtos([]*biz.ResourceProjectWorkload{}))

	got := toWorkloadProtos([]*biz.ResourceProjectWorkload{
		{Kind: "StatefulSet", Name: "db", PodCount: 3, CpuRequestMilli: 300, CpuUsageMilli: 100, MemRequestBytes: 1024, MemUsageBytes: 512},
		{Kind: "Deployment", Name: "web", PodCount: 1, CpuRequestMilli: 500},
	})
	if assert.Len(t, got, 2) {
		assert.Equal(t, "StatefulSet", got[0].Kind)
		assert.Equal(t, "db", got[0].Name)
		assert.Equal(t, int32(3), got[0].PodCount)
		assert.Equal(t, int64(300), got[0].CpuRequestMilli)
		assert.Equal(t, int64(100), got[0].CpuUsageMilli)
		assert.Equal(t, int64(1024), got[0].MemRequestBytes)
		assert.Equal(t, int64(512), got[0].MemUsageBytes)
		assert.Equal(t, "Deployment", got[1].Kind)
		assert.Equal(t, "web", got[1].Name)
		assert.Equal(t, int64(500), got[1].CpuRequestMilli)
	}
}

// Test_clusterSvc_ResourceBoard_Error 失败路径：管理集合 / 项目归属 / biz 聚合
// 任一环节失败均上抛错误，返回 nil 响应。
func Test_clusterSvc_ResourceBoard_Error(t *testing.T) {
	t.Run("管理集合失败", func(t *testing.T) {
		svc, mocks := newClusterSvcWithMocks(t)
		mocks.nsBiz.EXPECT().ListAllNames(gomock.Any()).Return(nil, errors.New("boom"))

		resp, err := svc.ResourceBoard(newAdminUserCtx(), &cluster.InfoRequest{})
		assert.Nil(t, resp)
		assert.Error(t, err)
	})

	t.Run("项目归属失败", func(t *testing.T) {
		svc, mocks := newClusterSvcWithMocks(t)
		mocks.nsBiz.EXPECT().ListAllNames(gomock.Any()).Return([]string{"ns-a"}, nil)
		mocks.projectBiz.EXPECT().ListAllProjectBriefs(gomock.Any()).Return(nil, errors.New("boom"))

		resp, err := svc.ResourceBoard(newAdminUserCtx(), &cluster.InfoRequest{})
		assert.Nil(t, resp)
		assert.Error(t, err)
	})

	t.Run("biz 聚合失败", func(t *testing.T) {
		svc, mocks := newClusterSvcWithMocks(t)
		mocks.nsBiz.EXPECT().ListAllNames(gomock.Any()).Return([]string{"ns-a"}, nil)
		mocks.projectBiz.EXPECT().ListAllProjectBriefs(gomock.Any()).Return(nil, nil)
		mocks.k8sBiz.EXPECT().ResourceBoard(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))

		resp, err := svc.ResourceBoard(newAdminUserCtx(), &cluster.InfoRequest{})
		assert.Nil(t, resp)
		assert.Error(t, err)
	})
}

// Test_clusterSvc_Authorize 授权门禁：ClusterInfo 免登录放行；非白名单方法
// admin 放行、普通用户拒绝。
func Test_clusterSvc_Authorize(t *testing.T) {
	svc, _ := newClusterSvcWithMocks(t)

	// 白名单方法（ClusterInfo）：直接放行，不要求 admin。
	ctx, err := svc.Authorize(newAdminUserCtx(), cluster.Cluster_ClusterInfo_FullMethodName)
	assert.NoError(t, err)
	assert.NotNil(t, ctx)

	// 非白名单方法：admin 放行。
	ctx, err = svc.Authorize(newAdminUserCtx(), cluster.Cluster_ClusterBoard_FullMethodName)
	assert.NoError(t, err)
	assert.NotNil(t, ctx)

	// 非白名单方法：普通用户拒绝。
	ctx, err = svc.Authorize(newOtherUserCtx(), cluster.Cluster_ClusterBoard_FullMethodName)
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
	assert.Nil(t, ctx)

	// DeployTrend（每日部署趋势）非公开方法：admin 放行、普通用户拒绝——与 ClusterBoard
	// 同门禁（allowlist 仅 ClusterInfo），此处显式钉死契约，防误加进免登录白名单。
	ctx, err = svc.Authorize(newAdminUserCtx(), cluster.Cluster_DeployTrend_FullMethodName)
	assert.NoError(t, err)
	assert.NotNil(t, ctx)
	ctx, err = svc.Authorize(newOtherUserCtx(), cluster.Cluster_DeployTrend_FullMethodName)
	assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
	assert.Nil(t, ctx)
}
