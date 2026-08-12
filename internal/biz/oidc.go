package biz

// oidc.go 定义 OIDC 授权码换发用例所需的 provider 抽象：
// OidcConfigItem 带 Provider/Config 配置，OIDC 类型归属 biz，transport 不应持
// oidc/oauth2 具体类型。defaultAuthProvider 是核心库适配器，业务层只依赖
// OidcAuthProvider 端口，便于单测替换。

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"golang.org/x/oauth2"
)

// idToken 是 OIDC id_token 的解码端口：核心库的 *oidc.IDToken 天然实现它。
type idToken interface {
	// Claims 把 id_token 载荷解码进 v（核心库 Claims 语义，透传）。
	Claims(v any) error
}

// OidcAuthProvider 封装 OIDC 授权码换发两步：用 code 换原始 id_token，再验签。
type OidcAuthProvider interface {
	// Exchange 用授权码向 IdP 换发令牌，返回原始 id_token 字符串。
	Exchange(ctx context.Context, code string) (string, error)
	// Verify 校验 id_token 签名/exp/iss/aud 并返回解码端口。
	Verify(ctx context.Context, token string) (idToken, error)
}

var _ OidcAuthProvider = (*defaultAuthProvider)(nil)

// defaultAuthProvider 是 oidc/oauth2 核心库的适配器实现。
type defaultAuthProvider struct {
	cfg      oauth2.Config
	provider *oidc.Provider
}

// NewDefaultAuthProvider 以 OIDC 客户端配置与 provider 构造适配器。
func NewDefaultAuthProvider(cfg oauth2.Config, provider *oidc.Provider) OidcAuthProvider {
	return &defaultAuthProvider{cfg: cfg, provider: provider}
}

// Exchange 用授权码向 IdP 换发令牌，返回原始 id_token 字符串。
func (d *defaultAuthProvider) Exchange(ctx context.Context, code string) (string, error) {
	token, err := d.cfg.Exchange(ctx, code)
	if err != nil {
		return "", err
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		// code 是一次性 OAuth2 凭证，不回显进错误信息，理由同 Exchange 方法。
		return "", errs.InvalidArgument("bad code")
	}
	return rawIDToken, nil
}

// Verify 校验 id_token 签名/exp/iss/aud 并返回解码端口。
func (d *defaultAuthProvider) Verify(ctx context.Context, token string) (idToken, error) {
	return d.provider.Verifier(&oidc.Config{ClientID: d.cfg.ClientID}).Verify(ctx, token)
}
