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
		if !TracingIgnoreFn(context.TODO(), url) {
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
	return !strings.HasPrefix(fullMethodName, "/api")
}
