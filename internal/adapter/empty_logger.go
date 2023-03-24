package adapter

import (
	"log"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
)

type emptyLogger struct{}

// NewEmptyLogger return contracts.LoggerInterface
func NewEmptyLogger() contracts.LoggerInterface {
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
