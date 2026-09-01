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
	return NewUserRepo(NewDataImpl(&NewDataParams{DB: entdb, Cfg: &config.Config{}}), timer.NewReal()), entdb
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

// Test_userRepo_List_SortAscDesc 排序方向：默认（空 sort）最近登录倒序在前；
// asc 最早登录在前；无论升降序，从未登录（last_login 为 NULL）者都垫底。
func Test_userRepo_List_SortAscDesc(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	now := time.Now()
	recentLogin := now.Add(-2 * time.Hour)
	oldLogin := now.Add(-30 * 24 * time.Hour)
	for _, u := range []struct {
		email     string
		lastLogin *time.Time
	}{
		{"recent@x.com", &recentLogin},
		{"never@x.com", nil},
		{"old@x.com", &oldLogin},
	} {
		_, err := entdb.User.Create().SetEmail(u.email).SetName(u.email).SetRoles([]string{}).SetNillableLastLogin(u.lastLogin).Save(ctx)
		require.NoError(t, err)
	}

	t.Run("desc default", func(t *testing.T) {
		out, err := repo.List(ctx, &biz.ListUserInput{Page: 1, PageSize: 10})
		require.NoError(t, err)
		require.Len(t, out.Items, 3)
		assert.Equal(t, "recent@x.com", out.Items[0].Email)
		assert.Equal(t, "old@x.com", out.Items[1].Email)
		assert.Equal(t, "never@x.com", out.Items[2].Email) // 从未登录垫底
	})

	t.Run("asc", func(t *testing.T) {
		out, err := repo.List(ctx, &biz.ListUserInput{Page: 1, PageSize: 10, Sort: "asc"})
		require.NoError(t, err)
		require.Len(t, out.Items, 3)
		assert.Equal(t, "old@x.com", out.Items[0].Email) // 最早登录在前
		assert.Equal(t, "recent@x.com", out.Items[1].Email)
		assert.Equal(t, "never@x.com", out.Items[2].Email) // 从未登录仍垫底
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
	assert.True(t, entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx).RolesOverride, "手动升级后应置角色接管标记")

	// 重复升级幂等无错
	require.NoError(t, repo.ToggleAdmin(ctx, "alice@x.com", true))

	require.NoError(t, repo.ToggleAdmin(ctx, "alice@x.com", false))
	assert.False(t, slices.Contains(entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx).Roles, biz.MarsAdmin))
	assert.True(t, entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx).RolesOverride, "降级后接管标记保持，SSO 不再覆盖")

	// 超级管理员不可降级
	err = repo.ToggleAdmin(ctx, biz.SuperAdminEmail, false)
	assert.Error(t, err)

	// 未找到用户
	err = repo.ToggleAdmin(ctx, "nobody@x.com", true)
	assert.Error(t, err)
}

// Test_userRepo_ResetRolesOverride 解除后台手动接管：置回 roles_override=false 恢复 SSO 角色
// 同步（已生效角色不被改动），幂等早退；未接管用户与不存在的邮箱各自安全处理。
func Test_userRepo_ResetRolesOverride(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	_, err := entdb.User.Create().SetEmail("alice@x.com").SetName("alice").SetRoles([]string{}).Save(ctx)
	require.NoError(t, err)

	// 先手动升级（接管）再解除：roles 保持（解除只清接管标记，不改变已生效角色），恢复 SSO 同步
	require.NoError(t, repo.ToggleAdmin(ctx, "alice@x.com", true))
	require.NoError(t, repo.ResetRolesOverride(ctx, "alice@x.com"))
	u := entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx)
	assert.True(t, slices.Contains(u.Roles, biz.MarsAdmin), "解除接管不改变已生效角色")
	assert.False(t, u.RolesOverride, "解除接管后 roles_override 置回 false")

	// 幂等：本就未接管时再次解除无错且保持 false
	require.NoError(t, repo.ResetRolesOverride(ctx, "alice@x.com"))
	assert.False(t, entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx).RolesOverride)

	// 邮箱小写归一：大写入参能匹配已归一存储的邮箱
	require.NoError(t, repo.ToggleAdmin(ctx, "alice@x.com", true))
	require.NoError(t, repo.ResetRolesOverride(ctx, "ALICE@X.COM"))
	assert.False(t, entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx).RolesOverride)

	// 不存在的邮箱 → NotFound 上抛
	err = repo.ResetRolesOverride(ctx, "nobody@x.com")
	assert.Error(t, err)
}

