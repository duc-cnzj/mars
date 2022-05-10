package contracts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserInfo_GetID(t *testing.T) {
	info := UserInfo{
		OpenIDClaims: OpenIDClaims{
			Sub: "sub",
		},
	}
	assert.Equal(t, "sub", info.GetID())
}

func TestUserInfo_IsAdmin(t *testing.T) {
	info := UserInfo{
		Roles: []string{""},
		OpenIDClaims: OpenIDClaims{
			Sub: "sub",
		},
	}
	assert.False(t, info.IsAdmin())
	info.Roles = []string{"admin", "xxx"}
	assert.True(t, info.IsAdmin())
}
