package transformer_test

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"github.com/duc-cnzj/mars/v6/internal/util/date"
	"github.com/stretchr/testify/assert"
)

// TestFromUser_Roles 验证角色归一化：含 mars_admin 归 admin，否则空数组（普通用户）。
func TestFromUser_Roles(t *testing.T) {
	m := transformer.FromUser(&biz.User{Email: "a@b.c", Roles: []string{biz.MarsAdmin, "user"}})
	assert.Equal(t, []string{"admin"}, m.Roles)

	m = transformer.FromUser(&biz.User{Email: "a@b.c", Roles: []string{"user"}})
	assert.Equal(t, []string{}, m.Roles)

	m = transformer.FromUser(&biz.User{Email: "a@b.c"})
	assert.Equal(t, []string{}, m.Roles)
}

// TestFromUser_LastLogin 验证 last_login 映射：nil 保持缺省，非 nil 转 RFC3339。
func TestFromUser_LastLogin(t *testing.T) {
	m := transformer.FromUser(&biz.User{Email: "a@b.c"})
	assert.Nil(t, m.LastLogin)

	now := time.Now()
	m = transformer.FromUser(&biz.User{Email: "a@b.c", LastLogin: &now})
	if assert.NotNil(t, m.LastLogin) {
		assert.Equal(t, date.ToRFC3339(&now), *m.LastLogin)
	}
}

// TestFromUser_Fields 验证基础字段映射。
func TestFromUser_Fields(t *testing.T) {
	created := time.Now()
	m := transformer.FromUser(&biz.User{
		ID:        7,
		Email:     "duc@mars.dev",
		Name:      "duc",
		CreatedAt: created,
	})
	assert.Equal(t, int32(7), m.Id)
	assert.Equal(t, "duc@mars.dev", m.Email)
	assert.Equal(t, "duc", m.Name)
	assert.Equal(t, date.ToRFC3339(&created), m.CreatedAt)
}

// TestFromUser_IsSuperAdmin 验证超管标识：内置超管固定邮箱为 true，其余为 false。
func TestFromUser_IsSuperAdmin(t *testing.T) {
	m := transformer.FromUser(&biz.User{Email: biz.SuperAdminEmail})
	assert.True(t, m.IsSuperAdmin)

	m = transformer.FromUser(&biz.User{Email: "ordinary@mars.com"})
	assert.False(t, m.IsSuperAdmin)
}

// TestFromUser_RolesOverride 验证接管标志透传：默认 false（角色来自 SSO 同步），
// 后台手动接管后为 true（前端据此显示来源 badge）。
func TestFromUser_RolesOverride(t *testing.T) {
	m := transformer.FromUser(&biz.User{Email: "a@b.c"})
	assert.False(t, m.RolesOverride)

	m = transformer.FromUser(&biz.User{Email: "a@b.c", RolesOverride: true})
	assert.True(t, m.RolesOverride)
}
