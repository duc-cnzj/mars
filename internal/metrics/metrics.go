/*
Package metrics

	Google SRE 定义了四个需要监控的关键指标。延迟（Latency），流量（Traffic），错误（Errors）和饱和度（Saturation）。
	正如google sre 所讨论的，如果您只能衡量服务的四个指标，请关注这四个指标。

	延迟 Latency
	> 延迟是服务处理传入请求和发送响应所用时间的度量。测量服务延迟有助于及早发现服务的缓慢。

	流量 Traffic
	> 流量可以更好地理解服务需求。通常称为服务 QPS（每秒查询数），流量是服务请求量的度量。此信号可帮助您决定何时需要扩大服务规模以应对不断增长的客户需求，或缩小服务规模以提高成本效益。

	错误 Errors
	> 错误是对客户端请求失败的度量。这些故障可以根据响应代码（HTTP 5XX 错误）轻松识别。
	> 在某些情况下，由于错误的结果数据或违反了约定，响应被认为是错误的。例如，您可能会收到HTTP 200 响应，但返回的数据不完整，或者响应时间超出了约定的 SLA。因此，除了响应码之外，可能还需要其他机制（代码逻辑）来捕获错误。

	饱和度 Saturation
	> 饱和度是服务器资源利用率的度量。这个信号告诉你服务资源的状态以及它们有多“满”。
	> 这些资源包括内存、cpu、网络 I/O 等。在资源利用率达到 100% 之前，服务性能也会缓慢下降。因此，有一个利用率目标很重要。延迟的增加是饱和度的一个很好的指标；测量延迟99线 有助于及早发现饱和度。
*/
package metrics

