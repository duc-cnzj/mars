package mlog

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDebug(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	SetLogger(l)
	l.EXPECT().Debug(gomock.Any()).Times(1)
	Debug()
}

func TestDebugf(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	SetLogger(l)
	l.EXPECT().Debugf("", "").Times(1)
	Debugf("", "")
}

func TestError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	SetLogger(l)
	l.EXPECT().Error(gomock.Any()).Times(1)
	Error()
}

func TestErrorf(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	SetLogger(l)
	l.EXPECT().Errorf("", "").Times(1)
	Errorf("", "")
}

func TestFatal(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	SetLogger(l)
	l.EXPECT().Fatal(gomock.Any()).Times(1)
	Fatal()
}

func TestFatalf(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	SetLogger(l)
	l.EXPECT().Fatalf("", "").Times(1)
	Fatalf("", "")
}

func TestInfo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	SetLogger(l)
	l.EXPECT().Info(gomock.Any()).Times(1)
	Info()
}

func TestInfof(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	SetLogger(l)
	l.EXPECT().Infof("", "").Times(1)
	Infof("", "")
}

func TestSetLogger(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	SetLogger(l)
	assert.Same(t, logger, l)
}

func TestWarning(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	SetLogger(l)
	l.EXPECT().Warning(gomock.Any()).Times(1)
	Warning()
}

func TestWarningf(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	SetLogger(l)
	l.EXPECT().Warningf("", "").Times(1)
	Warningf("", "")
}
