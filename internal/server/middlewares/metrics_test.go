package middlewares

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// TestMetricsUnaryServerInterceptor 覆盖 Unary 指标拦截器：成功 +GrpcRequestTotalSuccess，
// 失败 +GrpcRequestTotalFail 与 +GrpcErrorCount，并透传 handler 结果。
func TestMetricsUnaryServerInterceptor(t *testing.T) {
	res, err := MetricsUnaryServerInterceptor()(context.TODO(), nil, &grpc.UnaryServerInfo{FullMethod: "/metrics/ok"}, func(ctx context.Context, req any) (any, error) {
		return "aa", nil
	})
	assert.Nil(t, err)
	assert.Equal(t, "aa", res)
	assert.Equal(t, float64(1), testutil.ToFloat64(metrics.GrpcRequestTotalSuccess.WithLabelValues("/metrics/ok")))

	res2, err2 := MetricsUnaryServerInterceptor()(context.TODO(), nil, &grpc.UnaryServerInfo{FullMethod: "/metrics/fail"}, func(ctx context.Context, req any) (any, error) {
		return nil, errors.New("xxx")
	})
	assert.Equal(t, "xxx", err2.Error())
	assert.Nil(t, res2)
	assert.Equal(t, float64(1), testutil.ToFloat64(metrics.GrpcRequestTotalFail.WithLabelValues("/metrics/fail")))
	assert.Equal(t, float64(1), testutil.ToFloat64(metrics.GrpcErrorCount.WithLabelValues("/metrics/fail")))
}

type sstream struct {
	ctx context.Context
	grpc.ServerStream
}

func (s *sstream) Context() context.Context {
	return s.ctx
}

// TestMetricsStreamServerInterceptor 覆盖 Stream 指标拦截器：成功 +GrpcRequestTotalSuccess，
// 失败 +GrpcRequestTotalFail 与 +GrpcErrorCount，并透传 handler 错误。
func TestMetricsStreamServerInterceptor(t *testing.T) {
	err := MetricsStreamServerInterceptor()(nil, &sstream{ctx: context.TODO()}, &grpc.StreamServerInfo{FullMethod: "/metrics/sok"}, func(srv any, stream grpc.ServerStream) error {
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, float64(1), testutil.ToFloat64(metrics.GrpcRequestTotalSuccess.WithLabelValues("/metrics/sok")))

	err = MetricsStreamServerInterceptor()(nil, &sstream{ctx: context.TODO()}, &grpc.StreamServerInfo{FullMethod: "/metrics/sfail"}, func(srv any, stream grpc.ServerStream) error {
		return errors.New("xxx")
	})
	assert.Equal(t, "xxx", err.Error())
	assert.Equal(t, float64(1), testutil.ToFloat64(metrics.GrpcRequestTotalFail.WithLabelValues("/metrics/sfail")))
	assert.Equal(t, float64(1), testutil.ToFloat64(metrics.GrpcErrorCount.WithLabelValues("/metrics/sfail")))
}

// Test_accountGrpcResult 覆盖按成败累加指标的分支语义：成功只加 Success，
// 失败同时加 Fail 与 ErrorCount。
func Test_accountGrpcResult(t *testing.T) {
	accountGrpcResult("/acct/success", nil)
	assert.Equal(t, float64(1), testutil.ToFloat64(metrics.GrpcRequestTotalSuccess.WithLabelValues("/acct/success")))

	accountGrpcResult("/acct/fail", errors.New("x"))
	assert.Equal(t, float64(1), testutil.ToFloat64(metrics.GrpcRequestTotalFail.WithLabelValues("/acct/fail")))
	assert.Equal(t, float64(1), testutil.ToFloat64(metrics.GrpcErrorCount.WithLabelValues("/acct/fail")))
}
