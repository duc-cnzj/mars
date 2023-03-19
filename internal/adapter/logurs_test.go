package adapter

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewLogrusLogger(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().IsDebug().Return(false)
	assert.Implements(t, (*contracts.LoggerInterface)(nil), NewLogrusLogger(app))
}

func Test_callerField(t *testing.T) {
	field, _ := callerField()
	assert.Equal(t, "file", field)
}

func TestLogrusLogger_Debug(t *testing.T) {
	NewLogrusLogger(&app{isDebug: true}).Debug("aaa")
}

func TestLogrusLogger_Debugf(t *testing.T) {
	NewLogrusLogger(&app{isDebug: true}).Debugf("%v", "aaa")
}

func TestLogrusLogger_Error(t *testing.T) {
	NewLogrusLogger(&app{isDebug: true}).Error("aaa")
}

func TestLogrusLogger_Errorf(t *testing.T) {
	NewLogrusLogger(&app{isDebug: true}).Errorf("%v", "aaa")
}

func TestLogrusLogger_Info(t *testing.T) {
	NewLogrusLogger(&app{isDebug: true}).Info("aaa")
}

func TestLogrusLogger_Infof(t *testing.T) {
	NewLogrusLogger(&app{isDebug: true}).Infof("%v", "aaa")
}

func TestLogrusLogger_Warning(t *testing.T) {
	NewLogrusLogger(&app{isDebug: true}).Warning("aaa")
}

func TestLogrusLogger_Warningf(t *testing.T) {
	NewLogrusLogger(&app{isDebug: true}).Warningf("%v", "aaa")
}
