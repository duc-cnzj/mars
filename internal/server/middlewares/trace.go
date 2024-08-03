package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/duc-cnzj/mars/v4/internal/mlog"

	"go.opentelemetry.io/otel"

	marsauthorizor "github.com/duc-cnzj/mars/v4/internal/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	trace2 "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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

		ctx, span = trace2.NewNoopTracerProvider().Tracer("").Start(r.Context(), "")
		url := r.URL.String()
		if !tracingIgnoreFn(r.URL.Path) {
			ctxt := propagation.TraceContext{}
			ctx, span = otel.Tracer("mars").Start(ctxt.Extract(r.Context(), propagation.HeaderCarrier(r.Header)), fmt.Sprintf("[%s]: %s", r.Method, url))
			span.SetAttributes(grpcGatewayTag)
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

type gatewayCarrier metadata.MD

func (hc gatewayCarrier) Get(key string) string {
	vals := metadata.MD(hc).Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func (hc gatewayCarrier) Set(key string, value string) {
	metadata.MD(hc).Set(key, value)
}

func (hc gatewayCarrier) Keys() []string {
	keys := make([]string, 0, len(hc))
	for k := range hc {
		keys = append(keys, k)
	}
	return keys
}

func TraceUnaryClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start, span := otel.Tracer("mars").Start(ctx, "TraceUnaryClientInterceptor: "+method)
	defer span.End()
	span.SetAttributes(attribute.String("method", method))
	ctxt := propagation.TraceContext{}
	md := metadata.MD{}
	if outMD, found := metadata.FromOutgoingContext(ctx); found {
		md = outMD
	}
	ctxt.Inject(start, gatewayCarrier(md))
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD(md))

	err := invoker(ctx, method, req, reply, cc, opts...)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

func TraceUnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	ctxt := propagation.TraceContext{}
	md := metadata.MD{}
	if incomingContext, b := metadata.FromIncomingContext(ctx); b {
		ctx = ctxt.Extract(ctx, gatewayCarrier(incomingContext))
		md = incomingContext
	}
	start, span := otel.Tracer("mars").Start(ctx, info.FullMethod)
	defer span.End()
	incomingContext := metadata.NewIncomingContext(start, md)
	user := &marsauthorizor.UserInfo{}

	if u, _ := marsauthorizor.GetUser(incomingContext); u != nil {
		user = u
	}
	span.SetAttributes(attribute.String("user", user.Name))
	span.SetAttributes(attribute.String("email", user.Email))
	i, err := handler(incomingContext, req)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
	}
	return i, err
}
