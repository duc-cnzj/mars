package data

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// newEndedSpan 构造一个由 recorder 捕获的已结束 span 及其 endSpan 闭包，供断言。
// 用本地 TracerProvider 而非全局 otel.SetTracerProvider——规避 otel global 的
// setDelegate 只 swap 一次、多次设置会钉死首 provider 的测试间脆弱性。
func newEndedSpan() (*tracetest.SpanRecorder, func(error)) {
	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	_, span := tp.Tracer("test").Start(context.Background(), "x")
	return rec, func(err error) { endSpan(span, err) }
}

// TestEndSpan_NoError 验证 endSpan 对 nil 错误：span 已结束、状态保持 Unset、无异常事件。
func TestEndSpan_NoError(t *testing.T) {
	rec, end := newEndedSpan()
	end(nil)
	ended := rec.Ended()
	assert.Len(t, ended, 1)
	assert.Equal(t, codes.Unset, ended[0].Status().Code)
	assert.Empty(t, ended[0].Events())
	assert.False(t, ended[0].EndTime().IsZero())
}

// TestEndSpan_Error 验证 endSpan 对非 nil 错误：span 置 Error 状态、落一条 exception 事件。
func TestEndSpan_Error(t *testing.T) {
	rec, end := newEndedSpan()
	end(errors.New("boom"))
	ended := rec.Ended()
	assert.Len(t, ended, 1)
	assert.Equal(t, codes.Error, ended[0].Status().Code)
	assert.Equal(t, "boom", ended[0].Status().Description)
	events := ended[0].Events()
	assert.Len(t, events, 1)
	assert.Equal(t, semconv.ExceptionEventName, events[0].Name)
}

// TestTracerName 固化 tracer 命名约定，防止误改后失去 instrumentation scope 区分。
func TestTracerName(t *testing.T) {
	assert.Equal(t, "github.com/duc-cnzj/mars/v6/internal/data", tracerName)
}
