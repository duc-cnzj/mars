package data

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	entuser "github.com/duc-cnzj/mars/v6/internal/data/ent/user"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// newUserRepo 构造基于 sqlite 的 userRepo。
func newUserRepo(t *testing.T) (*userRepo, *ent.Client) {
	entdb, err := NewSqliteDB()
	require.NoError(t, err)
	t.Cleanup(func() { entdb.Close() })
	repo := NewUserRepo(NewDataImpl(&NewDataParams{DB: entdb, Cfg: &config.Config{}}), timer.NewReal())
	return repo.(*userRepo), entdb
}

// Test_userRepo_EnsureSynced_SeedsFromSources 内置管理员/命名空间成员两类源全部落库为
// users 投影，且重复调用幂等不产生重复行。登录用户由 SyncLoginUser 在登录热路径实时
// upsert，EnsureSynced 不再从登录事件兜底，故无登录源断言。
func Test_userRepo_EnsureSynced_SeedsFromSources(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	_, err := entdb.Member.Create().SetEmail("bob@x.com").Save(ctx)
	require.NoError(t, err)

	require.NoError(t, repo.EnsureSynced(ctx))

	byEmail := make(map[string]*ent.User)
	for _, u := range entdb.User.Query().AllX(ctx) {
		byEmail[u.Email] = u
	}
	require.Len(t, byEmail, 2, "应落库内置管理员 + 命名空间成员两行")

	admin := byEmail[biz.SuperAdminEmail]
	require.NotNil(t, admin)
	assert.Equal(t, biz.SuperAdminName, admin.Name)
	assert.Equal(t, []string{biz.MarsAdmin}, admin.Roles)
	assert.Nil(t, admin.LastLogin, "同步源不携带登录时间")

	bob := byEmail["bob@x.com"]
	require.NotNil(t, bob)
	assert.Equal(t, "bob", bob.Name, "成员源展示名回退邮箱本地部分")
	assert.Equal(t, []string{}, bob.Roles)
	assert.Nil(t, bob.LastLogin, "成员源无登录时间")
	assert.Less(t, admin.ID, bob.ID, "同步顺序：内置管理员恒先落库（ID 最小）")

	// 幂等：重复同步不新增行
	require.NoError(t, repo.EnsureSynced(ctx))
	assert.Len(t, entdb.User.Query().AllX(ctx), 2)
}

// Test_userRepo_EnsureSynced_SuperAdminFirst 多成员并发同步时内置管理员恒为第一条落库，
// ID 小于所有成员——回归锁定「同步用户时第一个先同步超级管理员」：collectSyncSources 若按
// map 遍历，Go 迭代随机序会把超管插到成员中间，ID 不再最小（此用例断言 >1 成员时仍成立）。
func Test_userRepo_EnsureSynced_SuperAdminFirst(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	for _, email := range []string{"a@x.com", "b@x.com", "c@x.com"} {
		_, err := entdb.Member.Create().SetEmail(email).Save(ctx)
		require.NoError(t, err)
	}
	require.NoError(t, repo.EnsureSynced(ctx))

	admin := entdb.User.Query().Where(entuser.EmailEQ(biz.SuperAdminEmail)).OnlyX(ctx)
	for _, u := range entdb.User.Query().AllX(ctx) {
		if u.Email == biz.SuperAdminEmail {
			continue
		}
		assert.Less(t, admin.ID, u.ID, "内置管理员应最先落库（ID 最小），got admin=%d member=%s id=%d", admin.ID, u.Email, u.ID)
	}
}

// Test_userRepo_EnsureSynced_PreservesRoles 已有用户手动升降级后再次同步不被冲回默认。
// last_login 由 SyncLoginUser 推进，EnsureSynced 不触碰登录时间（同步源不携带）。
func Test_userRepo_EnsureSynced_PreservesRoles(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	_, err := entdb.Member.Create().SetEmail("bob@x.com").Save(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.EnsureSynced(ctx))

	// 手动升级为管理员后，再次同步 roles 不被覆盖
	require.NoError(t, repo.ToggleAdmin(ctx, "bob@x.com", true))
	require.NoError(t, repo.EnsureSynced(ctx))

	bob := entdb.User.Query().Where(entuser.EmailEQ("bob@x.com")).OnlyX(ctx)
	assert.True(t, slices.Contains(bob.Roles, biz.MarsAdmin), "同步不得覆盖手动角色")
	assert.Nil(t, bob.LastLogin, "同步源不携带登录时间，不触碰 last_login")
}

