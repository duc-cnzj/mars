package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang/mock/gomock"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/duc-cnzj/mars-client/v4/auth"
	auth2 "github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/testutil"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestAuthSvc_AuthFuncOverride(t *testing.T) {
	svc := NewAuthSvc(nil, nil, "", nil)
	_, err := svc.AuthFuncOverride(context.TODO(), "")
	assert.Nil(t, err)
}

func TestAuthSvc_Info(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 1024)
	privateKey, _ := x509.MarshalPKCS8PrivateKey(key)
	bf := bytes.Buffer{}
	pem.Encode(&bf, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKey})
	authSvc := auth2.NewAuth(key, key.Public().(*rsa.PublicKey))
	sign, _ := authSvc.Sign(contracts.UserInfo{
		LogoutUrl: "xxx",
		Roles:     []string{"user"},
		OpenIDClaims: contracts.OpenIDClaims{
			Name:    "duc",
			Sub:     "2022",
			Email:   "1025434218@qq.com",
			Picture: "avatar.png",
		},
	})
	svc := NewAuthSvc(authSvc, nil, "", nil)
	ctx := context.TODO()
	info, err := svc.Info(ctx, &auth.InfoRequest{})
	assert.Nil(t, info)
	assert.Error(t, err)
	md := metadata.New(map[string]string{"Authorization": sign.Token})
	ctx = metadata.NewIncomingContext(ctx, md)
	info, err = svc.Info(ctx, &auth.InfoRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "duc", info.Name)
	assert.Equal(t, "1025434218@qq.com", info.Id)
	assert.Equal(t, "xxx", info.LogoutUrl)
	assert.Equal(t, "avatar.png", info.Avatar)
	assert.Equal(t, []string{"user"}, info.Roles)
}

func TestAuthSvc_Login(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	key, _ := rsa.GenerateKey(rand.Reader, 1024)
	privateKey, _ := x509.MarshalPKCS8PrivateKey(key)
	bf := bytes.Buffer{}
	pem.Encode(&bf, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKey})
	authSvc := auth2.NewAuth(key, key.Public().(*rsa.PublicKey))
	svc := NewAuthSvc(authSvc, nil, "admin", nil)
	app := testutil.MockApp(m)
	assertAuditLogFired(m, app)
	login, err := svc.Login(context.TODO(), &auth.LoginRequest{
		Username: "admin",
		Password: "admin",
	})
	assert.Nil(t, err)
	token, ok := authSvc.VerifyToken(login.Token)
	assert.Equal(t, "admin@mars.com", token.StandardClaims.Subject)
	assert.True(t, ok)
	assert.Equal(t, "管理员", token.UserInfo.Name)

	_, err = svc.Login(context.TODO(), &auth.LoginRequest{
		Username: "admin",
		Password: "xxxx",
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, fromError.Code())
}

func TestAuthSvc_Login_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	asvc := mock.NewMockAuthInterface(m)
	svc := NewAuthSvc(asvc, nil, "admin", nil)
	asvc.EXPECT().Sign(gomock.Any()).Return(nil, errors.New("xx")).Times(1)
	_, err := svc.Login(context.TODO(), &auth.LoginRequest{
		Username: "admin",
		Password: "admin",
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, fromError.Code())
}

func TestNewAuthSvc(t *testing.T) {
	assert.Implements(t, (*auth.AuthServer)(nil), NewAuthSvc(nil, nil, "", nil))
}

func TestAuthSvc_Settings(t *testing.T) {
	settings, err := (&AuthSvc{
		cfg: contracts.OidcConfig{
			"svc1": contracts.OidcConfigItem{
				Config: oauth2.Config{
					ClientID:     "xxx",
					ClientSecret: "aaa",
					Endpoint:     oauth2.Endpoint{},
					RedirectURL:  "/home",
					Scopes:       []string{"openid"},
				},
			},
			"svc2": contracts.OidcConfigItem{
				Config: oauth2.Config{
					ClientID:     "yyy",
					ClientSecret: "bbb",
					Endpoint:     oauth2.Endpoint{},
					RedirectURL:  "/redirect",
					Scopes:       []string{"openid"},
				},
			},
		},
	}).Settings(context.TODO(), &auth.SettingsRequest{})
	assert.Nil(t, err)

	assert.Len(t, settings.Items, 2)
}

