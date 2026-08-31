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
// EnsureSynced 中 DB 写/查失败的分支，无需真实 MySQL。
func newUserMockEntClient(t *testing.T) (*ent.Client, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	drv := entsql.OpenDB(dialect.MySQL, db)
	client := ent.NewClient(ent.Driver(drv))
	t.Cleanup(func() { _ = db.Close() })
	return client, mock
}

// membersQuery 匹配命名空间成员源查询（EnsureSynced 的第一条 SQL）。
var membersQuery = "SELECT .* FROM `members`"

// usersQuery 匹配用户全量查询（EnsureSynced 的第二条 SQL）。
var usersQuery = "SELECT .* FROM `users`"

// Test_userRepo_EnsureSynced_ExistingUsersQueryError 已有用户查询失败（非 NotFound）时
// 返回错误：DB 抖动不得被吞成空集合继续同步。
func Test_userRepo_EnsureSynced_ExistingUsersQueryError(t *testing.T) {
	client, mock := newUserMockEntClient(t)
	repo := NewUserRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}), timer.NewReal()).(*userRepo)
	ctx := context.TODO()

	mock.ExpectQuery(membersQuery).WillReturnRows(sqlmock.NewRows([]string{"email"}))
	mock.ExpectQuery(usersQuery).WillReturnError(errors.New("db boom"))

	err := repo.EnsureSynced(ctx)
	assert.ErrorContains(t, err, "list existing users for sync")
}

// Test_userRepo_EnsureSynced_UpdateSaveError 已有用户需补名/补登录时间时 UPDATE 失败
// 即返回错误：同步不得静默丢写。
func Test_userRepo_EnsureSynced_UpdateSaveError(t *testing.T) {
	client, mock := newUserMockEntClient(t)
	repo := NewUserRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}), timer.NewReal()).(*userRepo)
	ctx := context.TODO()

	mock.ExpectQuery(membersQuery).WillReturnRows(sqlmock.NewRows([]string{"email"}))
	// 已存在内置管理员但展示名为空 → 触发补名 UPDATE。
	now := time.Now()
	mock.ExpectQuery(usersQuery).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "name", "roles", "last_login"}).
		AddRow(1, now, now, biz.SuperAdminEmail, "", `["mars_admin"]`, nil))
	mock.ExpectExec("UPDATE `users`").WillReturnError(errors.New("db boom"))

	err := repo.EnsureSynced(ctx)
	assert.ErrorContains(t, err, "update user projection")
}

// Test_userRepo_EnsureSynced_MembersQueryError 成员源查询失败即返回错误：
// 同步不得把身份源缺失当作无成员静默继续。
func Test_userRepo_EnsureSynced_MembersQueryError(t *testing.T) {
	client, mock := newUserMockEntClient(t)
	repo := NewUserRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}), timer.NewReal()).(*userRepo)
	ctx := context.TODO()

	mock.ExpectQuery(membersQuery).WillReturnError(errors.New("db boom"))

	err := repo.EnsureSynced(ctx)
	assert.ErrorContains(t, err, "query members for user sync")
}

// Test_userRepo_EnsureSynced_CreateSaveError 新用户落库 INSERT 失败即返回错误。
func Test_userRepo_EnsureSynced_CreateSaveError(t *testing.T) {
	client, mock := newUserMockEntClient(t)
	repo := NewUserRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}), timer.NewReal()).(*userRepo)
	ctx := context.TODO()

	mock.ExpectQuery(membersQuery).WillReturnRows(sqlmock.NewRows([]string{"email"}))
	mock.ExpectQuery(usersQuery).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "name", "roles", "last_login"}))
	mock.ExpectExec("INSERT INTO `users`").WillReturnError(errors.New("db boom"))

	err := repo.EnsureSynced(ctx)
	assert.ErrorContains(t, err, "create user projection")
}

// Test_userRepo_SyncLoginUser_QueryError 邮箱查询失败（非 NotFound，如 DB 断开）即返回
// 错误：不得误走创建分支吞掉底层故障。
func Test_userRepo_SyncLoginUser_QueryError(t *testing.T) {
	client, mock := newUserMockEntClient(t)
	repo := NewUserRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}), timer.NewReal()).(*userRepo)
	ctx := context.TODO()

	mock.ExpectQuery(usersQuery).WillReturnError(errors.New("db boom"))

	err := repo.SyncLoginUser(ctx, "alice@x.com", "alice")
	assert.ErrorContains(t, err, "query user on login")
}

// Test_userRepo_SyncLoginUser_CreateSaveError 邮箱不存在（查询空行）走创建，INSERT 失败
// 即返回错误。
func Test_userRepo_SyncLoginUser_CreateSaveError(t *testing.T) {
	client, mock := newUserMockEntClient(t)
	repo := NewUserRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}), timer.NewReal()).(*userRepo)
	ctx := context.TODO()

	mock.ExpectQuery(usersQuery).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "name", "roles", "last_login"}))
	mock.ExpectExec("INSERT INTO `users`").WillReturnError(errors.New("db boom"))

	err := repo.SyncLoginUser(ctx, "alice@x.com", "alice")
	assert.ErrorContains(t, err, "create user projection on login")
}

// Test_userRepo_SyncLoginUser_UpdateSaveError 已存在用户需补名/补登录时间时 UPDATE 失败
// 即返回错误。
func Test_userRepo_SyncLoginUser_UpdateSaveError(t *testing.T) {
	client, mock := newUserMockEntClient(t)
	repo := NewUserRepo(NewDataImpl(&NewDataParams{Cfg: &config.Config{}, DB: client}), timer.NewReal()).(*userRepo)
	ctx := context.TODO()

	now := time.Now()
	mock.ExpectQuery(usersQuery).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "name", "roles", "last_login"}).
		AddRow(1, now, now, "alice@x.com", "", `[]`, nil))
	mock.ExpectExec("UPDATE `users`").WillReturnError(errors.New("db boom"))

	err := repo.SyncLoginUser(ctx, "alice@x.com", "alice")
	assert.ErrorContains(t, err, "update user projection on login")
}
