package adapter

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/stretchr/testify/assert"
)

func TestNewEmptyLogger(t *testing.T) {
	logger := NewEmptyLogger()
	assert.Implements(t, (*contracts.LoggerInterface)(nil), logger)
}

func Test_emptyLogger_Debug(t *testing.T) {
	logger := NewEmptyLogger()
	logger.Debug("")
}

func Test_emptyLogger_Debugf(t *testing.T) {
	logger := NewEmptyLogger()
	logger.Debugf("", "")
}

func Test_emptyLogger_Error(t *testing.T) {
	logger := NewEmptyLogger()
	logger.Error("")
}

func Test_emptyLogger_Errorf(t *testing.T) {
	logger := NewEmptyLogger()
	logger.Errorf("", "")
}

func Test_emptyLogger_Info(t *testing.T) {
	logger := NewEmptyLogger()
	logger.Info("")
}

func Test_emptyLogger_Infof(t *testing.T) {
	logger := NewEmptyLogger()
	logger.Infof("", "")
}

func Test_emptyLogger_Warning(t *testing.T) {
	logger := NewEmptyLogger()
	logger.Warning("")
}

func Test_emptyLogger_Warningf(t *testing.T) {
	logger := NewEmptyLogger()
	logger.Warningf("", "")
}
