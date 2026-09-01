package transformer

import (
	"github.com/duc-cnzj/mars/api/v6/proto/user"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/util/date"
)

// FromUser 把 biz.User 映射为后台用户管理展示模型：真实角色名归一化（含 mars_admin
// 归 admin，否则空数组表示普通用户），last_login 转 RFC3339（nil 保持缺省，表示从未登录）。
// IsSuperAdmin 按内置超管邮箱判定（单一事实来源 schematype.SuperAdminEmail）；
// RolesOverride 透传用户投影的接管标志（false=角色来自 SSO 登录同步，true=后台手动接管）。
func FromUser(u *biz.User) *user.UserModel {
	m := &user.UserModel{
		Id:            int32(u.ID),
		Email:         u.Email,
		Name:          u.Name,
		Roles:         normalizeUserRoles(u.Roles),
		CreatedAt:     date.ToRFC3339(&u.CreatedAt),
		IsSuperAdmin:  u.Email == biz.SuperAdminEmail,
		RolesOverride: u.RolesOverride,
	}
	if u.LastLogin != nil {
		lastLogin := date.ToRFC3339(u.LastLogin)
		m.LastLogin = &lastLogin
	}
	return m
}

// normalizeUserRoles 把真实角色名归一化为展示角色：含 mars_admin 即管理员，
// 否则空数组（普通用户由前端从空 roles 推导，不再下发 user 值）。
func normalizeUserRoles(roles []string) []string {
	for _, r := range roles {
		if r == biz.MarsAdmin {
			return []string{"admin"}
		}
	}
	return []string{}
}
