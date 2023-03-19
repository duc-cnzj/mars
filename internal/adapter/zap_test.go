package adapter

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/golang/mock/gomock"
)

func TestNewZapLogger(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().IsDebug().Return(true)
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).Times(1)
	NewZapLogger(app)
}

type app struct {
	isDebug bool
	cb      contracts.Callback
	contracts.ApplicationInterface
}

func (a *app) IsDebug() bool {
	return a.isDebug
}

func (a *app) RegisterAfterShutdownFunc(cb contracts.Callback) {
	a.cb = cb
}

func TestZapLogger_Debug(t *testing.T) {
	a := &app{isDebug: true}
	NewZapLogger(a).Debug("aaa")
	a.cb(a)
}

func TestZapLogger_Debugf(t *testing.T) {
	a := &app{isDebug: true}
	NewZapLogger(a).Debugf("%v", "aaa")
	a.cb(a)
}

func TestZapLogger_Error(t *testing.T) {
	a := &app{isDebug: true}
	NewZapLogger(a).Debugf("aaa")
	a.cb(a)
}

func TestZapLogger_Errorf(t *testing.T) {
	a := &app{isDebug: true}
	NewZapLogger(a).Errorf("%v", "aaa")
	a.cb(a)
}

func TestZapLogger_Info(t *testing.T) {
	a := &app{isDebug: true}
	NewZapLogger(a).Info("aaa")
	a.cb(a)
}

func TestZapLogger_Infof(t *testing.T) {
	a := &app{isDebug: true}
	NewZapLogger(a).Infof("%v", "aaa")
	a.cb(a)
}

func TestZapLogger_Warning(t *testing.T) {
	a := &app{isDebug: true}
	NewZapLogger(a).Warning("aaa")
	a.cb(a)
}

func TestZapLogger_Warningf(t *testing.T) {
	a := &app{isDebug: true}
	NewZapLogger(a).Warningf("%v", "aaa")
	a.cb(a)
}
