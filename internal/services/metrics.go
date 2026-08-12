package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/metrics"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/dustin/go-humanize"
)

var _ metrics.MetricsServer = (*metricsSvc)(nil)

// metricsSvc 是 metrics.MetricsServer 的 gRPC 实现：提供 Pod 资源占用排行、
// 命名空间/项目维度 CPU 内存统计与流式刷新，由 NewMetricsSvc 构造。
type metricsSvc struct {
	metrics.UnimplementedMetricsServer
	k8sBiz     biz.K8sBiz
	metricsBiz biz.MetricsBiz
	logger     mlog.Logger
	timer      timer.Timer
	accessBiz  biz.AccessBiz
}

// MetricsSvcDeps 收口 NewMetricsSvc 的构造依赖，由 wire 按字段注入。
// 字段名即依赖含义：新增/删除依赖只改这一处，不再改函数签名。
type MetricsSvcDeps struct {
	Timer      timer.Timer
	K8sBiz     biz.K8sBiz
	MetricsBiz biz.MetricsBiz
	Logger     mlog.Logger
	AccessBiz  biz.AccessBiz
}

// NewMetricsSvc 收口 metrics 服务的构造依赖，由 wire 按字段注入。
func NewMetricsSvc(deps MetricsSvcDeps) metrics.MetricsServer {
	logger := deps.Logger.WithModule("services/metrics")
	return &metricsSvc{
		k8sBiz:     deps.K8sBiz,
		metricsBiz: deps.MetricsBiz,
		logger:     logger,
		timer:      deps.Timer,
		accessBiz:  deps.AccessBiz,
	}
}

// tickDuration 是 StreamTopPod 的采样间隔。包级 var 而非 deps 字段：
// wire.Struct(new(MetricsSvcDeps), "*") 无法注入裸 time.Duration，测试通过覆盖该 var
// 注入小值以真实执行 ticker.C 周期路径（默认 5s 在测试超时窗口内永远不触发）。
var (
	tickDuration = 5 * time.Second
	timeSpan     = 5 * time.Second * 30
	length       = timeSpan / tickDuration
)

// TopPod 返回指定 pod 的实时 CPU/内存资源用量，响应前做命名空间级访问控制。
func (m *metricsSvc) TopPod(ctx context.Context, request *metrics.TopPodRequest) (*metrics.TopPodResponse, error) {
	// 与 CpuMemoryInProject/Namespace 对齐：原始 k8s namespace 名必须过访问控制，
	// 否则任意登录用户可枚举任意命名空间内 pod 的资源用量。
	if _, err := m.accessBiz.RequireNamespaceAccessByName(ctx, request.Namespace); err != nil {
		return nil, logError(ctx, m.logger, err)
	}
	sample, err := m.metricsBiz.PodSample(ctx, request.Namespace, request.Pod)
	if err != nil {
		// 分类（pod 消失 → NotFound、运行中采样失败 → Unavailable）已在 biz 收敛，
		// 这里不再重复 IsPodRunning 判定；运行中采样失败属可观测性事件，打日志留痕。
		if errors.Is(err, biz.ErrPodMetricsUnavailable) {
			m.logger.ErrorCtx(ctx, err)
		}
		return nil, err
	}

	return m.buildTopPodResponse(sample), nil
}

// StreamTopPod 周期推送指定 pod 的资源用量：先做命名空间级访问控制，
// 随后按 tickDuration 间隔持续下发，直到客户端断开。
func (m *metricsSvc) StreamTopPod(request *metrics.TopPodRequest, server metrics.Metrics_StreamTopPodServer) error {
	// 与 TopPod 一致：流式场景同样要求命名空间访问控制，防止未授权用户订阅任意 pod 资源用量。
	if _, err := m.accessBiz.RequireNamespaceAccessByName(server.Context(), request.Namespace); err != nil {
		return logError(server.Context(), m.logger, err)
	}
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()
	defer m.logger.DebugCtxf(server.Context(), "StreamTopPod exit")

	fn := func() error {
		sample, err := m.metricsBiz.PodSample(server.Context(), request.Namespace, request.Pod)
		if err != nil {
			// 分类决策在 biz：pod 仍在运行但采样失败 → 跳过本轮，下一轮 tick 继续采样（不打断流）；
			// pod 已消失 → NotFound，终止流并携带具体原因，与 TopPod 的错误语义一致。
			if errors.Is(err, biz.ErrPodMetricsUnavailable) {
				return nil
			}
			return err
		}

		if err := server.Send(m.buildTopPodResponse(sample)); err != nil {
			// logError 带上 stream 的 ctx（含请求元数据），与全项目日志标准一致。
			return logError(server.Context(), m.logger, err)
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
		}
	}
}

// CpuMemoryInProject 返回指定项目全部 pod 的 CPU/内存聚合用量，响应前做项目级访问控制。
func (m *metricsSvc) CpuMemoryInProject(ctx context.Context, request *metrics.CpuMemoryInProjectRequest) (*metrics.CpuMemoryInProjectResponse, error) {
	p, err := m.accessBiz.RequireProjectAccess(ctx, int(request.ProjectId))
	if err != nil {
		return nil, logError(ctx, m.logger, err)
	}
	cpu, memory := biz.ProjectCpuMemory(ctx, m.k8sBiz, p)

	return &metrics.CpuMemoryInProjectResponse{
		Cpu:    cpu,
		Memory: memory,
	}, nil
}

// CpuMemoryInNamespace 返回指定命名空间下全部 pod 的 CPU/内存聚合用量，
// 响应前做命名空间级访问控制。
func (m *metricsSvc) CpuMemoryInNamespace(ctx context.Context, request *metrics.CpuMemoryInNamespaceRequest) (*metrics.CpuMemoryInNamespaceResponse, error) {
	// 与 CpuMemoryInProject 一致：私有命名空间的资源用量只对 admin/创建者/成员可见。
	ns, nerr := m.accessBiz.RequireNamespaceAccessByID(ctx, int(request.NamespaceId))
	if nerr != nil {
		return nil, logError(ctx, m.logger, nerr)
	}

	cpu, memory := m.k8sBiz.GetCpuAndMemoryInNamespace(ctx, ns.Name)

	return &metrics.CpuMemoryInNamespaceResponse{
		Cpu:    cpu,
		Memory: memory,
	}, nil
}

// buildTopPodResponse 把 biz 单次 pod 资源采样渲染为 TopPodResponse。
// 采样与失败分类已下沉 biz.MetricsBiz，这里只做纯渲染（含单位归一化）。
func (m *metricsSvc) buildTopPodResponse(sample *biz.PodSample) *metrics.TopPodResponse {
	cpu, memory := sample.Cpu, sample.Memory
	// 统一以毫核 + " m" 单位输出。之前 <1000m 带单位（"500 m"）、≥1000m 输出裸数值（"1.500"），
	// 单位语义随数值漂移，UI 无法可靠区分两个量级。
	humanizeCPU := fmt.Sprintf("%d m", cpu.MilliValue())
	asInt64, ok := memory.AsInt64()
	if !ok {
		m.logger.Error("HumanizeMemory 计算失败: 数值无法转为 int64")
	}

	return &metrics.TopPodResponse{
		Cpu:            float64(cpu.MilliValue()),
		Memory:         float64(memory.ScaledValue(3)),
		HumanizeCpu:    humanizeCPU,
		HumanizeMemory: humanize.Bytes(uint64(asInt64)),
		Time:           m.timer.Now().Format("15:04:05"),
		Length:         int32(length),
	}
}
