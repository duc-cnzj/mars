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

	"github.com/duc-cnzj/mars/v5/internal/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

const system = "mars"

var appVersion = version.GetVersion().String()

var (
	BootstrapperStartMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem:   system,
		Name:        "bootstrapper_duration_seconds",
		Help:        "系统启动各阶段耗时",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"bootstrapper"})

	WebsocketConnectionsCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem:   system,
		Name:        "websocket_connections",
		Help:        "当前 websocket 连接数",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"username"})

	GrpcLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem:   system,
		Name:        "grpc_duration_seconds",
		Help:        "grpc 调用延迟",
		ConstLabels: prometheus.Labels{"version": appVersion},
		Buckets:     prometheus.ExponentialBuckets(0.01, 2, 17),
	}, []string{"method"})

	GrpcRequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   system,
		Name:        "grpc_request_total",
		Help:        "grpc 请求总数",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"method", "result"})
	GrpcRequestTotalFail    = GrpcRequestTotal.MustCurryWith(prometheus.Labels{"result": "fail"})
	GrpcRequestTotalSuccess = GrpcRequestTotal.MustCurryWith(prometheus.Labels{"result": "success"})

	GrpcErrorCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   system,
		Name:        "grpc_errors_total",
		Help:        "grpc 错误数量",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"method"})

	WebsocketRequestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem:   system,
		Name:        "websocket_request_duration_seconds",
		Help:        "websocket 调用延迟",
		ConstLabels: prometheus.Labels{"version": appVersion},
		Buckets:     prometheus.ExponentialBuckets(0.01, 2, 17),
	}, []string{"method"})

	WebsocketPanicCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   system,
		Name:        "websocket_request_panic_total",
		Help:        "websocket panic 错误数量",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"method"})

	WebsocketRequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   system,
		Name:        "websocket_request_total",
		Help:        "websocket 请求总数",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"method", "result"})
	WebsocketRequestTotalFail    = WebsocketRequestTotal.MustCurryWith(prometheus.Labels{"result": "panic"})
	WebsocketRequestTotalSuccess = WebsocketRequestTotal.MustCurryWith(prometheus.Labels{"result": "success"})

	CacheBytesGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem:   system,
		Name:        "cache_bytes",
		Help:        "cache bytes 统计",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"key"})

	CacheRememberDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem:   system,
		Name:        "cache_remember_duration_seconds",
		Help:        "cache Remember 调用时间",
		ConstLabels: prometheus.Labels{"version": appVersion},
		Buckets:     prometheus.ExponentialBuckets(0.01, 2, 17),
	}, []string{"key"})

	CronPanicCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   system,
		Name:        "cron_panic_total",
		Help:        "cron panic 错误数量",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"cron_name"})

	CronErrorCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   system,
		Name:        "cron_error_total",
		Help:        "cron error 错误数量",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"cron_name"})

	CronCommandCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem:   system,
		Name:        "cron_command_total",
		Help:        "cron command 总数",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"cron_name"})

	CronDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem:   system,
		Name:        "cron_duration_seconds",
		Help:        "cron 执行时间",
		ConstLabels: prometheus.Labels{"version": appVersion},
		Buckets:     prometheus.ExponentialBucketsRange(0.5, (1 * time.Hour).Seconds(), 20),
	}, []string{"cron_name"})

	K8sInformerFanOutListenerCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem:   system,
		Name:        "fanout_listener_count",
		Help:        "k8s fanout listener count",
		ConstLabels: prometheus.Labels{"version": appVersion},
	}, []string{"type"})
)

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
		WebsocketRequestLatency,
		WebsocketPanicCount,
		WebsocketRequestTotal,
		CacheBytesGauge,
		CacheRememberDuration,
		CronPanicCount,
		CronDuration,
		CronCommandCount,
		CronErrorCount,
		K8sInformerFanOutListenerCount,
	)

	return registry
}