// Test_userRepo_EnsureSynced_MemberEmailWithoutAtSign 成员邮箱不含 @ 时展示名
// 回退为邮箱本身（localPartOf 的兜底分支）。
func Test_userRepo_EnsureSynced_MemberEmailWithoutAtSign(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	_, err := entdb.Member.Create().SetEmail("plain").Save(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.EnsureSynced(ctx))

	member := entdb.User.Query().Where(entuser.EmailEQ("plain")).OnlyX(ctx)
	assert.Equal(t, "plain", member.Name, "无 @ 邮箱展示名回退为邮箱本身")
}

// Test_userRepo_EnsureSynced_MemberEmailHitsExistingSourceSkip 成员邮箱命中内置管理员源时
// 跳过不重复入列：collectSyncSources 的 sources 已存在分支不产生重复行、不改写角色。
func Test_userRepo_EnsureSynced_MemberEmailHitsExistingSourceSkip(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	_, err := entdb.Member.Create().SetEmail(biz.SuperAdminEmail).Save(ctx)
	require.NoError(t, err)

	require.NoError(t, repo.EnsureSynced(ctx))

	admins := entdb.User.Query().Where(entuser.EmailEQ(biz.SuperAdminEmail)).AllX(ctx)
	assert.Len(t, admins, 1, "成员邮箱命中已有源时不得产生重复行")
	assert.Equal(t, []string{biz.MarsAdmin}, admins[0].Roles, "角色不被成员源改写")
}

// Test_userRepo_EnsureSynced_MemberEmailLowercased 成员邮箱混合大小写同步时统一小写落库：
// collectSyncSources 与 SyncLoginUser（ToLower 归一）口径一致，杜绝同一逻辑身份因
// UNIQUE 索引大小写敏感分裂成 Bob@X.com + bob@x.com 两行——列表重复、统计虚高、
// 成员源行 last_login 永远推不进。
func Test_userRepo_EnsureSynced_MemberEmailLowercased(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	_, err := entdb.Member.Create().SetEmail("Bob@X.com").Save(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.EnsureSynced(ctx))

	// 成员源落库必须小写归一，不允许出现 Bob@X.com 行
	bobUpper := entdb.User.Query().Where(entuser.EmailEQ("Bob@X.com")).AllX(ctx)
	assert.Empty(t, bobUpper, "混合大小写邮箱不得原样落库")

	bob := entdb.User.Query().Where(entuser.EmailEQ("bob@x.com")).OnlyX(ctx)
	assert.NotNil(t, bob, "成员源应小写落库为 bob@x.com")
	assert.Equal(t, []string{}, bob.Roles, "成员源角色为空数组=普通用户")

	// 该用户走 OIDC 登录（大小写混合输入）：小写归一命中既有行，推进 last_login 而非新建重复行
	require.NoError(t, repo.SyncLoginUser(ctx, "Bob@X.com", "bob"))
	rows := entdb.User.Query().AllX(ctx)
	assert.Len(t, rows, 2, "超管 + bob@x.com，不允许分裂成两行")
	assert.NotNil(t, entdb.User.Query().Where(entuser.EmailEQ("bob@x.com")).OnlyX(ctx).LastLogin, "OIDC 登录应推进 last_login")
}

// Test_userRepo_EnsureSynced_NameOnlyFillsEmpty 展示名仅补空：已有非空展示名不被覆盖。
func Test_userRepo_EnsureSynced_NameOnlyFillsEmpty(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	_, err := entdb.Member.Create().SetEmail("bob@x.com").Save(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.EnsureSynced(ctx))

	// 清空展示名后重新同步应补回；改名后不应被覆盖
	_, err = entdb.User.Update().Where(entuser.EmailEQ("bob@x.com")).SetName("").Save(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.EnsureSynced(ctx))
	assert.Equal(t, "bob", entdb.User.Query().Where(entuser.EmailEQ("bob@x.com")).OnlyX(ctx).Name)

	_, err = entdb.User.Update().Where(entuser.EmailEQ("bob@x.com")).SetName("自定义名").Save(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.EnsureSynced(ctx))
	assert.Equal(t, "自定义名", entdb.User.Query().Where(entuser.EmailEQ("bob@x.com")).OnlyX(ctx).Name)
}

