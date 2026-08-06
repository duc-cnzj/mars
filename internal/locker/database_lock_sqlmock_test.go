package locker

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/DATA-DOG/go-sqlmock"

	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/stretchr/testify/assert"
)

// newMockEntClient 构造一个由 go-sqlmock 驱动的 ent client，
// 用于覆盖 databaseLock 中 DB 操作失败的分支，无需真实 MySQL。
func newMockEntClient(t *testing.T) (*ent.Client, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	drv := entsql.OpenDB(dialect.MySQL, db)
	client := ent.NewClient(ent.Driver(drv))
	t.Cleanup(func() { _ = db.Close() })
	return client, mock
}

// newMockDatabaseLock 构造一个后端为 go-sqlmock 的 databaseLock，
// 使全部 DB 路径（成功/失败）都不依赖真实 MySQL，保证覆盖可复现。
// lottery 默认 {0,1}：永不触发过期清理，保证测试确定性。
func newMockDatabaseLock(t *testing.T, lottery ...[2]int) (*databaseLock, *ent.Client, sqlmock.Sqlmock) {
	t.Helper()
	client, mock := newMockEntClient(t)
	l := [2]int{0, 1}
	if len(lottery) > 0 {
		l = lottery[0]
	}
	lock := NewDatabaseLock(timer.NewReal(), l, func() *ent.Client { return client }, mlog.NewForConfig(nil)).(*databaseLock)
	return lock, client, mock
}

// lockRows 构造一条 owner 为 lock.ID() 的 cache_locks 查询结果行。
func lockRows(lock *databaseLock) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "key", "owner", "expired_at"}).
		AddRow(1, "key", lock.ID(), time.Now())
}

func Test_databaseLock_ID(t *testing.T) {
	lock, _, _ := newMockDatabaseLock(t)
	assert.Len(t, lock.ID(), 40)
}

func Test_databaseLock_Type_String(t *testing.T) {
	lock, _, _ := newMockDatabaseLock(t)
	assert.Equal(t, "db", lock.Type())
}

func Test_databaseLock_createLock_Success(t *testing.T) {
	lock, client, mock := newMockDatabaseLock(t)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `cache_locks`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	assert.True(t, lock.createLock(client, "key", time.Now()))
}

func Test_databaseLock_createLock_Error(t *testing.T) {
	lock, client, mock := newMockDatabaseLock(t)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `cache_locks`")).
		WillReturnError(errors.New("duplicate key"))
	assert.False(t, lock.createLock(client, "key", time.Now()))
}

func Test_databaseLock_updateExpiredLock_Success(t *testing.T) {
	lock, client, mock := newMockDatabaseLock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `cache_locks`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	assert.True(t, lock.updateExpiredLock(client, "key", time.Now()))
}

func Test_databaseLock_updateExpiredLock_Error(t *testing.T) {
	lock, client, mock := newMockDatabaseLock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `cache_locks`")).
		WillReturnError(errors.New("db error"))
	assert.False(t, lock.updateExpiredLock(client, "key", time.Now()))
}

func Test_databaseLock_updateExpiredLock_NoAffectedRows(t *testing.T) {
	lock, client, mock := newMockDatabaseLock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `cache_locks`")).
		WillReturnResult(sqlmock.NewResult(1, 0))
	assert.False(t, lock.updateExpiredLock(client, "key", time.Now()))
}

func Test_databaseLock_cleanupExpiredLocks_Success(t *testing.T) {
	lock, client, mock := newMockDatabaseLock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `cache_locks`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	assert.NotPanics(t, func() { lock.cleanupExpiredLocks(client) })
}

func Test_databaseLock_cleanupExpiredLocks_Error(t *testing.T) {
	lock, client, mock := newMockDatabaseLock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `cache_locks`")).
		WillReturnError(errors.New("db error"))
	assert.NotPanics(t, func() { lock.cleanupExpiredLocks(client) })
}

func Test_databaseLock_Acquire_Create(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `cache_locks`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	assert.True(t, lock.Acquire("key", 60))
}

