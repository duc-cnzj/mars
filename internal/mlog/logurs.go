package mlog

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
)

type logrusLogger struct {
	logrus *logrus.Logger
	debug  bool
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
	z.logrus.WithField(callerField()).Debug(v...)
}

// Debugf printf debug msg
func (z *logrusLogger) Debugf(format string, v ...any) {
	z.logrus.WithField(callerField()).Debugf(format, v...)
}

// Warning print Warning msg
func (z *logrusLogger) Warning(v ...any) {
	z.logrus.WithField(callerField()).Warn(v...)
}

// Warningf prints Warning msg
func (z *logrusLogger) Warningf(format string, v ...any) {
	z.logrus.WithField(callerField()).Warnf(format, v...)
}

// Info print info msg
func (z *logrusLogger) Info(v ...any) {
	z.logrus.WithField(callerField()).Info(v...)
}

// Infof printf info msg
func (z *logrusLogger) Infof(format string, v ...any) {
	z.logrus.WithField(callerField()).Infof(format, v...)
}

// Error print err msg
func (z *logrusLogger) Error(v ...any) {
	z.logrus.WithField(callerField()).Error(v...)
}

// Errorf printf err msg
func (z *logrusLogger) Errorf(format string, v ...any) {
	z.logrus.WithField(callerField()).Errorf(format, v...)
}

// Fatal fatal err.
func (z *logrusLogger) Fatal(v ...any) {
	z.logrus.WithField(callerField()).Fatal(v...)
}

// Fatalf fatalf err.
func (z *logrusLogger) Fatalf(format string, v ...any) {
	z.logrus.WithField(callerField()).Fatalf(format, v...)
}

// callerField return caller key and file line.
func callerField() (string, string) {
	pc, _, _, _ := runtime.Caller(3)
	file, line := runtime.FuncForPC(pc).FileLine(pc)

	return "file", fmt.Sprintf("%s:%d.", file, line)
}
