package services

import (
	"context"

	"google.golang.org/grpc"

	"github.com/duc-cnzj/mars-client/v4/cluster"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/utils"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		cluster.RegisterClusterServer(s, new(ClusterSvc))
	})
	RegisterEndpoint(cluster.RegisterClusterHandlerFromEndpoint)
}

type ClusterSvc struct {
	cluster.UnimplementedClusterServer
}

func (c *ClusterSvc) ClusterInfo(ctx context.Context, req *cluster.ClusterInfoRequest) (*cluster.ClusterInfoResponse, error) {
	info := utils.ClusterInfo()

	return &cluster.ClusterInfoResponse{
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

func (c *ClusterSvc) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	mlog.Debug("client is calling method:", fullMethodName)
	return ctx, nil
}
