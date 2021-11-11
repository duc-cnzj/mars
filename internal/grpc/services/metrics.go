package services

import (
	"context"
	"fmt"
	"time"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/pkg/metrics"
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

func (m *Metrics) ProjectByID(request *metrics.ProjectByIDRequest, server metrics.Metrics_ProjectByIDServer) error {
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()
	defer mlog.Debug("ProjectByID exit")

	now := func() string {
		return time.Now().Format("15:04:05")
	}

	fn := func() error {
		podMetrics, err := app.K8sMetrics().MetricsV1beta1().PodMetricses(request.Namespace).Get(context.TODO(), request.Pod, metav1.GetOptions{})
		if err != nil {
			running, _ := utils.IsPodRunning(request.Namespace, request.Pod)
			if running {
				return nil
			}
			return err
		}

		cpu, memory := utils.GetCpuAndMemoryQuantity(*podMetrics)
		cpuM := cpu.MilliValue()
		var HumanizeCpu string = fmt.Sprintf("%v m", float64(cpu.MilliValue()))
		if cpuM > 1000 {
			HumanizeCpu = fmt.Sprintf("%.3f", float64(cpu.MilliValue())/1000)
		}
		asInt64, _ := memory.AsInt64()
		if err := server.Send(&metrics.ProjectByIDResponse{
			Cpu:            float64(cpu.MilliValue()),
			Memory:         float64(memory.ScaledValue(3)),
			HumanizeCpu:    HumanizeCpu,
			HumanizeMemory: humanize.Bytes(uint64(asInt64)),
			Time:           now(),
			Length:         int64(length),
		}); err != nil {
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
