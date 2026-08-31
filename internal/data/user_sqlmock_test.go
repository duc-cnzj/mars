package data

import (
	"context"
	"errors"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/DATA-DOG/go-sqlmock"
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
