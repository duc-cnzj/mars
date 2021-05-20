package utils

import (
	"context"
	"fmt"

	"k8s.io/metrics/pkg/apis/metrics/v1beta1"

	"github.com/dustin/go-humanize"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetCpuAndMemoryInNamespace(namespace string) (string, string) {
	metricses := K8sMetrics().MetricsV1beta1().PodMetricses(namespace)
	list, _ := metricses.List(context.Background(), metav1.ListOptions{})
	return analyseMetricsToCpuAndMemory(list)
}

func GetCpuAndMemoryInNamespaceByRelease(namespace, releaseName string) (string, string) {
	metricses := K8sMetrics().MetricsV1beta1().PodMetricses(namespace)
	list, _ := metricses.List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/instance=%s", releaseName),
	})
	return analyseMetricsToCpuAndMemory(list)
}

func analyseMetricsToCpuAndMemory(list *v1beta1.PodMetricsList) (string, string) {
	var cpu, memory *resource.Quantity

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

	var cpuStr, memoryStr string = "0 m", "0 MB"
	if cpu != nil {
		cpuStr = fmt.Sprintf("%d m", cpu.MilliValue())
	}
	if memory != nil {
		asInt64, _ := memory.AsInt64()
		memoryStr = humanize.Bytes(uint64(asInt64))
	}

	return cpuStr, memoryStr
}