type mockProvider struct {
	idtokenError  error
	exchangeError error
	verifyError   error
}

func (m *mockProvider) Exchange(ctx context.Context, code string) (string, error) {
	defer func() {
		m.exchangeError = nil
	}()
	if m.exchangeError != nil {
		return "", m.exchangeError
	}
	return "", nil
}

func (m *mockProvider) Verify(ctx context.Context, token string) (IDToken, error) {
	defer func() {
		m.verifyError = nil
	}()
	if m.verifyError != nil {
		return nil, m.verifyError
	}
	return &mockIDToken{
		err: m.idtokenError,
	}, nil
}

type mockIDToken struct {
	err error
}

func (m *mockIDToken) Claims(a any) error {
	if m.err != nil {
		return m.err
	}
	a.(*contracts.UserInfo).Sub = "mock"
	return nil
}

func TestAuthSvc_Exchange(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	asvc := mock.NewMockAuthInterface(m)
	asvc.EXPECT().Sign(gomock.Any()).Return(&contracts.SignData{
		Token:     "xx",
		ExpiredIn: 100,
	}, nil).Times(1)
	app := testutil.MockApp(m)
	assertAuditLogFired(m, app)
	exchange, err := NewAuthSvc(asvc, contracts.OidcConfig{
		"a": {
			Provider:           nil,
			Config:             oauth2.Config{},
			EndSessionEndpoint: "",
		},
		"b": {
			Provider:           nil,
			Config:             oauth2.Config{},
			EndSessionEndpoint: "",
		},
	}, "", func(cfg oauth2.Config, provider *oidc.Provider) OidcAuthProvider {
		return &mockProvider{}
	}).Exchange(context.TODO(), &auth.ExchangeRequest{
		Code: "xx",
	})
	assert.Nil(t, err)
	assert.Equal(t, "xx", exchange.Token)
	assert.Equal(t, int64(100), exchange.ExpiresIn)
}

func TestAuthSvc_Exchange_SignError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	asvc := mock.NewMockAuthInterface(m)
	asvc.EXPECT().Sign(gomock.Any()).Return(nil, errors.New("xxx")).Times(1)
	_, err := NewAuthSvc(asvc, contracts.OidcConfig{
		"a": {
			Provider:           nil,
			Config:             oauth2.Config{},
			EndSessionEndpoint: "",
		},
		"b": {
			Provider:           nil,
			Config:             oauth2.Config{},
			EndSessionEndpoint: "",
		},
	}, "", func(cfg oauth2.Config, provider *oidc.Provider) OidcAuthProvider {
		return &mockProvider{}
	}).Exchange(context.TODO(), &auth.ExchangeRequest{
		Code: "xx",
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, fromError.Code())
}

func TestAuthSvc_Exchange_Error1(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	fn1 := func(cfg oauth2.Config, provider *oidc.Provider) OidcAuthProvider {
		return &mockProvider{
			idtokenError: errors.New("xx"),
		}
	}
	fn2 := func(cfg oauth2.Config, provider *oidc.Provider) OidcAuthProvider {
		return &mockProvider{
			exchangeError: errors.New("ex err"),
		}
	}
	fn3 := func(cfg oauth2.Config, provider *oidc.Provider) OidcAuthProvider {
		return &mockProvider{
			verifyError: errors.New("verify err"),
		}
	}
	fnlist := []func(cfg oauth2.Config, provider *oidc.Provider) OidcAuthProvider{
		fn1, fn2, fn3,
	}
	for _, f := range fnlist {
		_, err := NewAuthSvc(nil, contracts.OidcConfig{
			"a": {
				Provider:           nil,
				Config:             oauth2.Config{},
				EndSessionEndpoint: "",
			},
			"b": {
				Provider:           nil,
				Config:             oauth2.Config{},
				EndSessionEndpoint: "",
			},
			"c": {
				Provider:           nil,
				Config:             oauth2.Config{},
				EndSessionEndpoint: "",
			},
		}, "", f).Exchange(context.TODO(), &auth.ExchangeRequest{
			Code: "xx",
		})
		fromError, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, fromError.Code())
	}
}

func TestNewDefaultAuthProvider(t *testing.T) {
	assert.Implements(t, (*OidcAuthProvider)(nil), NewDefaultAuthProvider(oauth2.Config{}, nil))
}
