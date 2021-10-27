package adapter

import (
	"fmt"
	"os"
	"runtime"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/sirupsen/logrus"
)

type LogrusLogger struct {
	app    contracts.ApplicationInterface
	logrus *logrus.Logger
}

func NewLogrusLogger(app contracts.ApplicationInterface) contracts.LoggerInterface {
	logger := logrus.New()

	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006/01/02 15:04:05",
	})

	if app.IsDebug() {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{})
		logger.SetLevel(logrus.InfoLevel)
	}

	return &LogrusLogger{
		app:    app,
		logrus: logger,
	}
}

func (z *LogrusLogger) Debug(v ...interface{}) {
	z.logrus.WithField(callerField()).Debug(v...)
}

func (z *LogrusLogger) Debugf(format string, v ...interface{}) {
	z.logrus.WithField(callerField()).Debugf(format, v...)
}

func (z *LogrusLogger) Warning(v ...interface{}) {
	z.logrus.WithField(callerField()).Warn(v...)
}

func (z *LogrusLogger) Warningf(format string, v ...interface{}) {
	z.logrus.WithField(callerField()).Warnf(format, v...)
}

func (z *LogrusLogger) Info(v ...interface{}) {
	z.logrus.WithField(callerField()).Info(v...)
}

func (z *LogrusLogger) Infof(format string, v ...interface{}) {
	z.logrus.WithField(callerField()).Infof(format, v...)
}

func (z *LogrusLogger) Error(v ...interface{}) {
	z.logrus.WithField(callerField()).Error(v...)
}

func (z *LogrusLogger) Errorf(format string, v ...interface{}) {
	z.logrus.WithField(callerField()).Errorf(format, v...)
}

func (z *LogrusLogger) Fatal(v ...interface{}) {
	z.logrus.WithField(callerField()).Fatal(v...)
}

func (z *LogrusLogger) Fatalf(format string, v ...interface{}) {
	z.logrus.WithField(callerField()).Fatalf(format, v...)
}

func callerField() (string, string) {
	pc, _, _, _ := runtime.Caller(3)
	file, line := runtime.FuncForPC(pc).FileLine(pc)

	return "file", fmt.Sprintf("%s:%d.", file, line)
}
