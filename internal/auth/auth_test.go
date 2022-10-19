package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/stretchr/testify/assert"
)

var (
	priKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	publicKey = &priKey.PublicKey
)

func TestAuth_Sign(t *testing.T) {
	auth := NewAuth(priKey, publicKey)
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
	auth := NewAuth(priKey, publicKey)
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
	assert.Implements(t, (*contracts.AuthInterface)(nil), NewAuth(nil, nil))
}
