package contracts

//go:generate mockgen -destination ../mock/mock_auth.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts AuthInterface

import (
	"context"
	"encoding/json"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/rbac"
	"github.com/golang-jwt/jwt"
)

const Expired = 8 * time.Hour

type Authenticator interface {
	VerifyToken(string) (*JwtClaims, bool)
}
type AuthInterface interface {
	Authenticator
	Sign(UserInfo) (*SignData, error)
}

type AuthorizeInterface interface {
	Authorize(ctx context.Context, fullMethodName string) (context.Context, error)
}

type JwtClaims struct {
	*jwt.StandardClaims
	UserInfo UserInfo `json:"user_info"`
}

type OidcClaims struct {
	LogoutUrl string `json:"logout_url"`
	OpenIDClaims
}

func (c OidcClaims) ToUserInfo() UserInfo {
	return UserInfo{
		LogoutUrl: c.LogoutUrl,
		Roles:     c.Roles,
		ID:        c.Sub,
		Email:     c.Email,
		Name:      c.Name,
		Picture:   c.Picture,
	}
}

type UserInfo struct {
	ID      string   `json:"id"`
	Email   string   `json:"email"`
	Name    string   `json:"name"`
	Picture string   `json:"picture"`
	Roles   []string `json:"roles"`

	LogoutUrl string `json:"logout_url"`
}

func (ui *UserInfo) Json() string {
	marshal, _ := json.Marshal(ui)
	return string(marshal)
}

func (ui UserInfo) GetID() string {
	return ui.ID
}

func (ui UserInfo) IsAdmin() bool {
	for _, role := range ui.Roles {
		if role == rbac.MarsAdmin {
			return true
		}
	}
	return false
}

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