func Test_databaseLock_Acquire_Takeover(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `cache_locks`")).
		WillReturnError(errors.New("duplicate key"))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `cache_locks`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	assert.True(t, lock.Acquire("key", 60))
}

func Test_databaseLock_Acquire_Fail(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `cache_locks`")).
		WillReturnError(errors.New("duplicate key"))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `cache_locks`")).
		WillReturnResult(sqlmock.NewResult(1, 0))
	assert.False(t, lock.Acquire("key", 60))
}

// Test_databaseLock_Acquire_Cleanup 用恒触发的 lottery 覆盖 Acquire 内过期锁清理路径。
func Test_databaseLock_Acquire_Cleanup(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t, [2]int{100, 1})
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `cache_locks`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `cache_locks`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	assert.True(t, lock.Acquire("key", 60))
}

func Test_databaseLock_Owner_Success(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `cache_locks`")).
		WillReturnRows(lockRows(lock))
	assert.Equal(t, lock.ID(), lock.Owner("key"))
}

func Test_databaseLock_Owner_Empty(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `cache_locks`")).
		WillReturnError(errors.New("not found"))
	assert.Empty(t, lock.Owner("key"))
}

func Test_databaseLock_Release_Success(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `cache_locks`")).
		WillReturnRows(lockRows(lock))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `cache_locks`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	assert.True(t, lock.Release("key"))
}

func Test_databaseLock_Release_NotOwner(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	other := NewDatabaseLock(timer.NewReal(), [2]int{0, 1}, nil, mlog.NewForConfig(nil))
	rows := sqlmock.NewRows([]string{"id", "key", "owner", "expired_at"}).
		AddRow(1, "key", other.ID(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `cache_locks`")).
		WillReturnRows(rows)
	assert.False(t, lock.Release("key"))
}

func Test_databaseLock_ForceRelease(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `cache_locks`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	assert.True(t, lock.ForceRelease("key"))
}

// Test_databaseLock_ForceRelease_DeleteError 覆盖 ForceRelease 中 DELETE 失败时返回 false 的分支。
func Test_databaseLock_ForceRelease_DeleteError(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `cache_locks`")).
		WillReturnError(errors.New("delete boom"))
	assert.False(t, lock.ForceRelease("key"))
}

func Test_databaseLock_RenewalAcquire_Success(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `cache_locks`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	releaseFn, ok := lock.RenewalAcquire("key", 60, 100)
	assert.True(t, ok)
	assert.NotNil(t, releaseFn)

	// release 会执行 Owner 查询 + DELETE。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `cache_locks`")).
		WillReturnRows(lockRows(lock))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `cache_locks`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	releaseFn()
}

func Test_databaseLock_RenewalAcquire_Fail(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `cache_locks`")).
		WillReturnError(errors.New("duplicate key"))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `cache_locks`")).
		WillReturnResult(sqlmock.NewResult(1, 0))
	fn, ok := lock.RenewalAcquire("key", 60, 100)
	assert.False(t, ok)
	assert.Nil(t, fn)
}

// Test_databaseLock_renewalExistKey_Success 用真实事务（Begin/fn/Commit）覆盖续期成功路径。
func Test_databaseLock_renewalExistKey_Success(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `cache_locks`")).
		WillReturnRows(lockRows(lock))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `cache_locks`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// ent 的 UpdateOne 在更新后会按 id 再加载一次该行。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id`, `key`, `owner`, `expired_at` FROM `cache_locks` WHERE `id`")).
		WillReturnRows(lockRows(lock))
	mock.ExpectCommit()
	assert.NoError(t, lock.renewalExistKey("key", 60))
}

// Test_databaseLock_renewalExistKey_QueryError 覆盖锁不存在时 Only 返回错误的 not-found 分支。
func Test_databaseLock_renewalExistKey_QueryError(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `cache_locks`")).
		WillReturnError(errors.New("not found"))
	mock.ExpectRollback()
	assert.Error(t, lock.renewalExistKey("key", 60))
}

// Test_databaseLock_renewalExistKey_NotOwner 覆盖锁已被他人持有时的 "not owner" 错误分支。
func Test_databaseLock_renewalExistKey_NotOwner(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	other := NewDatabaseLock(timer.NewReal(), [2]int{0, 1}, nil, mlog.NewForConfig(nil))
	rows := sqlmock.NewRows([]string{"id", "key", "owner", "expired_at"}).
		AddRow(1, "key", other.ID(), time.Now())
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `cache_locks`")).
		WillReturnRows(rows)
	mock.ExpectRollback()
	assert.ErrorContains(t, lock.renewalExistKey("key", 60), "not owner")
}

// Test_databaseLock_renewalRoutine_RenewError 覆盖续期出错（DB 未就绪，getDB 返回 nil）时
// 打印日志并退出 goroutine 的分支。
func Test_databaseLock_renewalRoutine_RenewError(t *testing.T) {
	lock := NewDatabaseLock(timer.NewReal(), [2]int{1, 2}, func() *ent.Client { return nil }, mlog.NewForConfig(nil)).(*databaseLock)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		lock.renewalRoutine(ctx, "key", 10, 1)
	}()

	select {
	case <-done:
		// renewalSeconds=1，ticker 触发一次后因续期失败退出。
	case <-time.After(3 * time.Second):
		t.Fatal("renewalRoutine 未在续期失败后退出")
	}
}

// Test_databaseLock_renewalRoutine_Cancel 覆盖 ctx 取消后 goroutine 退出且不再续期的分支。
// renewalSeconds=60 使 ticker 在测试窗口内不触发，纯靠 ctx 取消驱动退出，确定性无 flake。
func Test_databaseLock_renewalRoutine_Cancel(t *testing.T) {
	lock, _, _ := newMockDatabaseLock(t)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		lock.renewalRoutine(ctx, "key", 10, 60)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("renewalRoutine 未在 ctx 取消后退出")
	}
}

