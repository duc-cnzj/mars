package middlewares

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestGetUser(t *testing.T) {
	_, err := auth.GetUser(context.TODO())
	assert.Error(t, err)

	ctx := auth.SetUser(context.TODO(), &auth.UserInfo{
		Name: "duc",
	})
	user, err := auth.GetUser(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "duc", user.Name)
}

func TestMustGetUser(t *testing.T) {
	user := auth.MustGetUser(context.TODO())
	assert.Nil(t, user)
	ctx := auth.SetUser(context.TODO(), &auth.UserInfo{
		Name: "duc",
	})
	user = auth.MustGetUser(ctx)
	assert.Equal(t, "duc", user.Name)
}

func TestAuthStreamServerInterceptor(t *testing.T) {
	err := AuthStreamServerInterceptor()(&authServer{
		err: errors.New("xxx"),
	}, &ss{}, &grpc.StreamServerInfo{}, func(srv any, stream grpc.ServerStream) error {
		return nil
	})
	assert.Equal(t, "xxx", err.Error())

	called := 0
	err = AuthStreamServerInterceptor()(&authServer{}, &ss{}, &grpc.StreamServerInfo{}, func(srv any, stream grpc.ServerStream) error {
		called++
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, called)
	err = AuthStreamServerInterceptor()(nil, &ss{}, &grpc.StreamServerInfo{}, func(srv any, stream grpc.ServerStream) error {
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

func TestAuthUnaryServerInterceptor(t *testing.T) {
	called := 0
	as := &authServer{}
	_, err2 := AuthUnaryServerInterceptor()(context.TODO(), nil, &grpc.UnaryServerInfo{
		Server: as,
	}, func(ctx context.Context, req any) (any, error) {
		called++
		return nil, nil
	})
	assert.Nil(t, err2)
	assert.Equal(t, 1, called)
	assert.Equal(t, 1, as.called)

	ase := &authServer{err: errors.New("xxx")}
	_, err := AuthUnaryServerInterceptor()(context.TODO(), nil, &grpc.UnaryServerInfo{
		Server: ase,
	}, func(ctx context.Context, req any) (any, error) {
		called++
		return nil, nil
	})
	assert.Error(t, err)
	assert.Equal(t, 1, called)
	assert.Equal(t, 1, as.called)
}
