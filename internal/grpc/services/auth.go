package services

import (
	"context"
	"crypto/rsa"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
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
	priKey   *rsa.PrivateKey
	pubKey   *rsa.PublicKey
	cfg      contracts.OidcConfig
	adminPwd string
	auth.UnimplementedAuthServer
}

func NewAuth(priKey *rsa.PrivateKey, pubKey *rsa.PublicKey, cfg contracts.OidcConfig, adminPwd string) *Auth {
	return &Auth{priKey: priKey, pubKey: pubKey, cfg: cfg, adminPwd: adminPwd}
}

type JwtClaims struct {
	*jwt.StandardClaims
	UserInfo
}

type UserInfo struct {
	LogoutUrl string   `json:"logout_url"`
	Roles     []string `json:"roles"`

	OpenIDClaims
}

func (ui UserInfo) GetID() string {
	return ui.Sub
}

func (ui UserInfo) IsAdmin() bool {
	for _, role := range ui.Roles {
		if role == "admin" {
			return true
		}
	}
	return false
}

const Expired = 8 * time.Hour

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
func (a *Auth) Exchange(ctx context.Context, request *auth.ExchangeRequest) (*auth.LoginResponse, error) {
	var (
		idtoken  *oidc.IDToken
		err      error
		userinfo UserInfo
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
	var items = make([]*auth.OidcSetting, 0, len(a.cfg))
	for name, setting := range a.cfg {
		state := utils.RandomString(32)

		items = append(items, &auth.OidcSetting{
			Enabled:            true,
			Name:               name,
			Url:                setting.Config.AuthCodeURL(state),
			EndSessionEndpoint: setting.EndSessionEndpoint,
			State:              state,
		})
	}

	return &auth.SettingsResponse{Items: items}, nil
}

func (a *Auth) Login(ctx context.Context, request *auth.LoginRequest) (*auth.LoginResponse, error) {
	if request.Username == "admin" && request.Password == a.adminPwd {
		tokenString, err := a.sign(UserInfo{
			LogoutUrl: "",
			Roles:     []string{"admin"},
			OpenIDClaims: OpenIDClaims{
				Sub:   "1",
				Name:  "管理员",
				Email: "admin@mars.com",
			},
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
						Avatar:    c.Picture,
						Name:      c.Name,
						Email:     c.Email,
						LogoutUrl: c.LogoutUrl,
						Roles:     c.Roles,
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

type OpenIDClaims struct {
	Sub                 string                 `json:"sub"`
	Name                string                 `json:"name"`
	GivenName           string                 `json:"given_name"`
	FamilyName          string                 `json:"family_name"`
	MiddleName          string                 `json:"middle_name"`
	Nickname            string                 `json:"nickname"`
	PreferredUsername   string                 `json:"preferred_username"`
	Profile             string                 `json:"profile"`
	Picture             string                 `json:"picture"`
	Website             string                 `json:"website"`
	Email               string                 `json:"email"`
	EmailVerified       bool                   `json:"email_verified"`
	Gender              string                 `json:"gender"`
	Birthdate           string                 `json:"birthdate"`
	Zoneinfo            string                 `json:"zoneinfo"`
	Locale              string                 `json:"locale"`
	PhoneNumber         string                 `json:"phone_number"`
	PhoneNumberVerified bool                   `json:"phone_number_verified"`
	Address             map[string]interface{} `json:"address"`
	UpdatedAt           int                    `json:"updated_at"`
}
