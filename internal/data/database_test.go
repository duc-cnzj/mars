package data

import (
	"context"
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSlowTimer 恒返大间隔：让 slowLogDriver 的 elapsed > threshold 恒成立。
type fakeSlowTimer struct{}

func (fakeSlowTimer) Now() time.Time { return time.Time{} }
func (fakeSlowTimer) Since(time.Time) time.Duration {
	return time.Hour
}

// TestOpenDB 覆盖三种 driver 分支：sqlite 可用、mysql 惰性连接、未知 driver 报错。
func TestOpenDB(t *testing.T) {
	t.Run("sqlite", func(t *testing.T) {
		drv, err := OpenDB(&config.Config{DBDriver: "sqlite", DBDatabase: "file:opendb?mode=memory&cache=shared"})
		assert.NoError(t, err)
		assert.NotNil(t, drv)
	})

	t.Run("mysql returns driver without dialing", func(t *testing.T) {
		drv, err := OpenDB(&config.Config{DBDriver: "mysql", DBUsername: "u", DBPassword: "p", DBHost: "h", DBPort: "3306", DBDatabase: "d"})
		assert.NoError(t, err)
		assert.NotNil(t, drv)
	})

	t.Run("unsupported driver errors", func(t *testing.T) {
		_, err := OpenDB(&config.Config{DBDriver: "postgres"})
		assert.ErrorContains(t, err, "unsupported database driver postgres")
	})
}

// TestInitDB 覆盖 slowLog 开关两种分支，返回可用的 ent client。
func TestInitDB(t *testing.T) {
	for _, slogEnabled := range []bool{true, false} {
		drv, err := OpenDB(&config.Config{DBDriver: "sqlite", DBDatabase: filepath.Join(t.TempDir(), "initdb.db")})
		require.NoError(t, err)
		client := InitDB(drv, mlog.NewForConfig(nil), slogEnabled, time.Millisecond, fakeSlowTimer{})
		assert.NotNil(t, client)
		// Schema.Create 触发真实 SQL 查询，覆盖 ent.Log 闭包（logger.Debug(a...)）
		assert.NoError(t, client.Schema.Create(context.TODO()))
		client.Close()
	}
}

// TestSlowLogDriver 覆盖 Exec/Query 的慢查询日志分支与普通分支。
func TestSlowLogDriver(t *testing.T) {
	mk := func(threshold time.Duration) *slowLogDriver {
		drv, err := OpenDB(&config.Config{DBDriver: "sqlite", DBDatabase: filepath.Join(t.TempDir(), "slow.db")})
		require.NoError(t, err)
		return &slowLogDriver{
			Driver:        drv,
			slowThreshold: threshold,
			logger:        mlog.NewForConfig(nil),
			timer:         fakeSlowTimer{},
		}
	}

	t.Run("Exec logs slow query", func(t *testing.T) {
		d := mk(0)
		assert.NoError(t, d.Exec(context.TODO(), "CREATE TABLE slow_t (id int)", []any{}, nil))
	})

	t.Run("Exec within threshold no log", func(t *testing.T) {
		d := mk(2 * time.Hour)
		assert.NoError(t, d.Exec(context.TODO(), "CREATE TABLE fast_t (id int)", []any{}, nil))
	})

	t.Run("Query logs slow query", func(t *testing.T) {
		d := mk(0)
		assert.NoError(t, d.Exec(context.TODO(), "CREATE TABLE q_t (id int)", []any{}, nil))
		assert.NoError(t, d.Query(context.TODO(), "SELECT * FROM q_t", []any{}, &sql.Rows{}))
	})

	t.Run("Query within threshold no log", func(t *testing.T) {
		d := mk(2 * time.Hour)
		assert.NoError(t, d.Exec(context.TODO(), "CREATE TABLE qf_t (id int)", []any{}, nil))
		assert.NoError(t, d.Query(context.TODO(), "SELECT * FROM qf_t", []any{}, &sql.Rows{}))
	})

	t.Run("Exec propagates driver error", func(t *testing.T) {
		d := mk(0)
		assert.Error(t, d.Exec(context.TODO(), "INVALID SQL", []any{}, nil))
	})
}

// TestPackageClose 覆盖包级 Close() 关闭全局测试 DB（services 测试的辅助入口）。
func TestPackageClose(t *testing.T) {
	_, err := NewSqliteDB()
	assert.NoError(t, err)
	assert.NotPanics(t, Close)
}

// failDriver 是测试专用的故障注入 ent 驱动：包装底层 sqlite 驱动，
// 在指定次数的 Query/Exec 成功执行后注入错误，用于覆盖"第一段查询成功、
// 后续操作失败"这类 mid-operation 数据库错误分支。
//
//	qAfter / eAfter：允许成功执行的次数，-1 表示从不注入失败。
//	Query 与 Exec 各自独立计数；事务内的 Query/Exec 计入同一组计数器。
//	armed：迁移期（Schema.Create 发出的 FK pragma/建表语句）不计入计数，
//	武装后才开始计数，保证注入只针对被测 repo 方法的操作序列。
type failDriver struct {
	dialect.Driver
	qAfter int32
	eAfter int32
	qCount atomic.Int32
	eCount atomic.Int32
	armed  atomic.Bool
}

// Arm 武装故障注入：此后 Query/Exec 才开始计数与注入。
func (d *failDriver) Arm() {
	d.armed.Store(true)
}

func (d *failDriver) Query(ctx context.Context, query string, args, v any) error {
	if !d.armed.Load() {
		return d.Driver.Query(ctx, query, args, v)
	}
	n := d.qCount.Add(1)
	if d.qAfter >= 0 && n > d.qAfter {
		return errors.New("failDriver: injected query error")
	}
	return d.Driver.Query(ctx, query, args, v)
}

func (d *failDriver) Exec(ctx context.Context, query string, args, v any) error {
	if !d.armed.Load() {
		return d.Driver.Exec(ctx, query, args, v)
	}
	n := d.eCount.Add(1)
	if d.eAfter >= 0 && n > d.eAfter {
		return errors.New("failDriver: injected exec error")
	}
	return d.Driver.Exec(ctx, query, args, v)
}

func (d *failDriver) Tx(ctx context.Context) (dialect.Tx, error) {
	tx, err := d.Driver.Tx(ctx)
	if err != nil {
		return nil, err
	}
	return &failTx{Tx: tx, parent: d}, nil
}

type failTx struct {
	dialect.Tx
	parent *failDriver
}

func (t *failTx) Query(ctx context.Context, query string, args, v any) error {
	if !t.parent.armed.Load() {
		return t.Tx.Query(ctx, query, args, v)
	}
	n := t.parent.qCount.Add(1)
	if t.parent.qAfter >= 0 && n > t.parent.qAfter {
		return errors.New("failDriver: injected tx query error")
	}
	return t.Tx.Query(ctx, query, args, v)
}

func (t *failTx) Exec(ctx context.Context, query string, args, v any) error {
	if !t.parent.armed.Load() {
		return t.Tx.Exec(ctx, query, args, v)
	}
	n := t.parent.eCount.Add(1)
	if t.parent.eAfter >= 0 && n > t.parent.eAfter {
		return errors.New("failDriver: injected tx exec error")
	}
	return t.Tx.Exec(ctx, query, args, v)
}

// newFailDB 基于 sqlite + 故障注入驱动构造 ent client。
// 驱动保持未武装状态返回：Schema.Create 迁移期与测试 setup 阶段的
// 语句都直通不计入计数，调用方在 setup 完成后显式 fd.Arm()，
// 注入只针对被测 repo 方法的操作序列（计数从 0 开始，无需偏移）。
// 返回 client 与驱动句柄。
func newFailDB(t *testing.T, qAfter, eAfter int32) (*ent.Client, *failDriver) {
	t.Helper()
	sqlDrv, err := sql.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1&loc=Local")
	require.NoError(t, err)
	fd := &failDriver{Driver: sqlDrv, qAfter: qAfter, eAfter: eAfter}
	client := ent.NewClient(ent.Driver(fd))
	require.NoError(t, client.Schema.Create(context.TODO()))
	t.Cleanup(func() { client.Close() })
	return client, fd
}
