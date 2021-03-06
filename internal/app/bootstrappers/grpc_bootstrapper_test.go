package bootstrappers

import (
	"context"
	"testing"

	auth2 "github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/testutil"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
)

func TestAuthenticate(t *testing.T) {
	_, err := Authenticate(context.TODO())
	assert.Error(t, err)
	md := metadata.New(map[string]string{"authorization": "Bearer xxx"})

	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	auth := mock.NewMockAuthInterface(m)
	app.EXPECT().Auth().Return(auth).Times(2)
	auth.EXPECT().VerifyToken("xxx").Return(nil, false)
	incomingContext := metadata.NewIncomingContext(context.TODO(), md)
	_, err = Authenticate(incomingContext)
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, fromError.Code())

	auth.EXPECT().VerifyToken("xxx").Return(&contracts.JwtClaims{
		UserInfo: contracts.UserInfo{
			OpenIDClaims: contracts.OpenIDClaims{Name: "duc"},
		},
	}, true)
	ctx2, err := Authenticate(incomingContext)
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

func Test_grpcRunner_Run(t *testing.T) {}

func Test_grpcRunner_Shutdown(t *testing.T) {}

func Test_traceWithOpName(t *testing.T) {}
