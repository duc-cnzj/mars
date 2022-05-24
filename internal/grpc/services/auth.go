package services

import (
	"context"
	"sort"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/duc-cnzj/mars-client/v4/auth"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/utils"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		auth.RegisterAuthServer(s, NewAuthSvc(app.Auth(), app.Oidc(), app.Config().AdminPassword, NewDefaultAuthProvider))
	})
	RegisterEndpoint(auth.RegisterAuthHandlerFromEndpoint)
}

type AuthSvc struct {
	cfg             contracts.OidcConfig
	adminPwd        string
	authsvc         contracts.AuthInterface
	NewProviderFunc func(cfg oauth2.Config, provider *oidc.Provider) OidcAuthProvider

	auth.UnimplementedAuthServer
}

func NewAuthSvc(authsvc contracts.AuthInterface, cfg contracts.OidcConfig, adminPwd string, NewProviderFunc func(cfg oauth2.Config, provider *oidc.Provider) OidcAuthProvider) *AuthSvc {
	return &AuthSvc{authsvc: authsvc, cfg: cfg, adminPwd: adminPwd, NewProviderFunc: NewProviderFunc}
}

type OidcAuthProvider interface {
	Exchange(ctx context.Context, code string) (string, error)
	Verify(ctx context.Context, token string) (IDToken, error)
}

type IDToken interface {
	Claims(any) error
}

func (a *AuthSvc) Login(ctx context.Context, request *auth.LoginRequest) (*auth.LoginResponse, error) {
	if request.Username == "admin" && request.Password == a.adminPwd {
		data, err := a.authsvc.Sign(contracts.UserInfo{
			LogoutUrl: "",
			Roles:     []string{"admin"},
			OpenIDClaims: contracts.OpenIDClaims{
				Sub:   "1",
				Name:  "管理员",
				Email: "admin@mars.com",
			},
		})
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, err.Error())
		}
		return &auth.LoginResponse{
			Token:     data.Token,
			ExpiresIn: data.ExpiredIn,
		}, nil
	}

	return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated.")
}

func (a *AuthSvc) Info(ctx context.Context, req *auth.InfoRequest) (*auth.InfoResponse, error) {
	incomingContext, ok := metadata.FromIncomingContext(ctx)
	if ok {
		tokenSlice := incomingContext.Get("Authorization")
		if len(tokenSlice) == 1 {
			if c, b := a.authsvc.VerifyToken(tokenSlice[0]); b {
				return &auth.InfoResponse{
					Id:        c.StandardClaims.Subject,
					Avatar:    c.UserInfo.Picture,
					Name:      c.UserInfo.Name,
					Email:     c.UserInfo.Email,
					LogoutUrl: c.UserInfo.LogoutUrl,
					Roles:     c.UserInfo.Roles,
				}, nil
			}
		}
	}

	return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated.")
}

func (a *AuthSvc) Settings(ctx context.Context, request *auth.SettingsRequest) (*auth.SettingsResponse, error) {
	var items = make([]*auth.SettingsResponse_OidcSetting, 0, len(a.cfg))
	for name, setting := range a.cfg {
		state := utils.RandomString(32)

		items = append(items, &auth.SettingsResponse_OidcSetting{
			Enabled:            true,
			Name:               name,
			Url:                setting.Config.AuthCodeURL(state),
			EndSessionEndpoint: setting.EndSessionEndpoint,
			State:              state,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return &auth.SettingsResponse{Items: items}, nil
}

func (a *AuthSvc) Exchange(ctx context.Context, request *auth.ExchangeRequest) (*auth.ExchangeResponse, error) {
	var (
		userinfo contracts.UserInfo
		parsed   bool
	)

	for _, item := range a.cfg {
		var (
			token   string
			err     error
			idtoken IDToken
		)
		p := a.NewProviderFunc(item.Config, item.Provider)
		token, err = p.Exchange(context.TODO(), request.Code)
		if err != nil {
			mlog.Debug(err)
			continue
		}
		if idtoken, err = p.Verify(context.TODO(), token); err != nil {
			mlog.Debug(err)
			continue
		}
		if err = idtoken.Claims(&userinfo); err != nil {
			mlog.Debug(err)
			continue
		}
		parsed = true
		userinfo.LogoutUrl = item.EndSessionEndpoint
	}

	if !parsed {
		return nil, status.Errorf(codes.InvalidArgument, "invalid code: "+request.Code)
	}

	userinfo.Roles = []string{}

	mlog.Debug(userinfo)
	data, err := a.authsvc.Sign(userinfo)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &auth.ExchangeResponse{
		Token:     data.Token,
		ExpiresIn: data.ExpiredIn,
	}, nil
}

func (a *AuthSvc) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	mlog.Debug("client is calling method:", fullMethodName)
	return ctx, nil
}

type DefaultAuthProvider struct {
	cfg      oauth2.Config
	provider *oidc.Provider
}

func NewDefaultAuthProvider(cfg oauth2.Config, provider *oidc.Provider) OidcAuthProvider {
	return &DefaultAuthProvider{cfg: cfg, provider: provider}
}

func (d *DefaultAuthProvider) Exchange(ctx context.Context, code string) (string, error) {
	token, err := d.cfg.Exchange(ctx, code)
	if err != nil {
		return "", err
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return "", status.Errorf(codes.InvalidArgument, "bad code: "+code)
	}
	return rawIDToken, nil
}

func (d *DefaultAuthProvider) Verify(ctx context.Context, token string) (IDToken, error) {
	return d.provider.Verifier(&oidc.Config{ClientID: d.cfg.ClientID}).Verify(ctx, token)
}
