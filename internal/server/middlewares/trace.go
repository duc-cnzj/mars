package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/opentracing/opentracing-go/ext"
	"go.opentelemetry.io/otel/trace/noop"

	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	trace2 "go.opentelemetry.io/otel/trace"
)

var grpcGatewayTag = attribute.String("component", "grpc-gateway")

// TracingWrapper
// [W3C Tracing Headers](https://www.w3.org/TR/trace-context/)
func TracingWrapper(logger mlog.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx  context.Context
			span trace2.Span
		)

		ctx, span = noop.NewTracerProvider().Tracer("").Start(r.Context(), "")
		url := r.URL.String()
		if !tracingIgnoreFn(r.URL.Path) {
			ctxt := propagation.TraceContext{}
			ctx, span = otel.Tracer("mars").Start(ctxt.Extract(r.Context(), propagation.HeaderCarrier(r.Header)), fmt.Sprintf("[%s]: %s", r.Method, url))
			span.SetAttributes(grpcGatewayTag)
			span.SetAttributes(attribute.String(string(ext.SpanKind), "grpc-gateway"))
			span.SetAttributes(attribute.String("url", url))
			span.SetAttributes(attribute.String("method", r.Method))
			span.SetAttributes(attribute.String("user-agent", r.UserAgent()))
		}

		defer func() {
			pattern := GetPatternHeader(w)
			if pattern != "" {
				span.SetName(pattern)
			}

			span.End()
		}()
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func tracingIgnoreFn(fullMethodName string) bool {
	if fullMethodName == "/ws" {
		return true
	}

	if !strings.HasPrefix(fullMethodName, "/api") {
		return true
	}

	// /api/metrics/namespace/{namespace}/pods/{pod}/stream
	if strings.HasPrefix(fullMethodName, "/api/metrics/namespace/") && strings.HasSuffix(fullMethodName, "/stream") {
		return true
	}

	// /api/containers/namespaces/{namespace}/pods/{pod}/containers/{container}/stream_logs
	if strings.HasPrefix(fullMethodName, "/api/containers/namespaces/") && strings.HasSuffix(fullMethodName, "/stream_logs") {
		return true
	}

	return false
}
