package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"strings"
	"time"

	"github.com/duc-cnzj/mars/v5/internal/util/timer"

	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/ent/accesstoken"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/schematype"
	"github.com/golang-jwt/jwt"
)

const Expired = 12 * time.Hour

type Authenticator interface {
	VerifyToken(string) (*JwtClaims, bool)
}
type Auth interface {
	Authenticator
	Sign(*UserInfo) (*SignData, error)
}

type JwtClaims struct {
	*jwt.StandardClaims
	UserInfo *UserInfo `json:"user_info"`
}

type OidcClaims struct {
	LogoutUrl string `json:"logout_url"`
	OpenIDClaims
}

func (c OidcClaims) ToUserInfo() *UserInfo {
	return &UserInfo{
		LogoutUrl: c.LogoutUrl,
		Roles:     c.Roles,
		ID:        c.Sub,
		Email:     c.Email,
		Name:      c.Name,
		Picture:   c.Picture,
	}
}

type UserInfo = schematype.UserInfo

type OpenIDClaims struct {
	Sub                 string         `json:"sub"`
	Name                string         `json:"name"`
	GivenName           string         `json:"given_name"`
	FamilyName          string         `json:"family_name"`
	MiddleName          string         `json:"middle_name"`
	Nickname            string         `json:"nickname"`
	PreferredUsername   string         `json:"preferred_username"`
	Profile             string         `json:"profile"`
	Picture             string         `json:"picture"`
	Website             string         `json:"website"`
	Email               string         `json:"email"`
	EmailVerified       bool           `json:"email_verified"`
	Gender              string         `json:"gender"`
	Birthdate           string         `json:"birthdate"`
	Zoneinfo            string         `json:"zoneinfo"`
	Locale              string         `json:"locale"`
	PhoneNumber         string         `json:"phone_number"`
	PhoneNumberVerified bool           `json:"phone_number_verified"`
	Address             map[string]any `json:"address"`
	UpdatedAt           int            `json:"updated_at"`

	// Roles 自定义权限
	Roles []string `json:"roles"`
}

type SignData struct {
	Token     string
	ExpiredIn int64
}

var _ Auth = (*Authn)(nil)

type Authn struct {
	Authns   []Authenticator
	signFunc func(info *UserInfo) (*SignData, error)
}

func NewAuthn(data data.Data, timer timer.Timer) (Auth, error) {
	pem, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(data.Config().PrivateKey))
	if err != nil {
		return nil, err
	}
	auth := NewJwtAuth(pem, pem.Public().(*rsa.PublicKey), timer)

	return &Authn{
		Authns: []Authenticator{
			auth,
			NewAccessTokenAuth(data, timer),
		},
		signFunc: auth.Sign,
	}, nil
}

func (a *Authn) VerifyToken(s string) (*JwtClaims, bool) {
	for _, authn := range a.Authns {
		if token, ok := authn.VerifyToken(s); ok {
			return token, true
		}
	}

	return nil, false
}

func (a *Authn) Sign(info *UserInfo) (*SignData, error) {
	return a.signFunc(info)
}

type jwtAuth struct {
	priKey *rsa.PrivateKey
	pubKey *rsa.PublicKey
	timer  timer.Timer
}

func NewJwtAuth(priKey *rsa.PrivateKey, pubKey *rsa.PublicKey, timer timer.Timer) Auth {
	return &jwtAuth{priKey: priKey, pubKey: pubKey, timer: timer}
}

func (a *jwtAuth) VerifyToken(t string) (*JwtClaims, bool) {
	var token string = t
	if len(t) > 6 && strings.EqualFold("bearer", t[0:6]) {
		token = strings.TrimSpace(t[6:])
	}
	if token != "" {
		parse, err := jwt.ParseWithClaims(token, &JwtClaims{}, func(token *jwt.Token) (any, error) {
			return a.pubKey, nil
		})
		if err == nil && parse.Valid {
			return parse.Claims.(*JwtClaims), true
		}
	}

	return nil, false
}

func (a *jwtAuth) Sign(info *UserInfo) (*SignData, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &JwtClaims{
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: a.timer.Now().Add(Expired).Unix(),
			Issuer:    "mars",
			IssuedAt:  a.timer.Now().Unix(),
			Subject:   info.Email,
		},
		UserInfo: info,
	})

	signedString, err := token.SignedString(a.priKey)
	if err != nil {
		return nil, err
	}
	return &SignData{
		Token:     signedString,
		ExpiredIn: int64(Expired.Seconds()),
	}, nil
}

type accessTokenAuth struct {
	data  data.Data
	timer timer.Timer
}

func NewAccessTokenAuth(data data.Data, timer timer.Timer) Authenticator {
	return &accessTokenAuth{data: data, timer: timer}
}

func (a *accessTokenAuth) VerifyToken(t string) (*JwtClaims, bool) {
	var token string = t
	if len(t) > 6 && strings.EqualFold("bearer", t[0:6]) {
		token = strings.TrimSpace(t[6:])
	}
	if token != "" {
		if first, err := a.data.DB().AccessToken.Query().Where(accesstoken.Token(token)).First(context.TODO()); err == nil {
			first.Update().SetLastUsedAt(a.timer.Now()).Save(context.TODO())
			return &JwtClaims{UserInfo: &first.UserInfo}, true
		}
	}

	return nil, false
}

type ctxTokenInfo struct{}

func SetUser(ctx context.Context, info *UserInfo) context.Context {
	return context.WithValue(ctx, &ctxTokenInfo{}, info)
}

func GetUser(ctx context.Context) (*UserInfo, error) {
	if info, ok := ctx.Value(&ctxTokenInfo{}).(*UserInfo); ok {
		return info, nil
	}

	return nil, errors.New("user not found")
}

func MustGetUser(ctx context.Context) *UserInfo {
	info, _ := ctx.Value(&ctxTokenInfo{}).(*UserInfo)
	return info
}
