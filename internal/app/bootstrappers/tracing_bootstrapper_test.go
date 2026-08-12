package bootstrappers

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	otlprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/mock/gomock"
)

func TestTracingBootstrapper_Bootstrap(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := app.NewMockApp(m)
	app.EXPECT().Config().Return(&config.Config{})
	app.EXPECT().Logger().Return(mlog.NewForConfig(nil))
	assert.Nil(t, (&TracingBootstrapper{}).Bootstrap(app))
}

func TestTracingBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{"trace"}, (&TracingBootstrapper{}).Tags())
}

func TestNewPropagator(t *testing.T) {
	assert.NotNil(t, newPropagator())
}

func TestTracingBootstrapper_Bootstrap_WithEndpoint(t *testing.T) {
	// setupOTelSDK 会写 otel 全局状态，测后恢复。
	oldProp := otel.GetTextMapPropagator()
	oldTP := otel.GetTracerProvider()
	oldMP := otel.GetMeterProvider()
	t.Cleanup(func() {
		otel.SetTextMapPropagator(oldProp)
		otel.SetTracerProvider(oldTP)
		otel.SetMeterProvider(oldMP)
	})

	m := gomock.NewController(t)
	defer m.Finish()
	app := app.NewMockApp(m)
	app.EXPECT().Config().Return(&config.Config{TracingEndpoint: "localhost:4317"})
	app.EXPECT().PrometheusRegistry().Return(prometheus2.NewRegistry())
	app.EXPECT().Logger().Return(mlog.NewForConfig(nil))
	// 注册后立即执行闭包，覆盖 after-shutdown 钩子体（真实 shutdown 时触发，shutdown 收敛错误不 panic）。
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).Do(func(f func()) { f() })
	assert.Nil(t, (&TracingBootstrapper{}).Bootstrap(app))
}

type spanExporterStub struct{}

func (spanExporterStub) ExportSpans(context.Context, []trace.ReadOnlySpan) error { return nil }
func (spanExporterStub) Shutdown(context.Context) error                          { return nil }

func TestNewTraceProvider(t *testing.T) {
	p := newTraceProvider(resource.Default(), spanExporterStub{})
	assert.NotNil(t, p)
	assert.NoError(t, p.Shutdown(context.Background()))
}

func TestNewMeterProvider(t *testing.T) {
	reg := prometheus2.NewRegistry()
	reader, err := otlprom.New(otlprom.WithRegisterer(reg))
	assert.NoError(t, err)
	mp := newMeterProvider(resource.Default(), reader)
	assert.NotNil(t, mp)
	assert.NoError(t, mp.Shutdown(context.Background()))
}

func TestSetupOTelSDK(t *testing.T) {
	// setupOTelSDK 会写 otel 全局 propagator/tracer/meter provider，测后必须恢复。
	oldProp := otel.GetTextMapPropagator()
	oldTP := otel.GetTracerProvider()
	oldMP := otel.GetMeterProvider()
	t.Cleanup(func() {
		otel.SetTextMapPropagator(oldProp)
		otel.SetTracerProvider(oldTP)
		otel.SetMeterProvider(oldMP)
	})

	shutdown := setupOTelSDK(context.Background(), "localhost:4317", prometheus2.NewRegistry())
	assert.NotNil(t, shutdown)
	// shutdown 收集 trace+meter provider 的清理；endpoint 不可达时 exporter 关闭可能报错，
	// 但 shutdown 用 errors.Join 收敛、不 panic。
	assert.NotPanics(t, func() { _ = shutdown(context.Background()) })
}

func TestMustNew_ReturnsValue(t *testing.T) {
	v := mustNew("ok", func() (int, error) { return 42, nil })
	assert.Equal(t, 42, v)
}

func TestMustNew_PanicsOnError(t *testing.T) {
	// 注入失败构造器，验证 bootstrap 阶段 fatal 快速失败语义。
	assert.PanicsWithError(t, "mars: setup boom failed: boom", func() {
		mustNew("boom", func() (int, error) { return 0, errors.New("boom") })
	})
}

func TestWarnOnShutdownError(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	calls := 0
	// 关闭无错：闭包正常执行，不 panic。
	ok := warnOnShutdownError(logger, func(context.Context) error { calls++; return nil })
	assert.NotPanics(t, func() { ok() })
	assert.Equal(t, 1, calls)
	// 关闭报错：走 Warning 告警分支，不 panic。
	fail := warnOnShutdownError(logger, func(context.Context) error { calls++; return errors.New("shutdown boom") })
	assert.NotPanics(t, func() { fail() })
	assert.Equal(t, 2, calls)
}
