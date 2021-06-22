package adapter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/pkg/mlog"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type GormLoggerAdapter struct {
	level logger.LogLevel
}

func (g *GormLoggerAdapter) LogMode(level logger.LogLevel) logger.Interface {
	g.level = level

	return g
}

func (g *GormLoggerAdapter) Info(ctx context.Context, s string, i ...interface{}) {
	if g.level >= logger.Info {
		mlog.Infof(s, i...)
	}
}

func (g *GormLoggerAdapter) Warn(ctx context.Context, s string, i ...interface{}) {
	if g.level >= logger.Warn {
		mlog.Warningf(s, i...)
	}
}

func (g *GormLoggerAdapter) Error(ctx context.Context, s string, i ...interface{}) {
	if g.level >= logger.Error {
		mlog.Errorf(s, i...)
	}
}

func (g *GormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	const (
		traceStr      = "%s [%.3fms] [rows:%v] %s"
		traceWarnStr  = "%s %s [%.3fms] [rows:%v] %s"
		traceErrStr   = "%s %s [%.3fms] [rows:%v] %s"
		slowThreshold = 200 * time.Millisecond
	)
	if g.level > logger.Silent {
		elapsed := time.Since(begin)
		switch {
		case err != nil && g.level >= logger.Error:
			if errors.Is(err, gorm.ErrRecordNotFound) {
				sql, rows := fc()
				if rows == -1 {
					mlog.Debugf(traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
				} else {
					mlog.Debugf(traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
				}
				return
			}
			sql, rows := fc()
			if rows == -1 {
				mlog.Errorf(traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				mlog.Errorf(traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		case elapsed > slowThreshold && g.level >= logger.Warn:
			sql, rows := fc()
			slowLog := fmt.Sprintf("SLOW SQL >= %v", slowThreshold)
			if rows == -1 {
				mlog.Warningf(traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				mlog.Warningf(traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		case g.level == logger.Info:
			sql, rows := fc()
			if rows == -1 {
				mlog.Infof(traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				mlog.Infof(traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		}
	}
}
