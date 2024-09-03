package mlog

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogrusLogger(t *testing.T) {
	logger := NewLogrusLogger(true)
	assert.NotNil(t, logger)
}

func TestLogrusLoggerWithModule(t *testing.T) {
	logger := NewLogrusLogger(true)
	loggerWithModule := logger.WithModule("testModule")
	assert.NotNil(t, loggerWithModule)
}

func TestLogrusLoggerDebug(t *testing.T) {
	logger := NewLogrusLogger(true)
	logger.Debug("debug message")
}

func TestLogrusLoggerDebugf(t *testing.T) {
	logger := NewLogrusLogger(true)
	logger.Debugf("debugf message %s", "test")
}

func TestLogrusLoggerDebugCtx(t *testing.T) {
	logger := NewLogrusLogger(true)
	ctx := context.TODO()
	logger.DebugCtx(ctx, "debug message with context")
}

func TestLogrusLoggerDebugCtxf(t *testing.T) {
	logger := NewLogrusLogger(true)
	ctx := context.TODO()
	logger.DebugCtxf(ctx, "debugf message with context %s", "test")
}

func TestLogrusLoggerInfo(t *testing.T) {
	logger := NewLogrusLogger(true)
	logger.Info("info message")
}

func TestLogrusLoggerInfof(t *testing.T) {
	logger := NewLogrusLogger(true)
	logger.Infof("infof message %s", "test")
}

func TestLogrusLoggerInfoCtx(t *testing.T) {
	logger := NewLogrusLogger(true)
	ctx := context.TODO()
	logger.InfoCtx(ctx, "info message with context")
}

func TestLogrusLoggerInfoCtxf(t *testing.T) {
	logger := NewLogrusLogger(true)
	ctx := context.TODO()
	logger.InfoCtxf(ctx, "infof message with context %s", "test")
}

func TestLogrusLoggerWarning(t *testing.T) {
	logger := NewLogrusLogger(true)
	logger.Warning("warning message")
}

func TestLogrusLoggerWarningf(t *testing.T) {
	logger := NewLogrusLogger(true)
	logger.Warningf("warningf message %s", "test")
}

func TestLogrusLoggerWarningCtx(t *testing.T) {
	logger := NewLogrusLogger(true)
	ctx := context.TODO()
	logger.WarningCtx(ctx, "warning message with context")
}

func TestLogrusLoggerWarningCtxf(t *testing.T) {
	logger := NewLogrusLogger(true)
	ctx := context.TODO()
	logger.WarningCtxf(ctx, "warningf message with context %s", "test")
}

func TestLogrusLoggerError(t *testing.T) {
	logger := NewLogrusLogger(true)
	logger.Error("error message")
}

func TestLogrusLoggerErrorf(t *testing.T) {
	logger := NewLogrusLogger(true)
	logger.Errorf("errorf message %s", "test")
}

func TestLogrusLoggerErrorCtx(t *testing.T) {
	logger := NewLogrusLogger(true)
	ctx := context.TODO()
	logger.ErrorCtx(ctx, "error message with context")
}

func TestLogrusLoggerErrorCtxf(t *testing.T) {
	logger := NewLogrusLogger(true)
	ctx := context.TODO()
	logger.ErrorCtxf(ctx, "errorf message with context %s", "test")
}

func Test_logrusLogger_HandlePanic(t *testing.T) {
	defer func() {
		a := recover()
		assert.Equal(t, "test panic", a)
	}()
	logger := NewLogrusLogger(true)
	defer logger.HandlePanic("test panic")
	panic("test panic")
}

func Test_logrusLogger_HandlePanicWithCallback(t *testing.T) {
	defer func() {
		a := recover()
		assert.Equal(t, "test panic", a)
	}()
	logger := NewLogrusLogger(true)
	defer logger.HandlePanicWithCallback("test panic", func(err error) {
		assert.Equal(t, "test panic", err.Error())
	})
	panic("test panic")
}

func Test_logrusLogger_HandlePanicWithCallback2(t *testing.T) {
	defer func() {
		a := recover()
		assert.Equal(t, "x", a.(error).Error())
	}()
	logger := NewLogrusLogger(true)
	defer logger.HandlePanicWithCallback("test panic", func(err error) {
		assert.Equal(t, "x", err.Error())
	})
	panic(errors.New("x"))
}

func Test_logrusLogger_Flush(t *testing.T) {
	logger := NewLogrusLogger(true)
	assert.Nil(t, logger.Flush())
}

func Test_logrusLogger_fields(t *testing.T) {
	logger := NewLogrusLogger(true).(*logrusLogger)
	logger.module = "x"
	assert.NotNil(t, logger.fields())
	assert.Len(t, logger.fields(), 2)
}
