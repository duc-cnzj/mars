package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v6/proto/cluster"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
)

var _ cluster.ClusterServer = (*clusterSvc)(nil)

// clusterSvc 是 cluster.ClusterServer 的 gRPC 实现：返回集群信息概览，由 NewClusterSvc 构造。
type clusterSvc struct {
	cluster.UnimplementedClusterServer

	k8sBiz biz.K8sBiz
	logger mlog.Logger
}

// ClusterSvcDeps 收口 NewClusterSvc 的构造依赖，由 wire 按字段注入。
type ClusterSvcDeps struct {
	K8sBiz biz.K8sBiz
	Logger mlog.Logger
}

// NewClusterSvc 收口集群信息服务的构造依赖，由 wire 按字段注入。
func NewClusterSvc(deps ClusterSvcDeps) cluster.ClusterServer {
	return &clusterSvc{k8sBiz: deps.K8sBiz, logger: deps.Logger.WithModule("services/cluster")}
}

// ClusterInfo 返回当前 k8s 集群的基础信息，为免登录公开接口（白名单见 middlewares.PublicMethods）。
func (c *clusterSvc) ClusterInfo(ctx context.Context, req *cluster.InfoRequest) (*cluster.InfoResponse, error) {
	return &cluster.InfoResponse{
		Item: transformer.FromClusterInfo(c.k8sBiz.ClusterInfo()),
	}, nil
}