// Test_userRepo_List_SearchAdminOnlyStats 覆盖搜索/仅管理员过滤/全量统计/分页。
func Test_userRepo_List_SearchAdminOnlyStats(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	for _, u := range []struct {
		email, name string
		roles       []string
	}{
		{"alice@x.com", "alice", []string{}},
		{"carol@x.com", "carol", []string{}},
		{"boss@x.com", "boss", []string{biz.MarsAdmin}},
	} {
		_, err := entdb.User.Create().SetEmail(u.email).SetName(u.name).SetRoles(u.roles).Save(ctx)
		require.NoError(t, err)
	}

	t.Run("search matches email or name", func(t *testing.T) {
		out, err := repo.List(ctx, &biz.ListUserInput{Page: 1, PageSize: 10, Search: "ali"})
		require.NoError(t, err)
		if assert.Len(t, out.Items, 1) {
			assert.Equal(t, "alice@x.com", out.Items[0].Email)
		}
		assert.Equal(t, int32(1), out.Pag.Count)
		// 统计按全量口径，不受搜索影响
		assert.Equal(t, int32(3), out.Stats.Total)
	})

	t.Run("admin only", func(t *testing.T) {
		out, err := repo.List(ctx, &biz.ListUserInput{Page: 1, PageSize: 10, AdminOnly: true})
		require.NoError(t, err)
		if assert.Len(t, out.Items, 1) {
			assert.Equal(t, "boss@x.com", out.Items[0].Email)
		}
		assert.Equal(t, int32(1), out.Pag.Count)
	})

	t.Run("full list stats", func(t *testing.T) {
		out, err := repo.List(ctx, &biz.ListUserInput{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Len(t, out.Items, 3)
		assert.Equal(t, int32(3), out.Pag.Count)
		assert.Equal(t, int32(3), out.Stats.Total)
		assert.Equal(t, int32(1), out.Stats.Admins)
		assert.Equal(t, int32(2), out.Stats.Regular)
	})
}

// Test_userRepo_ToggleAdmin 覆盖升降级/幂等/超级管理员保护/未找到。
func Test_userRepo_ToggleAdmin(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	_, err := entdb.User.Create().SetEmail("alice@x.com").SetName("alice").SetRoles([]string{}).Save(ctx)
	require.NoError(t, err)

	require.NoError(t, repo.ToggleAdmin(ctx, "alice@x.com", true))
	assert.True(t, slices.Contains(entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx).Roles, biz.MarsAdmin))

	// 重复升级幂等无错
	require.NoError(t, repo.ToggleAdmin(ctx, "alice@x.com", true))

	require.NoError(t, repo.ToggleAdmin(ctx, "alice@x.com", false))
	assert.False(t, slices.Contains(entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx).Roles, biz.MarsAdmin))

	// 超级管理员不可降级
	err = repo.ToggleAdmin(ctx, biz.SuperAdminEmail, false)
	assert.Error(t, err)

	// 未找到用户
	err = repo.ToggleAdmin(ctx, "nobody@x.com", true)
	assert.Error(t, err)
}

// Test_userRepo_List_ErrorBranch 关闭 DB 下 List 返回错误而非 panic。
func Test_userRepo_List_ErrorBranch(t *testing.T) {
	repo := NewUserRepo(NewDataImpl(&NewDataParams{DB: mustClosedDB(t), Cfg: &config.Config{}}), timer.NewReal()).(*userRepo)
	_, err := repo.List(context.TODO(), &biz.ListUserInput{Page: 1, PageSize: 10})
	assert.Error(t, err)
}

// Test_userRepo_EnsureSynced_ErrorBranch 关闭 DB 下同步源查询失败即返回错误。
func Test_userRepo_EnsureSynced_ErrorBranch(t *testing.T) {
	repo := NewUserRepo(NewDataImpl(&NewDataParams{DB: mustClosedDB(t), Cfg: &config.Config{}}), timer.NewReal()).(*userRepo)
	assert.Error(t, repo.EnsureSynced(context.TODO()))
}

// TestToUser 覆盖 nil 与实体两种转换。
func TestToUser(t *testing.T) {
	assert.Nil(t, toUser(nil))
	now := time.Now()
	u := toUser(&ent.User{ID: 1, Email: "a@b.c", Name: "a", Roles: []string{}, LastLogin: &now, CreatedAt: now})
	assert.Equal(t, 1, u.ID)
	assert.Equal(t, "a@b.c", u.Email)
	assert.Equal(t, "a", u.Name)
	assert.Empty(t, u.Roles)
	assert.Equal(t, &now, u.LastLogin)
}

// TestToggleRole 覆盖追加/移除/幂等去重。
func TestToggleRole(t *testing.T) {
	assert.Equal(t, []string{"mars_admin"}, toggleRole([]string{}, "mars_admin", true))
	assert.Equal(t, []string{"mars_admin"}, toggleRole([]string{"mars_admin"}, "mars_admin", true), "已存在时保持去重")
	assert.Empty(t, toggleRole([]string{"mars_admin"}, "mars_admin", false))
	assert.Empty(t, toggleRole([]string{}, "mars_admin", false), "不存在时移除为幂等")
	assert.Empty(t, toggleRole(nil, "mars_admin", false))
	// 历史遗留的非 mars_admin 角色（如旧版写入的 "user"）原样保留，不因升降级被清掉
	assert.Equal(t, []string{"other", "mars_admin"}, toggleRole([]string{"other"}, "mars_admin", true))
	assert.Equal(t, []string{"other"}, toggleRole([]string{"other", "mars_admin"}, "mars_admin", false))
}

