package biz

import (
	"context"
	"strings"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
)

// User 是后台用户管理的领域模型：email 为唯一身份标识，roles 持久化真实角色名
// （含 mars_admin=管理员；空数组=普通用户，即不含任何角色），与鉴权判定的 Roles 判据同源。
// RolesOverride 标记角色是否已被后台手动管理接管（true=手工升降级，SSO 不再覆盖），
// 供用户管理页展示角色来源（SSO 自动 / 后台手动）。
type User struct {
	ID    int
	Email string
	Name  string
	Roles []string
	// RolesOverride 表示该用户角色是否已被后台手动管理接管：false=登录时按 SSO 角色同步；
	// true=后台手动升降级后 SSO 不再覆盖。
	RolesOverride bool
	LastLogin     *time.Time
	CreatedAt     time.Time
}

// UserStats 是用户统计（全量口径，不受搜索/角色过滤影响）。
type UserStats struct {
	Total   int32
	Admins  int32
	Regular int32
}

// ListUserInput 是用户分页列表输入：Search 按邮箱/展示名模糊匹配，
// AdminOnly 只看管理员（roles 含 mars_admin），Sort 指定最近登录排序方向。
type ListUserInput struct {
	Page, PageSize int32
	Search         string
	AdminOnly      bool
	// Sort 排序方向：空 = 最近登录倒序（desc）；asc/desc = 指定最近登录升/降序。
	Sort string
}

// ListUserResult 是用户分页列表结果，携带统计。
type ListUserResult struct {
	Items []*User
	Pag   *pagination.Pagination
	Stats UserStats
}

// UserRepo 是用户投影的持久化端口（data 层实现）：登录 upsert 与查询/角色管理。
type UserRepo interface {
	// SyncLoginUser 登录成功时按邮箱 upsert 用户投影（不存在则创建），roles 为该次
	// 登录身份携带的角色（SSO id_token / 内置超管），data 层据此决定角色同步策略。
	SyncLoginUser(ctx context.Context, email, name string, roles []string) error
	// List 分页查询用户，支持搜索与仅管理员过滤，返回全量统计。
	List(ctx context.Context, input *ListUserInput) (*ListUserResult, error)
	// ToggleAdmin 设置/移除指定用户的管理员角色（mars_admin）。
	ToggleAdmin(ctx context.Context, email string, admin bool) error
	// ResetRolesOverride 解除后台手动接管（roles_override 置回 false），该用户恢复按 SSO
	// 角色同步；供超管把误接管/不再需要手动管理的用户交还给 SSO。
	ResetRolesOverride(ctx context.Context, email string) error
}

// UserBiz 是后台用户管理的业务接口：登录 upsert、查询与角色管理。
type UserBiz interface {
	// SyncLoginUser 登录成功时同步用户投影（不存在则创建），email 不能为空；
	// roles 为登录身份携带的角色，透传给 repo 做 SSO 角色同步。
	SyncLoginUser(ctx context.Context, email, name string, roles []string) error
	// List 分页查询用户列表（支持搜索/仅管理员过滤）。
	List(ctx context.Context, input *ListUserInput) (*ListUserResult, error)
	// ToggleAdmin 设置/移除指定用户的管理员角色（mars_admin），email 不能为空；
	// 仅内置超级管理员可调用（普通管理员只能查看不能修改）。
	ToggleAdmin(ctx context.Context, email string, admin bool) error
	// ResetRolesOverride 解除指定用户的后台手动接管（roles_override 置回 false），该用户
	// 恢复按 SSO 角色同步；仅内置超级管理员可调用，email 不能为空。
	ResetRolesOverride(ctx context.Context, email string) error
}

// userBiz 是 UserBiz 的实现：包装 UserRepo 提供投影同步/登录 upsert/查询/角色管理。
type userBiz struct {
	repo UserRepo
}

// NewUserBiz 构造用户管理业务实现。
func NewUserBiz(repo UserRepo) UserBiz {
	return &userBiz{repo: repo}
}

// List 分页查询用户投影：直接透传 repo 查询。
func (u *userBiz) List(ctx context.Context, input *ListUserInput) (*ListUserResult, error) {
	return u.repo.List(ctx, input)
}

// SyncLoginUser 登录成功时同步用户投影：email 空报 InvalidArgument，否则把邮箱与登录
// 身份携带的角色（SSO id_token / 内置超管）透传给 repo 做 upsert 与角色同步。
func (u *userBiz) SyncLoginUser(ctx context.Context, email, name string, roles []string) error {
	if strings.TrimSpace(email) == "" {
		return errs.InvalidArgument("email 不能为空")
	}
	return u.repo.SyncLoginUser(ctx, strings.TrimSpace(email), name, roles)
}

// ToggleAdmin 翻转指定用户的管理员角色：仅内置超级管理员可调用（普通管理员只能查看
// 后台 users 不能修改），email 空报 InvalidArgument，否则把 email/admin 透传给 repo。
func (u *userBiz) ToggleAdmin(ctx context.Context, email string, admin bool) error {
	// 权限门卫：只有超级管理员才能修改其他用户的权限，普通管理员仅可查看。
	if !MustGetUser(ctx).IsSuperAdmin() {
		return errs.WrapPermissionDenied(errs.ErrorPermissionDenied, "切换用户管理员角色")
	}
	if strings.TrimSpace(email) == "" {
		return errs.InvalidArgument("email 不能为空")
	}
	return u.repo.ToggleAdmin(ctx, email, admin)
}

// ResetRolesOverride 解除指定用户的后台手动接管（roles_override 置回 false，恢复 SSO 角色
// 同步）：与 ToggleAdmin 同级权限门卫——只有超级管理员可操作，普通管理员仅可查看；email 空
// 报 InvalidArgument，否则把 email 透传给 repo。
func (u *userBiz) ResetRolesOverride(ctx context.Context, email string) error {
	// 权限门卫：解除接管同属角色管理，只有超级管理员能操作。
	if !MustGetUser(ctx).IsSuperAdmin() {
		return errs.WrapPermissionDenied(errs.ErrorPermissionDenied, "解除用户角色接管")
	}
	if strings.TrimSpace(email) == "" {
		return errs.InvalidArgument("email 不能为空")
	}
	return u.repo.ResetRolesOverride(ctx, email)
}
