package bootstrappers

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	auth2 "github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthenticate(t *testing.T) {
	_, err := authenticate(context.TODO())
	assert.Error(t, err)
	md := metadata.New(map[string]string{"authorization": "Bearer xxx"})

	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	auth := mock.NewMockAuthInterface(m)
	app.EXPECT().Auth().Return(auth).Times(2)
	auth.EXPECT().VerifyToken("xxx").Return(nil, false)
	incomingContext := metadata.NewIncomingContext(context.TODO(), md)
	_, err = authenticate(incomingContext)
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, fromError.Code())

	auth.EXPECT().VerifyToken("xxx").Return(&contracts.JwtClaims{
		UserInfo: contracts.UserInfo{
			Name: "duc",
		},
	}, true)
	ctx2, err := authenticate(incomingContext)
	assert.Nil(t, err)
	assert.Equal(t, "duc", auth2.MustGetUser(ctx2).Name)
}

func TestGrpcBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().AddServer(gomock.Any()).Times(1)
	app.EXPECT().Config().Times(1).Return(&config.Config{})
	(&GrpcBootstrapper{}).Bootstrap(app)
}

func Test_grpcRunner_Run(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Auth().Times(1)
	app.EXPECT().Config().Return(&config.Config{}).Times(1)
	app.EXPECT().Oidc().Times(1)
	port, _ := config.GetFreePort()
	assert.Nil(t, (&grpcRunner{server: &mockGrpcServer{}, endpoint: fmt.Sprintf("0.0.0.0:%d", port)}).Run(context.TODO()))
}

type mockGrpcServer struct{}

func (m *mockGrpcServer) Serve(lis net.Listener) error {
	return lis.Close()
}

func (m *mockGrpcServer) GracefulStop() {
	time.Sleep(1 * time.Second)
}

func Test_grpcRunner_Shutdown(t *testing.T) {
	assert.Nil(t, (&grpcRunner{}).Shutdown(context.TODO()))
	assert.Nil(t, (&grpcRunner{server: &mockGrpcServer{}}).Shutdown(context.TODO()))
	cancel, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	assert.Error(t, (&grpcRunner{server: &mockGrpcServer{}}).Shutdown(cancel))
}

func TestGrpcBootstrapper_Tags(t *testing.T) {
	assert.Equal(t, []string{"api", "grpc"}, (&GrpcBootstrapper{}).Tags())
}

func Test_recoveryHandler(t *testing.T) {
	assert.Nil(t, recoveryHandler(nil))
}

func Test_grpcRunner_initServer(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Auth().Times(1)
	app.EXPECT().Config().Return(&config.Config{}).Times(1)
	app.EXPECT().Oidc().Times(1)
	assert.NotNil(t, (&grpcRunner{}).initServer())
}
