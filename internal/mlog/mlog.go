package mlog

import (
	"fmt"
	"runtime"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/sirupsen/logrus"
)

var logger contracts.LoggerInterface = logrus.New()

func SetLogger(l contracts.LoggerInterface) {
	logger = l
}

func Debug(v ...interface{}) {
	logger.Debug(v...)
}

func Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

func Warning(v ...interface{}) {
	v = append(v, getCaller(true))

	logger.Warning(v...)
}

func Warningf(format string, v ...interface{}) {
	format = format + "%s"
	v = append(v, getCaller(true))

	logger.Warningf(format, v...)
}

func Info(v ...interface{}) {
	v = append(v, getCaller(true))

	logger.Info(v...)
}

func Infof(format string, v ...interface{}) {
	format = format + "%s"
	v = append(v, getCaller(true))

	logger.Infof(format, v...)
}

func Error(v ...interface{}) {
	v = append(v, getCaller(false))

	logger.Error(v...)
}

func Errorf(format string, v ...interface{}) {
	format = format + "%s"
	v = append(v, getCaller(false))

	logger.Errorf(format, v...)
}

func Fatal(v ...interface{}) {
	v = append(v, getCaller(false))

	logger.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	format = format + "%s"
	v = append(v, getCaller(false))

	logger.Fatalf(format, v...)
}

func getCaller(simple bool) string {
	pc, _, _, _ := runtime.Caller(2)
	file, line := runtime.FuncForPC(pc).FileLine(pc)

	if simple {
		return fmt.Sprintf("\tfile=%s:%d.", file, line)
	}

	return fmt.Sprintf("\n[stacktrace]\n#0: func=%v \n#1: file=%s \n#2: line=%d.", runtime.FuncForPC(pc).Name(), file, line)
}
