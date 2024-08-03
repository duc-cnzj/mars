package services

import (
	"context"
	"fmt"
	"sort"

	"github.com/duc-cnzj/mars/v4/internal/ent/schema/schematype"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/duc-cnzj/mars/api/v4/auth"
	"github.com/duc-cnzj/mars/api/v4/types"
	auth2 "github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/util/rand"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var _ auth.AuthServer = (*authSvc)(nil)

type authSvc struct {
	guest

	cfg      data.OidcConfig
	adminPwd string
	authsvc  auth2.Auth
	logger   mlog.Logger

	auth.UnimplementedAuthServer
	eventRepo repo.EventRepo
}

func NewAuthSvc(eventRepo repo.EventRepo, logger mlog.Logger, authsvc auth2.Auth, data *data.Data) auth.AuthServer {
	return &authSvc{
		logger:    logger,
		eventRepo: eventRepo,
		cfg:       data.Oidc,
		adminPwd:  data.Cfg.AdminPassword,
		authsvc:   authsvc,
	}
}

func (a *authSvc) Login(ctx context.Context, request *auth.LoginRequest) (*auth.LoginResponse, error) {
	if request.Username == "admin" && request.Password == a.adminPwd {
		userinfo := &auth2.UserInfo{
			LogoutUrl: "",
			Roles:     []string{schematype.MarsAdmin},
			ID:        "1",
			Name:      "管理员",
			Email:     "1025434218@qq.com",
		}
		data, err := a.authsvc.Sign(userinfo)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, err.Error())
		}
		a.eventRepo.AuditLog(types.EventActionType_Login, userinfo.Name, fmt.Sprintf("用户 '%s' email: '%s' 登录了系统", userinfo.Name, userinfo.Email))

		return &auth.LoginResponse{
			Token:     data.Token,
			ExpiresIn: data.ExpiredIn,
		}, nil
	}

	return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated.")
}

func (a *authSvc) Info(ctx context.Context, req *auth.InfoRequest) (*auth.InfoResponse, error) {
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

func (a *authSvc) Settings(ctx context.Context, request *auth.SettingsRequest) (*auth.SettingsResponse, error) {
	var items = make([]*auth.SettingsResponse_OidcSetting, 0, len(a.cfg))
	for name, setting := range a.cfg {
		state := rand.String(32)

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

func (a *authSvc) Exchange(ctx context.Context, request *auth.ExchangeRequest) (*auth.ExchangeResponse, error) {
	var (
		logger     = a.logger
		oidcClaims auth2.OidcClaims
		parsed     bool
	)

	for _, item := range a.cfg {
		var (
			token   string
			err     error
			idtoken idToken
		)
		p := NewDefaultAuthProvider(item.Config, item.Provider)
		token, err = p.Exchange(context.TODO(), request.Code)
		if err != nil {
			logger.Error(err)
			continue
		}
		if idtoken, err = p.Verify(context.TODO(), token); err != nil {
			logger.Error(err)
			continue
		}
		if err = idtoken.Claims(&oidcClaims); err != nil {
			logger.Error(err)
			continue
		}
		parsed = true
		oidcClaims.LogoutUrl = item.EndSessionEndpoint
	}

	if !parsed {
		return nil, status.Errorf(codes.InvalidArgument, "invalid code: "+request.Code)
	}

	userinfo := oidcClaims.ToUserInfo()

	logger.Debug(userinfo)
	data, err := a.authsvc.Sign(userinfo)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	a.eventRepo.AuditLog(types.EventActionType_Login, userinfo.Name, fmt.Sprintf("用户 '%s' email: '%s' 登录了系统", userinfo.Name, userinfo.Email))

	return &auth.ExchangeResponse{
		Token:     data.Token,
		ExpiresIn: data.ExpiredIn,
	}, nil
}

type idToken interface {
	Claims(any) error
}

type OidcAuthProvider interface {
	Exchange(ctx context.Context, code string) (string, error)
	Verify(ctx context.Context, token string) (idToken, error)
}

var _ OidcAuthProvider = (*defaultAuthProvider)(nil)

type defaultAuthProvider struct {
	cfg      oauth2.Config
	provider *oidc.Provider
}

func NewDefaultAuthProvider(cfg oauth2.Config, provider *oidc.Provider) OidcAuthProvider {
	return &defaultAuthProvider{cfg: cfg, provider: provider}
}

func (d *defaultAuthProvider) Exchange(ctx context.Context, code string) (string, error) {
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

func (d *defaultAuthProvider) Verify(ctx context.Context, token string) (idToken, error) {
	return d.provider.Verifier(&oidc.Config{ClientID: d.cfg.ClientID}).Verify(ctx, token)
}
