package data

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"

	_ "github.com/duc-cnzj/mars/v6/internal/data/ent/runtime"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

// slowLogDriver 包装 ent dialect.Driver：当 Exec/Query 执行耗时超过阈值时记录一条慢查询日志。
type slowLogDriver struct {
	dialect.Driver
	slowThreshold time.Duration
	logger        mlog.Logger
	timer         timer.Timer
}

// logSlow 执行耗时超过阈值时输出慢查询日志（含查询、参数与耗时），否则静默。
func (d *slowLogDriver) logSlow(query string, args any, elapsed time.Duration) {
	if elapsed > d.slowThreshold {
		d.logger.Infof("slow query: %s, args: %v, took: %s", query, args, elapsed)
	}
}

// Exec 执行写操作并透传错误；超阈值时经 logSlow 记录慢查询日志。
func (d *slowLogDriver) Exec(ctx context.Context, query string, args, v any) error {
	start := d.timer.Now()
	err := d.Driver.Exec(ctx, query, args, v)
	d.logSlow(query, args, d.timer.Since(start))
	return err
}

// Query 执行读操作并透传错误；超阈值时经 logSlow 记录慢查询日志。
func (d *slowLogDriver) Query(ctx context.Context, query string, args, v any) error {
	start := d.timer.Now()
	err := d.Driver.Query(ctx, query, args, v)
	d.logSlow(query, args, d.timer.Since(start))
	return err
}

// OpenDB 按配置的驱动名打开数据库连接（sqlite 用共享缓存文件、mysql 用 DSN 惰性连接）。
// 不支持的驱动是显式配置校验失败，用语义构造器映射为 InvalidArgument(400)，
// 上层 errs.Wrap 会保留该状态码，避免客户端把"驱动不支持"误判成服务器内部错误。
func OpenDB(config *config.Config) (*sql.Driver, error) {
	switch config.DBDriver {
	case "sqlite":
		return sql.Open("sqlite3", fmt.Sprintf("file:%v?cache=shared&_fk=1", config.DBDatabase))
	case "mysql":
		return sql.Open("mysql", config.DSN())
	}
	return nil, errs.WrapInvalidArgument(fmt.Errorf("unsupported database driver %v", config.DBDriver), "open database")
}

// InitDB 基于已打开的驱动构造 ent client，统一配置连接池参数与可选慢查询日志包装。
// drv 恒为 OpenDB 返回的 *sql.Driver（内部契约）；签名不带 error——drv 断言与
// ent.NewClient 均无错误路径，去掉误导性的恒 nil 返回（零死代码）。
func InitDB(drv dialect.Driver, logger mlog.Logger, slogLogEnabled bool, slowLogThreshold time.Duration, timer timer.Timer) *ent.Client {
	db := drv.(*sql.Driver).DB()
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)
	if slogLogEnabled {
		drv = &slowLogDriver{
			timer:         timer,
			Driver:        drv,
			slowThreshold: slowLogThreshold,
			logger:        logger.WithModule("SlowLog"),
		}
	}
	return ent.NewClient(
		ent.Driver(drv),
		ent.Log(func(a ...any) {
			logger.Debug(a...)
		}),
	)
}
