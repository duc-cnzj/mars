package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v6/proto/cluster"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
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
	logger       mlog.Logger
}

// ClusterSvcDeps 收口 NewClusterSvc 的构造依赖，由 wire 按字段注入。
type ClusterSvcDeps struct {
	K8sBiz       biz.K8sBiz
	NamespaceBiz biz.NamespaceBiz
	ProjectBiz   biz.ProjectBiz
	AccessBiz    biz.AccessBiz
	Logger       mlog.Logger
}

// NewClusterSvc 收口集群信息服务的构造依赖，由 wire 按字段注入。
func NewClusterSvc(deps ClusterSvcDeps) cluster.ClusterServer {
	return &clusterSvc{
		k8sBiz:       deps.K8sBiz,
		namespaceBiz: deps.NamespaceBiz,
		projectBiz:   deps.ProjectBiz,
		accessBiz:    deps.AccessBiz,
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
// biz 控制 Top Pod 排行维度（cpu 默认 / mem）。
func (c *clusterSvc) ClusterBoard(ctx context.Context, req *cluster.BoardRequest) (*cluster.BoardResponse, error) {
	managedNames, err := c.namespaceBiz.ListAllNames(ctx)
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
// 的空间）。命名空间集合由 namespaceBiz 提供，项目归属由 projectBiz.ListAll 提供，
// biz 层按项目 PodSelectors 匹配 pod 拆分后返回。
func (c *clusterSvc) ResourceBoard(ctx context.Context, req *cluster.InfoRequest) (*cluster.ResourceBoardResponse, error) {
	managedNames, err := c.namespaceBiz.ListAllNames(ctx)
	if err != nil {
		return nil, logError(ctx, c.logger, err)
	}
	projects, err := c.projectBiz.ListAll(ctx)
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
