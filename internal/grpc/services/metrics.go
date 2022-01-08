package services

import (
	"context"
	"fmt"
	"time"

	"k8s.io/metrics/pkg/apis/metrics/v1beta1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/duc-cnzj/mars/client/metrics"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/dustin/go-humanize"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Metrics struct {
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

func (m *Metrics) Show(ctx context.Context, request *metrics.MetricsShowRequest) (*metrics.MetricsShowResponse, error) {
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

func (m *Metrics) StreamShow(request *metrics.MetricsShowRequest, server metrics.Metrics_StreamShowServer) error {
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

func (m *Metrics) metrics(podMetrics *v1beta1.PodMetrics) *metrics.MetricsShowResponse {
	cpu, memory := utils.GetCpuAndMemoryQuantity(*podMetrics)
	cpuM := cpu.MilliValue()
	var HumanizeCpu string = fmt.Sprintf("%v m", float64(cpu.MilliValue()))
	if cpuM > 1000 {
		HumanizeCpu = fmt.Sprintf("%.3f", float64(cpu.MilliValue())/1000)
	}
	asInt64, _ := memory.AsInt64()

	return &metrics.MetricsShowResponse{
		Cpu:            float64(cpu.MilliValue()),
		Memory:         float64(memory.ScaledValue(3)),
		HumanizeCpu:    HumanizeCpu,
		HumanizeMemory: humanize.Bytes(uint64(asInt64)),
		Time:           now(),
		Length:         int64(length),
	}
}
