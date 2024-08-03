package middlewares

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

func MetricsServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func(t time.Time) {
		//user := &marsauthorizor.UserInfo{}
		//if u, err := marsauthorizor.GetUser(ctx); err == nil {
		//	user = u
		//}
		//mlog.Infof("[Grpc]: user: %v, visit: %v, use: %s.", user.Name, info.FullMethod, time.Since(t))
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

func MetricsStreamServerInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	defer func(t time.Time) {
		//user, e := marsauthorizor.GetUser(ss.Context())
		//if e == nil {
		//mlog.Infof("[Grpc]: user: %v, visit: %v, use: %s.", user.Name, info.FullMethod, time.Since(t))
		//}
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
