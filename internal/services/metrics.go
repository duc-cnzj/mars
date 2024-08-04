package services

import (
	"context"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/api/v4/metrics"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/dustin/go-humanize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

var _ metrics.MetricsServer = (*metricsSvc)(nil)

type metricsSvc struct {
	metrics.UnimplementedMetricsServer
	k8sRepo  repo.K8sRepo
	logger   mlog.Logger
	projRepo repo.ProjectRepo
	nsRepo   repo.NamespaceRepo
}

func NewMetricsSvc(k8sRepo repo.K8sRepo, logger mlog.Logger, projRepo repo.ProjectRepo, nsRepo repo.NamespaceRepo) metrics.MetricsServer {
	return &metricsSvc{k8sRepo: k8sRepo, logger: logger, projRepo: projRepo, nsRepo: nsRepo}
}

var (
	tickDuration = 5 * time.Second
	timeSpan     = 5 * time.Second * 30
	length       = timeSpan / tickDuration
)

var now = func() string {
	return time.Now().Format("15:04:05")
}

func (m *metricsSvc) TopPod(ctx context.Context, request *metrics.TopPodRequest) (*metrics.TopPodResponse, error) {
	podMetrics, err := m.k8sRepo.GetPodMetrics(ctx, request.Namespace, request.Pod)
	if err != nil {
		running, reason := m.k8sRepo.IsPodRunning(request.Namespace, request.Pod)
		if !running {
			return nil, status.Error(codes.NotFound, reason)
		}
		return nil, err
	}

	return m.metrics(podMetrics), nil
}

func (m *metricsSvc) StreamTopPod(request *metrics.TopPodRequest, server metrics.Metrics_StreamTopPodServer) error {
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()
	defer m.logger.Debug("ProjectByID exit")

	fn := func() error {
		podMetrics, err := m.k8sRepo.GetPodMetrics(server.Context(), request.Namespace, request.Pod)
		if err != nil {
			running, _ := m.k8sRepo.IsPodRunning(request.Namespace, request.Pod)
			if running {
				return nil
			}
			return err
		}

		if err := server.Send(m.metrics(podMetrics)); err != nil {
			m.logger.Error(err)
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
			return nil
		case <-ticker.C:
			if err := fn(); err != nil {
				return err
			}
			//case <-m.app.Done():
			//	return nil
		}
	}
}

func (m *metricsSvc) CpuMemoryInProject(ctx context.Context, request *metrics.CpuMemoryInProjectRequest) (*metrics.CpuMemoryInProjectResponse, error) {
	p, err := m.projRepo.Show(ctx, int(request.ProjectId))
	if err != nil {
		return nil, err
	}
	cpu, memory := m.k8sRepo.GetCpuAndMemory(m.projRepo.GetAllPodMetrics(p))

	return &metrics.CpuMemoryInProjectResponse{
		Cpu:    cpu,
		Memory: memory,
	}, nil
}

func (m *metricsSvc) CpuMemoryInNamespace(ctx context.Context, request *metrics.CpuMemoryInNamespaceRequest) (*metrics.CpuMemoryInNamespaceResponse, error) {
	ns, err := m.nsRepo.Show(ctx, int(request.NamespaceId))
	if err != nil {
		return nil, err
	}

	cpu, memory := m.k8sRepo.GetCpuAndMemoryInNamespace(ns.Name)

	return &metrics.CpuMemoryInNamespaceResponse{
		Cpu:    cpu,
		Memory: memory,
	}, nil
}

func (m *metricsSvc) metrics(podMetrics *v1beta1.PodMetrics) *metrics.TopPodResponse {
	cpu, memory := m.k8sRepo.GetCpuAndMemoryQuantity(*podMetrics)
	cpuM := cpu.MilliValue()
	var HumanizeCpu string = fmt.Sprintf("%v m", float64(cpu.MilliValue()))
	if cpuM > 1000 {
		HumanizeCpu = fmt.Sprintf("%.3f", float64(cpu.MilliValue())/1000)
	}
	asInt64, _ := memory.AsInt64()

	return &metrics.TopPodResponse{
		Cpu:            float64(cpu.MilliValue()),
		Memory:         float64(memory.ScaledValue(3)),
		HumanizeCpu:    HumanizeCpu,
		HumanizeMemory: humanize.Bytes(uint64(asInt64)),
		Time:           now(),
		Length:         int32(length),
	}
}
