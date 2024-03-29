package bootstrappers

import (
	"context"
	"net"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/version"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const serviceName = "mars"

type TracingBootstrapper struct{}

func (t *TracingBootstrapper) Tags() []string {
	return []string{"trace"}
}

func (t *TracingBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	cfg := app.Config()
	if cfg.JaegerAgentHostPort != "" {
		host, port, err := net.SplitHostPort(cfg.JaegerAgentHostPort)
		if err != nil {
			return err
		}
		jaeexp, err := newJaegerExporter(host, port)
		if err != nil {
			return err
		}
		opts := []trace.TracerProviderOption{
			trace.WithBatcher(jaeexp),
			trace.WithResource(newResource()),
		}
		if !app.IsDebug() {
			// [采样器参考](https://github.com/open-telemetry/docs-cn/blob/main/specification/trace/sdk.md)
			opts = append(opts, trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(0.3))))
		}
		tp := trace.NewTracerProvider(opts...)
		otel.SetTracerProvider(tp)
		app.RegisterAfterShutdownFunc(func(app contracts.ApplicationInterface) {
			mlog.Info("shutdown tracer")
			timeout, cancelFunc := context.WithTimeout(context.TODO(), 3*time.Second)
			defer cancelFunc()
			if err := tp.Shutdown(timeout); err != nil {
				mlog.Error(err)
			}
		})
	}
	otel.SetErrorHandler(&errorHandler{})
	tracer := otel.Tracer("mars")
	app.SetTracer(tracer)

	return nil
}

type errorHandler struct{}

func (e *errorHandler) Handle(err error) {
	mlog.Warning(err)
}

func newResource() *resource.Resource {
	v := version.GetVersion()
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String(v.String()),
		attribute.String("system.build_date", v.BuildDate),
		attribute.String("system.git_commit", v.GitCommit),
		attribute.String("system.git_branch", v.GitBranch),
		attribute.String("system.go_version", v.GoVersion),
		attribute.String("system.platform", v.Platform),
	)
}

func newJaegerExporter(host, port string) (trace.SpanExporter, error) {
	return jaeger.New(
		jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(host),
			jaeger.WithAgentPort(port),
		),
	)
}
