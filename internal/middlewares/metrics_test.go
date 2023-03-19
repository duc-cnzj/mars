package middlewares

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestMetricsServerInterceptor(t *testing.T) {
	ctx := auth.SetUser(context.TODO(), &contracts.UserInfo{
		Name: "duc",
	})
	res, err := MetricsServerInterceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/api/xxx"}, func(ctx context.Context, req any) (any, error) {
		return "aa", nil
	})
	assert.Nil(t, err)
	assert.Equal(t, "aa", res)
	res2, err2 := MetricsServerInterceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/api/xxx"}, func(ctx context.Context, req any) (any, error) {
		return nil, errors.New("xxx")
	})
	assert.Equal(t, "xxx", err2.Error())
	assert.Nil(t, res2)
}

type sstream struct {
	ctx context.Context
	grpc.ServerStream
}

func (s *sstream) Context() context.Context {
	return s.ctx
}

func TestMetricsStreamServerInterceptor(t *testing.T) {
	ctx := auth.SetUser(context.TODO(), &contracts.UserInfo{
		Name: "duc",
	})
	err := MetricsStreamServerInterceptor(nil, &sstream{ctx: ctx}, &grpc.StreamServerInfo{FullMethod: "/api/xx"}, func(srv any, stream grpc.ServerStream) error {
		return nil
	})
	assert.Nil(t, err)
	err = MetricsStreamServerInterceptor(nil, &sstream{ctx: ctx}, &grpc.StreamServerInfo{FullMethod: "/api/xx"}, func(srv any, stream grpc.ServerStream) error {
		return errors.New("xxx")
	})
	assert.Equal(t, "xxx", err.Error())
}
