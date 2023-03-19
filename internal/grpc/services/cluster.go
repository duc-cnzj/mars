package services

import (
	"context"

	"google.golang.org/grpc"

	"github.com/duc-cnzj/mars-client/v4/cluster"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/utils"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		cluster.RegisterClusterServer(s, new(ClusterSvc))
	})
	RegisterEndpoint(cluster.RegisterClusterHandlerFromEndpoint)
}

type ClusterSvc struct {
	Guest

	cluster.UnimplementedClusterServer
}

func (c *ClusterSvc) ClusterInfo(ctx context.Context, req *cluster.InfoRequest) (*cluster.InfoResponse, error) {
	info := utils.ClusterInfo()

	return &cluster.InfoResponse{
		Status:            info.Status,
		FreeMemory:        info.FreeMemory,
		FreeCpu:           info.FreeCpu,
		FreeRequestMemory: info.FreeRequestMemory,
		FreeRequestCpu:    info.FreeRequestCpu,
		TotalMemory:       info.TotalMemory,
		TotalCpu:          info.TotalCpu,
		UsageMemoryRate:   info.UsageMemoryRate,
		UsageCpuRate:      info.UsageCpuRate,
		RequestMemoryRate: info.RequestMemoryRate,
		RequestCpuRate:    info.RequestCpuRate,
	}, nil
}
