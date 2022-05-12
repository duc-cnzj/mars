package utils

import (
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandlePanic(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
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
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	app.EXPECT().IsDebug().Return(true)
	defer func() {
		e := recover()
		assert.Equal(t, "err", e)
	}()
	defer HandlePanic("panic")
	panic("err")
}
