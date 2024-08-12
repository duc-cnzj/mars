package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v4/cluster"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/transformer"
)

var _ cluster.ClusterServer = (*clusterSvc)(nil)

type clusterSvc struct {
	guest
	repo repo.K8sRepo
	cluster.UnimplementedClusterServer
}

func NewClusterSvc(repo repo.K8sRepo) cluster.ClusterServer {
	return &clusterSvc{repo: repo}
}

func (c *clusterSvc) ClusterInfo(ctx context.Context, req *cluster.InfoRequest) (*cluster.InfoResponse, error) {
	info := c.repo.ClusterInfo()

	return &cluster.InfoResponse{
		Item: transformer.FromClusterInfo(info),
	}, nil
}
