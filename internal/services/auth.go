package services

import (
	"context"
	"fmt"
	"sort"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/duc-cnzj/mars/api/v5/auth"
	"github.com/duc-cnzj/mars/api/v5/types"
	auth2 "github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/util/rand"
	"github.com/spf13/cast"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var _ auth.AuthServer = (*authSvc)(nil)

type authSvc struct {
	auth.UnimplementedAuthServer

	guest

	logger    mlog.Logger
	authRepo  repo.AuthRepo
	eventRepo repo.EventRepo
}

func NewAuthSvc(eventRepo repo.EventRepo, logger mlog.Logger, authRepo repo.AuthRepo) auth.AuthServer {
	return &authSvc{
		logger:    logger.WithModule("services/auth"),
		eventRepo: eventRepo,
		authRepo:  authRepo,
	}
}

func (a *authSvc) Login(ctx context.Context, request *auth.LoginRequest) (*auth.LoginResponse, error) {
	loginResp, err := a.authRepo.Login(ctx, &repo.LoginInput{
		Username: request.Username,
		Password: request.Password,
	})
	if err != nil {
		return nil, err
	}

	a.eventRepo.AuditLog(
		types.EventActionType_Login,
		loginResp.UserInfo.Name,
		fmt.Sprintf("用户 '%s' email: '%s' 登录了系统", loginResp.UserInfo.Name, loginResp.UserInfo.Email),
	)

	return &auth.LoginResponse{
		Token:     loginResp.Token,
		ExpiresIn: loginResp.ExpiredIn,
	}, nil
}

func (a *authSvc) Info(ctx context.Context, req *auth.InfoRequest) (*auth.InfoResponse, error) {
	incomingContext, ok := metadata.FromIncomingContext(ctx)
	if ok {
		tokenSlice := incomingContext.Get("Authorization")
		if len(tokenSlice) == 1 {
			if c, err := a.authRepo.VerifyToken(ctx, tokenSlice[0]); err == nil {
				return &auth.InfoResponse{
					Id:        cast.ToInt32(c.ID),
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

func (a *authSvc) Settings(ctx context.Context, request *auth.SettingsRequest) (*auth.SettingsResponse, error) {
	settings, _ := a.authRepo.Settings(ctx)
	var items = make([]*auth.SettingsResponse_OidcSetting, 0, len(settings))
	for name, setting := range settings {
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

	settings, _ := a.authRepo.Settings(ctx)
	for _, item := range settings {
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

	data, err := a.authRepo.Sign(ctx, userinfo)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	a.eventRepo.AuditLogWithRequest(
		types.EventActionType_Login,
		userinfo.Name,
		fmt.Sprintf("用户 '%s' email: '%s' 登录了系统", userinfo.Name, userinfo.Email),
		request,
	)

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

var ErrorPermissionDenied = repo.ToError(403, "没有权限执行该操作")

var MustGetUser = auth2.MustGetUser

type guest struct{}

func (c guest) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	//mlog.Debug("client is calling method:", fullMethodName)
	return ctx, nil
}
