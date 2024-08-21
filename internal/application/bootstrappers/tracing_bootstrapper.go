package bootstrappers

import (
	"context"
	"errors"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/version"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const serviceName = "mars"

type TracingBootstrapper struct{}

func (t *TracingBootstrapper) Tags() []string {
	return []string{"trace"}
}

func (t *TracingBootstrapper) Bootstrap(app application.App) error {
	config := app.Config()
	if config.TracingEndpoint == "" {
		app.Logger().Warning("TracingEndpoint is not set, skipping tracing bootstrapping")
		return nil
	}
	shutdownFuncs, err := setupOTelSDK(context.Background(), config.TracingEndpoint, app.PrometheusRegistry())
	if err != nil {
		return err
	}
	app.RegisterAfterShutdownFunc(func(app application.App) {
		if err := shutdownFuncs(context.TODO()); err != nil {
			app.Logger().Warning(err)
		}
	})
	return nil
}

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context, grpcEndpoint string, promReg *prometheus2.Registry) (shutdown func(ctx2 context.Context) error, err error) {
	v := version.GetVersion()
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersionKey.String(v.String()),
		),
	)
	if err != nil {
		panic(err)
	}

	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(grpcEndpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithReconnectionPeriod(time.Second*5),
	)
	if err != nil {
		panic(err)
	}

	// Set up trace provider.
	tracerProvider, err := newTraceProvider(r, exporter)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	e, err := prometheus.New(prometheus.WithRegisterer(promReg))
	meterProvider, err := newMeterProvider(r, e)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	//loggerProvider, err := newLoggerProvider()
	//if err != nil {
	//	handleErr(err)
	//	return
	//}
	//shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	//global.SetLoggerProvider(loggerProvider)

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(r *resource.Resource, exp sdktrace.SpanExporter) (*trace.TracerProvider, error) {
	traceProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(r),
		trace.WithBatcher(exp),
	)
	return traceProvider, nil
}

func newMeterProvider(res *resource.Resource, reader sdkmetric.Reader) (*sdkmetric.MeterProvider, error) {
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)
	return meterProvider, nil
}

//func newLoggerProvider() (*log.LoggerProvider, error) {
//logExporter, err := stdoutlog.New()
//if err != nil {
//	return nil, err
//}
//loggerProvider := log.NewLoggerProvider(
//	log.WithProcessor(log.NewBatchProcessor(logExporter)),
//)
//return loggerProvider, nil
//}
