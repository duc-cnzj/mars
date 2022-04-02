package services

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"

	"github.com/duc-cnzj/mars-client/v4/metrics"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/dustin/go-humanize"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	AddServerFunc(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		metrics.RegisterMetricsServer(s, new(MetricsSvc))
	})
	AddEndpointFunc(metrics.RegisterMetricsHandlerFromEndpoint)
}

type MetricsSvc struct {
	metrics.UnsafeMetricsServer
}

var (
	tickDuration = 5 * time.Second
	timeSpan     = 5 * time.Second * 30
	length       = timeSpan / tickDuration
)

var now = func() string {
	return time.Now().Format("15:04:05")
}

func (m *MetricsSvc) TopPod(ctx context.Context, request *metrics.MetricsTopPodRequest) (*metrics.MetricsTopPodResponse, error) {
	podMetrics, err := app.K8sMetrics().MetricsV1beta1().PodMetricses(request.Namespace).Get(context.TODO(), request.Pod, metav1.GetOptions{})
	if err != nil {
		running, reason := utils.IsPodRunning(request.Namespace, request.Pod)
		if !running {
			return nil, status.Error(codes.NotFound, reason)
		}
		return nil, err
	}

	return m.metrics(podMetrics), nil
}

func (m *MetricsSvc) StreamTopPod(request *metrics.MetricsTopPodRequest, server metrics.Metrics_StreamTopPodServer) error {
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()
	defer mlog.Debug("ProjectByID exit")

	fn := func() error {
		podMetrics, err := app.K8sMetrics().MetricsV1beta1().PodMetricses(request.Namespace).Get(context.TODO(), request.Pod, metav1.GetOptions{})
		if err != nil {
			running, _ := utils.IsPodRunning(request.Namespace, request.Pod)
			if running {
				return nil
			}
			return err
		}

		if err := server.Send(m.metrics(podMetrics)); err != nil {
			mlog.Error(err)
			return err
		}
		return nil
	}

	if err := fn(); err != nil {
		return err
	}
	for {
		select {
		case <-server.Context().Done():
			return server.Context().Err()
		case <-ticker.C:
			if err := fn(); err != nil {
				return err
			}
		case <-app.App().Done():
			return nil
		}
	}
}

func (m *MetricsSvc) CpuMemoryInProject(ctx context.Context, request *metrics.MetricsCpuMemoryInProjectRequest) (*metrics.MetricsCpuMemoryInProjectResponse, error) {
	var p models.Project
	if err := app.DB().Where("`id` = ?", request.ProjectId).First(&p).Error; err != nil {
		return nil, err
	}
	cpu, memory := utils.GetCpuAndMemory(p.GetAllPodMetrics())

	return &metrics.MetricsCpuMemoryInProjectResponse{
		Cpu:    cpu,
		Memory: memory,
	}, nil
}

func (m *MetricsSvc) CpuMemoryInNamespace(ctx context.Context, request *metrics.MetricsCpuMemoryInNamespaceRequest) (*metrics.MetricsCpuMemoryInNamespaceResponse, error) {
	var ns models.Namespace
	if err := app.DB().Preload("Projects").Where("`id` = ?", request.NamespaceId).First(&ns).Error; err != nil {
		return nil, err
	}

	cpu, memory := utils.GetCpuAndMemoryInNamespace(ns.Name)

	return &metrics.MetricsCpuMemoryInNamespaceResponse{
		Cpu:    cpu,
		Memory: memory,
	}, nil
}

func (m *MetricsSvc) metrics(podMetrics *v1beta1.PodMetrics) *metrics.MetricsTopPodResponse {
	cpu, memory := utils.GetCpuAndMemoryQuantity(*podMetrics)
	cpuM := cpu.MilliValue()
	var HumanizeCpu string = fmt.Sprintf("%v m", float64(cpu.MilliValue()))
	if cpuM > 1000 {
		HumanizeCpu = fmt.Sprintf("%.3f", float64(cpu.MilliValue())/1000)
	}
	asInt64, _ := memory.AsInt64()

	return &metrics.MetricsTopPodResponse{
		Cpu:            float64(cpu.MilliValue()),
		Memory:         float64(memory.ScaledValue(3)),
		HumanizeCpu:    HumanizeCpu,
		HumanizeMemory: humanize.Bytes(uint64(asInt64)),
		Time:           now(),
		Length:         int64(length),
	}
}
