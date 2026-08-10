package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"google.golang.org/grpc"
)

// RouteLogger 是 HTTP 访问日志中间件：debug 级记录请求方法与路径及耗时。
func RouteLogger(logger mlog.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(t time.Time) {
			logger.Debugf("[Http]: method: %v, url: %v, use %v", r.Method, r.URL, time.Since(t))
		}(time.Now())
		h.ServeHTTP(w, r)
	})
}

// LoggerUnaryServerInterceptor 是 gRPC 请求日志拦截器：请求体实现了 Stringer 时
// debug 级记录方法名与请求内容。
func LoggerUnaryServerInterceptor(logger mlog.Logger) func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if request, ok := req.(interface {
			String() string
		}); ok {
			logger.Debugf("[request logger]: method=%s body=%v", info.FullMethod, request.String())
		}
		return handler(ctx, req)
	}
}
