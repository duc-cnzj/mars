package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v6/proto/cluster"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"go.opentelemetry.io/otel"
)

var _ cluster.ClusterServer = (*clusterSvc)(nil)

// clusterSvc 是 cluster.ClusterServer 的 gRPC 实现：返回集群信息概览与管理员看板
// /空间资源聚合，由 NewClusterSvc 构造。
type clusterSvc struct {
	cluster.UnimplementedClusterServer

	k8sBiz       biz.K8sBiz
	namespaceBiz biz.NamespaceBiz
	projectBiz   biz.ProjectBiz
	accessBiz    biz.AccessBiz
	changelogBiz biz.ChangelogBiz
	logger       mlog.Logger
}

// ClusterSvcDeps 收口 NewClusterSvc 的构造依赖，由 wire 按字段注入。
type ClusterSvcDeps struct {
	K8sBiz       biz.K8sBiz
	NamespaceBiz biz.NamespaceBiz
	ProjectBiz   biz.ProjectBiz
	AccessBiz    biz.AccessBiz
	ChangelogBiz biz.ChangelogBiz
	Logger       mlog.Logger
}

// NewClusterSvc 收口集群信息服务的构造依赖，由 wire 按字段注入。
func NewClusterSvc(deps ClusterSvcDeps) cluster.ClusterServer {
	return &clusterSvc{
		k8sBiz:       deps.K8sBiz,
		namespaceBiz: deps.NamespaceBiz,
		projectBiz:   deps.ProjectBiz,
		accessBiz:    deps.AccessBiz,
		changelogBiz: deps.ChangelogBiz,
		logger:       deps.Logger.WithModule("services/cluster"),
	}
}

// Authorize 校验访问权限：ClusterInfo 是免登录公开方法放行，ClusterBoard 需 admin。
func (c *clusterSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	return c.accessBiz.RequireAdmin(ctx, fullMethodName, cluster.Cluster_ClusterInfo_FullMethodName)
}

// ClusterInfo 返回当前 k8s 集群的基础信息，为免登录公开接口（白名单见 biz.IsPublicMethod）。
func (c *clusterSvc) ClusterInfo(ctx context.Context, req *cluster.InfoRequest) (*cluster.InfoResponse, error) {
	return &cluster.InfoResponse{
		Item: transformer.FromClusterInfo(c.k8sBiz.ClusterInfo()),
	}, nil
}

// ClusterBoard 返回集群看板聚合快照：总览 + 节点明细 + 命名空间用量 + Top Pod，
// 为管理员专用接口。命名空间排行/Top Pod 只保留 mars 自己管理的命名空间及其 Pod
// （由 namespaceBiz 提供管理空间名集合，biz 层过滤后 TopN）；req.top_sort 透传给
// biz 控制 Top Pod 排行维度（cpu 默认 / mem）。管理空间名查询包一层 span，trace
// 面板区分「DB 查管理空间慢」还是「快照拉取/聚合慢」。
func (c *clusterSvc) ClusterBoard(ctx context.Context, req *cluster.BoardRequest) (*cluster.BoardResponse, error) {
	nsCtx, nsSpan := otel.Tracer("").Start(ctx, "clusterSvc/ClusterBoard/listNames")
	managedNames, err := c.namespaceBiz.ListAllNames(nsCtx)
	nsSpan.End()
	if err != nil {
		return nil, logError(ctx, c.logger, err)
	}
	board, err := c.k8sBiz.ClusterBoard(ctx, managedNames, req.GetTopSort())
	if err != nil {
		return nil, logError(ctx, c.logger, err)
	}
	resp := &cluster.BoardResponse{Overview: transformer.FromClusterInfo(board.Overview)}
	for _, n := range board.Nodes {
		resp.Nodes = append(resp.Nodes, &cluster.BoardNode{
			Name:        n.Name,
			Role:        n.Role,
			Status:      n.Status,
			CpuCapacity: n.CpuCapacity,
			CpuUsage:    n.CpuUsage,
			CpuRequest:  n.CpuRequest,
			MemCapacity: n.MemCapacity,
			MemUsage:    n.MemUsage,
			MemRequest:  n.MemRequest,
		})
	}
	for _, ns := range board.Namespaces {
		resp.Namespaces = append(resp.Namespaces, &cluster.BoardNamespace{
			Name:        ns.Name,
			CpuMilli:    ns.CpuMilli,
			MemoryBytes: ns.MemoryBytes,
			PodCount:    ns.PodCount,
		})
	}
	for _, p := range board.Pods {
		resp.Pods = append(resp.Pods, &cluster.BoardPod{
			Namespace:   p.Namespace,
			Pod:         p.Pod,
			CpuMilli:    p.CpuMilli,
			MemoryBytes: p.MemoryBytes,
		})
	}
	return resp, nil
}

