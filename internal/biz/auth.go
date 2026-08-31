package biz

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

// LoginInput 是登录请求输入。
type LoginInput struct {
	Username string
	Password string
}

// LoginResponse 是登录结果：JWT token、过期秒数与会话用户信息。
type LoginResponse struct {
	Token     string
	ExpiredIn int64

	UserInfo *UserInfo
}

const (
	// SuperAdminEmail 是内置超级管理员身份的固定邮箱，单一事实来源在 schematype 包
	// （与 UserInfo.IsSuperAdmin 判定共用，改邮箱只需动一处）。
	SuperAdminEmail = schematype.SuperAdminEmail

	// SuperAdminName 是内置超级管理员的展示名。
	SuperAdminName = "超级管理员"
)

// adminUserInfo 是内置超级管理员身份的固定用户信息（登录绕过 OIDC 直接返回）。
var adminUserInfo = &UserInfo{
	LogoutUrl: "",
	Roles:     []string{MarsAdmin},
	ID:        "1",
	Name:      SuperAdminName,
	Email:     SuperAdminEmail,
}

// AuthBiz 是认证业务逻辑的接口。
type AuthBiz interface {
	// Login 校验用户名密码并签发会话 token。
	Login(ctx context.Context, input *LoginInput) (*LoginResponse, error)
	// VerifyToken 校验会话 token 并还原用户信息。
	VerifyToken(ctx context.Context, token string) (*UserInfo, error)
	// Settings 返回已配置的 OIDC provider 映射。
	Settings(ctx context.Context) (OidcConfig, error)
	// Sign 为用户信息签发会话 token。
	Sign(ctx context.Context, input *UserInfo) (*LoginResponse, error)
	// Exchange 遍历已配置的 OIDC provider，用一次性授权码换取用户信息。
	// 全部 provider 都失败时返回 InvalidArgument，且不回显 code（一次性凭证）。
	Exchange(ctx context.Context, code string) (*UserInfo, error)
}

// AuthConfigProvider 是 AuthBiz 的配置取数窄接口：定义在消费方（biz），
// 由 data 门面实现（data→biz 依赖已存在，无循环）。只暴露 AuthBiz 需要的
// 两个配置，不把整个 data 门面塞给 biz。
type AuthConfigProvider interface {
	// AdminPassword 返回内置 admin 账号的登录密码。
	AdminPassword() string
	// OidcConfig 返回已装配的 OIDC provider 配置。
	OidcConfig() OidcConfig
}

type authBiz struct {
	auth          Auth
	adminPassword string
	oidcConfig    func() OidcConfig
	logger        mlog.Logger
}

// NewAuthBiz 构造 auth biz：注入认证器与配置（admin 密码、OIDC provider 配置）。
func NewAuthBiz(auth Auth, cfg AuthConfigProvider, logger mlog.Logger) AuthBiz {
	return &authBiz{
		auth:          auth,
		adminPassword: cfg.AdminPassword(),
		oidcConfig:    func() OidcConfig { return cfg.OidcConfig() },
		logger:        logger,
	}
}

// Login 校验 admin 账号密码后签发 token，返回登录响应。
func (a *authBiz) Login(ctx context.Context, input *LoginInput) (*LoginResponse, error) {
	if input == nil {
		return nil, errs.WrapInvalidArgument(errors.New("login input 不能为空"), "login")
	}
	if input.Username != "admin" || a.adminPassword != input.Password {
		return nil, errs.Unauthenticated("用户名或密码错误")
	}
	signData, err := a.auth.Sign(adminUserInfo)
	if err != nil {
		return nil, errs.WrapUnauthenticated(err, "auth sign")
	}

	return &LoginResponse{
		Token:     signData.Token,
		ExpiredIn: signData.ExpiredIn,
		UserInfo:  adminUserInfo,
	}, nil
}

// VerifyToken 校验 token 并返回用户信息，email 统一转小写。
func (a *authBiz) VerifyToken(ctx context.Context, token string) (*UserInfo, error) {
	verifyToken, b := a.auth.VerifyToken(token)
	if !b {
		return nil, errs.WrapUnauthenticated(errors.New("token验证失败"), "verify token")
	}
	// email 统一小写
	verifyToken.UserInfo.Email = strings.ToLower(verifyToken.UserInfo.Email)
	return verifyToken.UserInfo, nil
}

// Settings 返回已装配的 OIDC provider 配置（provider 未配置 OIDC 时返回 nil）。
func (a *authBiz) Settings(ctx context.Context) (OidcConfig, error) {
	// oidcConfig 在 NewAuthBiz 里恒为包装闭包，nil 分支不可达已删。
	return a.oidcConfig(), nil
}

// Sign 校验输入后为用户签发 token，返回登录响应。
func (a *authBiz) Sign(ctx context.Context, input *UserInfo) (*LoginResponse, error) {
	if input == nil {
		return nil, errs.WrapInvalidArgument(errors.New("sign user 不能为空"), "sign")
	}
	signData, err := a.auth.Sign(input)
	if err != nil {
		return nil, errs.Unauthenticated(err.Error())
	}

	return &LoginResponse{
		Token:     signData.Token,
		ExpiredIn: signData.ExpiredIn,
		UserInfo:  input,
	}, nil
}

// Exchange 遍历 OIDC provider 用一次性授权码换取用户信息（逻辑自 services/auth.go 下沉）。
// 单个 provider 的换发/验签/claims 解码失败均静默跳过，继续尝试下一个——biz 不打印错误日志，
// 全部失败时以 InvalidArgument 上抛，由最上层 services 统一打印（聚合结果，不回显具体 provider
// 的中间错误，避免泄露 code 相关细节）。
func (a *authBiz) Exchange(ctx context.Context, code string) (*UserInfo, error) {
	// Settings 具体实现恒返回 nil error（oidcConfig 无错误返回面），忽略之，
	// 否则该 err 分支是永远不可达的死代码。
	settings, _ := a.Settings(ctx)
	var (
		oidcClaims OidcClaims
		parsed     bool
	)
	for name, item := range settings {
		// provider 未初始化（如 OIDC 发现未完成）时跳过，避免 nil 解引用 panic。
		if item.Provider == nil {
			a.logger.WarningCtx(ctx, fmt.Sprintf("oidc provider '%s' is nil, skip", name))
			continue
		}
		p := NewDefaultAuthProvider(item.Config, item.Provider)
		token, err := p.Exchange(ctx, code)
		if err != nil {
			continue
		}
		idtoken, err := p.Verify(ctx, token)
		if err != nil {
			continue
		}
		// Verify 只校验签名/exp/iss/aud，不校验自定义 claims 的字段类型；
		// 若 provider 返回类型不匹配的 claims（如 email_verified 为字符串），
		// Claims 解码会失败，此时跳过该 provider 而不是带着脏 claims 继续。
		if err := idtoken.Claims(&oidcClaims); err != nil {
			continue
		}
		parsed = true
		oidcClaims.LogoutUrl = item.EndSessionEndpoint
		break
	}
	if !parsed {
		return nil, errs.InvalidArgument("invalid code")
	}
	return oidcClaims.ToUserInfo(), nil
}
