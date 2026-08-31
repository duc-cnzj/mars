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
	"github.com/duc-cnzj/mars/v6/internal/data/ent/member"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/user"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
)

var _ biz.UserRepo = (*userRepo)(nil)

// userRepo 是后台用户投影的持久化实现：维护 users 表与真实身份源
// （内置管理员/命名空间成员）的对账，以及分页查询与角色升降级。
type userRepo struct {
	data  dataStore
	timer timer.Timer
}

// NewUserRepo 构造用户投影 repo。
func NewUserRepo(data dataStore, timer timer.Timer) biz.UserRepo {
	return &userRepo{data: data, timer: timer}
}

// syncSource 是用户同步源的中间结构：一个邮箱对应一条待落库投影。
type syncSource struct {
	email string
	name  string
	roles []string
}

// EnsureSynced 把真实身份源同步为 users 投影，幂等可重复调用：
// 源 = 内置管理员（SuperAdminEmail 恒为管理员）+ 命名空间成员（补漏）。
// 登录用户由 SyncLoginUser 在登录热路径实时 upsert（不存在创建、存在推进 last_login），
// 同步源不再从登录事件全表扫描兜底——那会让每次同步全表扫 event 表（O(全部登录事件)），
// 事件表随平台使用持续增长，同步成本线性放大；SyncLoginUser 失败是 DB 故障级罕见瞬态，
// 下次登录自动补回，不值得为此付出全扫代价。由 UserBiz.Sync 触发（页面「同步用户」按钮）。
// 已有用户仅补空 name，绝不覆盖 roles（尊重管理员后台手动升降级，避免同步把角色冲回默认）。
// 落库顺序确定：内置管理员恒为首条（最先创建、ID 最小），随后成员按 ID 升序——不依赖
// map 迭代随机序（见 collectSyncSources）。
func (r *userRepo) EnsureSynced(ctx context.Context) (err error) {
	ctx, span := tracer.Start(ctx, "userRepo/EnsureSynced")
	defer func() { endSpan(span, err) }()
	db := r.data.DB()

	sources, err := r.collectSyncSources(ctx)
	if err != nil {
		return err
	}

	existing, err := db.User.Query().All(ctx)
	if err != nil {
		return errs.Wrap(err, "list existing users for sync")
	}
	byEmail := make(map[string]*ent.User, len(existing))
	for _, u := range existing {
		byEmail[u.Email] = u
	}

	for _, src := range sources {
		if cur, ok := byEmail[src.email]; ok {
			update := db.User.UpdateOneID(cur.ID)
			// 同步源（管理员/成员）不携带登录时间：只补空 name，last_login 由 SyncLoginUser 推进。
			if updateUserProjection(update, cur, nil, src.name) {
				if _, err := update.Save(ctx); err != nil {
					return errs.Wrap(err, "update user projection")
				}
			}
			continue
		}
		if _, err := db.User.Create().
			SetEmail(src.email).
			SetName(src.name).
			SetRoles(src.roles).
			Save(ctx); err != nil {
			return errs.Wrap(err, "create user projection")
		}
	}
	return nil
}

// SyncLoginUser 登录成功时按邮箱 upsert 用户投影（幂等可重复调用）：
// 已存在 → 仅推进最近登录（不倒退）+ 补空展示名（不覆盖非空）；不存在 → 创建
// （超级管理员恒 mars_admin，其余角色为空数组；展示名取登录名，空则回退邮箱本地部分；
// 最近登录 = 当前时刻）。与 EnsureSynced 的源语义一致，但只处理单邮箱（登录热路径，
// 不做全量对账）；roles 一律不覆盖，尊重管理员后台手动升降级。
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
// 仅补更晚的最近登录（已有登录时间不倒退，未登录用户补首次登录时间）+ 补空展示名
// （非空不被覆盖，来源为登录名或同步源固定名）。返回是否有字段被实际修改；调用方
// 负责 Save 与错误包裹，以保留各自错误消息语义。EnsureSynced（lastLogin 恒 nil，只
// 补空 name）与 SyncLoginUser（推进最近登录）共用。
func updateUserProjection(update *ent.UserUpdateOne, cur *ent.User, lastLogin *time.Time, name string) bool {
	changed := false
	if lastLogin != nil && (cur.LastLogin == nil || lastLogin.After(*cur.LastLogin)) {
		update.SetLastLogin(*lastLogin)
		changed = true
	}
	if cur.Name == "" && name != "" {
		update.SetName(name)
		changed = true
	}
	return changed
}

// collectSyncSources 汇总身份源并保证确定性顺序：内置管理员恒为首条，随后命名空间成员
// 按 ID 升序稳定补漏（未在管理员源中出现的邮箱，展示名回退邮箱本地部分）。返回有序切片
// 而非 map——EnsureSynced 按序落库，杜绝 Go map 迭代随机序把超管插到成员中间（超管必须
// 最先创建、ID 最小，作为平台拥有者锚定用户表）。登录用户由 SyncLoginUser 实时 upsert，
// 不再从登录事件全表扫描兜底（性能权衡见 EnsureSynced 注释）。
func (r *userRepo) collectSyncSources(ctx context.Context) ([]*syncSource, error) {
	db := r.data.DB()
	sources := make([]*syncSource, 0)

	// 源一：内置管理员，恒定角色且不可降级，恒为同步首条。
	sources = append(sources, &syncSource{
		email: biz.SuperAdminEmail,
		name:  biz.SuperAdminName,
		roles: []string{biz.MarsAdmin},
	})

	// 源二：命名空间成员补漏（未在管理员源中出现的邮箱），按 ID 升序保证稳定顺序。
	members, err := db.Member.Query().
		Select(member.FieldEmail).
		Where(member.EmailNEQ("")).
		Order(ent.Asc(member.FieldID)).
		All(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "query members for user sync")
	}
	// 邮箱统一小写归一（对齐 SyncLoginUser 的 ToLower 与 VerifyToken）：成员源与登录源
	// 归一化口径不一致会把同一逻辑身份分裂成两行——UNIQUE 索引大小写敏感，Bob@X.com 与
	// bob@x.com 互不冲突同时落库，列表重复、统计虚高、last_login 永远推不进成员源行。
	// 空串/纯空白成员邮箱跳过。
	seen := map[string]bool{biz.SuperAdminEmail: true}
	for _, m := range members {
		email := strings.ToLower(strings.TrimSpace(m.Email))
		if email == "" || seen[email] {
			continue
		}
		seen[email] = true
		sources = append(sources, &syncSource{
			email: email,
			name:  localPartOf(email),
			roles: []string{},
		})
	}
	return sources, nil
}

// localPartOf 取邮箱 @ 前的本地部分作展示名回退（真实展示名优先取事件 OIDC 名）。
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

	users, err := query.Clone().
		Order(ent.Desc(user.FieldLastLogin), ent.Desc(user.FieldID)).
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
