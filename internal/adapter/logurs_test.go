package adapter

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewLogrusLogger(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	app.EXPECT().IsDebug().Return(false)
	assert.Implements(t, (*contracts.LoggerInterface)(nil), NewLogrusLogger(app))
}

func Test_callerField(t *testing.T) {
	field, _ := callerField()
	assert.Equal(t, "file", field)
}
