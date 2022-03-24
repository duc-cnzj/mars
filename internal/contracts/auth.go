package contracts

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt"
)

const Expired = 8 * time.Hour

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
}

type SignData struct {
	Token     string
	ExpiredIn int64
}

type AuthInterface interface {
	VerifyToken(string) (*JwtClaims, bool)
	Sign(UserInfo) (*SignData, error)
}

type AuthorizeInterface interface {
	Authorize(ctx context.Context, fullMethodName string) (context.Context, error)
}
