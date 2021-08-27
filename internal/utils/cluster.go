package utils

import (
	"context"
	"fmt"

	"github.com/dustin/go-humanize"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type ClusterStatus string

const (
	StatusBad     ClusterStatus = "bad"
	StatusNotGood ClusterStatus = "not good"
	StatusHealth  ClusterStatus = "health"
)

type InfoResponse struct {
	Status ClusterStatus `json:"status"`

	FreeMemory string `json:"free_memory"`
	FreeCpu    string `json:"free_cpu"`

	TotalMemory string `json:"total_memory"`
	TotalCpu    string `json:"total_cpu"`

	UsageMemoryRate string `json:"usage_memory_rate"`
	UsageCpuRate    string `json:"usage_cpu_rate"`
}

func ClusterInfo() *InfoResponse {
	selector := labels.Everything()
	var nodes []v1.Node

	// 获取已经使用的 cpu, memory
	nodeList, _ := K8sClient().Client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	nodes = append(nodes, nodeList.Items...)
	allocatable := make(map[string]v1.ResourceList)

	var (
		totalCpu    = &resource.Quantity{}
		totalMemory = &resource.Quantity{}
	)

	for _, n := range nodes {
		allocatable[n.Name] = n.Status.Allocatable
		totalCpu.Add(n.Status.Allocatable.Cpu().DeepCopy())
		totalMemory.Add(n.Status.Allocatable.Memory().DeepCopy())
	}

	var (
		usedCpu    = &resource.Quantity{}
		usedMemory = &resource.Quantity{}
	)

	list, _ := K8sMetrics().MetricsV1beta1().NodeMetricses().List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})

	for _, item := range list.Items {
		usedCpu.Add(item.Usage.Cpu().DeepCopy())
		usedMemory.Add(item.Usage.Memory().DeepCopy())
	}

	freeMemory := totalMemory.DeepCopy()
	freeMemory.Sub(*usedMemory)
	freeCpu := totalCpu.DeepCopy()
	freeCpu.Sub(*usedCpu)

	rateMemory := float64(usedMemory.Value()) / float64(totalMemory.Value()) * 100
	rateCpu := float64(usedCpu.Value()) / float64(totalCpu.Value()) * 100

	var status = StatusHealth
	if rateMemory > 60 {
		status = StatusNotGood
	}
	if rateMemory > 80 {
		status = StatusBad
	}

	return &InfoResponse{
		Status:          status,
		FreeMemory:      humanize.IBytes(uint64(freeMemory.Value())),
		FreeCpu:         fmt.Sprintf("%.2f core", float64(freeCpu.MilliValue())/1000),
		TotalMemory:     humanize.IBytes(uint64(totalMemory.Value())),
		TotalCpu:        fmt.Sprintf("%.2f core", float64(totalCpu.MilliValue())/1000),
		UsageMemoryRate: fmt.Sprintf("%.0f%%", rateMemory),
		UsageCpuRate:    fmt.Sprintf("%.0f%%", rateCpu),
	}
}
