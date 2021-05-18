package utils

import (
	"context"
	"fmt"

	"github.com/dustin/go-humanize"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetCpuAndMemoryInNamespace(namespace string) (string, string) {
	var cpu, memory *resource.Quantity
	metricses := K8sMetrics().MetricsV1beta1().PodMetricses(namespace)
	list, _ := metricses.List(context.Background(), metav1.ListOptions{})
	for _, item := range list.Items {
		for _, container := range item.Containers {
			if cpu == nil {
				cpu = container.Usage.Cpu()
			}

			if memory == nil {
				memory = container.Usage.Memory()
			}

			cpu.Add(*container.Usage.Cpu())
			memory.Add(*container.Usage.Memory())
		}
	}

	var cpuStr string
	var memoryStr string
	if cpu != nil {
		cpuStr = fmt.Sprintf("%d m", cpu.MilliValue())
	}
	if memory != nil {
		asInt64, _ := memory.AsInt64()
		memoryStr = humanize.Bytes(uint64(asInt64))
	}

	return cpuStr, memoryStr
}
