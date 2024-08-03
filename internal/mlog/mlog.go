package mlog

import (
	"log"

	"github.com/duc-cnzj/mars/v4/internal/config"
)

type Logger interface {
	Debug(v ...any)
	Debugf(format string, v ...any)
	Warning(v ...any)
	Warningf(format string, v ...any)
	Info(v ...any)
	Infof(format string, v ...any)
	Error(v ...any)
	Errorf(format string, v ...any)
	Fatal(v ...any)
	Fatalf(format string, v ...any)
	Flush() error

	HandlePanic(title string)
	HandlePanicWithCallback(title string, callback func(error))
}

func NewLogger(cfg *config.Config) Logger {
	switch cfg.LogChannel {
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
