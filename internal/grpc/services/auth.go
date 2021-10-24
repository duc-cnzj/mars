package services

import (
	"context"
	"crypto/rsa"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/pkg/auth"
	"github.com/golang-jwt/jwt"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Auth struct {
	priKey *rsa.PrivateKey
	pubKey *rsa.PublicKey
	cfg    *contracts.OidcConfig
	auth.UnimplementedAuthServer
}

func NewAuth(priKey *rsa.PrivateKey, pubKey *rsa.PublicKey, cfg *contracts.OidcConfig) *Auth {
	return &Auth{priKey: priKey, pubKey: pubKey, cfg: cfg}
}

type JwtClaims struct {
	*jwt.StandardClaims
	UserInfo
}

type UserInfo struct {
	LogoutUrl string `json:"logout_url"`
	Sub       string `json:"sub"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Username  string `json:"username"`
}

func (ui UserInfo) GetID() int64 {
	atoi, _ := strconv.Atoi(ui.Sub)
	return int64(atoi)
}

const Expired = 8 * time.Hour

func (a *Auth) Exchange(ctx context.Context, request *auth.ExchangeRequest) (*auth.LoginResponse, error) {
	var (
		idtoken *oauth2.Token
		err     error
	)
	if idtoken, err = a.cfg.Config.Exchange(context.TODO(), request.Code); err != nil {
		return nil, err
	}
	verifier := a.cfg.Provider.Verifier(&oidc.Config{ClientID: a.cfg.Config.ClientID})
	rawIDToken, ok := idtoken.Extra("id_token").(string)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "bad code: "+request.Code)
	}

	verify, err := verifier.Verify(context.TODO(), rawIDToken)
	if err != nil {
		return nil, err
	}

	var userinfo UserInfo
	verify.Claims(&userinfo)
	userinfo.LogoutUrl = a.cfg.EndSessionEndpoint

	mlog.Debug(userinfo)
	tokenString, err := a.sign(userinfo)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &auth.LoginResponse{
		Token:     tokenString,
		ExpiresIn: int64(Expired),
	}, nil
}

func (a *Auth) sign(userinfo UserInfo) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &JwtClaims{
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: time.Now().Add(Expired).Unix(),
			Issuer:    "mars",
		},
		UserInfo: userinfo,
	})

	return token.SignedString(a.priKey)
}

func (a *Auth) Settings(ctx context.Context, empty *emptypb.Empty) (*auth.SettingsResponse, error) {
	if app.Config().OidcEnabled {
		state := utils.RandomString(32)

		return &auth.SettingsResponse{
			SsoEnabled:         true,
			Url:                a.cfg.Config.AuthCodeURL(state),
			EndSessionEndpoint: a.cfg.EndSessionEndpoint,
			State:              state,
		}, nil
	}

	return &auth.SettingsResponse{}, nil
}

func (a *Auth) Login(ctx context.Context, request *auth.LoginRequest) (*auth.LoginResponse, error) {
	if request.Username == "admin" && request.Password == app.Config().AdminPassword {
		tokenString, err := a.sign(UserInfo{
			Sub:      "1",
			Name:     "管理员",
			Email:    "admin@mars.com",
			Username: "admin",
		})
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, err.Error())
		}
		return &auth.LoginResponse{
			Token:     tokenString,
			ExpiresIn: int64(Expired),
		}, nil
	}

	return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated.")
}

func (a *Auth) Info(ctx context.Context, empty *emptypb.Empty) (*auth.InfoResponse, error) {
	incomingContext, ok := metadata.FromIncomingContext(ctx)
	if ok {
		tokenSlice := incomingContext.Get("Authorization")
		if len(tokenSlice) == 1 {
			token := strings.TrimSpace(strings.TrimLeft(tokenSlice[0], "Bearer"))
			parse, err := jwt.ParseWithClaims(token, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
				return a.pubKey, nil
			})
			if err == nil && parse.Valid {
				c, ok := parse.Claims.(*JwtClaims)
				if ok {
					return &auth.InfoResponse{
						Id:        c.GetID(),
						Username:  c.Username,
						Name:      c.Name,
						Email:     c.Email,
						LogoutUrl: c.LogoutUrl,
					}, nil
				}
			}
		}
	}

	return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated.")
}

func (a *Auth) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	mlog.Debug("client is calling method:", fullMethodName)
	return ctx, nil
}
