package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	marsauthorizor "github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/contracts"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var grpcGatewayTag = attribute.String("component", "grpc-gateway")

// TracingWrapper
// [W3C Tracing Headers](https://www.w3.org/TR/trace-context/)
func TracingWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !TracingIgnoreFn(r.URL.Path) {
			url := r.URL.String()
			ctxt := propagation.TraceContext{}
			start, span := app.Tracer().Start(ctxt.Extract(r.Context(), propagation.HeaderCarrier(r.Header)), fmt.Sprintf("[%s]: %s", r.Method, url))
			span.SetAttributes(grpcGatewayTag)
			span.SetAttributes(attribute.String("url", url))
			span.SetAttributes(attribute.String("method", r.Method))
			span.SetAttributes(attribute.String("user-agent", r.UserAgent()))
			defer span.End()
			r = r.WithContext(start)
		}
		h.ServeHTTP(w, r)
	})
}

func TracingIgnoreFn(fullMethodName string) bool {
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

type GatewayCarrier metadata.MD

func (hc GatewayCarrier) Get(key string) string {
	vals := metadata.MD(hc).Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func (hc GatewayCarrier) Set(key string, value string) {
	metadata.MD(hc).Set(key, value)
}

func (hc GatewayCarrier) Keys() []string {
	keys := make([]string, 0, len(hc))
	for k := range hc {
		keys = append(keys, k)
	}
	return keys
}

func TraceUnaryClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start, span := app.Tracer().Start(ctx, "TraceUnaryClientInterceptor: "+method)
	defer span.End()
	span.SetAttributes(attribute.String("method", method))
	ctxt := propagation.TraceContext{}
	md := metadata.MD{}
	if outMD, found := metadata.FromOutgoingContext(ctx); found {
		ctxt.Inject(start, GatewayCarrier(outMD))
		md = outMD
	}
	ctxt.Inject(start, GatewayCarrier(md))
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
		ctx = ctxt.Extract(ctx, GatewayCarrier(incomingContext))
		md = incomingContext
	}
	start, span := app.Tracer().Start(ctx, info.FullMethod)
	defer span.End()
	incomingContext := metadata.NewIncomingContext(start, md)
	user := &contracts.UserInfo{}

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

func TraceStreamServerInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()
	ctxt := propagation.TraceContext{}
	md := metadata.MD{}
	if incomingMD, b := metadata.FromIncomingContext(ctx); b {
		ctx = ctxt.Extract(ctx, GatewayCarrier(incomingMD))
		md = incomingMD
	}
	start, span := app.Tracer().Start(ctx, info.FullMethod)
	defer span.End()
	span.SetAttributes(attribute.Bool("is_server_stream", info.IsServerStream))
	span.SetAttributes(attribute.Bool("is_client_stream", info.IsClientStream))
	incomingContext := metadata.NewIncomingContext(start, md)
	user := &contracts.UserInfo{}
	if u, _ := marsauthorizor.GetUser(incomingContext); u != nil {
		user = u
	}
	span.SetAttributes(attribute.String("user", user.Name))
	span.SetAttributes(attribute.String("email", user.Email))
	err := handler(srv, ss)
	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
	}

	return err
}
