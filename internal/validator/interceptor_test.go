package validator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestStreamServerInterceptor(t *testing.T) {
	assert.IsType(t, (grpc.StreamServerInterceptor)(nil), StreamServerInterceptor())
}

func TestUnaryServerInterceptor(t *testing.T) {
	assert.IsType(t, (grpc.UnaryServerInterceptor)(nil), UnaryServerInterceptor())
}

type ss struct {
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
	return nil
}

type v struct {
	called bool
}

func (v *v) Validate() error {
	v.called = true
	return nil
}

func Test_recvWrapper_RecvMsg(t *testing.T) {
	r := recvWrapper{ServerStream: &ss{}}
	vv := &v{}
	r.RecvMsg(vv)
	assert.True(t, vv.called)
}
