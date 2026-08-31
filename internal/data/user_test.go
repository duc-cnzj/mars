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

	require.NoError(t, repo.SyncLoginUser(ctx, "alice@x.com", "oidc-alias"))

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
	require.NoError(t, repo.SyncLoginUser(ctx, "linkaijian@uco.com", "林开建（AI）"))

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
