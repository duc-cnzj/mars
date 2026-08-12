package biz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOidcClaims_ToUserInfo 验证 OIDC 声明转用户信息：邮箱转小写，
// Sub/Name/Roles/LogoutUrl 原样透传。
func TestOidcClaims_ToUserInfo(t *testing.T) {
	c := OidcClaims{
		LogoutUrl: "aaa.com",
		OpenIDClaims: OpenIDClaims{
			Sub:   "1",
			Name:  "duc",
			Email: "Duc@q.c",
			Roles: []string{"admin"},
		},
	}
	info := c.ToUserInfo()
	assert.Equal(t, "1", info.ID)
	assert.Equal(t, "duc", info.Name)
	assert.Equal(t, "duc@q.c", info.Email)
	assert.Equal(t, []string{"admin"}, info.Roles)
	assert.Equal(t, "aaa.com", info.LogoutUrl)

	c2 := OidcClaims{
		LogoutUrl: "aaa.com",
		OpenIDClaims: OpenIDClaims{
			Sub:   "1",
			Name:  "duc",
			Email: "123adb@q.c",
			Roles: []string{"admin"},
		},
	}
	info2 := c2.ToUserInfo()
	assert.Equal(t, "123adb@q.c", info2.Email)
}
