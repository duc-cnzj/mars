package data

import (
	"context"
	"crypto/rsa"
	"errors"
	"strings"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/golang-jwt/jwt"
)

// authn 是 biz.Auth 的组合实现：把 JWT 验签与 TokenManager（DB 访问令牌）两类
// Authenticator 合并为一次 VerifyToken 的兜底链，签名委托给 jwtAuth。
// 它只被 biz.NewAuthBiz 消费——是鉴权核心的内部细节，不再上浮到应用门面。
type authn struct {
	authns   []biz.Authenticator
	signFunc func(info *biz.UserInfo) (*biz.SignData, error)
}

var _ biz.Auth = (*authn)(nil)

// NewAuthn 从配置中的 RSA 私钥与 TokenManager 构造 biz.Auth：私钥用于签发，公钥
// 验签，TokenManager 作为访问令牌的兜底校验源。wire 经 WireDataSet 提供。
func NewAuthn(tm biz.TokenManager, cfg *config.Config, timer timer.Timer) (biz.Auth, error) {
	pem, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(cfg.PrivateKey))
	if err != nil {
		return nil, errs.Wrap(err, "parse rsa private key")
	}
	auth := newJwtAuth(pem, pem.Public().(*rsa.PublicKey), timer)

	return &authn{
		authns: []biz.Authenticator{
			auth,
			newTokenManagerAuth(tm, timer),
		},
		signFunc: auth.Sign,
	}, nil
}

// VerifyToken 依次尝试各 Authenticator，任一通过即返回 claims。
func (a *authn) VerifyToken(s string) (*biz.JwtClaims, bool) {
	for _, authn := range a.authns {
		if token, ok := authn.VerifyToken(s); ok {
			return token, true
		}
	}
	return nil, false
}

// Sign 委托给 jwtAuth 的签名实现。
func (a *authn) Sign(info *biz.UserInfo) (*biz.SignData, error) {
	return a.signFunc(info)
}

// stripBearer 去掉 "bearer " 前缀（大小写不敏感，前缀后可有空格），非 bearer 原样返回。
func stripBearer(t string) string {
	if len(t) > 6 && strings.EqualFold("bearer", t[0:6]) {
		return strings.TrimSpace(t[6:])
	}
	return t
}

// jwtAuth 是 RS256 JWT 的签发与验签实现。
type jwtAuth struct {
	priKey *rsa.PrivateKey
	pubKey *rsa.PublicKey
	timer  timer.Timer
}

// newJwtAuth 用 RSA 公私钥与计时器构造 jwtAuth 实例。
func newJwtAuth(priKey *rsa.PrivateKey, pubKey *rsa.PublicKey, timer timer.Timer) *jwtAuth {
	return &jwtAuth{priKey: priKey, pubKey: pubKey, timer: timer}
}

// VerifyToken 校验 RS256 签名并容忍 "bearer " 前缀（大小写不敏感）。
func (a *jwtAuth) VerifyToken(t string) (*biz.JwtClaims, bool) {
	if token := stripBearer(t); token != "" {
		parse, err := jwt.ParseWithClaims(token, &biz.JwtClaims{}, func(token *jwt.Token) (any, error) {
			// 显式限定 RS256，防止 alg confusion（如 HS256 用公钥当密钥）。
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return a.pubKey, nil
		})
		if err == nil && parse.Valid {
			return parse.Claims.(*biz.JwtClaims), true
		}
	}
	return nil, false
}

// Sign 用 RSA 私钥签发 RS256 JWT，过期时间对齐 biz.Expired。
func (a *jwtAuth) Sign(info *biz.UserInfo) (*biz.SignData, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &biz.JwtClaims{
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: a.timer.Now().Add(biz.Expired).Unix(),
			Issuer:    "mars",
			IssuedAt:  a.timer.Now().Unix(),
			Subject:   info.Email,
		},
		UserInfo: info,
	})

	signedString, err := token.SignedString(a.priKey)
	if err != nil {
		return nil, errs.Wrap(err, "sign jwt token")
	}
	return &biz.SignData{
		Token:     signedString,
		ExpiredIn: int64(biz.Expired.Seconds()),
	}, nil
}

// tokenManagerAuth 把 biz.TokenManager（DB 访问令牌校验）适配为 Authenticator。
type tokenManagerAuth struct {
	tm    biz.TokenManager
	timer timer.Timer
}

// newTokenManagerAuth 用 TokenManager 构造其 Authenticator 适配实现。
func newTokenManagerAuth(tm biz.TokenManager, timer timer.Timer) biz.Authenticator {
	return &tokenManagerAuth{tm: tm, timer: timer}
}

// VerifyToken 把 token 交给 TokenManager 校验，命中则转成 claims。
func (a *tokenManagerAuth) VerifyToken(t string) (*biz.JwtClaims, bool) {
	if token := stripBearer(t); token != "" {
		if info, ok := a.tm.VerifyAndTouch(context.TODO(), token, a.timer.Now()); ok {
			return &biz.JwtClaims{UserInfo: info}, true
		}
	}
	return nil, false
}
