package services

import (
	"context"
	"errors"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	auth2 "github.com/duc-cnzj/mars/api/v5/auth"
	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/metadata"
)

func TestNewAuthSvc(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewAuthSvc(repo.NewMockEventRepo(m), mlog.NewForConfig(nil), repo.NewMockAuthRepo(m))
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
	svc := NewAuthSvc(eventRepo, mlog.NewForConfig(nil), authRepo)
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
	svc := NewAuthSvc(eventRepo, mlog.NewForConfig(nil), authRepo)
	resp, err := svc.Info(context.TODO(), nil)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestAuthSvc_Login_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	authRepo := repo.NewMockAuthRepo(m)
	svc := NewAuthSvc(eventRepo, mlog.NewForConfig(nil), authRepo)

	eventRepo.EXPECT().AuditLog(gomock.Any(), gomock.Any(), gomock.Any())
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
	svc := NewAuthSvc(eventRepo, mlog.NewForConfig(nil), authRepo)

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
	svc := NewAuthSvc(eventRepo, mlog.NewForConfig(nil), authRepo)

	authRepo.EXPECT().Settings(gomock.Any()).Return(data.OidcConfig{}, nil)

	_, err := svc.Settings(context.Background(), &auth2.SettingsRequest{})
	assert.NoError(t, err)
}

func Test_guest_AuthFuncOverride(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	svc := NewAuthSvc(repo.NewMockEventRepo(m), mlog.NewForConfig(nil), repo.NewMockAuthRepo(m))

	_, err := svc.(*authSvc).AuthFuncOverride(context.Background(), "TestMethod")
	assert.NoError(t, err)
}

func TestNewDefaultAuthProvider(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	provider := &oidc.Provider{}
	cfg := oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://localhost:8080/auth",
			TokenURL: "http://localhost:8080/token",
		},
	}

	authProvider := NewDefaultAuthProvider(cfg, provider)

	assert.NotNil(t, authProvider)
	assert.IsType(t, &defaultAuthProvider{}, authProvider)
}

func TestAuthSvc_Settings_NoSettings(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	authRepo := repo.NewMockAuthRepo(m)
	svc := NewAuthSvc(eventRepo, mlog.NewForConfig(nil), authRepo)

	authRepo.EXPECT().Settings(gomock.Any()).Return(nil, nil)

	resp, err := svc.Settings(context.Background(), &auth2.SettingsRequest{})
	assert.NoError(t, err)
	assert.Empty(t, resp.Items)
}

func TestAuthSvc_Settings_ErrorFetchingSettings(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	authRepo := repo.NewMockAuthRepo(m)
	svc := NewAuthSvc(eventRepo, mlog.NewForConfig(nil), authRepo)

	authRepo.EXPECT().Settings(gomock.Any()).Return(data.OidcConfig{
		"b": data.OidcConfigItem{
			Config:             oauth2.Config{},
			EndSessionEndpoint: "",
		},
		"a": data.OidcConfigItem{
			Config:             oauth2.Config{},
			EndSessionEndpoint: "",
		},
	}, nil)

	res, err := svc.Settings(context.Background(), &auth2.SettingsRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(res.Items))
	assert.Equal(t, "a", res.Items[0].Name)
	assert.Equal(t, "b", res.Items[1].Name)
}
