package mlog

import (
	"errors"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ Logger = (*zapLogger)(nil)

type zapLogger struct {
	sugar *zap.SugaredLogger
	debug bool
}

// NewZapLogger impl contracts.Logger.
func NewZapLogger(debug bool) Logger {
	var (
		logger *zap.Logger
		cfg    zap.Config
	)

	opts := []zap.Option{zap.AddStacktrace(zapcore.ErrorLevel), zap.AddCallerSkip(2)}
	if debug {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
	}
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.CallerKey = "file"
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	logger, _ = cfg.Build(opts...)

	return &zapLogger{sugar: logger.Sugar(), debug: debug}
}

func (z *zapLogger) HandlePanic(title string) {
	bf := make([]byte, 1024*5)
	n := runtime.Stack(bf, false)
	bf = bf[:n]

	err := recover()
	if err != nil {
		z.Errorf("[Panic]: title: %v, err: %v --- [%s]", title, err, string(bf))
		if z.debug {
			panic(err)
		}
	}
}

func (z *zapLogger) HandlePanicWithCallback(title string, callback func(error)) {
	bf := make([]byte, 1024*5)
	n := runtime.Stack(bf, false)
	bf = bf[:n]

	err := recover()
	if err != nil {
		switch e := err.(type) {
		case error:
			callback(e)
		case string:
			callback(errors.New(e))
		}
		z.Errorf("[Panic]: title: %v, err: %v --- [%s]", title, err, string(bf))
		if z.debug {
			panic(err)
		}
	}
}

func (z *zapLogger) Flush() error {
	return z.sugar.Sync()
}

// Debug print debug msg
func (z *zapLogger) Debug(v ...any) {
	z.sugar.Debug(v...)
}

// Debugf printf debug msg
func (z *zapLogger) Debugf(format string, v ...any) {
	z.sugar.Debugf(format, v...)
}

// Warning print Warning msg
func (z *zapLogger) Warning(v ...any) {
	z.sugar.Warn(v...)
}

// Warningf prints Warning msg
func (z *zapLogger) Warningf(format string, v ...any) {
	z.sugar.Warnf(format, v...)
}

// Info print info msg
func (z *zapLogger) Info(v ...any) {
	z.sugar.Info(v...)
}

// Infof printf info msg
func (z *zapLogger) Infof(format string, v ...any) {
	z.sugar.Infof(format, v...)
}

// Error print err msg
func (z *zapLogger) Error(v ...any) {
	z.sugar.Error(v...)
}

// Errorf printf err msg
func (z *zapLogger) Errorf(format string, v ...any) {
	z.sugar.Errorf(format, v...)
}

// Fatal fatal err.
func (z *zapLogger) Fatal(v ...any) {
	z.sugar.Fatal(v...)
}

// Fatalf fatalf err.
func (z *zapLogger) Fatalf(format string, v ...any) {
	z.sugar.Fatalf(format, v...)
}
