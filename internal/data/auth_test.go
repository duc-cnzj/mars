package data

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"math/big"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

var (
	priKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	publicKey = &priKey.PublicKey
)

type mockTokenManager struct {
	fn func(ctx context.Context, token string, now time.Time) (*biz.UserInfo, bool)
}

func (m *mockTokenManager) VerifyAndTouch(ctx context.Context, token string, now time.Time) (*biz.UserInfo, bool) {
	return m.fn(ctx, token, now)
}

type mockAuthenticator struct {
	fn func(s string) (*biz.JwtClaims, bool)
}

func (m *mockAuthenticator) VerifyToken(s string) (*biz.JwtClaims, bool) {
	return m.fn(s)
}

func TestAuth_Sign(t *testing.T) {
	auth := newJwtAuth(priKey, publicKey, timer.NewReal())
	sign, err := auth.Sign(&biz.UserInfo{
		LogoutUrl: "xxx",
		Roles:     []string{schematype.MarsAdmin},
		ID:        "1",
		Email:     "1025434218@qq.com",
		Name:      "duc",
	})
	assert.Nil(t, err)
	token, b := auth.VerifyToken(sign.Token)
	assert.True(t, b)
	assert.Equal(t, "mars", token.Issuer)
	assert.Equal(t, "duc", token.UserInfo.Name)
	assert.Equal(t, "1025434218@qq.com", token.Subject)
	assert.Equal(t, []string{schematype.MarsAdmin}, token.UserInfo.Roles)
	assert.Equal(t, "xxx", token.UserInfo.LogoutUrl)

	pk := &rsa.PrivateKey{
		PublicKey: rsa.PublicKey{
			N: big.NewInt(1),
		},
	}
	assert.Less(t, pk.Size(), 11)
	authError := newJwtAuth(pk, nil, timer.NewReal())
	_, err = authError.Sign(&biz.UserInfo{
		LogoutUrl: "xxx",
		Roles:     []string{schematype.MarsAdmin},
		ID:        "1",
		Email:     "1025434218@qq.com",
		Name:      "duc",
	})
	assert.Error(t, err)
}

func TestAuth_VerifyToken(t *testing.T) {
	auth := newJwtAuth(priKey, publicKey, timer.NewReal())
	sign, _ := auth.Sign(&biz.UserInfo{
		LogoutUrl: "xxx",
		Roles:     []string{schematype.MarsAdmin},
		ID:        "1",
		Name:      "duc",
	})
	_, b := auth.VerifyToken(sign.Token)
	assert.True(t, b)
	_, b = auth.VerifyToken("Bearer " + sign.Token)
	assert.True(t, b)
	_, b = auth.VerifyToken("bearer " + sign.Token)
	assert.True(t, b)
	_, b = auth.VerifyToken("bearer" + sign.Token)
	assert.True(t, b)
	_, b = auth.VerifyToken("")
	assert.False(t, b)
}

// TestJwtAuth_VerifyToken_RejectsNonRSA 覆盖 alg 校验分支：HS256 签名的 token
// 在 keyfunc 处被显式拒绝，防止 alg confusion（HS256 用公钥当密钥）。
func TestJwtAuth_VerifyToken_RejectsNonRSA(t *testing.T) {
	auth := newJwtAuth(priKey, publicKey, timer.NewReal())
	hmac, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &biz.JwtClaims{
		StandardClaims: &jwt.StandardClaims{Subject: "x"},
	}).SignedString([]byte("secret"))
	assert.NoError(t, err)
	_, ok := auth.VerifyToken(hmac)
	assert.False(t, ok)
}

func TestNewAuth(t *testing.T) {
	assert.Implements(t, (*biz.Auth)(nil), newJwtAuth(nil, nil, timer.NewReal()))
}

func TestNewAccessTokenAuth(t *testing.T) {
	assert.Implements(t, (*biz.Authenticator)(nil), newTokenManagerAuth(nil, timer.NewReal()))
}

