package mlog

import (
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/sirupsen/logrus"
)

var logger contracts.LoggerInterface = logrus.New()

func SetLogger(l contracts.LoggerInterface) {
	logger = l
}

func Debug(v ...any) {
	logger.Debug(v...)
}

func Debugf(format string, v ...any) {
	logger.Debugf(format, v...)
}

func Warning(v ...any) {
	logger.Warning(v...)
}

func Warningf(format string, v ...any) {
	logger.Warningf(format, v...)
}

func Info(v ...any) {
	logger.Info(v...)
}

func Infof(format string, v ...any) {
	logger.Infof(format, v...)
}

func Error(v ...any) {
	logger.Error(v...)
}

func Errorf(format string, v ...any) {
	logger.Errorf(format, v...)
}

func Fatal(v ...any) {
	logger.Fatal(v...)
}

func Fatalf(format string, v ...any) {
	logger.Fatalf(format, v...)
}
