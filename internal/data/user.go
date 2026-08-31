package data

import (
	"context"
	"slices"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/user"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
)

var _ biz.UserRepo = (*userRepo)(nil)

// userRepo 是后台用户投影的持久化实现：维护 users 表的登录 upsert、分页查询与角色升降级。
type userRepo struct {
	data  dataStore
	timer timer.Timer
}

// NewUserRepo 构造用户投影 repo。
func NewUserRepo(data dataStore, timer timer.Timer) biz.UserRepo {
	return &userRepo{data: data, timer: timer}
}

// SyncLoginUser 登录成功时按邮箱 upsert 用户投影（幂等可重复调用）：
// 已存在 → 仅推进最近登录（不倒退）+ 补空展示名/升级 email 本地部分默认名（手动设置
// 的非空名不被覆盖）；不存在 → 创建（超级管理员恒 mars_admin，其余角色为空数组；展示名
// 取登录名，空则回退邮箱本地部分；最近登录 = 当前时刻）。只处理单邮箱（登录热路径）；
// roles 一律不覆盖，尊重管理员后台手动升降级。
func (r *userRepo) SyncLoginUser(ctx context.Context, email, name string) (err error) {
	ctx, span := tracer.Start(ctx, "userRepo/SyncLoginUser")
	defer func() { endSpan(span, err) }()
	db := r.data.DB()

	// 邮箱统一小写归一（对齐 VerifyToken），空邮箱无身份可同步。
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return errs.InvalidArgument("email 不能为空")
	}
	now := r.timer.Now()

	cur, err := db.User.Query().Where(user.EmailEQ(email)).First(ctx)
	if err == nil {
		update := db.User.UpdateOneID(cur.ID)
		if !updateUserProjection(update, cur, &now, name) {
			return nil
		}
		_, err = update.Save(ctx)
		return errs.Wrap(err, "update user projection on login")
	}
	// 邮箱不存在走创建；其余查询失败（如 DB 断开）归 500 上抛，不误走创建分支。
	if !errs.IsNotFound(err) {
		return errs.Wrap(err, "query user on login")
	}

	roles := []string{}
	if email == biz.SuperAdminEmail {
		roles = []string{biz.MarsAdmin}
	}
	displayName := name
	if displayName == "" {
		displayName = localPartOf(email)
	}
	_, err = db.User.Create().
		SetEmail(email).
		SetName(displayName).
		SetRoles(roles).
		SetNillableLastLogin(&now).
		Save(ctx)
	return errs.Wrap(err, "create user projection on login")
}

// updateUserProjection 把 lastLogin/name 投影规则应用到既有用户（不覆盖 roles）：
// 仅补更晚的最近登录（已有登录时间不倒退，未登录用户补首次登录时间）+ 展示名规则：
// 补空，并把「email 本地部分」自动默认名升级为真实登录名（本地部分是空登录名的
// 回退默认值，不是人改的，可安全升级）；与本地部分不同的手动名不被覆盖。返回是否有
// 字段被实际修改；调用方负责 Save 与错误包裹，以保留各自错误消息语义。SyncLoginUser
// 登录热路径（推进最近登录 + 补名）调用。
func updateUserProjection(update *ent.UserUpdateOne, cur *ent.User, lastLogin *time.Time, name string) bool {
	changed := false
	if lastLogin != nil && (cur.LastLogin == nil || lastLogin.After(*cur.LastLogin)) {
		update.SetLastLogin(*lastLogin)
		changed = true
	}
	if name != "" && name != cur.Name && (cur.Name == "" || cur.Name == localPartOf(cur.Email)) {
		update.SetName(name)
		changed = true
	}
	return changed
}

// localPartOf 取邮箱 @ 前的本地部分作展示名回退（真实展示名优先取 OIDC 登录名）。
func localPartOf(email string) string {
	if i := strings.IndexByte(email, '@'); i > 0 {
		return email[:i]
	}
	return email
}

// adminPredicate 是「roles 数组包含 mars_admin」的查询谓词：用 MySQL JSON_CONTAINS
// 判定数组成员，兼容 roles 为 JSON 数组的存储结构。
func adminPredicate() func(*sql.Selector) {
	return func(s *sql.Selector) {
		s.Where(sqljson.ValueContains(user.FieldRoles, biz.MarsAdmin))
	}
}

