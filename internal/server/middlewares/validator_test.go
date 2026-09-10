package middlewares

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestStreamServerInterceptor(t *testing.T) {
	t.Parallel()
	assert.IsType(t, (grpc.StreamServerInterceptor)(nil), ValidatorStreamServerInterceptor())
	called := false
	ValidatorStreamServerInterceptor()("", nil, nil, func(srv any, stream grpc.ServerStream) error {
		assert.IsType(t, (*recvWrapper)(nil), stream)
		called = true
		return nil
	})
	assert.True(t, called)
}

type mockValidator struct {
	err error
}

func (m *mockValidator) Validate() error {
	return m.err
}

func TestUnaryServerInterceptor(t *testing.T) {
	t.Parallel()
	assert.IsType(t, (grpc.UnaryServerInterceptor)(nil), ValidatorUnaryServerInterceptor())

	called := 0
	ValidatorUnaryServerInterceptor()(context.TODO(), &mockValidator{}, nil, func(ctx context.Context, req any) (any, error) {
		called++
		return nil, nil
	})
	assert.Equal(t, 1, called)
	_, err := ValidatorUnaryServerInterceptor()(context.TODO(), &mockValidator{err: errors.New("xxx")}, nil, func(ctx context.Context, req any) (any, error) {
		called++
		return nil, nil
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, 1, called)
	assert.Equal(t, codes.InvalidArgument, fromError.Code())
	assert.Equal(t, "xxx", fromError.Message())

	// 请求未实现 Validator：不校验，直接透传 handler。
	called = 0
	resp, err := ValidatorUnaryServerInterceptor()(context.TODO(), "plain-string", nil, func(ctx context.Context, req any) (any, error) {
		called++
		return "ok", nil
	})
	assert.Nil(t, err)
	assert.Equal(t, "ok", resp)
	assert.Equal(t, 1, called)
}

type ss struct {
	recvErr error
}

func (s *ss) SetHeader(md metadata.MD) error {
	return nil
}

func (s *ss) SendHeader(md metadata.MD) error {
	return nil
}

func (s *ss) SetTrailer(md metadata.MD) {
}

func (s *ss) Context() context.Context {
	return context.TODO()
}

func (s *ss) SendMsg(m any) error {
	return nil
}

func (s *ss) RecvMsg(m any) error {
	return s.recvErr
}

type v struct {
	err    error
	called bool
}

func (v *v) Validate() error {
	v.called = true
	return v.err
}

func Test_recvWrapper_RecvMsg(t *testing.T) {
	t.Parallel()
	r := recvWrapper{ServerStream: &ss{}}
	vv := &v{}
	r.RecvMsg(vv)
	assert.True(t, vv.called)

	r1 := recvWrapper{ServerStream: &ss{}}
	vv1 := &v{
		err: errors.New("xxx"),
	}
	recvErr := r1.RecvMsg(vv1)
	// 回归防护：stream 校验失败必须映射成 InvalidArgument，与 unary 路径一致。
	// 若 RecvMsg 原样上抛 Validate 的 MultiError（非 status 错误），这里会落成 Unknown。
	st, ok := status.FromError(recvErr)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Equal(t, "xxx", st.Message())
	assert.True(t, vv1.called)

	r2 := recvWrapper{ServerStream: &ss{
		recvErr: errors.New("xxx"),
	}}
	vv2 := &v{}
	assert.Equal(t, "xxx", r2.RecvMsg(vv2).Error())
	assert.False(t, vv2.called)
}
