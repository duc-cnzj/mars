package validator

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestStreamServerInterceptor(t *testing.T) {
	assert.IsType(t, (grpc.StreamServerInterceptor)(nil), StreamServerInterceptor())
	called := false
	StreamServerInterceptor()("", nil, nil, func(srv any, stream grpc.ServerStream) error {
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
	assert.IsType(t, (grpc.UnaryServerInterceptor)(nil), UnaryServerInterceptor())

	called := 0
	UnaryServerInterceptor()(context.TODO(), &mockValidator{}, nil, func(ctx context.Context, req any) (any, error) {
		called++
		return nil, nil
	})
	assert.Equal(t, 1, called)
	_, err := UnaryServerInterceptor()(context.TODO(), &mockValidator{err: errors.New("xxx")}, nil, func(ctx context.Context, req any) (any, error) {
		called++
		return nil, nil
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, 1, called)
	assert.Equal(t, "xxx", fromError.Message())
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
	return nil
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
	r := recvWrapper{ServerStream: &ss{}}
	vv := &v{}
	r.RecvMsg(vv)
	assert.True(t, vv.called)

	r1 := recvWrapper{ServerStream: &ss{}}
	vv1 := &v{
		err: errors.New("xxx"),
	}
	assert.Equal(t, "xxx", r1.RecvMsg(vv1).Error())
	assert.True(t, vv1.called)

	r2 := recvWrapper{ServerStream: &ss{
		recvErr: errors.New("xxx"),
	}}
	vv2 := &v{}
	assert.Equal(t, "xxx", r2.RecvMsg(vv2).Error())
	assert.False(t, vv2.called)
}
