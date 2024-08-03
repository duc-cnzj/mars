package bootstrappers

import "github.com/duc-cnzj/mars/v4/internal/application"

const serviceName = "mars"

type TracingBootstrapper struct{}

func (t *TracingBootstrapper) Tags() []string {
	return []string{"trace"}
}

func (t *TracingBootstrapper) Bootstrap(appli application.App) error {
	//cfg := appli.Config()
	//if cfg.JaegerAgentHostPort != "" {
	//	host, port, err := net.SplitHostPort(cfg.JaegerAgentHostPort)
	//	if err != nil {
	//		return err
	//	}
	//	jaeexp, err := newJaegerExporter(host, port)
	//	if err != nil {
	//		return err
	//	}
	//	opts := []trace.TracerProviderOption{
	//		trace.WithBatcher(jaeexp),
	//		trace.WithResource(newResource()),
	//	}
	//	if !appli.IsDebug() {
	//		// [采样器参考](https://github.com/open-telemetry/docs-cn/blob/main/specification/trace/sdk.md)
	//		opts = append(opts, trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(0.3))))
	//	}
	//	tp := trace.NewTracerProvider(opts...)
	//	otel.SetTracerProvider(tp)
	//	appli.RegisterAfterShutdownFunc(func(app app.App) {
	//		appli.Logger().Info("shutdown tracer")
	//		timeout, cancelFunc := context.WithTimeout(context.TODO(), 3*time.Second)
	//		defer cancelFunc()
	//		if err := tp.Shutdown(timeout); err != nil {
	//			appli.Logger().Error(err)
	//		}
	//	})
	//}
	//otel.SetErrorHandler(&errorHandler{})
	//tracer := otel.Tracer("mars")

	return nil
}

//
//type errorHandler struct{}
//
//func (e *errorHandler) Handle(err error) {
//	mlog.Warning(err)
//}
//
//func newResource() *resource.Resource {
//	v := version.GetVersion()
//	return resource.NewWithAttributes(
//		semconv.SchemaURL,
//		semconv.ServiceNameKey.String(serviceName),
//		semconv.ServiceVersionKey.String(v.String()),
//		attribute.String("system.build_date", v.BuildDate),
//		attribute.String("system.git_commit", v.GitCommit),
//		attribute.String("system.git_branch", v.GitBranch),
//		attribute.String("system.go_version", v.GoVersion),
//		attribute.String("system.platform", v.Platform),
//	)
//}
//
//func newJaegerExporter(host, port string) (trace.SpanExporter, error) {
//	return jaeger.New(
//		jaeger.WithAgentEndpoint(
//			jaeger.WithAgentHost(host),
//			jaeger.WithAgentPort(port),
//		),
//	)
//}