// ResourceBoard 返回空间资源聚合快照：每个管理命名空间的 Pod requests/实际用量
// 占比及项目明细，为管理员专用接口（定位「申请了很多 requests 却用不到多少资源」
// 的空间）。命名空间集合由 namespaceBiz 提供，项目归属由 projectBiz.ListAllProjectBriefs 提供，
// biz 层按项目 PodSelectors 匹配 pod 拆分后返回。两个 DB 前置查询各包一层 span，
// trace 面板区分「DB 查管理空间/项目慢」还是「快照拉取/聚合慢」。
func (c *clusterSvc) ResourceBoard(ctx context.Context, req *cluster.InfoRequest) (*cluster.ResourceBoardResponse, error) {
	nsCtx, nsSpan := otel.Tracer("").Start(ctx, "clusterSvc/ResourceBoard/listNames")
	managedNames, err := c.namespaceBiz.ListAllNames(nsCtx)
	nsSpan.End()
	if err != nil {
		return nil, logError(ctx, c.logger, err)
	}
	projCtx, projSpan := otel.Tracer("").Start(ctx, "clusterSvc/ResourceBoard/listProjects")
	projects, err := c.projectBiz.ListAllProjectBriefs(projCtx)
	projSpan.End()
	if err != nil {
		return nil, logError(ctx, c.logger, err)
	}
	board, err := c.k8sBiz.ResourceBoard(ctx, managedNames, projects)
	if err != nil {
		return nil, logError(ctx, c.logger, err)
	}
	resp := &cluster.ResourceBoardResponse{}
	for _, ns := range board.Namespaces {
		item := &cluster.ResourceNamespace{
			Name:            ns.Name,
			PodCount:        ns.PodCount,
			CpuRequestMilli: ns.CpuRequestMilli,
			CpuUsageMilli:   ns.CpuUsageMilli,
			MemRequestBytes: ns.MemRequestBytes,
			MemUsageBytes:   ns.MemUsageBytes,
		}
		for _, p := range ns.Projects {
			item.Projects = append(item.Projects, &cluster.ResourceProject{
				Name:            p.Name,
				PodCount:        p.PodCount,
				CpuRequestMilli: p.CpuRequestMilli,
				CpuUsageMilli:   p.CpuUsageMilli,
				MemRequestBytes: p.MemRequestBytes,
				MemUsageBytes:   p.MemUsageBytes,
				Workloads:       toWorkloadProtos(p.Workloads),
			})
		}
		resp.Namespaces = append(resp.Namespaces, item)
	}
	return resp, nil
}

// DeployTrend 返回近 N 天每日部署次数（默认 30、上限 90）：透传 changelog biz 的按天聚合
// 结果并映射为 proto，供集群总览页「每日部署趋势」曲线。数据源 = changelog（每次部署一条），
// 服务端时区分桶、无部署补 0；管理员接口（Authorize 兜底，仅 ClusterInfo 公开）。
func (c *clusterSvc) DeployTrend(ctx context.Context, req *cluster.DeployTrendRequest) (*cluster.DeployTrendResponse, error) {
	days := biz.DeployTrendDefaultDays
	if req != nil && req.GetDays() > 0 {
		days = int(req.GetDays())
	}
	if days > biz.DeployTrendMaxDays {
		days = biz.DeployTrendMaxDays
	}
	items, err := c.changelogBiz.DeployDailyCounts(ctx, days)
	if err != nil {
		return nil, logError(ctx, c.logger, err)
	}
	resp := &cluster.DeployTrendResponse{Days: int32(days)}
	for _, it := range items {
		resp.Items = append(resp.Items, &cluster.DeployTrendPoint{Date: it.Date, Count: int32(it.Count)})
	}
	if len(items) > 0 {
		resp.StartDate = items[0].Date
		resp.EndDate = items[len(items)-1].Date
	}
	return resp, nil
}

// toWorkloadProtos 把 biz 层工作负载明细映射为 proto 消息；nil/空列表原样返回空表。
func toWorkloadProtos(workloads []*biz.ResourceProjectWorkload) []*cluster.ResourceProjectWorkload {
	result := make([]*cluster.ResourceProjectWorkload, 0, len(workloads))
	for _, w := range workloads {
		result = append(result, &cluster.ResourceProjectWorkload{
			Kind:            w.Kind,
			Name:            w.Name,
			PodCount:        w.PodCount,
			CpuRequestMilli: w.CpuRequestMilli,
			CpuUsageMilli:   w.CpuUsageMilli,
			MemRequestBytes: w.MemRequestBytes,
			MemUsageBytes:   w.MemUsageBytes,
		})
	}
	return result
}
