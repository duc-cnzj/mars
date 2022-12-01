package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	priKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	publicKey = &priKey.PublicKey
)

func TestAuth_Sign(t *testing.T) {
	auth := NewJwtAuth(priKey, publicKey)
	sign, err := auth.Sign(contracts.UserInfo{
		LogoutUrl: "xxx",
		Roles:     []string{"admin"},
		ID:        "1",
		Email:     "1025434218@qq.com",
		Name:      "duc",
	})
	assert.Nil(t, err)
	token, b := auth.VerifyToken(sign.Token)
	assert.True(t, b)
	assert.Equal(t, "mars", token.StandardClaims.Issuer)
	assert.Equal(t, "duc", token.UserInfo.Name)
	assert.Equal(t, "1025434218@qq.com", token.StandardClaims.Subject)
	assert.Equal(t, []string{"admin"}, token.UserInfo.Roles)
	assert.Equal(t, "xxx", token.UserInfo.LogoutUrl)
}

func TestAuth_VerifyToken(t *testing.T) {
	auth := NewJwtAuth(priKey, publicKey)
	sign, _ := auth.Sign(contracts.UserInfo{
		LogoutUrl: "xxx",
		Roles:     []string{"admin"},
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

func TestNewAuth(t *testing.T) {
	assert.Implements(t, (*contracts.AuthInterface)(nil), NewJwtAuth(nil, nil))
}

func TestNewAccessTokenAuth(t *testing.T) {
	assert.Implements(t, (*contracts.Authenticator)(nil), NewAccessTokenAuth(nil))
}

func TestAccessTokenAuth_VerifyToken(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, fn := testutil.SetGormDB(m, app)
	defer fn()
	db.AutoMigrate(&models.AccessToken{})
	_, b := NewAccessTokenAuth(app).VerifyToken("xxxx")
	assert.False(t, b)

	at := models.NewAccessToken("my token", time.Now().Add(10*time.Second), &contracts.UserInfo{
		ID:        "xx",
		Email:     "admin@admin.com",
		Name:      "duc",
		Picture:   "xx",
		Roles:     []string{"admin"},
		LogoutUrl: "https://xxx",
	})
	db.Create(at)
	assert.True(t, at.LastUsedAt.IsZero())
	u, b := NewAccessTokenAuth(app).VerifyToken(at.Token)
	assert.True(t, b)
	assert.Equal(t, "xx", u.UserInfo.ID)
	assert.Equal(t, "admin@admin.com", u.UserInfo.Email)
	assert.Equal(t, "duc", u.UserInfo.Name)
	assert.Equal(t, "xx", u.UserInfo.Picture)
	assert.Equal(t, []string{"admin"}, u.UserInfo.Roles)
	assert.Equal(t, "https://xxx", u.UserInfo.LogoutUrl)

	var at2 models.AccessToken
	db.Where("`token` = ?", at.Token).First(&at2)
	assert.False(t, at2.LastUsedAt.IsZero())

	_, bb := NewAccessTokenAuth(app).VerifyToken("bearer " + at.Token)
	assert.True(t, bb)
}

type autha struct{}

func (a *autha) VerifyToken(s string) (*contracts.JwtClaims, bool) {
	return nil, false
}

type authb struct{}

func (a *authb) VerifyToken(s string) (*contracts.JwtClaims, bool) {
	return &contracts.JwtClaims{UserInfo: contracts.UserInfo{Email: "duc@duc.com"}}, true
}

func TestAuthn_VerifyToken(t *testing.T) {
	x, b := NewAuthn(nil, &autha{}, &authb{}).VerifyToken("xx")
	assert.True(t, b)
	assert.Equal(t, "duc@duc.com", x.UserInfo.Email)

	_, bb := NewAuthn(nil, &autha{}).VerifyToken("xx")
	assert.False(t, bb)
}

func TestAuthn_Sign(t *testing.T) {
	_, err := NewAuthn(func(info contracts.UserInfo) (*contracts.SignData, error) {
		return nil, errors.New("xxx")
	}).Sign(contracts.UserInfo{})
	assert.Equal(t, "xxx", err.Error())
}

func TestNewAuthn(t *testing.T) {
	assert.Implements(t, (*contracts.AuthInterface)(nil), NewAuthn(nil))
}
