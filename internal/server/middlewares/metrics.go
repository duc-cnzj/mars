package middlewares

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"

	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

// MetricsServerInterceptor 是 Unary 指标拦截器：记录请求耗时、调用用户与方法名，
// 并按成败累加 GrpcRequestTotalFail/Success 与 GrpcErrorCount 指标。
func MetricsServerInterceptor(logger mlog.Logger) func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func(t time.Time) {
			user := &biz.UserInfo{}
			if u, err := biz.GetUser(ctx); err == nil {
				user = u
			}
			logger.Infof("[Grpc]: user: %v, visit: %v, use: %s.", user.Name, info.FullMethod, time.Since(t))
			metrics.GrpcLatency.With(prometheus.Labels{"method": info.FullMethod}).Observe(time.Since(t).Seconds())
		}(time.Now())

		i, err := handler(ctx, req)
		if err != nil {
			metrics.GrpcRequestTotalFail.With(prometheus.Labels{"method": info.FullMethod}).Inc()
			metrics.GrpcErrorCount.With(prometheus.Labels{"method": info.FullMethod}).Inc()
		} else {
			metrics.GrpcRequestTotalSuccess.With(prometheus.Labels{"method": info.FullMethod}).Inc()
		}

		return i, err
	}
}

// MetricsStreamServerInterceptor 是 Stream 指标拦截器：语义同 Unary 版，
// 请求成败同样累加 GrpcRequestTotalFail/Success 与 GrpcErrorCount。
func MetricsStreamServerInterceptor(logger mlog.Logger) func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		defer func(t time.Time) {
			user, e := biz.GetUser(ss.Context())
			if e == nil {
				logger.Infof("[Grpc]: user: %v, visit: %v, use: %s.", user.Name, info.FullMethod, time.Since(t))
			}
		}(time.Now())

		e := handler(srv, ss)
		if e != nil {
			metrics.GrpcRequestTotalFail.With(prometheus.Labels{"method": info.FullMethod}).Inc()
			metrics.GrpcErrorCount.With(prometheus.Labels{"method": info.FullMethod}).Inc()
		} else {
			metrics.GrpcRequestTotalSuccess.With(prometheus.Labels{"method": info.FullMethod}).Inc()
		}

		return e
	}
}
