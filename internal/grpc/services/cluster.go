package services

import (
	"context"

	"github.com/duc-cnzj/mars/internal/mlog"

	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/pkg/cluster"
)

type Cluster struct {
	cluster.UnimplementedClusterServer
}

func (c *Cluster) ClusterInfo(ctx context.Context, req *cluster.ClusterInfoRequest) (*cluster.ClusterInfoResponse, error) {
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

func (c *Cluster) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	mlog.Debug("client is calling method:", fullMethodName)
	return ctx, nil
}
