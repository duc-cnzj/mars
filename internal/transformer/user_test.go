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
