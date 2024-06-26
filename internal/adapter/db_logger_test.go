package adapter

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/mock"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGormLoggerAdapter(t *testing.T) {
	assert.Implements(t, (*logger.Interface)(nil), &gormLoggerAdapter{})
}

func TestGormLoggerAdapter_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Errorf("aaa").Times(1)
	NewGormLoggerAdapter(GormLoggerWithSlowLog(true, defaultSlowThreshold)).LogMode(logger.Info).Error(context.TODO(), "aaa")
}

func TestGormLoggerAdapter_Info(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Infof("aaa").Times(1)
	NewGormLoggerAdapter(GormLoggerWithSlowLog(true, defaultSlowThreshold)).LogMode(logger.Info).Info(context.TODO(), "aaa")
}

func TestGormLoggerAdapter_LogMode(t *testing.T) {
	ada := &gormLoggerAdapter{level: logger.Info}
	ada.LogMode(logger.Error)
	assert.Equal(t, logger.Error, ada.level)
}

type loggerMock struct {
	errs   []string
	infos  []string
	debugs []string
	warns  []string
	contracts.LoggerInterface
}

func (l *loggerMock) Debugf(format string, v ...any) {
	l.debugs = append(l.debugs, fmt.Sprintf(format, v...))
}

func (l *loggerMock) Infof(format string, v ...any) {
	l.infos = append(l.infos, fmt.Sprintf(format, v...))
}

func (l *loggerMock) Warningf(format string, v ...any) {
	l.warns = append(l.warns, fmt.Sprintf(format, v...))
}

func (l *loggerMock) Errorf(format string, v ...any) {
	l.errs = append(l.errs, fmt.Sprintf(format, v...))
}

func TestGormLoggerAdapter_Trace(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := &loggerMock{}
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())

	time1 := time.Now()
	NewGormLoggerAdapter(GormLoggerWithSlowLog(true, defaultSlowThreshold)).LogMode(logger.Info).Trace(context.TODO(), time1, func() (string, int64) {
		return "aaa", 100
	}, gorm.ErrRecordNotFound)
	assert.Regexp(t, `\[SQL\]: record not found \[(.*?)ms\] \[rows:100\] aaa \S+$`, l.debugs[0])

	NewGormLoggerAdapter(GormLoggerWithSlowLog(true, defaultSlowThreshold)).LogMode(logger.Info).Trace(context.TODO(), time1, func() (string, int64) {
		return "aaa", -1
	}, gorm.ErrRecordNotFound)
	assert.Regexp(t, `\[SQL\]: record not found \[(.*?)ms\] \[rows:-\] aaa \S+$`, l.debugs[1])

	NewGormLoggerAdapter(GormLoggerWithSlowLog(true, defaultSlowThreshold)).LogMode(logger.Info).Trace(context.TODO(), time1, func() (string, int64) {
		return "aaa", -1
	}, errors.New("xxx"))
	assert.Regexp(t, `\[SQL\]: xxx \[(.*?)ms\] \[rows:-\] aaa \S+$`, l.errs[0])

	NewGormLoggerAdapter(GormLoggerWithSlowLog(true, defaultSlowThreshold)).LogMode(logger.Info).Trace(context.TODO(), time1, func() (string, int64) {
		return "aaa", 100
	}, errors.New("xxx"))
	assert.Regexp(t, `\[SQL\]: xxx \[(.*?)ms\] \[rows:100\] aaa \S+$`, l.errs[1])

	NewGormLoggerAdapter(GormLoggerWithSlowLog(true, defaultSlowThreshold)).LogMode(logger.Info).Trace(context.TODO(), time1, func() (string, int64) {
		return "aaa", 100
	}, errors.New("xxx for key 'cache_locks.PRIMARY'"))
	assert.Regexp(t, `\[SQL\]: xxx for key 'cache_locks.PRIMARY' \[(.*?)ms\] \[rows:100\] aaa \S+$`, l.debugs[2])

	time2 := time.Now().Add(-1 * time.Second)
	NewGormLoggerAdapter(GormLoggerWithSlowLog(true, defaultSlowThreshold)).LogMode(logger.Error).Trace(context.TODO(), time2, func() (string, int64) {
		return "show sql...", 100
	}, nil)
	assert.Regexp(t, `\[SQL\]: \(SLOW SQL\) >= 200ms \[(.*?)ms\] \[rows:100\] show sql\.\.\. \S+$`, l.warns[0])

	NewGormLoggerAdapter(GormLoggerWithSlowLog(true, defaultSlowThreshold)).LogMode(logger.Info).Trace(context.TODO(), time2, func() (string, int64) {
		return "show sql...", -1
	}, nil)
	assert.Regexp(t, `\[SQL\]: \(SLOW SQL\) >= 200ms \[(.*?)ms\] \[rows:-\] show sql\.\.\. \S+$`, l.warns[1])

	NewGormLoggerAdapter(GormLoggerWithSlowLog(true, defaultSlowThreshold)).LogMode(logger.Info).Trace(context.TODO(), time1, func() (string, int64) {
		return "sql...", -1
	}, nil)
	assert.Regexp(t, `\[SQL\]: \[(.*?)ms\] \[rows:-\] sql\.\.\. \S+$`, l.infos[0])

	NewGormLoggerAdapter(GormLoggerWithSlowLog(true, defaultSlowThreshold)).LogMode(logger.Info).Trace(context.TODO(), time1, func() (string, int64) {
		return "sql...", 100
	}, nil)
	assert.Regexp(t, `\[SQL\]: \[(.*?)ms\] \[rows:100\] sql\.\.\. \S+$`, l.infos[1])
}

func TestGormLoggerAdapter_Warn(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Warningf("aaa").Times(1)
	NewGormLoggerAdapter(GormLoggerWithSlowLog(true, defaultSlowThreshold)).LogMode(logger.Info).Warn(context.TODO(), "aaa")
}

func TestNewGormLoggerAdapter(t *testing.T) {
	l := NewGormLoggerAdapter(GormLoggerWithSlowLog(true, 1*time.Millisecond)).(*gormLoggerAdapter)
	assert.Equal(t, 1*time.Millisecond, l.slowLogThreshold)
	assert.True(t, l.slowLogEnabled)
	assert.Equal(t, logger.Warn, l.level)
	l2 := NewGormLoggerAdapter().(*gormLoggerAdapter)
	assert.False(t, l2.slowLogEnabled)
	assert.Equal(t, defaultSlowThreshold, l2.slowLogThreshold)
	assert.Equal(t, logger.Warn, l2.level)
}