// Test_databaseLock_NilDB 覆盖 DB 未初始化（getDB 返回 nil）时所有 DB 方法
// 优雅降级而不 panic 的分支：Acquire=false / Owner="" / Release=false /
// ForceRelease=false / renewalExistKey=error。这是 wire 构造期早于 InitDB
// 场景下的安全兜底。
func Test_databaseLock_NilDB(t *testing.T) {
	lock := NewDatabaseLock(timer.NewReal(), [2]int{1, 2}, func() *ent.Client { return nil }, mlog.NewForConfig(nil)).(*databaseLock)

	assert.False(t, lock.Acquire("key", 60))
	assert.Empty(t, lock.Owner("key"))
	assert.False(t, lock.Release("key"))
	assert.False(t, lock.ForceRelease("key"))
	assert.ErrorContains(t, lock.renewalExistKey("key", 60), "db not initialized")
}

// Test_databaseLock_Release_QueryError 覆盖 Release 中持有者查询失败的错误分支。
func Test_databaseLock_Release_QueryError(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `cache_locks`")).
		WillReturnError(errors.New("not found"))
	assert.False(t, lock.Release("key"))
}

// Test_databaseLock_Release_DeleteError 覆盖 Release 中 DELETE 失败时返回 false 的分支。
func Test_databaseLock_Release_DeleteError(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `cache_locks`")).
		WillReturnRows(lockRows(lock))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `cache_locks`")).
		WillReturnError(errors.New("delete boom"))
	assert.False(t, lock.Release("key"))
}

// Test_databaseLock_renewalExistKey_BeginError 覆盖 db.Tx 启动事务失败的错误分支。
func Test_databaseLock_renewalExistKey_BeginError(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	mock.ExpectBegin().WillReturnError(errors.New("begin error"))
	assert.Error(t, lock.renewalExistKey("key", 60))
}

// Test_databaseLock_renewalExistKey_UpdateError 覆盖续期 UPDATE 失败的错误分支。
func Test_databaseLock_renewalExistKey_UpdateError(t *testing.T) {
	lock, _, mock := newMockDatabaseLock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `cache_locks`")).
		WillReturnRows(lockRows(lock))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `cache_locks`")).
		WillReturnError(errors.New("update boom"))
	mock.ExpectRollback()
	assert.Error(t, lock.renewalExistKey("key", 60))
}