func TestAccessTokenAuth_VerifyToken(t *testing.T) {
	_, b := newTokenManagerAuth(nil, timer.NewReal()).VerifyToken("")
	assert.False(t, b)

	mockTM := &mockTokenManager{
		fn: func(ctx context.Context, token string, now time.Time) (*biz.UserInfo, bool) {
			return &biz.UserInfo{
				ID:        "xx",
				Email:     "admin@admin.com",
				Name:      "duc",
				Picture:   "xx",
				Roles:     []string{schematype.MarsAdmin},
				LogoutUrl: "https://xxx",
			}, true
		},
	}
	u, b := newTokenManagerAuth(mockTM, timer.NewReal()).
		VerifyToken("my token")
	assert.True(t, b)
	assert.Equal(t, "xx", u.UserInfo.ID)
	assert.Equal(t, "admin@admin.com", u.UserInfo.Email)
	assert.Equal(t, "duc", u.UserInfo.Name)
	assert.Equal(t, "xx", u.UserInfo.Picture)
	assert.Equal(t, []string{schematype.MarsAdmin}, u.UserInfo.Roles)
	assert.Equal(t, "https://xxx", u.UserInfo.LogoutUrl)

	_, bb := newTokenManagerAuth(mockTM, timer.NewReal()).VerifyToken("bearer my token")
	assert.True(t, bb)

	mockTM2 := &mockTokenManager{
		fn: func(ctx context.Context, token string, now time.Time) (*biz.UserInfo, bool) {
			return nil, false
		},
	}
	_, bb = newTokenManagerAuth(mockTM2, timer.NewReal()).VerifyToken("unknown")
	assert.False(t, bb)
}

func TestNewAuthn(t *testing.T) {
	cfg := &config.Config{PrivateKey: `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQCdx5ZBeL3P3lH2fU/8yd4E1L880DjaKCnnnQkya+kOE7kkJNtP
xW4WIKsBgXUPtXUYk/uA5AkklJ/1ssiTbkM/G5J54ThsACarhiNijUznD81c7g0Q
6pbHYGAHU91wQgpcIv39cOKZVpFkEfIwgBMIKUvupBpGyXMU4YALVV23CQIDAQAB
AoGARo+kzeDumlDlvONr6zRoOybd45eHZWEC5JchLtB9qJL/gH+PKQy1X+X6NDEu
JflTxcsgdhMFV7u0EdCDzRNJtPKP/cU8hww0J2l3ZKTGzbbQnLIBFD3In8sEc9xe
3ikEjqs0EgSh3uY5XEq8qzuX3cI+FNlGyOwzM+ZcN7nWfPUCQQDOURX82COQIfAT
RjTshDQ55J/DUPPHyzpTER9OZNXYKp0IBBNzYyhJ6SHQHSuxHfL8W1FVHhmIsIBW
GQWo0y7zAkEAw8ZPJ4QH5otMsIgIfwMuPX0rO+QxwmJ6eg9ADuFr5zv6HizjAVVP
dKXuUU0gnemD4DncgiV2jZ0v2RzHK1aZEwJAR6G7gpgAcPB3jBmaEmwsPdV06rlW
io2y6FhPiEZWQME62CeiITPSLyc0SC94lfwR+zAxYt4ae2zcgggaAO2hpQJAecA5
d7S3iRu2XM6sofijaCAQpBV9EItX6dLUHqz4Av0cxmlZ33ljiYKr3CngD/SqS+cQ
CGwt91H68MXh40TeuwJARxz1VMLq7hKo8J4scAW/YrBTE4N6malYjYoR2HFs+YwL
cSE/4A4yfzTjN2r5GuJr8rTU7gU4Su9C8dLC0htWCA==
-----END RSA PRIVATE KEY-----`}
	authn, _ := NewAuthn(nil, cfg, timer.NewReal())
	assert.Implements(t, (*biz.Auth)(nil), authn)
}

func TestNewAuthn_InvalidPrivateKey(t *testing.T) {
	// 无效 PEM：ParseRSAPrivateKeyFromPEM 返回错误，NewAuthn 直接透传。
	authn, err := NewAuthn(nil, &config.Config{PrivateKey: "not-a-valid-pem"}, timer.NewReal())
	assert.Nil(t, authn)
	assert.Error(t, err)
}

func TestAuthn_VerifyToken(t *testing.T) {
	auth := &mockAuthenticator{
		fn: func(s string) (*biz.JwtClaims, bool) {
			if s == "a" {
				return nil, true
			}
			return nil, false
		},
	}
	a := &authn{authns: []biz.Authenticator{auth}}
	_, b := a.VerifyToken("a")
	assert.True(t, b)

	_, bb := a.VerifyToken("b")
	assert.False(t, bb)
}

func TestAuthn_Sign(t *testing.T) {
	called := false
	a := &authn{signFunc: func(info *biz.UserInfo) (*biz.SignData, error) {
		called = true
		return nil, nil
	}}
	sign, err := a.Sign(nil)
	assert.Nil(t, sign)
	assert.Nil(t, err)
	assert.True(t, called)
}
