package contracts

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/rbac"
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
	info.Roles = []string{rbac.MarsAdmin, "xxx"}
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
			Roles:   []string{rbac.MarsAdmin, "xxx"},
		},
	}

	assert.Equal(t, o.LogoutUrl, o.ToUserInfo().LogoutUrl)
	assert.Equal(t, o.Picture, o.ToUserInfo().Picture)
	assert.Equal(t, o.Sub, o.ToUserInfo().ID)
	assert.Equal(t, o.Email, o.ToUserInfo().Email)
	assert.Equal(t, o.Name, o.ToUserInfo().Name)
	assert.Equal(t, []string{rbac.MarsAdmin, "xxx"}, o.ToUserInfo().Roles)
	assert.True(t, o.ToUserInfo().IsAdmin())
}

func TestUserInfo_Json(t *testing.T) {
	uinfo := UserInfo{
		ID:        "1",
		Email:     "xx@xx.com",
		Name:      "duc",
		Picture:   "pic",
		Roles:     []string{rbac.MarsAdmin},
		LogoutUrl: "https://xx/logout",
	}
	assert.Equal(t, `{"id":"1","email":"xx@xx.com","name":"duc","picture":"pic","roles":["mars_admin"],"logout_url":"https://xx/logout"}`, uinfo.Json())
}
