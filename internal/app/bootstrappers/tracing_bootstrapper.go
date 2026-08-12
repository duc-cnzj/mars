package bootstrappers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/version"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// serviceName 是上报给 collector 的 OpenTelemetry 服务名。
const serviceName = "mars"

// TracingBootstrapper 启动 opentelemetry 链路追踪流水线。
type TracingBootstrapper struct{}

// Tags 实现 Bootstrapper 接口的 Tags。
func (t *TracingBootstrapper) Tags() []string {
	return []string{"trace"}
}

// Bootstrap 实现 Bootstrapper 接口的 Bootstrap。
func (t *TracingBootstrapper) Bootstrap(deps app.BootstrapDeps) error {
	config := deps.Config()
	if config.TracingEndpoint == "" {
		deps.Logger().Warning("TracingEndpoint is not set, skipping tracing bootstrapping")
		return nil
	}
	shutdownFuncs := setupOTelSDK(context.TODO(), config.TracingEndpoint, deps.PrometheusRegistry())
	deps.RegisterAfterShutdownFunc(warnOnShutdownError(deps.Logger(), shutdownFuncs))
	return nil
}

// warnOnShutdownError 包一层 after-shutdown 钩子：provider 关闭报错只告警不 panic。
// 独立成可测试 helper，避免闭包内分支不可测。
func warnOnShutdownError(logger mlog.Logger, fn func(context.Context) error) func() {
	return func() {
		if err := fn(context.TODO()); err != nil {
			logger.Warning(err)
		}
	}
}

// setupOTelSDK 启动 OpenTelemetry 流水线。
// 返回的 shutdown 负责关闭 trace+meter provider；不可恢复的配置错误直接 panic（bootstrap 阶段应快速失败）。
func setupOTelSDK(ctx context.Context, grpcEndpoint string, promReg *prometheus2.Registry) (shutdown func(context.Context) error) {
	v := version.GetVersion()
	r := mustNew("opentelemetry resource", func() (*resource.Resource, error) {
		def := resource.Default()
		// 用 Default() 自身的 SchemaURL 建 service resource，避免 otel 升级后
		// semconv 与 sdk 版本 schema URL 不一致导致 Merge 冲突 panic。
		return resource.Merge(
			def,
			resource.NewWithAttributes(
				def.SchemaURL(),
				semconv.ServiceName(serviceName),
				semconv.ServiceVersionKey.String(v.String()),
			),
		)
	})

	var shutdownFuncs []func(context.Context) error

	// shutdown 调用通过 shutdownFuncs 注册的清理函数。
	// 各调用的错误被合并。
	// 每个注册的清理函数只被调用一次。
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// 设置 propagator。
	otel.SetTextMapPropagator(newPropagator())

	exporter := mustNew("otlp trace exporter", func() (trace.SpanExporter, error) {
		return otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithEndpoint(grpcEndpoint),
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithReconnectionPeriod(time.Second*5),
		)
	})

	// 设置 trace provider。
	tracerProvider := newTraceProvider(r, exporter)
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	e := mustNew("prometheus exporter", func() (sdkmetric.Reader, error) {
		return prometheus.New(prometheus.WithRegisterer(promReg))
	})
	meterProvider := newMeterProvider(r, e)
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	return
}

// mustNew 构造期 fatal 守卫：bootstrap 阶段遇到不可恢复的配置错误直接 panic 快速失败。
// 与 template.Must / regexp.MustCompile 同款惯用法，panic 分支可通过注入失败构造器测试。
func mustNew[T any](name string, fn func() (T, error)) T {
	v, err := fn()
	if err != nil {
		panic(fmt.Errorf("mars: setup %s failed: %w", name, err))
	}
	return v
}

// newPropagator 返回携带 trace + baggage 上下文的复合 text map propagator。
func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// newTraceProvider 构建由给定 exporter 支撑的全采样 tracer provider。
func newTraceProvider(r *resource.Resource, exp trace.SpanExporter) *trace.TracerProvider {
	return trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(r),
		trace.WithBatcher(exp),
	)
}

// newMeterProvider 构建经给定 reader 读取指标的 meter provider。
func newMeterProvider(res *resource.Resource, reader sdkmetric.Reader) *sdkmetric.MeterProvider {
	return sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)
}
