package mlog

import (
	"fmt"
	"runtime"

	"github.com/DuC-cnZj/mars/pkg/contracts"
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
	logger.Warning(v...)
}

func Warningf(format string, v ...interface{}) {
	logger.Warningf(format, v...)
}

func Info(v ...interface{}) {
	logger.Info(v...)
}

func Infof(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

func Error(v ...interface{}) {
	v = append([]interface{}{getCaller()}, v...)

	logger.Error(v...)
}

func Errorf(format string, v ...interface{}) {
	format = "%s" + format
	v = append([]interface{}{getCaller()}, v...)

	logger.Errorf(format, v...)
}

func Fatal(v ...interface{}) {
	v = append([]interface{}{getCaller()}, v...)
	logger.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	format = "%s" + format
	v = append([]interface{}{getCaller()}, v...)

	logger.Fatalf(format, v...)
}

func getCaller() string {
	pc, _, _, _ := runtime.Caller(2)
	file, line := runtime.FuncForPC(pc).FileLine(pc)

	return fmt.Sprintf("func=%v file=%s line=%d |\t", runtime.FuncForPC(pc).Name(), file, line)
}
