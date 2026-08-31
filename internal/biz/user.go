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
type User struct {
	ID        int
	Email     string
	Name      string
	Roles     []string
	LastLogin *time.Time
	CreatedAt time.Time
}

// UserStats 是用户统计（全量口径，不受搜索/角色过滤影响）。
type UserStats struct {
	Total   int32
	Admins  int32
	Regular int32
}

// ListUserInput 是用户分页列表输入：Search 按邮箱/展示名模糊匹配，
// AdminOnly 只看管理员（roles 含 mars_admin）。
type ListUserInput struct {
	Page, PageSize int32
	Search         string
	AdminOnly      bool
}

// ListUserResult 是用户分页列表结果，携带统计。
type ListUserResult struct {
	Items []*User
	Pag   *pagination.Pagination
	Stats UserStats
}

// UserRepo 是用户投影的持久化端口（data 层实现）：登录 upsert 与查询/角色管理。
type UserRepo interface {
	// SyncLoginUser 登录成功时按邮箱 upsert 用户投影（不存在则创建）。
	SyncLoginUser(ctx context.Context, email, name string) error
	// List 分页查询用户，支持搜索与仅管理员过滤，返回全量统计。
	List(ctx context.Context, input *ListUserInput) (*ListUserResult, error)
	// ToggleAdmin 设置/移除指定用户的管理员角色（mars_admin）。
	ToggleAdmin(ctx context.Context, email string, admin bool) error
}

// UserBiz 是后台用户管理的业务接口：登录 upsert、查询与角色管理。
type UserBiz interface {
	// SyncLoginUser 登录成功时同步用户投影（不存在则创建），email 不能为空。
	SyncLoginUser(ctx context.Context, email, name string) error
	// List 分页查询用户列表（支持搜索/仅管理员过滤）。
	List(ctx context.Context, input *ListUserInput) (*ListUserResult, error)
	// ToggleAdmin 设置/移除指定用户的管理员角色（mars_admin），email 不能为空。
	ToggleAdmin(ctx context.Context, email string, admin bool) error
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

// SyncLoginUser 登录成功时同步用户投影：email 空报 InvalidArgument，否则透传 repo
// upsert（不存在则创建、存在则推进最近登录，不覆盖角色）。
func (u *userBiz) SyncLoginUser(ctx context.Context, email, name string) error {
	if strings.TrimSpace(email) == "" {
		return errs.InvalidArgument("email 不能为空")
	}
	return u.repo.SyncLoginUser(ctx, strings.TrimSpace(email), name)
}

// ToggleAdmin 翻转指定用户的管理员角色。
func (u *userBiz) ToggleAdmin(ctx context.Context, email string, admin bool) error {
	if strings.TrimSpace(email) == "" {
		return errs.InvalidArgument("email 不能为空")
	}
	return u.repo.ToggleAdmin(ctx, email, admin)
}
