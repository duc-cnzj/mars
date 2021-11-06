package utils

import (
	"context"

	"github.com/opentracing/opentracing-go"
)

func TracingChildOfFn(ctx context.Context, name string, fn func()) {
	defer func(span opentracing.Span) {
		span.Finish()
	}(opentracing.GlobalTracer().StartSpan("[Tracer]: "+name, opentracing.ChildOf(opentracing.SpanFromContext(ctx).Context())))

	fn()
}

func TracingFromCtx(ctx context.Context, name string) func() {
	span := opentracing.GlobalTracer().StartSpan("[Tracer]: "+name, opentracing.ChildOf(opentracing.SpanFromContext(ctx).Context()))

	return func() { span.Finish() }
}
