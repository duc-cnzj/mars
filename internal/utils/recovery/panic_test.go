package recovery

import (
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestHandlePanic(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().IsDebug().Return(false)
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)
	defer HandlePanic("panic")
	panic("err")
}

func TestHandlePanic1(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)
	app := testutil.MockApp(m)
	app.EXPECT().IsDebug().Return(true)
	defer func() {
		e := recover()
		assert.Equal(t, "err", e)
	}()
	defer HandlePanic("panic")
	panic("err")
}

func TestHandlePanicWithCallback(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)
	app := testutil.MockApp(m)
	app.EXPECT().IsDebug().Return(true)
	var cbCalled bool
	var cbErr error
	outError := errors.New("err")

	defer func() {
		e := recover()
		assert.Equal(t, outError, e)
		assert.True(t, cbCalled)
		assert.Equal(t, outError, cbErr)

	}()
	defer HandlePanicWithCallback("panic", func(err error) {
		cbCalled = true
		cbErr = err
	})
	panic(outError)
}

func TestHandlePanicWithCallback1(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().IsDebug().Return(false)
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)
	var cbCalled bool
	var cbErr error
	outError := errors.New("err")
	defer func() {
		assert.True(t, cbCalled)
		assert.Equal(t, outError, cbErr)
	}()
	defer HandlePanicWithCallback("panic", func(err error) {
		cbErr = err
		cbCalled = true
	})
	panic(outError)
}

func TestHandlePanicWithCallback2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().IsDebug().Return(false)
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)
	var cbCalled bool
	var cbErr error
	outError := "err str"
	defer func() {
		assert.True(t, cbCalled)
		assert.Equal(t, errors.New(outError), cbErr)
	}()
	defer HandlePanicWithCallback("panic", func(err error) {
		cbErr = err
		cbCalled = true
	})
	panic(outError)
}
