package middlewares

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
)

// TestAccessLogUnaryServerInterceptor 覆盖 Unary 访问日志：有用户/无用户都打日志，
// 无用户（公开方法）用户名为空串；返回值透传 handler。
func TestAccessLogUnaryServerInterceptor(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	logger := mlog.NewMockLogger(m)

	ctx := biz.SetUser(context.TODO(), &biz.UserInfo{Name: "duc"})
	logger.EXPECT().Infof("[Grpc]: user: %v, visit: %v, use: %s.", "duc", "/access/ok", gomock.Any()).Times(1)
	res, err := AccessLogUnaryServerInterceptor(logger)(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/access/ok"}, func(ctx context.Context, req any) (any, error) {
		return "aa", nil
	})
	assert.Nil(t, err)
	assert.Equal(t, "aa", res)

	logger.EXPECT().Infof("[Grpc]: user: %v, visit: %v, use: %s.", "", "/access/public", gomock.Any()).Times(1)
	_, err = AccessLogUnaryServerInterceptor(logger)(context.TODO(), nil, &grpc.UnaryServerInfo{FullMethod: "/access/public"}, func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})
	assert.Nil(t, err)
}

// TestAccessLogStreamServerInterceptor 覆盖 Stream 访问日志：流会话结束时打日志，
// 有用户/无用户均打，use 为整段会话时长。
func TestAccessLogStreamServerInterceptor(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	logger := mlog.NewMockLogger(m)

	ctx := biz.SetUser(context.TODO(), &biz.UserInfo{Name: "duc"})
	logger.EXPECT().Infof("[Grpc]: user: %v, visit: %v, use: %s.", "duc", "/access/sok", gomock.Any()).Times(1)
	err := AccessLogStreamServerInterceptor(logger)(nil, &sstream{ctx: ctx}, &grpc.StreamServerInfo{FullMethod: "/access/sok"}, func(srv any, stream grpc.ServerStream) error {
		return nil
	})
	assert.Nil(t, err)

	logger.EXPECT().Infof("[Grpc]: user: %v, visit: %v, use: %s.", "", "/access/spublic", gomock.Any()).Times(1)
	err = AccessLogStreamServerInterceptor(logger)(nil, &sstream{ctx: context.TODO()}, &grpc.StreamServerInfo{FullMethod: "/access/spublic"}, func(srv any, stream grpc.ServerStream) error {
		return nil
	})
	assert.Nil(t, err)
}

// Test_grpcUser 覆盖用户解析两分支：已注入返回原用户，未注入返回空 UserInfo 而非 nil。
func Test_grpcUser(t *testing.T) {
	ctx := biz.SetUser(context.TODO(), &biz.UserInfo{Name: "duc"})
	assert.Equal(t, "duc", grpcUser(ctx).Name)

	assert.Equal(t, "", grpcUser(context.TODO()).Name)
}