import (
	"time"

	"github.com/duc-cnzj/mars/v6/internal/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

const system = "mars"

var appVersion = version.GetVersion().String()

var (
	// BootstrapperStartMetrics 记录系统启动各阶段的耗时，标签 bootstrapper 为阶段名。
	BootstrapperStartMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem:   system,
		Name:        "bootstrapper_duration_seconds",
		Help:        "系统启动各阶段耗时",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"bootstrapper"})

	// WebsocketConnectionsCount 记录当前 websocket 连接数，标签 username 为连接用户名。
	WebsocketConnectionsCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem:   system,
		Name:        "websocket_connections",
		Help:        "当前 websocket 连接数",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"username"})

	// GrpcLatency 记录 grpc 调用延迟直方图，标签 method 为完整方法名。
	GrpcLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem:   system,
		Name:        "grpc_duration_seconds",
		Help:        "grpc 调用延迟",
		ConstLabels: prometheus.Labels{"version": appVersion},
		Buckets:     prometheus.ExponentialBuckets(0.01, 2, 17),
	}, []string{"method"})

	// GrpcRequestTotal 记录 grpc 请求总数，result 维度区分 fail/success。
	GrpcRequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   system,
		Name:        "grpc_request_total",
		Help:        "grpc 请求总数",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"method", "result"})
	// GrpcRequestTotalFail 是 GrpcRequestTotal 预 curry result=fail 的便捷句柄。
	GrpcRequestTotalFail = GrpcRequestTotal.MustCurryWith(prometheus.Labels{"result": "fail"})
	// GrpcRequestTotalSuccess 是 GrpcRequestTotal 预 curry result=success 的便捷句柄。
	GrpcRequestTotalSuccess = GrpcRequestTotal.MustCurryWith(prometheus.Labels{"result": "success"})

	// GrpcErrorCount 记录 grpc 错误数量。
	GrpcErrorCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   system,
		Name:        "grpc_errors_total",
		Help:        "grpc 错误数量",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"method"})

	// WebsocketRequestLatency 记录 websocket 调用延迟直方图。
	WebsocketRequestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem:   system,
		Name:        "websocket_request_duration_seconds",
		Help:        "websocket 调用延迟",
		ConstLabels: prometheus.Labels{"version": appVersion},
		Buckets:     prometheus.ExponentialBuckets(0.01, 2, 17),
	}, []string{"method"})

	// WebsocketPanicCount 记录 websocket 请求 panic 错误数量。
	WebsocketPanicCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   system,
		Name:        "websocket_request_panic_total",
		Help:        "websocket panic 错误数量",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"method"})

	// WebsocketRequestTotal 记录 websocket 请求总数，result 维度区分 panic/success。
	// 与 grpc 侧 result=fail 不同：websocket 只把 recover 到的 panic 视为失败
	// （见 internal/services/websocket/controller.go 的 dispatchEvent），
	// 因此 Fail 用 "panic" 标签精确描述其触发时机，这是有意的语义差异。
	WebsocketRequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   system,
		Name:        "websocket_request_total",
		Help:        "websocket 请求总数",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"method", "result"})
	// WebsocketRequestTotalFail 是 WebsocketRequestTotal 预 curry result=panic 的便捷句柄。
	WebsocketRequestTotalFail = WebsocketRequestTotal.MustCurryWith(prometheus.Labels{"result": "panic"})
	// WebsocketRequestTotalSuccess 是 WebsocketRequestTotal 预 curry result=success 的便捷句柄。
	WebsocketRequestTotalSuccess = WebsocketRequestTotal.MustCurryWith(prometheus.Labels{"result": "success"})

	// CacheBytesGauge 记录 cache 字节数统计。
	CacheBytesGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem:   system,
		Name:        "cache_bytes",
		Help:        "cache bytes 统计",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"key"})

	// CacheRememberDuration 记录 cache Remember 调用耗时。
	CacheRememberDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem:   system,
		Name:        "cache_remember_duration_seconds",
		Help:        "cache Remember 调用时间",
		ConstLabels: prometheus.Labels{"version": appVersion},
		Buckets:     prometheus.ExponentialBuckets(0.01, 2, 17),
	}, []string{"key"})

	// CronPanicCount 记录 cron 任务 panic 错误数量。
	CronPanicCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   system,
		Name:        "cron_panic_total",
		Help:        "cron panic 错误数量",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"cron_name"})

	// CronErrorCount 记录 cron 任务错误数量。
	CronErrorCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   system,
		Name:        "cron_error_total",
		Help:        "cron error 错误数量",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"cron_name"})

	// CronCommandCount 记录 cron 命令执行总数。
	CronCommandCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   system,
		Name:        "cron_command_total",
		Help:        "cron command 总数",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"cron_name"})

	// CronDuration 记录 cron 任务执行时长直方图。
	CronDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem:   system,
		Name:        "cron_duration_seconds",
		Help:        "cron 执行时间",
		ConstLabels: prometheus.Labels{"version": appVersion},
		Buckets:     prometheus.ExponentialBucketsRange(0.5, (1 * time.Hour).Seconds(), 20),
	}, []string{"cron_name"})

	// K8sInformerFanOutListenerCount 记录 k8s fanout 各类型监听者数量。
	K8sInformerFanOutListenerCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem:   system,
		Name:        "fanout_listener_count",
		Help:        "k8s fanout 各类型监听者数量",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"type"})

	// FanOutChannelLength 记录各 fanout 通道当前积压长度。
	FanOutChannelLength = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem:   system,
		Name:        "fanout_channel_length",
		Help:        "fanout 通道当前积压长度",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"name"})

	// AIUsageEventTotal 记录前端上报的 AI 使用漏斗事件（入口点击、首 token、
	// 对话完成/失败）。event 为服务端白名单枚举，刻意不带用户维度，
	// 避免指标基数随用户数膨胀。
	AIUsageEventTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   system,
		Name:        "ai_usage_event_total",
		Help:        "AI 使用漏斗事件数",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"event"})

	// AIChatRequestsTotal 记录服务端 AI 对话请求数（一轮 chat 或 confirm 计一次），
	// result 区分正常结束（含暂停等待确认）与失败。
	AIChatRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   system,
		Name:        "ai_chat_requests_total",
		Help:        "服务端 AI 对话请求数",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"result"})

	// AIChatFirstTokenSeconds 记录从请求开始到模型首段文本的耗时，
	// 直接反映用户感知的「多久开始出字」。
	AIChatFirstTokenSeconds = prometheus.NewHistogram(prometheus.HistogramOpts{
		Subsystem:   system,
		Name:        "ai_chat_first_token_seconds",
		Help:        "AI 对话首 token 耗时（秒）",
		ConstLabels: prometheus.Labels{"version": appVersion},
		Buckets:     []float64{0.1, 0.3, 0.5, 1, 2, 5, 10, 20, 30, 60},
	})

	// AIChatToolCallsTotal 记录对话中执行的工具调用数，tool 为工具名、result 区分
	// 成功与业务失败（isError）。
	AIChatToolCallsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   system,
		Name:        "ai_chat_tool_calls_total",
		Help:        "服务端 AI 对话工具调用数",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"tool", "result"})
)

// AI 对话指标的结果标签值。
const (
	// AIChatResultOK 表示本轮对话正常结束或按预期暂停等待确认。
	AIChatResultOK = "ok"
	// AIChatResultError 表示上游、工具协议或轮数上限导致的失败。
	AIChatResultError = "error"
)

// NewRegistry 构造并注册全部应用指标 + Go/进程运行时收集器，返回可用的 prometheus.Registry。
func NewRegistry() *prometheus.Registry {
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		WebsocketConnectionsCount,
		BootstrapperStartMetrics,
		GrpcLatency,
		GrpcErrorCount,
		GrpcRequestTotal,
		FanOutChannelLength,
		WebsocketRequestLatency,
		WebsocketPanicCount,
		WebsocketRequestTotal,
		CacheBytesGauge,
		CacheRememberDuration,
		CronPanicCount,
		CronDuration,
		CronCommandCount,
		CronErrorCount,
		AIUsageEventTotal,
		AIChatRequestsTotal,
		AIChatFirstTokenSeconds,
		AIChatToolCallsTotal,
		K8sInformerFanOutListenerCount,
	)

	return registry
}
