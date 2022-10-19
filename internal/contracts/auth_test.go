package contracts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserInfo_GetID(t *testing.T) {
	info := UserInfo{
		ID: "sub",
	}
	assert.Equal(t, "sub", info.GetID())
}

func TestUserInfo_IsAdmin(t *testing.T) {
	info := UserInfo{
		Roles: []string{""},
		ID:    "sub",
	}
	assert.False(t, info.IsAdmin())
	info.Roles = []string{"admin", "xxx"}
	assert.True(t, info.IsAdmin())
}

func TestOidcClaims_ToUserInfo(t *testing.T) {
	o := OidcClaims{
		LogoutUrl: "xxx",
		OpenIDClaims: OpenIDClaims{
			Sub:     "sub",
			Name:    "duc",
			Email:   "Email",
			Picture: "avatar.png",
		},
	}

	assert.Equal(t, []string{}, o.ToUserInfo().Roles)
	assert.Equal(t, o.LogoutUrl, o.ToUserInfo().LogoutUrl)
	assert.Equal(t, o.Picture, o.ToUserInfo().Picture)
	assert.Equal(t, o.Sub, o.ToUserInfo().ID)
	assert.Equal(t, o.Email, o.ToUserInfo().Email)
	assert.Equal(t, o.Name, o.ToUserInfo().Name)
}