// Test_userRepo_SyncLoginUser_CreateNew 新邮箱登录创建投影：roles=[]（空角色=普通用户），展示名取登录名，
// last_login 为登录时刻，且邮箱统一小写归一（对齐 VerifyToken）。
func Test_userRepo_SyncLoginUser_CreateNew(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	after := time.Now()
	require.NoError(t, repo.SyncLoginUser(ctx, "ALICE@X.com", "alice"))

	u := entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx)
	assert.Equal(t, "alice", u.Name)
	assert.Equal(t, []string{}, u.Roles, "新用户默认普通用户")
	require.NotNil(t, u.LastLogin)
	assert.WithinDuration(t, after, *u.LastLogin, time.Minute, "last_login 应为登录时刻")

	// 幂等：重复同步不新增行
	require.NoError(t, repo.SyncLoginUser(ctx, "alice@x.com", "alice"))
	assert.Len(t, entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).AllX(ctx), 1)
}

// Test_userRepo_SyncLoginUser_CreateSuperAdmin 超级管理员邮箱登录创建时恒为管理员角色。
func Test_userRepo_SyncLoginUser_CreateSuperAdmin(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	require.NoError(t, repo.SyncLoginUser(ctx, biz.SuperAdminEmail, "boss"))

	u := entdb.User.Query().Where(entuser.EmailEQ(biz.SuperAdminEmail)).OnlyX(ctx)
	assert.Equal(t, []string{biz.MarsAdmin}, u.Roles, "超级管理员登录创建恒为管理员角色")
}

// Test_userRepo_SyncLoginUser_EmptyNameFallsBackToLocalPart 登录名为空时新建用户展示名
// 回退邮箱本地部分。
func Test_userRepo_SyncLoginUser_EmptyNameFallsBackToLocalPart(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	require.NoError(t, repo.SyncLoginUser(ctx, "ghost@x.com", ""))

	u := entdb.User.Query().Where(entuser.EmailEQ("ghost@x.com")).OnlyX(ctx)
	assert.Equal(t, "ghost", u.Name, "空登录名应回退邮箱本地部分")
}

// Test_userRepo_SyncLoginUser_UpdateExisting 已存在用户：推进 last_login + 补空展示名，
// 手动升级的管理员角色不被同步覆盖。
func Test_userRepo_SyncLoginUser_UpdateExisting(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	old := time.Now().Add(-time.Hour)
	_, err := entdb.User.Create().
		SetEmail("alice@x.com").
		SetName("").
		SetRoles([]string{}).
		SetNillableLastLogin(&old).
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.ToggleAdmin(ctx, "alice@x.com", true))

	require.NoError(t, repo.SyncLoginUser(ctx, "alice@x.com", "alice"))

	u := entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx)
	assert.Equal(t, "alice", u.Name, "空展示名应补全")
	assert.True(t, slices.Contains(u.Roles, biz.MarsAdmin), "同步不得覆盖手动角色")
	require.NotNil(t, u.LastLogin)
	assert.True(t, u.LastLogin.After(old), "last_login 应前进到登录时刻")
}

// Test_userRepo_SyncLoginUser_DoesNotOverrideName 已有非空展示名不被登录名覆盖。
func Test_userRepo_SyncLoginUser_DoesNotOverrideName(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	_, err := entdb.User.Create().
		SetEmail("alice@x.com").
		SetName("自定义名").
		SetRoles([]string{}).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, repo.SyncLoginUser(ctx, "alice@x.com", "oidc-alias"))

	u := entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx)
	assert.Equal(t, "自定义名", u.Name, "非空展示名不被覆盖")
}

// Test_userRepo_SyncLoginUser_DoesNotRewindLastLogin 已有更晚登录时间不倒退。
func Test_userRepo_SyncLoginUser_DoesNotRewindLastLogin(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	future := time.Now().Add(time.Hour)
	_, err := entdb.User.Create().
		SetEmail("alice@x.com").
		SetName("alice").
		SetRoles([]string{}).
		SetNillableLastLogin(&future).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, repo.SyncLoginUser(ctx, "alice@x.com", "alice"))

	u := entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx)
	assert.WithinDuration(t, future, *u.LastLogin, time.Millisecond, "last_login 不得倒退")
}

// Test_userRepo_SyncLoginUser_EmptyEmail 空邮箱是确定语义错误：返回 InvalidArgument。
func Test_userRepo_SyncLoginUser_EmptyEmail(t *testing.T) {
	repo, _ := newUserRepo(t)

	err := repo.SyncLoginUser(context.TODO(), "  ", "alice")
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "空邮箱应判定为参数不合法，got %v", err)
}
