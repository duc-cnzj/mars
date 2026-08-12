package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestMetricsBizPodSample(t *testing.T) {
	ctx := context.TODO()

	t.Run("success returns normalized sample", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		k8s := NewMockK8sBiz(m)
		k8s.EXPECT().GetPodMetrics(ctx, "ns", "pod").Return(&v1beta1.PodMetrics{}, nil)
		cpu := resource.NewMilliQuantity(100, resource.DecimalSI)
		mem := resource.NewQuantity(1024, resource.BinarySI)
		k8s.EXPECT().GetCpuAndMemoryQuantity(gomock.Any()).Return(cpu, mem)

		sample, err := NewMetricsBiz(k8s).PodSample(ctx, "ns", "pod")
		assert.NoError(t, err)
		assert.Equal(t, cpu, sample.Cpu)
		assert.Equal(t, mem, sample.Memory)
	})

	t.Run("pod gone maps to NotFound with reason", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		k8s := NewMockK8sBiz(m)
		k8s.EXPECT().GetPodMetrics(ctx, "ns", "pod").Return(nil, errors.New("metrics endpoint down"))
		k8s.EXPECT().IsPodRunning("ns", "pod").Return(false, "pod terminated")

		_, err := NewMetricsBiz(k8s).PodSample(ctx, "ns", "pod")
		assert.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
		assert.Equal(t, "pod terminated", status.Convert(err).Message())
	})

	t.Run("running but sampling failed maps to ErrPodMetricsUnavailable", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		k8s := NewMockK8sBiz(m)
		underlying := errors.New("metrics-api timeout")
		k8s.EXPECT().GetPodMetrics(ctx, "ns", "pod").Return(nil, underlying)
		k8s.EXPECT().IsPodRunning("ns", "pod").Return(true, "")

		_, err := NewMetricsBiz(k8s).PodSample(ctx, "ns", "pod")
		assert.ErrorIs(t, err, ErrPodMetricsUnavailable)
		// 保留底层错误细节，transport 侧 ErrorCtx 可看到采样失败的具体原因。
		assert.Contains(t, err.Error(), "metrics-api timeout")
	})
}
