package middlewares

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

// MetricsUnaryServerInterceptor 是 Unary 指标拦截器：记录请求耗时（GrpcLatency），
// 并按成败累加 GrpcRequestTotalFail/Success 与 GrpcErrorCount 指标。
func MetricsUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func(t time.Time) {
			metrics.GrpcLatency.With(prometheus.Labels{"method": info.FullMethod}).Observe(time.Since(t).Seconds())
		}(time.Now())

		resp, err = handler(ctx, req)
		accountGrpcResult(info.FullMethod, err)
		return resp, err
	}
}

// MetricsStreamServerInterceptor 是 Stream 指标拦截器：按成败累加
// GrpcRequestTotalFail/Success 与 GrpcErrorCount。与 Unary 版差异是不观测
// GrpcLatency——长连接流（如日志流）持续秒级，按请求计耗时无意义。
func MetricsStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, ss)
		accountGrpcResult(info.FullMethod, err)
		return err
	}
}

// accountGrpcResult 按成败累加 gRPC 指标：成功 +GrpcRequestTotalSuccess，
// 失败 +GrpcRequestTotalFail 与 +GrpcErrorCount，method 标签取 fullMethodName。
func accountGrpcResult(fullMethodName string, err error) {
	if err != nil {
		metrics.GrpcRequestTotalFail.With(prometheus.Labels{"method": fullMethodName}).Inc()
		metrics.GrpcErrorCount.With(prometheus.Labels{"method": fullMethodName}).Inc()
	} else {
		metrics.GrpcRequestTotalSuccess.With(prometheus.Labels{"method": fullMethodName}).Inc()
	}
}
