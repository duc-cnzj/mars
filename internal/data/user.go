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

// 编译期断言：userRepo 同时实现 biz.EffectiveRolesProvider——authBiz 经该窄接口
// 读取 users 表接管状态计算生效角色（对齐 AuthConfigProvider 的窄接口取数模式）。
var _ biz.EffectiveRolesProvider = (*userRepo)(nil)

// userRepo 是后台用户投影的持久化实现：维护 users 表的登录 upsert、分页查询与角色升降级。
type userRepo struct {
	data  dataStore
	timer timer.Timer
}

// NewUserRepo 构造用户投影 repo。返回具体类型（对齐 data.NewData→*dataImpl 的
// 装配惯例）：wire 依赖具体类型一次实例满足多个窄接口——userRepo 同时实现
// biz.UserRepo（用户管理）与 biz.EffectiveRolesProvider（生效角色解析），
// 同一实例注入 NewUserBiz/NewAuthBiz，避免两接口各建实例读同一张表。
func NewUserRepo(data dataStore, timer timer.Timer) *userRepo {
	return &userRepo{data: data, timer: timer}
}

// SyncLoginUser 登录成功时按邮箱 upsert 用户投影（幂等可重复调用）：
// 已存在 → 推进最近登录 + 补空展示名（手动设置的非空名不被覆盖），且当该用户未被后台
// 手动管理（roles_override=false）时按登录身份同步角色（SSO id_token 带来的 mars_admin
// 得以写进 users 表）；不存在 → 创建（角色取登录身份，空则普通用户；展示名取登录名，
// 空则回退邮箱本地部分；最近登录 = 当前时刻）。只处理单邮箱（登录热路径）；roles_override
// 为 true 的用户（已被后台手动升降级接管）角色不被 SSO 覆盖，尊重手动管理。
func (r *userRepo) SyncLoginUser(ctx context.Context, email, name string, roles []string) (err error) {
	ctx, span := tracer.Start(ctx, "userRepo/SyncLoginUser")
	defer func() { endSpan(span, err) }()
	db := r.data.DB()

	// 邮箱统一小写归一（对齐 VerifyToken），空邮箱无身份可同步。
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return errs.InvalidArgument("email 不能为空")
	}
	now := r.timer.Now()
	// 登录身份携带的角色（SSO id_token / 内置超管），nil 归一为空数组统一 JSON 存储；
	// 超级管理员恒 mars_admin——登录身份已带，此守卫兜底防 SSO 侧把它降级成普通用户。
	roles = normalizeRoles(roles)
	if email == biz.SuperAdminEmail && !slices.Contains(roles, biz.MarsAdmin) {
		roles = append(roles, biz.MarsAdmin)
	}

	cur, err := db.User.Query().Where(user.EmailEQ(email)).First(ctx)
	if err == nil {
		update := db.User.UpdateOneID(cur.ID)
		changed := updateUserProjection(update, cur, &now, name)
		// SSO 角色同步：仅未被后台手动接管（roles_override=false）时按登录身份覆盖角色；
		// 被接管（true）则尊重手动升降级，即使 SSO 下次仍带 mars_admin 也不洗掉手动降权。
		if !cur.RolesOverride && !slices.Equal(cur.Roles, roles) {
			update.SetRoles(roles)
			changed = true
		}
		if !changed {
			return nil
		}
		_, err = update.Save(ctx)
		return errs.Wrap(err, "update user projection on login")
	}
	// 邮箱不存在走创建；其余查询失败（如 DB 断开）归 500 上抛，不误走创建分支。
	if !errs.IsNotFound(err) {
		return errs.Wrap(err, "query user on login")
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

// EffectiveRoles 计算登录用户在鉴权时的生效角色：后台手动接管（roles_override=true）
// 的用户以 users 表手工角色为准（升降级真正生效，SSO 不再覆盖）；未接管或尚未落投影
// （首次登录前窗口）回落登录身份携带的 SSO 角色。超级管理员恒 mars_admin（守卫兜底防
// 接管状态误伤超管）。供鉴权入口在验签后调用，实现「后台可控制 SSO 带来的管理员权限」：
// 后台降权后该用户即使 JWT 仍带 mars_admin，生效角色也不含管理员。用户表读取失败返回
// 错误（由调用方决定回落策略），不在此处静默吞错。
func (r *userRepo) EffectiveRoles(ctx context.Context, email string, ssoRoles []string) ([]string, error) {
	ctx, span := tracer.Start(ctx, "userRepo/EffectiveRoles")
	defer func() { endSpan(span, nil) }()
	db := r.data.DB()
	// 邮箱统一小写归一（对齐 SyncLoginUser/VerifyToken）。
	email = strings.ToLower(strings.TrimSpace(email))
	// SSO 角色归一：nil 转空数组；超级管理员恒 mars_admin——登录身份已带，此守卫兜底。
	ssoRoles = normalizeRoles(ssoRoles)
	if email == biz.SuperAdminEmail && !slices.Contains(ssoRoles, biz.MarsAdmin) {
		ssoRoles = append(ssoRoles, biz.MarsAdmin)
	}

	cur, err := db.User.Query().Where(user.EmailEQ(email)).First(ctx)
	if err != nil {
		// 用户尚未落投影（首次登录前窗口）：跟随登录身份角色，不视为故障。
		if !errs.IsNotFound(err) {
			return nil, errs.Wrap(err, "query user for effective roles")
		}
		return ssoRoles, nil
	}
	// 未被后台手动接管：生效角色 = SSO 登录身份角色（users 表角色在每次登录时已同步）。
	if !cur.RolesOverride {
		return ssoRoles, nil
	}
	// 已被后台手动接管：生效角色 = users 表手工角色（升降级结果），SSO 不再覆盖。
	return normalizeRoles(cur.Roles), nil
}

// normalizeRoles 把登录身份携带的角色归一为确定值：nil 转空切片，统一 JSON 以数组而非
// null 存储，保证与 schema 默认值 [] 一致。
func normalizeRoles(roles []string) []string {
	if roles == nil {
		return []string{}
	}
	return roles
}

// updateUserProjection 把 lastLogin/name 投影规则应用到既有用户（不处理 roles，角色同步
// 由调用方 SyncLoginUser 按 roles_override 判定单独进行）：
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

// toUser 把 ent.User 转换为 biz.User（nil 安全）。RolesOverride 透传手动接管标记，
// 供用户管理页展示「角色来源」（SSO 自动 / 后台手动）。
func toUser(u *ent.User) *biz.User {
	if u == nil {
		return nil
	}
	return &biz.User{
		ID:            u.ID,
		Email:         u.Email,
		Name:          u.Name,
		Roles:         u.Roles,
		RolesOverride: u.RolesOverride,
		LastLogin:     u.LastLogin,
		CreatedAt:     u.CreatedAt,
	}
}

// ToggleAdmin 设置/移除指定用户的管理员角色（mars_admin）；超级管理员不可降级。
// 手动升降级即声明接管该用户角色：置 roles_override=true，此后 SSO 登录不再覆盖其角色，
// 即使 SSO 下次仍带 mars_admin，手动降权也不被洗掉。
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
	// 角色未变且已是手动接管状态：无字段需写，幂等早退。
	if slices.Equal(roles, u.Roles) && u.RolesOverride {
		return nil
	}
	_, err = db.User.UpdateOneID(u.ID).SetRoles(roles).SetRolesOverride(true).Save(ctx)
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

// ResetRolesOverride 解除后台手动接管：把 roles_override 置回 false，该用户从下一次登录起
// 恢复按 SSO 角色同步（SSO 重新成为角色来源）。幂等：本就未接管（roles_override=false）直接
// 返回；邮箱不存在按 NotFound 上抛。超级管理员无接管概念（biz 门卫已拦截降级/接管），不受影响。
func (r *userRepo) ResetRolesOverride(ctx context.Context, email string) (err error) {
	ctx, span := tracer.Start(ctx, "userRepo/ResetRolesOverride")
	defer func() { endSpan(span, err) }()
	// 邮箱统一小写归一（对齐 SyncLoginUser/EffectiveRoles）。
	email = strings.ToLower(strings.TrimSpace(email))
	db := r.data.DB()
	u, err := db.User.Query().Where(user.EmailEQ(email)).First(ctx)
	if err != nil {
		return errs.Wrap(err, "query user")
	}
	if !u.RolesOverride {
		return nil
	}
	_, err = db.User.UpdateOneID(u.ID).SetRolesOverride(false).Save(ctx)
	return errs.Wrap(err, "update user roles override")
}
