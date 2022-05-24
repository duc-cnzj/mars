package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestGetUser(t *testing.T) {
	_, err := GetUser(context.TODO())
	assert.Error(t, err)

	ctx := SetUser(context.TODO(), &contracts.UserInfo{
		OpenIDClaims: contracts.OpenIDClaims{Name: "duc"},
	})
	user, err := GetUser(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "duc", user.Name)
}

func TestMustGetUser(t *testing.T) {
	user := MustGetUser(context.TODO())
	assert.Nil(t, user)
	ctx := SetUser(context.TODO(), &contracts.UserInfo{
		OpenIDClaims: contracts.OpenIDClaims{Name: "duc"},
	})
	user = MustGetUser(ctx)
	assert.Equal(t, "duc", user.Name)
}

type ss struct {
	grpc.ServerStream
}

func (s *ss) Context() context.Context {
	return context.TODO()
}

func TestStreamServerInterceptor(t *testing.T) {
	err := StreamServerInterceptor()(&authServer{
		err: errors.New("xxx"),
	}, &ss{}, &grpc.StreamServerInfo{}, func(srv any, stream grpc.ServerStream) error {
		return nil
	})
	assert.Equal(t, "xxx", err.Error())

	called := 0
	err = StreamServerInterceptor()(&authServer{}, &ss{}, &grpc.StreamServerInfo{}, func(srv any, stream grpc.ServerStream) error {
		called++
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, called)
	err = StreamServerInterceptor()(nil, &ss{}, &grpc.StreamServerInfo{}, func(srv any, stream grpc.ServerStream) error {
		called++
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, called)
}

type authServer struct {
	err    error
	called int
}

func (a *authServer) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	a.called++
	if a.err != nil {
		return nil, a.err
	}
	return nil, nil
}

func TestUnaryServerInterceptor(t *testing.T) {
	called := 0
	as := &authServer{}
	_, err2 := UnaryServerInterceptor()(context.TODO(), nil, &grpc.UnaryServerInfo{
		Server: as,
	}, func(ctx context.Context, req any) (any, error) {
		called++
		return nil, nil
	})
	assert.Nil(t, err2)
	assert.Equal(t, 1, called)
	assert.Equal(t, 1, as.called)

	ase := &authServer{err: errors.New("xxx")}
	_, err := UnaryServerInterceptor()(context.TODO(), nil, &grpc.UnaryServerInfo{
		Server: ase,
	}, func(ctx context.Context, req any) (any, error) {
		called++
		return nil, nil
	})
	assert.Error(t, err)
	assert.Equal(t, 1, called)
	assert.Equal(t, 1, as.called)
}