// Test_userRepo_List_ErrorBranch 关闭 DB 下 List 返回错误而非 panic。
func Test_userRepo_List_ErrorBranch(t *testing.T) {
	repo := NewUserRepo(NewDataImpl(&NewDataParams{DB: mustClosedDB(t), Cfg: &config.Config{}}), timer.NewReal())
	_, err := repo.List(context.TODO(), &biz.ListUserInput{Page: 1, PageSize: 10})
	assert.Error(t, err)
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

// Test_userRepo_SyncLoginUser_CreateNew 新邮箱登录创建投影：角色取登录身份（此处为空=
// 普通用户），展示名取登录名，last_login 为登录时刻，且邮箱统一小写归一（对齐 VerifyToken）。
func Test_userRepo_SyncLoginUser_CreateNew(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	after := time.Now()
	require.NoError(t, repo.SyncLoginUser(ctx, "ALICE@X.com", "alice", []string{}))

	u := entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx)
	assert.Equal(t, "alice", u.Name)
	assert.Equal(t, []string{}, u.Roles, "无角色身份的登录创建为普通用户")
	assert.False(t, u.RolesOverride, "新建用户默认未手动接管，跟随 SSO")
	require.NotNil(t, u.LastLogin)
	assert.WithinDuration(t, after, *u.LastLogin, time.Minute, "last_login 应为登录时刻")

	// 幂等：重复同步不新增行
	require.NoError(t, repo.SyncLoginUser(ctx, "alice@x.com", "alice", []string{}))
	assert.Len(t, entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).AllX(ctx), 1)
}

// Test_userRepo_SyncLoginUser_CreateSeedsSSORoles 新邮箱登录创建投影时，SSO id_token
// 携带的角色（mars_admin）种入 users 表，使用户管理页能反映 SSO 身份。
func Test_userRepo_SyncLoginUser_CreateSeedsSSORoles(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	require.NoError(t, repo.SyncLoginUser(ctx, "admin@x.com", "sso-admin", []string{biz.MarsAdmin}))

	u := entdb.User.Query().Where(entuser.EmailEQ("admin@x.com")).OnlyX(ctx)
	assert.Equal(t, []string{biz.MarsAdmin}, u.Roles, "新建用户应种入 SSO 带来的 mars_admin")
	assert.False(t, u.RolesOverride)
}

// Test_userRepo_SyncLoginUser_CreateSuperAdmin 超级管理员邮箱登录创建时恒为管理员角色
// （即使登录身份角色为空，守卫也保证超管不降级）。
func Test_userRepo_SyncLoginUser_CreateSuperAdmin(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	require.NoError(t, repo.SyncLoginUser(ctx, biz.SuperAdminEmail, "boss", []string{}))

	u := entdb.User.Query().Where(entuser.EmailEQ(biz.SuperAdminEmail)).OnlyX(ctx)
	assert.Equal(t, []string{biz.MarsAdmin}, u.Roles, "超级管理员登录创建恒为管理员角色")
}

// Test_userRepo_SyncLoginUser_EmptyNameFallsBackToLocalPart 登录名为空时新建用户展示名
// 回退邮箱本地部分。
func Test_userRepo_SyncLoginUser_EmptyNameFallsBackToLocalPart(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	require.NoError(t, repo.SyncLoginUser(ctx, "ghost@x.com", "", []string{}))

	u := entdb.User.Query().Where(entuser.EmailEQ("ghost@x.com")).OnlyX(ctx)
	assert.Equal(t, "ghost", u.Name, "空登录名应回退邮箱本地部分")
}

// Test_userRepo_SyncLoginUser_UpdateExisting 已存在用户：推进 last_login + 补空展示名，
// 后台手动管理接管（roles_override=true）的角色不被登录同步覆盖。
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

	// SSO 登录身份不带管理员角色：手动升级的 admin 不被覆盖
	require.NoError(t, repo.SyncLoginUser(ctx, "alice@x.com", "alice", []string{}))

	u := entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx)
	assert.Equal(t, "alice", u.Name, "空展示名应补全")
	assert.True(t, slices.Contains(u.Roles, biz.MarsAdmin), "同步不得覆盖手动角色")
	assert.True(t, u.RolesOverride, "手动升降级后应标记角色接管")
	require.NotNil(t, u.LastLogin)
	assert.True(t, u.LastLogin.After(old), "last_login 应前进到登录时刻")
}

// Test_userRepo_SyncLoginUser_SyncsRolesWhenNotOverridden 未被手动接管的用户（
// roles_override=false）登录时按登录身份同步角色：SSO 带 mars_admin 即写入 users 表。
func Test_userRepo_SyncLoginUser_SyncsRolesWhenNotOverridden(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	// 既有投影：普通用户，未手动接管
	_, err := entdb.User.Create().
		SetEmail("alice@x.com").
		SetName("alice").
		SetRoles([]string{}).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, repo.SyncLoginUser(ctx, "alice@x.com", "alice", []string{biz.MarsAdmin}))

	u := entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx)
	assert.Equal(t, []string{biz.MarsAdmin}, u.Roles, "未接管用户应同步 SSO 带来的 mars_admin")
	assert.False(t, u.RolesOverride)
}

