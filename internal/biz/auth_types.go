package biz

import (
	"context"
	"strings"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/golang-jwt/jwt"
)

// Expired 是签发的 JWT 会话 token 的过期时长。
const Expired = 12 * time.Hour

// MarsAdmin 是内置管理员角色名。
const MarsAdmin = schematype.MarsAdmin

// Authenticator 校验 JWT 并还原声明，返回声明与有效性。
type Authenticator interface {
	// VerifyToken 校验 token，合法则返回其中的 JWT claims。
	VerifyToken(string) (*JwtClaims, bool)
}

// Auth 组合 token 校验与签发能力。
type Auth interface {
	Authenticator
	// Sign 为用户信息签发 JWT token。
	Sign(*UserInfo) (*SignData, error)
}

// JwtClaims 是 JWT 载荷结构：标准声明 + 内嵌用户信息。
type JwtClaims struct {
	*jwt.StandardClaims
	UserInfo *UserInfo `json:"user_info"`
}

// OidcClaims 是 OIDC 用户信息声明：标准字段 + 登出地址。
type OidcClaims struct {
	LogoutUrl string `json:"logout_url"`
	OpenIDClaims
}

// ToUserInfo 把 OIDC claims 转换为用户信息，email 统一转小写。
func (c OidcClaims) ToUserInfo() *UserInfo {
	return &UserInfo{
		LogoutUrl: c.LogoutUrl,
		Roles:     c.Roles,
		ID:        c.Sub,
		Email:     strings.ToLower(c.Email),
		Name:      c.Name,
		Picture:   c.Picture,
	}
}

// UserInfo 是登录用户信息模型（复用 schematype 定义）。
type UserInfo = schematype.UserInfo

// OpenIDClaims 是 OpenID Connect 标准声明集合。
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
	CustomAttributes    map[string]any `json:"custom_attributes"`
	Address             map[string]any `json:"address"`
	UpdatedAt           int            `json:"updated_at"`

	Roles []string `json:"roles"`
}

// SignData 是 JWT 签发结果：token 原文与过期秒数。
type SignData struct {
	Token     string
	ExpiredIn int64
}

// TokenManager 是 access token 校验接口。
type TokenManager interface {
	// VerifyAndTouch 校验 access token 并回写最近使用时间，返回对应用户。
	VerifyAndTouch(ctx context.Context, token string, now time.Time) (*UserInfo, bool)
}
