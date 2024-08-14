package services

import (
	"context"
	"errors"
	"testing"

	auth2 "github.com/duc-cnzj/mars/api/v4/auth"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/metadata"
)

func TestNewAuthSvc(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewAuthSvc(repo.NewMockEventRepo(m), mlog.NewLogger(nil), repo.NewMockAuthRepo(m))
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.(*authSvc).logger)
	assert.NotNil(t, svc.(*authSvc).eventRepo)
	assert.NotNil(t, svc.(*authSvc).authRepo)
}

func Test_authSvc_Info(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	authRepo := repo.NewMockAuthRepo(m)
	svc := NewAuthSvc(eventRepo, mlog.NewLogger(nil), authRepo)
	authRepo.EXPECT().VerifyToken(gomock.Any(), "token").Return(&auth.UserInfo{}, nil)
	resp, err := svc.Info(metadata.NewIncomingContext(context.TODO(), metadata.Pairs("Authorization", "token")), nil)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func Test_authSvc_Info_Fail(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	authRepo := repo.NewMockAuthRepo(m)
	svc := NewAuthSvc(eventRepo, mlog.NewLogger(nil), authRepo)
	resp, err := svc.Info(context.TODO(), nil)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestAuthSvc_Login_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	authRepo := repo.NewMockAuthRepo(m)
	svc := NewAuthSvc(eventRepo, mlog.NewLogger(nil), authRepo)

	eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	authRepo.EXPECT().Login(gomock.Any(), &repo.LoginInput{
		Username: "test",
		Password: "password",
	}).Return(&repo.LoginResponse{
		Token:     "test-token",
		ExpiredIn: 100,
		UserInfo:  &auth.UserInfo{},
	}, nil)

	resp, err := svc.Login(context.TODO(), &auth2.LoginRequest{
		Username: "test",
		Password: "password",
	})
	assert.NoError(t, err)
	assert.Equal(t, "test-token", resp.Token)
}

func TestAuthSvc_Login_Failure(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	authRepo := repo.NewMockAuthRepo(m)
	svc := NewAuthSvc(eventRepo, mlog.NewLogger(nil), authRepo)

	authRepo.EXPECT().Login(gomock.Any(), &repo.LoginInput{
		Username: "test",
		Password: "password",
	}).Return(nil, errors.New("error"))

	_, err := svc.Login(context.Background(), &auth2.LoginRequest{
		Username: "test",
		Password: "password",
	})
	assert.Error(t, err)
}

func TestAuthSvc_Settings_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	authRepo := repo.NewMockAuthRepo(m)
	svc := NewAuthSvc(eventRepo, mlog.NewLogger(nil), authRepo)

	authRepo.EXPECT().Settings(gomock.Any()).Return(data.OidcConfig{}, nil)

	_, err := svc.Settings(context.Background(), &auth2.SettingsRequest{})
	assert.NoError(t, err)
}

func Test_guest_AuthFuncOverride(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewAuthSvc(repo.NewMockEventRepo(m), mlog.NewLogger(nil), repo.NewMockAuthRepo(m))

	_, err := svc.(*authSvc).AuthFuncOverride(context.Background(), "TestMethod")
	assert.NoError(t, err)
}
