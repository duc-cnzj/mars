package data

import (
	"context"
	"errors"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newUserMockEntClient 构造由 go-sqlmock 驱动的 ent client，用于覆盖
// SyncLoginUser 中 DB 写/查失败的分支，无需真实 MySQL。
func newUserMockEntClient(t *testing.T) (*ent.Client, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	drv := entsql.OpenDB(dialect.MySQL, db)
	client := ent.NewClient(ent.Driver(drv))
	t.Cleanup(func() { _ = db.Close() })
	return client, mock
}

// usersQuery 匹配用户全量查询（SyncLoginUser 邮箱命中查询）。
var usersQuery = "SELECT .* FROM `users`"

// Test_userRepo_SyncLoginUser_QueryError 邮箱查询失败（非 NotFound，如 DB 断开）即返回
// 错误：不得误走创建分支吞掉底层故障。
func Test_userRepo_SyncLoginUser_QueryError(t *testing.T) {
	client, mock := newUserMockEntClient(t)
	repo := NewUserRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}), timer.NewReal())
	ctx := context.TODO()

	mock.ExpectQuery(usersQuery).WillReturnError(errors.New("db boom"))

	err := repo.SyncLoginUser(ctx, "alice@x.com", "alice", []string{})
	assert.ErrorContains(t, err, "query user on login")
}

// Test_userRepo_SyncLoginUser_CreateSaveError 邮箱不存在（查询空行）走创建，INSERT 失败
// 即返回错误。
func Test_userRepo_SyncLoginUser_CreateSaveError(t *testing.T) {
	client, mock := newUserMockEntClient(t)
	repo := NewUserRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}), timer.NewReal())
	ctx := context.TODO()

	mock.ExpectQuery(usersQuery).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "name", "roles", "last_login", "roles_override"}))
	mock.ExpectExec("INSERT INTO `users`").WillReturnError(errors.New("db boom"))

	err := repo.SyncLoginUser(ctx, "alice@x.com", "alice", []string{})
	assert.ErrorContains(t, err, "create user projection on login")
}

// Test_userRepo_SyncLoginUser_UpdateSaveError 已存在用户需补名/补登录时间时 UPDATE 失败
// 即返回错误。
func Test_userRepo_SyncLoginUser_UpdateSaveError(t *testing.T) {
	client, mock := newUserMockEntClient(t)
	repo := NewUserRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}), timer.NewReal())
	ctx := context.TODO()

	now := time.Now()
	mock.ExpectQuery(usersQuery).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "name", "roles", "last_login", "roles_override"}).
		AddRow(1, now, now, "alice@x.com", "", `[]`, nil, false))
	mock.ExpectExec("UPDATE `users`").WillReturnError(errors.New("db boom"))

	err := repo.SyncLoginUser(ctx, "alice@x.com", "alice", []string{})
	assert.ErrorContains(t, err, "update user projection on login")
}

// userRowCols 是用户行查询的列集合（含 roles_override），供 EffectiveRoles 测试复用。
var userRowCols = []string{"id", "created_at", "updated_at", "email", "name", "roles", "last_login", "roles_override"}

// Test_userRepo_EffectiveRoles_NotOverridden 未被后台接管：生效角色回落登录身份携带的
// SSO 角色（users 表角色在登录时已同步，但以当前登录身份为准）。
func Test_userRepo_EffectiveRoles_NotOverridden(t *testing.T) {
	client, mock := newUserMockEntClient(t)
	repo := NewUserRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}), timer.NewReal())
	ctx := context.TODO()

	now := time.Now()
	mock.ExpectQuery(usersQuery).WillReturnRows(sqlmock.NewRows(userRowCols).
		AddRow(1, now, now, "alice@x.com", "alice", `[]`, nil, false))

	roles, err := repo.EffectiveRoles(ctx, "alice@x.com", []string{biz.MarsAdmin})
	assert.NoError(t, err)
	assert.Equal(t, []string{biz.MarsAdmin}, roles, "未接管时生效角色 = SSO 登录角色")
}

// Test_userRepo_EffectiveRoles_Overridden 已被后台手动接管：生效角色取 users 表手工角色，
// 忽略登录身份角色（SSO 即使下次仍带 mars_admin 也不覆盖降权）。
func Test_userRepo_EffectiveRoles_Overridden(t *testing.T) {
	client, mock := newUserMockEntClient(t)
	repo := NewUserRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}), timer.NewReal())
	ctx := context.TODO()

	now := time.Now()
	mock.ExpectQuery(usersQuery).WillReturnRows(sqlmock.NewRows(userRowCols).
		AddRow(1, now, now, "alice@x.com", "alice", `["mars_admin"]`, nil, true))

	roles, err := repo.EffectiveRoles(ctx, "alice@x.com", []string{})
	assert.NoError(t, err)
	assert.Equal(t, []string{biz.MarsAdmin}, roles, "接管后生效角色 = users 表手工角色")
}

// Test_userRepo_EffectiveRoles_UserNotFound 用户尚未落投影（首次登录前窗口）：回落登录身份
// 角色，不视为故障。
func Test_userRepo_EffectiveRoles_UserNotFound(t *testing.T) {
	client, mock := newUserMockEntClient(t)
	repo := NewUserRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}), timer.NewReal())
	ctx := context.TODO()

	mock.ExpectQuery(usersQuery).WillReturnRows(sqlmock.NewRows(userRowCols))

	roles, err := repo.EffectiveRoles(ctx, "alice@x.com", []string{biz.MarsAdmin})
	assert.NoError(t, err)
	assert.Equal(t, []string{biz.MarsAdmin}, roles)
}

// Test_userRepo_EffectiveRoles_QueryError 用户表读取失败（DB 断开等非 NotFound）：上抛错误，
// 由调用方决定回落策略，不在此处静默吞错。
func Test_userRepo_EffectiveRoles_QueryError(t *testing.T) {
	client, mock := newUserMockEntClient(t)
	repo := NewUserRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}), timer.NewReal())
	ctx := context.TODO()

	mock.ExpectQuery(usersQuery).WillReturnError(errors.New("db boom"))

	_, err := repo.EffectiveRoles(ctx, "alice@x.com", []string{})
	assert.ErrorContains(t, err, "query user for effective roles")
}

// Test_userRepo_EffectiveRoles_SuperAdminGuard 超级管理员恒 mars_admin：即使登录身份未带
// 管理员角色也兜底补上（SSO 侧无法降级内置超管），守卫在 DB 查询前生效。
func Test_userRepo_EffectiveRoles_SuperAdminGuard(t *testing.T) {
	client, mock := newUserMockEntClient(t)
	repo := NewUserRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}), timer.NewReal())
	ctx := context.TODO()

	// 超管尚未落投影：守卫已把 mars_admin 补进登录角色，空查询回落该角色。
	mock.ExpectQuery(usersQuery).WillReturnRows(sqlmock.NewRows(userRowCols))

	roles, err := repo.EffectiveRoles(ctx, biz.SuperAdminEmail, []string{})
	assert.NoError(t, err)
	assert.Contains(t, roles, biz.MarsAdmin)
}
