package mlog

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZapLoggerDebugMode(t *testing.T) {
	logger := NewZapLogger(true)
	assert.True(t, logger.(*zapLogger).debug)
}

func TestZapLoggerProductionMode(t *testing.T) {
	logger := NewZapLogger(false)
	assert.False(t, logger.(*zapLogger).debug)
}

func TestZapLoggerDebugf(t *testing.T) {
	logger := NewZapLogger(true)
	logger.Debugf("debug message with format: %s", "test")
}

func TestZapLoggerWarningf(t *testing.T) {
	logger := NewZapLogger(true)
	logger.Warningf("warning message with format: %s", "test")
}

func TestZapLoggerInfof(t *testing.T) {
	logger := NewZapLogger(true)
	logger.Infof("info message with format: %s", "test")
}

func TestZapLoggerErrorf(t *testing.T) {
	logger := NewZapLogger(true)
	logger.Errorf("error message with format: %s", "test")
}

func TestZapLoggerHandlePanic(t *testing.T) {
	defer func() {
		a := recover()
		assert.Equal(t, "test panic", a)
	}()
	logger := NewZapLogger(true)
	defer logger.HandlePanic("test panic")
	panic("test panic")
}

func TestZapLoggerHandlePanicWithCallback(t *testing.T) {
	defer func() {
		a := recover()
		assert.Equal(t, "test panic", a)
	}()
	logger := NewZapLogger(true)
	defer logger.HandlePanicWithCallback("test panic", func(err error) {
		assert.Equal(t, errors.New("test panic"), err)
	})
	panic("test panic")
}

func TestZapLoggerHandlePanicWithCallback2(t *testing.T) {
	defer func() {
		a := recover()
		assert.Equal(t, "test panic", a.(error).Error())
	}()
	logger := NewZapLogger(true)
	defer logger.HandlePanicWithCallback("test panic", func(err error) {
		assert.Equal(t, errors.New("test panic"), err)
	})
	panic(errors.New("test panic"))
}

func TestZapLoggerFlush(t *testing.T) {
	logger := NewZapLogger(true)
	logger.Flush()
}

func TestZapLoggerDebug(t *testing.T) {
	logger := NewZapLogger(true)
	logger.Debug("debug message")
}

func TestZapLoggerDebugCtx(t *testing.T) {
	logger := NewZapLogger(true)
	ctx := context.TODO()
	logger.DebugCtx(ctx, "debug message with context")
}

func TestZapLoggerWarning(t *testing.T) {
	logger := NewZapLogger(true)
	logger.Warning("warning message")
}

func TestZapLoggerWarningCtx(t *testing.T) {
	logger := NewZapLogger(true)
	ctx := context.TODO()
	logger.WarningCtx(ctx, "warning message with context")
}

func TestZapLoggerInfo(t *testing.T) {
	logger := NewZapLogger(true)
	logger.Info("info message")
}

func TestZapLoggerInfoCtx(t *testing.T) {
	logger := NewZapLogger(true)
	ctx := context.TODO()
	logger.InfoCtx(ctx, "info message with context")
}

func TestZapLoggerError(t *testing.T) {
	logger := NewZapLogger(true)
	logger.Error("error message")
}

func TestZapLoggerErrorCtx(t *testing.T) {
	logger := NewZapLogger(true)
	ctx := context.TODO()
	logger.ErrorCtx(ctx, "error message with context")
}

func Test_zapLogger_DebugCtxf(t *testing.T) {
	logger := NewZapLogger(true)
	ctx := context.TODO()
	logger.DebugCtxf(ctx, "debugf message with context %s", "test")
}

func Test_zapLogger_WarningCtxf(t *testing.T) {
	logger := NewZapLogger(true)
	ctx := context.TODO()
	logger.WarningCtxf(ctx, "warningf message with context %s", "test")
}

func Test_zapLogger_InfoCtxf(t *testing.T) {
	logger := NewZapLogger(true)
	ctx := context.TODO()
	logger.InfoCtxf(ctx, "infof message with context %s", "test")
}

func Test_zapLogger_ErrorCtxf(t *testing.T) {
	logger := NewZapLogger(true)
	ctx := context.TODO()
	logger.ErrorCtxf(ctx, "errorf message with context %s", "test")
}

func Test_zapLogger_WithModule(t *testing.T) {
	logger := NewZapLogger(true)
	assert.NotNil(t, logger.WithModule("test"))
}
