package schematype

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMarsAdmin_Value 验证管理员角色名的字面量契约。
func TestMarsAdmin_Value(t *testing.T) {
	assert.Equal(t, "mars_admin", MarsAdmin)
}

// TestUploadType_Values 验证上传类型常量与底层值的映射契约。
func TestUploadType_Values(t *testing.T) {
	assert.Equal(t, UploadType("local"), Local)
	assert.Equal(t, UploadType("s3"), S3)
}

func TestUserInfo_IsAdmin_WithAdminRole(t *testing.T) {
	ui := &UserInfo{Roles: []string{"developer", string(MarsAdmin)}}
	assert.True(t, ui.IsAdmin())
}

func TestUserInfo_IsAdmin_AdminRoleFirst(t *testing.T) {
	ui := &UserInfo{Roles: []string{string(MarsAdmin), "developer"}}
	assert.True(t, ui.IsAdmin())
}

func TestUserInfo_IsAdmin_WithoutAdminRole(t *testing.T) {
	ui := &UserInfo{Roles: []string{"developer", "operator"}}
	assert.False(t, ui.IsAdmin())
}

func TestUserInfo_IsAdmin_EmptyRoles(t *testing.T) {
	ui := &UserInfo{Roles: []string{}}
	assert.False(t, ui.IsAdmin())
}

func TestUserInfo_IsAdmin_NilRoles(t *testing.T) {
	ui := &UserInfo{}
	assert.False(t, ui.IsAdmin())
}

// TestUserInfo_JSON_Contract 验证 UserInfo 的 json 字段名与 OIDC userinfo
// 契约一致（id/email/name/picture/roles/logout_url），防止 tag 漂移破坏认证链路。
func TestUserInfo_JSON_Contract(t *testing.T) {
	ui := &UserInfo{
		ID:        "sub-1",
		Email:     "a@b.com",
		Name:      "Alice",
		Picture:   "https://img/p.png",
		Roles:     []string{"developer"},
		LogoutUrl: "https://logout",
	}
	raw, err := json.Marshal(ui)
	assert.NoError(t, err)

	var m map[string]any
	assert.NoError(t, json.Unmarshal(raw, &m))
	assert.Equal(t, map[string]any{
		"id":         "sub-1",
		"email":      "a@b.com",
		"name":       "Alice",
		"picture":    "https://img/p.png",
		"roles":      []any{"developer"},
		"logout_url": "https://logout",
	}, m)

	var back UserInfo
	assert.NoError(t, json.Unmarshal(raw, &back))
	assert.Equal(t, *ui, back)
}

// TestUserInfo_IsSuperAdmin_SuperAdminEmail 内置超管固定邮箱命中 IsSuperAdmin。
func TestUserInfo_IsSuperAdmin_SuperAdminEmail(t *testing.T) {
	ui := &UserInfo{Email: SuperAdminEmail}
	assert.True(t, ui.IsSuperAdmin())
}

// TestUserInfo_IsSuperAdmin_OtherEmail 普通邮箱（含管理员身份）不命中超管判定。
func TestUserInfo_IsSuperAdmin_OtherEmail(t *testing.T) {
	ui := &UserInfo{Email: "someone@else.com", Roles: []string{string(MarsAdmin)}}
	assert.False(t, ui.IsSuperAdmin())
}

// TestUserInfo_IsSuperAdmin_EmptyEmail 空邮箱（未登录/未知身份）不命中超管判定。
func TestUserInfo_IsSuperAdmin_EmptyEmail(t *testing.T) {
	ui := &UserInfo{}
	assert.False(t, ui.IsSuperAdmin())
}
