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
		auth.RegisterAuthServer(s, NewAuthSvc(app.Auth(), app.Oidc(), app.Config().AdminPassword))
	})
	RegisterEndpoint(auth.RegisterAuthHandlerFromEndpoint)
}

type AuthSvc struct {
	cfg      contracts.OidcConfig
	adminPwd string
	authsvc  contracts.AuthInterface
	auth.UnimplementedAuthServer
}

func NewAuthSvc(authsvc contracts.AuthInterface, cfg contracts.OidcConfig, adminPwd string) *AuthSvc {
	return &AuthSvc{authsvc: authsvc, cfg: cfg, adminPwd: adminPwd}
}

func verify(cfg oauth2.Config, provider *oidc.Provider, code string) (*oidc.IDToken, error) {
	var (
		token *oauth2.Token
		err   error
	)
	if token, err = cfg.Exchange(context.TODO(), code); err != nil {
		return nil, err
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "bad code: "+code)
	}

	idtoken, err := verifier.Verify(context.TODO(), rawIDToken)
	if err != nil {
		return nil, err
	}

	return idtoken, nil
}

func (a *AuthSvc) Login(ctx context.Context, request *auth.AuthLoginRequest) (*auth.AuthLoginResponse, error) {
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
		return &auth.AuthLoginResponse{
			Token:     data.Token,
			ExpiresIn: data.ExpiredIn,
		}, nil
	}

	return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated.")
}

func (a *AuthSvc) Info(ctx context.Context, req *auth.AuthInfoRequest) (*auth.AuthInfoResponse, error) {
	incomingContext, ok := metadata.FromIncomingContext(ctx)
	if ok {
		tokenSlice := incomingContext.Get("Authorization")
		if len(tokenSlice) == 1 {
			if c, b := a.authsvc.VerifyToken(tokenSlice[0]); b {
				return &auth.AuthInfoResponse{
					Id:        c.GetID(),
					Avatar:    c.Picture,
					Name:      c.Name,
					Email:     c.Email,
					LogoutUrl: c.LogoutUrl,
					Roles:     c.Roles,
				}, nil
			}
		}
	}

	return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated.")
}

func (a *AuthSvc) Settings(ctx context.Context, request *auth.AuthSettingsRequest) (*auth.AuthSettingsResponse, error) {
	var items = make([]*auth.AuthSettingsResponse_OidcSetting, 0, len(a.cfg))
	for name, setting := range a.cfg {
		state := utils.RandomString(32)

		items = append(items, &auth.AuthSettingsResponse_OidcSetting{
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

	return &auth.AuthSettingsResponse{Items: items}, nil
}

func (a *AuthSvc) Exchange(ctx context.Context, request *auth.AuthExchangeRequest) (*auth.AuthExchangeResponse, error) {
	var (
		idtoken  *oidc.IDToken
		err      error
		userinfo contracts.UserInfo
		parsed   bool
	)

	for _, item := range a.cfg {
		if idtoken, err = verify(item.Config, item.Provider, request.Code); err != nil {
			continue
		}
		if err := idtoken.Claims(&userinfo); err != nil {
			return nil, err
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

	return &auth.AuthExchangeResponse{
		Token:     data.Token,
		ExpiresIn: data.ExpiredIn,
	}, nil
}

func (a *AuthSvc) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	mlog.Debug("client is calling method:", fullMethodName)
	return ctx, nil
}
