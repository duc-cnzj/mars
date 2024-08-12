package transformer

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/stretchr/testify/assert"
)

func TestFromClusterInfoWithValidInfo(t *testing.T) {
	info := &repo.ClusterInfo{
		Status:            "Running",
		FreeMemory:        "1024",
		FreeCpu:           "2",
		FreeRequestMemory: "512",
		FreeRequestCpu:    "1",
		TotalMemory:       "2048",
		TotalCpu:          "4",
		UsageMemoryRate:   "0.5",
		UsageCpuRate:      "0.5",
		RequestMemoryRate: "0.25",
		RequestCpuRate:    "0.25",
	}

	result := FromClusterInfo(info)

	assert.Equal(t, info.Status, result.Status)
	assert.Equal(t, info.FreeMemory, result.FreeMemory)
	assert.Equal(t, info.FreeCpu, result.FreeCpu)
	assert.Equal(t, info.FreeRequestMemory, result.FreeRequestMemory)
	assert.Equal(t, info.FreeRequestCpu, result.FreeRequestCpu)
	assert.Equal(t, info.TotalMemory, result.TotalMemory)
	assert.Equal(t, info.TotalCpu, result.TotalCpu)
	assert.Equal(t, info.UsageMemoryRate, result.UsageMemoryRate)
	assert.Equal(t, info.UsageCpuRate, result.UsageCpuRate)
	assert.Equal(t, info.RequestMemoryRate, result.RequestMemoryRate)
	assert.Equal(t, info.RequestCpuRate, result.RequestCpuRate)
}

func TestFromClusterInfoWithNilInfo(t *testing.T) {
	result := FromClusterInfo(nil)
	assert.Nil(t, result)
}
