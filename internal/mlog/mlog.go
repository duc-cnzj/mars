package mlog

//go:generate mockgen -destination ./mock_logger.go -package mlog github.com/duc-cnzj/mars/v5/internal/mlog Logger
import (
	"context"

	"github.com/duc-cnzj/mars/v5/internal/config"
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

func NewForConfig(cfg *config.Config) Logger {
	var (
		channel string
		debug   bool
	)
	if cfg != nil {
		channel = cfg.LogChannel
		debug = cfg.Debug
	}
	switch channel {
	case "zap":
		return NewZapLogger(debug)
	case "logrus":
		fallthrough
	default:
		return NewLogrusLogger(debug)
	}
}
