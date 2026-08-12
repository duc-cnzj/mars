package biz

import (
	"context"
	"errors"
	"fmt"

	"github.com/duc-cnzj/mars/v6/internal/errs"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ErrPodMetricsUnavailable 标识"pod 仍在运行但本次采样失败"的确定性状态。
// 与 NotFound（pod 已消失）区分：流式场景跳过本轮继续采样，一次性场景原样上抛。
// 用哨兵错误 + errors.Is 判定，把 TopPod/StreamTopPod 的"运行中采样失败"分类决策收敛到 biz。
var ErrPodMetricsUnavailable = errors.New("pod metrics unavailable")

// PodSample 是单次 pod 资源采样的归一化结果，承载 CPU/内存量。
// 由 MetricsBiz.PodSample 产出，transport 只负责把它渲染成响应。
type PodSample struct {
	Cpu    *resource.Quantity
	Memory *resource.Quantity
}

// MetricsBiz 收口 pod 资源采样的业务规则：加载 metrics + 按 pod 运行态分类失败。
// transport（metricsSvc）只感知三个语义——成功拿到采样、pod 消失（NotFound）、
// 运行中暂不可采样（ErrPodMetricsUnavailable），不再散落 IsPodRunning 判定。
type MetricsBiz interface {
	// PodSample 采样指定 pod 的 CPU/内存用量并按运行态分类失败。
	PodSample(ctx context.Context, namespace, pod string) (*PodSample, error)
}

// metricsBiz 是 MetricsBiz 的默认实现，依赖 K8sBiz 加载 metrics 与 pod 运行态。
type metricsBiz struct {
	k8sBiz K8sBiz
}

// NewMetricsBiz 构造 pod 资源采样业务对象，依赖 K8sBiz 提供 metrics 与运行态判定。
func NewMetricsBiz(k8sBiz K8sBiz) MetricsBiz {
	return &metricsBiz{k8sBiz: k8sBiz}
}

// PodSample 采样指定 pod 的 CPU/内存用量，并按失败原因分类：
//   - 采样失败且 pod 已不在运行 → errs.NotFound(reason)，与"资源不存在"语义一致；
//   - 采样失败但 pod 仍在运行 → 包装 ErrPodMetricsUnavailable（保留底层错误细节），
//     由调用方决定跳过本轮（流式）还是上抛（一次性）；
//   - 采样成功 → 返回归一化后的 PodSample。
func (b *metricsBiz) PodSample(ctx context.Context, namespace, pod string) (*PodSample, error) {
	podMetrics, err := b.k8sBiz.GetPodMetrics(ctx, namespace, pod)
	if err != nil {
		running, reason := b.k8sBiz.IsPodRunning(namespace, pod)
		if !running {
			return nil, errs.NotFound(reason)
		}
		return nil, fmt.Errorf("%w: %v", ErrPodMetricsUnavailable, err)
	}
	cpu, memory := b.k8sBiz.GetCpuAndMemoryQuantity(*podMetrics)
	return &PodSample{Cpu: cpu, Memory: memory}, nil
}
