package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var grpcGatewayTag = opentracing.Tag{Key: string(ext.Component), Value: "grpc-gateway"}

func TracingWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.String()
		if !TracingIgnoreFn(context.TODO(), r.URL.Path) {
			parentSpanContext, err := opentracing.GlobalTracer().Extract(
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(r.Header))
			if err == nil || err == opentracing.ErrSpanContextNotFound {
				serverSpan := opentracing.GlobalTracer().StartSpan(
					url,
					ext.RPCServerOption(parentSpanContext),
					grpcGatewayTag,
					opentracing.Tags{"url": url},
				)
				r = r.WithContext(opentracing.ContextWithSpan(r.Context(), serverSpan))
				defer serverSpan.Finish()
			}
		}
		h.ServeHTTP(w, r)
	})
}

func TracingIgnoreFn(ctx context.Context, fullMethodName string) bool {
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
