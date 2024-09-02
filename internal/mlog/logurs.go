package mlog

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

type logrusLogger struct {
	logrus *logrus.Logger
	debug  bool
	module string
}

var _ Logger = (*logrusLogger)(nil)

// NewLogrusLogger return contracts.Logger
func NewLogrusLogger(debug bool) Logger {
	logger := logrus.New()

	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006/01/02 15:04:05",
	})

	if debug {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{})
		logger.SetLevel(logrus.InfoLevel)
	}

	return &logrusLogger{
		debug:  debug,
		logrus: logger,
	}
}

func (z *logrusLogger) WithModule(module string) Logger {
	return &logrusLogger{
		debug:  z.debug,
		module: module,
		logrus: z.logrus,
	}
}

func (z *logrusLogger) HandlePanic(title string) {
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

func (z *logrusLogger) HandlePanicWithCallback(title string, callback func(error)) {
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

func (z *logrusLogger) Flush() error {
	return nil
}

// Debug print debug msg
func (z *logrusLogger) Debug(v ...any) {
	z.logrus.WithFields(z.fields()).Debug(v...)
}

// Debugf printf debug msg
func (z *logrusLogger) Debugf(format string, v ...any) {
	z.logrus.WithFields(z.fields()).Debugf(format, v...)
}

func (z *logrusLogger) DebugCtx(ctx context.Context, v ...any) {
	z.logWithContext(ctx).WithFields(z.fields()).Debug(v...)
}

func (z *logrusLogger) DebugCtxf(ctx context.Context, format string, v ...any) {
	z.logWithContext(ctx).WithFields(z.fields()).Debugf(format, v...)
}

// Warning print Warning msg
func (z *logrusLogger) Warning(v ...any) {
	z.logrus.WithFields(z.fields()).Warn(v...)
}

// Warningf prints Warning msg
func (z *logrusLogger) Warningf(format string, v ...any) {
	z.logrus.WithFields(z.fields()).Warnf(format, v...)
}

func (z *logrusLogger) WarningCtx(ctx context.Context, v ...any) {
	z.logWithContext(ctx).WithFields(z.fields()).Warn(v...)
}

func (z *logrusLogger) WarningCtxf(ctx context.Context, format string, v ...any) {
	z.logWithContext(ctx).WithFields(z.fields()).Warnf(format, v...)
}

// Info print info msg
func (z *logrusLogger) Info(v ...any) {
	z.logrus.WithFields(z.fields()).Info(v...)
}

// Infof printf info msg
func (z *logrusLogger) Infof(format string, v ...any) {
	z.logrus.WithFields(z.fields()).Infof(format, v...)
}

func (z *logrusLogger) InfoCtx(ctx context.Context, v ...any) {
	z.logWithContext(ctx).WithFields(z.fields()).Info(v...)
}

func (z *logrusLogger) InfoCtxf(ctx context.Context, format string, v ...any) {
	z.logWithContext(ctx).WithFields(z.fields()).Infof(format, v...)

}

// Error print err msg
func (z *logrusLogger) Error(v ...any) {
	z.logrus.WithFields(z.fields()).Error(v...)
}

// Errorf printf err msg
func (z *logrusLogger) Errorf(format string, v ...any) {
	z.logrus.WithFields(z.fields()).Errorf(format, v...)
}

func (z *logrusLogger) ErrorCtx(ctx context.Context, v ...any) {
	z.logWithContext(ctx).WithFields(z.fields()).Error(v...)
}

func (z *logrusLogger) ErrorCtxf(ctx context.Context, format string, v ...any) {
	z.logWithContext(ctx).WithFields(z.fields()).Errorf(format, v...)
}

// Fatal fatal err.
func (z *logrusLogger) Fatal(v ...any) {
	z.logrus.WithFields(z.fields()).Fatal(v...)
}

// Fatalf fatalf err.
func (z *logrusLogger) Fatalf(format string, v ...any) {
	z.logrus.WithFields(z.fields()).Fatalf(format, v...)
}

func (z *logrusLogger) FatalCtx(ctx context.Context, v ...any) {
	z.logWithContext(ctx).WithFields(z.fields()).Fatal(v...)
}

func (z *logrusLogger) FatalCtxf(ctx context.Context, format string, v ...any) {
	z.logWithContext(ctx).WithFields(z.fields()).Fatalf(format, v...)
}

// fields return caller key and file line.
func (z *logrusLogger) fields() logrus.Fields {
	pc, _, _, _ := runtime.Caller(2)
	file, line := runtime.FuncForPC(pc).FileLine(pc)

	f := map[string]any{
		"file": fmt.Sprintf("%s:%d.", file, line),
	}
	if z.module != "" {
		f["module"] = z.module
	}
	return f
}

func (z *logrusLogger) logWithContext(ctx context.Context) *logrus.Entry {
	spanContext := trace.SpanContextFromContext(ctx)
	return z.logrus.WithField("SpanID", spanContext.SpanID()).WithField("TraceID", spanContext.TraceID())
}