// List 分页查询用户投影：支持按邮箱/展示名模糊搜索与仅管理员过滤；
// 统计（total/admins/regular）按全量口径计算，不受搜索/过滤影响。
func (r *userRepo) List(ctx context.Context, input *biz.ListUserInput) (out *biz.ListUserResult, err error) {
	ctx, span := tracer.Start(ctx, "userRepo/List")
	defer func() { endSpan(span, err) }()
	db := r.data.DB()

	query := db.User.Query()
	if s := strings.TrimSpace(input.Search); s != "" {
		query = query.Where(user.Or(user.EmailContainsFold(s), user.NameContainsFold(s)))
	}
	if input.AdminOnly {
		query = query.Where(adminPredicate())
	}

	// 排序：默认最近登录倒序（最近登录在前）；asc 显式指定升序（最早登录在前）。
	// 无论升降序，last_login 为 NULL（从未登录）的用户都垫底：以 `last_login IS NULL`
	// 作首要排序列（非 NULL=0 在前、NULL=1 在后），MySQL/SQLite 均兼容（不用
	// Postgres 专有的 NULLS LAST）；非法 sort 值回落 desc 默认。
	loginOrder := sql.OrderDesc()
	if input.Sort == "asc" {
		loginOrder = sql.OrderAsc()
	}
	query = query.Order(
		func(s *sql.Selector) {
			s.OrderExprFunc(func(b *sql.Builder) {
				b.Ident(user.FieldLastLogin).WriteOp(sql.OpIsNull)
			})
		},
		sql.OrderByField(user.FieldLastLogin, loginOrder).ToFunc(),
		ent.Desc(user.FieldID),
	)

	users, err := query.Clone().
		Offset(pagination.GetPageOffset(input.Page, input.PageSize)).
		Limit(int(input.PageSize)).
		All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "list users")
	}
	count := query.Clone().CountX(ctx)

	total := db.User.Query().CountX(ctx)
	admins := db.User.Query().Where(adminPredicate()).CountX(ctx)
	return &biz.ListUserResult{
		Items: slice.Map(users, toUser),
		Pag:   pagination.NewPagination(input.Page, input.PageSize, count),
		Stats: biz.UserStats{Total: int32(total), Admins: int32(admins), Regular: int32(total - admins)},
	}, nil
}

// toUser 把 ent.User 转换为 biz.User（nil 安全）。
func toUser(u *ent.User) *biz.User {
	if u == nil {
		return nil
	}
	return &biz.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Roles:     u.Roles,
		LastLogin: u.LastLogin,
		CreatedAt: u.CreatedAt,
	}
}

// ToggleAdmin 设置/移除指定用户的管理员角色（mars_admin）；超级管理员不可降级。
func (r *userRepo) ToggleAdmin(ctx context.Context, email string, admin bool) (err error) {
	ctx, span := tracer.Start(ctx, "userRepo/ToggleAdmin")
	defer func() { endSpan(span, err) }()
	if email == biz.SuperAdminEmail {
		return errs.InvalidArgument("超级管理员不可降级")
	}
	db := r.data.DB()
	u, err := db.User.Query().Where(user.EmailEQ(email)).First(ctx)
	if err != nil {
		return errs.Wrap(err, "query user")
	}
	roles := toggleRole(u.Roles, biz.MarsAdmin, admin)
	if slices.Equal(roles, u.Roles) {
		return nil
	}
	_, err = db.User.UpdateOneID(u.ID).SetRoles(roles).Save(ctx)
	return errs.Wrap(err, "update user role")
}

// toggleRole 返回把 role 增/删（present=true 追加，false 移除）后的新角色切片，保持既有顺序。
func toggleRole(roles []string, role string, present bool) []string {
	out := make([]string, 0, len(roles)+1)
	added := false
	for _, r := range roles {
		if r == role {
			if present {
				out = append(out, r)
				added = true
			}
			continue
		}
		out = append(out, r)
	}
	if present && !added {
		out = append(out, role)
	}
	return out
}
