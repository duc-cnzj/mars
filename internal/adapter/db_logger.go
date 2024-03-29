package adapter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/mlog"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const defaultSlowThreshold = 200 * time.Millisecond

type gormLoggerAdapter struct {
	level logger.LogLevel

	slowLogEnabled   bool
	slowLogThreshold time.Duration
}

type GormLoggerOption func(*gormLoggerAdapter)

// GormLoggerWithSlowLog slow log switch.
func GormLoggerWithSlowLog(enabled bool, slowThreshold time.Duration) func(*gormLoggerAdapter) {
	return func(l *gormLoggerAdapter) {
		l.slowLogEnabled = enabled
		l.slowLogThreshold = slowThreshold
	}
}

// NewGormLoggerAdapter return logger.Interface.
func NewGormLoggerAdapter(opts ...GormLoggerOption) logger.Interface {
	l := &gormLoggerAdapter{level: logger.Warn, slowLogEnabled: false, slowLogThreshold: defaultSlowThreshold}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// LogMode log mode
func (g *gormLoggerAdapter) LogMode(level logger.LogLevel) logger.Interface {
	g.level = level

	return g
}

// Info print info
func (g *gormLoggerAdapter) Info(ctx context.Context, s string, i ...any) {
	if g.level >= logger.Info {
		mlog.Infof(s, i...)
	}
}

// Warn print warn messages
func (g *gormLoggerAdapter) Warn(ctx context.Context, s string, i ...any) {
	if g.level >= logger.Warn {
		mlog.Warningf(s, i...)
	}
}

// Error print error messages
func (g *gormLoggerAdapter) Error(ctx context.Context, s string, i ...any) {
	if g.level >= logger.Error {
		mlog.Errorf(s, i...)
	}
}

// Trace print sql message
func (g *gormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	const (
		traceStr     = "[SQL]: [%.3fms] [rows:%v] %s %s"
		traceWarnStr = "[SQL]: %s [%.3fms] [rows:%v] %s %s"
		traceErrStr  = "[SQL]: %s [%.3fms] [rows:%v] %s %s"
	)

	if g.level > logger.Silent {
		elapsed := time.Since(begin)
		switch {
		case err != nil && g.level >= logger.Error:
			if errors.Is(err, gorm.ErrRecordNotFound) {
				sql, rows := fc()
				if rows == -1 {
					mlog.Debugf(traceErrStr, err, float64(elapsed.Nanoseconds())/1e6, "-", sql, utils.FileWithLineNum())
				} else {
					mlog.Debugf(traceErrStr, err, float64(elapsed.Nanoseconds())/1e6, rows, sql, utils.FileWithLineNum())
				}
				return
			}
			sql, rows := fc()
			if rows == -1 {
				mlog.Errorf(traceErrStr, err, float64(elapsed.Nanoseconds())/1e6, "-", sql, utils.FileWithLineNum())
			} else {
				if strings.Contains(err.Error(), "for key 'cache_locks.PRIMARY'") {
					mlog.Debugf(traceErrStr, err, float64(elapsed.Nanoseconds())/1e6, rows, sql, utils.FileWithLineNum())
					return
				}
				mlog.Errorf(traceErrStr, err, float64(elapsed.Nanoseconds())/1e6, rows, sql, utils.FileWithLineNum())
			}
		case elapsed > g.slowLogThreshold && g.slowLogEnabled:
			sql, rows := fc()
			slowLog := fmt.Sprintf("(SLOW SQL) >= %v", g.slowLogThreshold)
			if rows == -1 {
				mlog.Warningf(traceWarnStr, slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql, utils.FileWithLineNum())
			} else {
				mlog.Warningf(traceWarnStr, slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql, utils.FileWithLineNum())
			}
		case g.level == logger.Info:
			sql, rows := fc()
			if rows == -1 {
				mlog.Infof(traceStr, float64(elapsed.Nanoseconds())/1e6, "-", sql, utils.FileWithLineNum())
			} else {
				mlog.Infof(traceStr, float64(elapsed.Nanoseconds())/1e6, rows, sql, utils.FileWithLineNum())
			}
		}
	}
}
