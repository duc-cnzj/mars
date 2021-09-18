package utils

import (
	"context"
	"fmt"
	"strings"

	app "github.com/duc-cnzj/mars/internal/app/helper"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/dustin/go-humanize"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

type ClusterStatus string

const (
	StatusBad     ClusterStatus = "bad"
	StatusNotGood ClusterStatus = "not good"
	StatusHealth  ClusterStatus = "health"
)

type InfoResponse struct {
	// 健康状况
	Status ClusterStatus `json:"status"`

	// 可用内存
	FreeMemory string `json:"free_memory"`
	// 可用 cpu
	FreeCpu string `json:"free_cpu"`

	// 可分配内存
	FreeRequestMemory string `json:"free_request_memory"`
	// 可分配 cpu
	FreeRequestCpu string `json:"free_request_cpu"`

	// 总共的可调度的内存
	TotalMemory string `json:"total_memory"`
	// 总共的可调度的 cpu
	TotalCpu string `json:"total_cpu"`

	// 内存使用率
	UsageMemoryRate string `json:"usage_memory_rate"`
	// cpu 使用率
	UsageCpuRate string `json:"usage_cpu_rate"`

	// 内存分配率
	RequestMemoryRate string `json:"request_memory_rate"`
	// cpu 分配率
	RequestCpuRate string `json:"request_cpu_rate"`
}

func ClusterInfo() *InfoResponse {
	selector := labels.Everything()
	var nodes []v1.Node

	// 获取已经使用的 cpu, memory
	nodeList, _ := app.K8sClient().Client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	nodes = append(nodes, nodeList.Items...)
	allocatable := make(map[string]v1.ResourceList)

	var (
		totalCpu    = &resource.Quantity{}
		totalMemory = &resource.Quantity{}
	)

	var (
		workerNodes    []v1.Node
		noExecuteNodes []v1.Node
	)

	for _, node := range nodes {
		noExecute := false
		for _, taint := range node.Spec.Taints {
			if taint.Effect == v1.TaintEffectNoExecute {
				noExecute = true
				break
			}
		}
		if !noExecute {
			workerNodes = append(workerNodes, node)
		} else {
			noExecuteNodes = append(workerNodes, node)
		}
	}

	for _, n := range workerNodes {
		allocatable[n.Name] = n.Status.Allocatable
		totalCpu.Add(n.Status.Allocatable.Cpu().DeepCopy())
		totalMemory.Add(n.Status.Allocatable.Memory().DeepCopy())
	}

	requestCpu, requestMemory := getNodeRequestCpuAndMemory(noExecuteNodes)
	var (
		usedCpu    = &resource.Quantity{}
		usedMemory = &resource.Quantity{}
	)

	list, _ := app.K8sMetrics().MetricsV1beta1().NodeMetricses().List(context.TODO(), metav1.ListOptions{
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

	freeRequestMemory := totalMemory.DeepCopy()
	freeRequestMemory.Sub(*requestMemory)
	freeRequestCpu := totalCpu.DeepCopy()
	freeRequestCpu.Sub(*requestCpu)

	rateMemory := float64(usedMemory.Value()) / float64(totalMemory.Value()) * 100
	rateCpu := float64(usedCpu.Value()) / float64(totalCpu.Value()) * 100
	rateRequestMemory := float64(requestMemory.Value()) / float64(totalMemory.Value()) * 100
	rateRequestCpu := float64(requestCpu.Value()) / float64(totalCpu.Value()) * 100

	var status = StatusHealth
	if rateRequestMemory > 60 || rateRequestCpu > 60 {
		status = StatusNotGood
	}
	if rateRequestMemory > 80 || rateRequestCpu > 80 {
		status = StatusBad
	}

	return &InfoResponse{
		Status:            status,
		FreeRequestMemory: humanize.IBytes(uint64(freeRequestMemory.Value())),
		FreeRequestCpu:    fmt.Sprintf("%.2f core", float64(freeRequestCpu.MilliValue())/1000),
		FreeMemory:        humanize.IBytes(uint64(freeMemory.Value())),
		FreeCpu:           fmt.Sprintf("%.2f core", float64(freeCpu.MilliValue())/1000),
		TotalMemory:       humanize.IBytes(uint64(totalMemory.Value())),
		TotalCpu:          fmt.Sprintf("%.2f core", float64(totalCpu.MilliValue())/1000),
		UsageMemoryRate:   fmt.Sprintf("%.1f%%", rateMemory),
		UsageCpuRate:      fmt.Sprintf("%.1f%%", rateCpu),
		RequestCpuRate:    fmt.Sprintf("%.1f%%", rateRequestCpu),
		RequestMemoryRate: fmt.Sprintf("%.1f%%", rateRequestMemory),
	}
}

func getNodeRequestCpuAndMemory(noExecuteNodes []v1.Node) (*resource.Quantity, *resource.Quantity) {
	var (
		requestCpu    = &resource.Quantity{}
		requestMemory = &resource.Quantity{}
	)

	var nodeSelector []string = []string{
		"status.phase!=" + string(v1.PodSucceeded),
		"status.phase!=" + string(v1.PodFailed),
	}
	for _, node := range noExecuteNodes {
		nodeSelector = append(nodeSelector, "spec.nodeName!="+node.Name)
	}
	fieldSelector, _ := fields.ParseSelector(strings.Join(nodeSelector, ","))
	nodeNonTerminatedPodsList, err := app.K8sClientSet().CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{FieldSelector: fieldSelector.String()})
	if err != nil {
		mlog.Error(err)
		return requestCpu, requestMemory
	}
	for _, item := range nodeNonTerminatedPodsList.Items {
		for _, container := range item.Spec.Containers {
			requestCpu.Add(container.Resources.Requests.Cpu().DeepCopy())
			requestMemory.Add(container.Resources.Requests.Memory().DeepCopy())
		}
	}

	return requestCpu, requestMemory
}
