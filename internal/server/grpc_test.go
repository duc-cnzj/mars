package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGrpcRunner_RecoveryHandler(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	logger := mlog.NewMockLogger(m)
	auth := auth.NewMockAuth(m)

	runner := &grpcRunner{
		logger: logger,
		auth:   auth,
	}

	// Test case: recoveryHandler logs the error
	err := errors.New("test error")
	logger.EXPECT().Errorf("[Grpc]: recovery error: \n%v", err).Times(1)

	assert.Nil(t, runner.recoveryHandler(err))
}

func TestAuthenticate(t *testing.T) {
	_, err := authenticate(context.TODO(), nil)
	assert.Error(t, err)
	md := metadata.New(map[string]string{"authorization": "Bearer xxx"})

	m := gomock.NewController(t)
	defer m.Finish()
	authS := auth.NewMockAuth(m)
	authS.EXPECT().VerifyToken("xxx").Return(nil, false)
	incomingContext := metadata.NewIncomingContext(context.TODO(), md)
	_, err = authenticate(incomingContext, authS)
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, fromError.Code())

	authS.EXPECT().VerifyToken("xxx").Return(&auth.JwtClaims{
		UserInfo: &auth.UserInfo{
			Name: "duc",
		},
	}, true)
	ctx2, err := authenticate(incomingContext, authS)
	assert.Nil(t, err)
	assert.Equal(t, "duc", auth.MustGetUser(ctx2).Name)
}

func TestNewGrpcRunner(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	appMock := application.NewMockApp(m)
	appMock.EXPECT().GrpcRegistry().Return(nil).Times(1)
	appMock.EXPECT().Logger().Return(mlog.NewForConfig(nil)).Times(1)
	appMock.EXPECT().Auth().Return(auth.NewMockAuth(m)).Times(1)

	runner := NewGrpcRunner("test-endpoint", appMock)

	assert.NotNil(t, runner)
	assert.Equal(t, "test-endpoint", runner.(*grpcRunner).endpoint)
}

func TestGrpcRunner_Shutdown(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	server := NewMockGrpcServerImp(m)
	runner := &grpcRunner{
		logger: mlog.NewForConfig(nil),
		server: server,
	}

	server.EXPECT().GracefulStop().Times(2)
	// Test case: Shutdown completes before context deadline
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	err := runner.Shutdown(ctx)
	assert.Nil(t, err)

	// Test case: Shutdown does not complete before context deadline
	ctx, cancel = context.WithCancel(context.TODO())
	cancel()

	err = runner.Shutdown(ctx)
	assert.NotNil(t, err)
	time.Sleep(time.Second)
}

func Test_grpcRunner_initServer(t *testing.T) {
	var ss any
	(&grpcRunner{
		grpcRegistry: &application.GrpcRegistry{
			RegistryFunc: func(s grpc.ServiceRegistrar) {
				ss = s
			},
		},
	}).initServer()

	assert.NotNil(t, ss)
}
