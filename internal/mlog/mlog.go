package mlog

import (
	"context"
	"log"

	"github.com/duc-cnzj/mars/v4/internal/config"
)

type Logger interface {
	WithModule(module string) Logger

	Debug(v ...any)
	Debugf(format string, v ...any)

	DebugCtx(ctx context.Context, v ...any)
	DebugCtxf(ctx context.Context, format string, v ...any)

	Warning(v ...any)
	Warningf(format string, v ...any)

	WarningCtx(ctx context.Context, v ...any)
	WarningCtxf(ctx context.Context, format string, v ...any)

	Info(v ...any)
	Infof(format string, v ...any)

	InfoCtx(ctx context.Context, v ...any)
	InfoCtxf(ctx context.Context, format string, v ...any)

	Error(v ...any)
	Errorf(format string, v ...any)

	ErrorCtx(ctx context.Context, v ...any)
	ErrorCtxf(ctx context.Context, format string, v ...any)

	Fatal(v ...any)
	Fatalf(format string, v ...any)

	FatalCtx(ctx context.Context, v ...any)
	FatalCtxf(ctx context.Context, format string, v ...any)

	Flush() error

	HandlePanic(title string)
	HandlePanicWithCallback(title string, callback func(error))
}

func NewLogger(cfg *config.Config) Logger {
	var channel string
	if cfg != nil {
		channel = cfg.LogChannel
	}
	switch channel {
	case "logrus":
		return NewLogrusLogger(cfg.Debug)
	case "zap":
		return NewZapLogger(cfg.Debug)
	default:
		return NewEmptyLogger()
	}
}

var _ Logger = (*emptyLogger)(nil)

type emptyLogger struct{}

func (e *emptyLogger) WithModule(module string) Logger {
	return e
}

func (e *emptyLogger) FatalCtx(ctx context.Context, v ...any) {
}

func (e *emptyLogger) FatalCtxf(ctx context.Context, format string, v ...any) {
}

func (e *emptyLogger) DebugCtx(ctx context.Context, v ...any) {
}

func (e *emptyLogger) DebugCtxf(ctx context.Context, format string, v ...any) {
}

func (e *emptyLogger) WarningCtx(ctx context.Context, v ...any) {
}

func (e *emptyLogger) WarningCtxf(ctx context.Context, format string, v ...any) {
}

func (e *emptyLogger) InfoCtx(ctx context.Context, v ...any) {
}

func (e *emptyLogger) InfoCtxf(ctx context.Context, format string, v ...any) {
}

func (e *emptyLogger) ErrorCtx(ctx context.Context, v ...any) {
}

func (e *emptyLogger) ErrorCtxf(ctx context.Context, format string, v ...any) {
}

// NewEmptyLogger return contracts.LoggerInterface
func NewEmptyLogger() Logger {
	return &emptyLogger{}
}

// Debug print debug msg
func (e *emptyLogger) Debug(v ...any) {}

// Debugf printf debug msg
func (e *emptyLogger) Debugf(format string, v ...any) {}

// Warning print Warning msg
func (e *emptyLogger) Warning(v ...any) {}

// Warningf prints Warning msg
func (e *emptyLogger) Warningf(format string, v ...any) {}

// Info print info msg
func (e *emptyLogger) Info(v ...any) {}

// Infof printf info msg
func (e *emptyLogger) Infof(format string, v ...any) {}

// Error print err msg
func (e *emptyLogger) Error(v ...any) {}

// Errorf printf err msg
func (e *emptyLogger) Errorf(format string, v ...any) {}

// Fatal fatal err.
func (e *emptyLogger) Fatal(v ...any) {
	log.Fatal(v...)
}

// Fatalf fatalf err.
func (e *emptyLogger) Fatalf(format string, v ...any) {
	log.Fatalf(format, v...)
}

func (e *emptyLogger) Flush() error {
	return nil
}

func (e *emptyLogger) HandlePanic(title string) {
}

func (e *emptyLogger) HandlePanicWithCallback(title string, callback func(error)) {
}