// Test_userRepo_SyncLoginUser_OverrideKeepsManualRoles 已被手动接管的用户登录时不覆盖
// 角色：即使 SSO 再次带 mars_admin，后台手动降权也不被洗掉。
func Test_userRepo_SyncLoginUser_OverrideKeepsManualRoles(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	// 手动降权：先建为管理员再降为普通用户（ToggleAdmin 置接管标记）
	_, err := entdb.User.Create().
		SetEmail("alice@x.com").
		SetName("alice").
		SetRoles([]string{biz.MarsAdmin}).
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.ToggleAdmin(ctx, "alice@x.com", false))

	// SSO 登录身份带 mars_admin：手动降权优先，不重新升权
	require.NoError(t, repo.SyncLoginUser(ctx, "alice@x.com", "alice", []string{biz.MarsAdmin}))

	u := entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx)
	assert.NotContains(t, u.Roles, biz.MarsAdmin, "手动接管后 SSO 再带 admin 也不覆盖降权")
	assert.True(t, u.RolesOverride)
}

// Test_userRepo_SyncLoginUser_DoesNotOverrideName 已有非空展示名（手动设置，≠ email
// 本地部分）不被登录名覆盖。
func Test_userRepo_SyncLoginUser_DoesNotOverrideName(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	_, err := entdb.User.Create().
		SetEmail("alice@x.com").
		SetName("自定义名").
		SetRoles([]string{}).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, repo.SyncLoginUser(ctx, "alice@x.com", "oidc-alias", []string{}))

	u := entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx)
	assert.Equal(t, "自定义名", u.Name, "非空展示名不被覆盖")
}

// Test_userRepo_SyncLoginUser_UpgradesLocalPartName 既有用户名为「email 本地部分」
// （空登录名的自动默认值）时，带真实登录名登录应升级展示名并推进最近登录。
func Test_userRepo_SyncLoginUser_UpgradesLocalPartName(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	// 既有投影：展示名为 email 本地部分（空登录名自动默认值落库）
	_, err := entdb.User.Create().
		SetEmail("linkaijian@uco.com").
		SetName("linkaijian").
		SetRoles([]string{}).
		Save(ctx)
	require.NoError(t, err)

	// 带真实 OIDC 登录名再次登录：本地部分默认名升级为真实名
	require.NoError(t, repo.SyncLoginUser(ctx, "linkaijian@uco.com", "林开建（AI）", []string{}))

	u := entdb.User.Query().Where(entuser.EmailEQ("linkaijian@uco.com")).OnlyX(ctx)
	assert.Equal(t, "林开建（AI）", u.Name, "email 本地部分默认名应升级为真实登录名")
	assert.NotNil(t, u.LastLogin, "登录应推进 last_login")
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

	require.NoError(t, repo.SyncLoginUser(ctx, "alice@x.com", "alice", []string{}))

	u := entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx)
	assert.WithinDuration(t, future, *u.LastLogin, time.Millisecond, "last_login 不得倒退")
}

// Test_userRepo_SyncLoginUser_NilRoles SSO id_token 无 roles claim 时角色为 nil：归一为
// 空数组存储（统一 JSON 数组而非 null），不破坏「普通用户」语义。
func Test_userRepo_SyncLoginUser_NilRoles(t *testing.T) {
	repo, entdb := newUserRepo(t)
	ctx := context.TODO()

	require.NoError(t, repo.SyncLoginUser(ctx, "alice@x.com", "alice", nil))

	u := entdb.User.Query().Where(entuser.EmailEQ("alice@x.com")).OnlyX(ctx)
	assert.Equal(t, []string{}, u.Roles, "nil roles 应归一为空数组")
}

// TestLocalPartOf 覆盖邮箱本地部分提取：含 @ 取前缀，无 @ 原样返回。
func TestLocalPartOf(t *testing.T) {
	assert.Equal(t, "alice", localPartOf("alice@x.com"))
	assert.Equal(t, "nouser", localPartOf("nouser"), "无 @ 的邮箱原样返回")
	assert.Equal(t, "", localPartOf(""), "空串原样返回")
}

// Test_userRepo_SyncLoginUser_EmptyEmail 空邮箱是确定语义错误：返回 InvalidArgument。
func Test_userRepo_SyncLoginUser_EmptyEmail(t *testing.T) {
	repo, _ := newUserRepo(t)

	err := repo.SyncLoginUser(context.TODO(), "  ", "alice", []string{})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "空邮箱应判定为参数不合法，got %v", err)
}
