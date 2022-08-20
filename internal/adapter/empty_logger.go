package adapter

import "log"

type emptyLogger struct{}

func NewEmptyLogger() *emptyLogger {
	return &emptyLogger{}
}

func (e *emptyLogger) Debug(v ...any) {}

func (e *emptyLogger) Debugf(format string, v ...any) {}

func (e *emptyLogger) Warning(v ...any) {}

func (e *emptyLogger) Warningf(format string, v ...any) {}

func (e *emptyLogger) Info(v ...any) {}

func (e *emptyLogger) Infof(format string, v ...any) {}

func (e *emptyLogger) Error(v ...any) {}

func (e *emptyLogger) Errorf(format string, v ...any) {}

func (e *emptyLogger) Fatal(v ...any) {
	log.Fatal(v...)
}

func (e *emptyLogger) Fatalf(format string, v ...any) {
	log.Fatalf(format, v...)
}
