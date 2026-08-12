package transformer

import (
	"github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz"
)

// FromClusterInfo 把 biz.ClusterInfo 转换为 websocket ClusterInfo。
func FromClusterInfo(info *biz.ClusterInfo) *websocket.ClusterInfo {
	if info == nil {
		return nil
	}
	return &websocket.ClusterInfo{
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
	}
}
