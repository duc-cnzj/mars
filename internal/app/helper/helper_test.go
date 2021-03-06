package app

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type testApp struct {
	authCalled       bool
	cacheCalled      bool
	configCalled     bool
	dbCalled         bool
	eventCalled      bool
	k8sCalled        bool
	k8smetricsCalled bool
	dbmanager        contracts.DBManager
	oidcCalled       bool
	uploaderCalled   bool
	sfCalled         bool

	contracts.ApplicationInterface
}

func (a *testApp) Oidc() contracts.OidcConfig {
	a.oidcCalled = true
	return nil
}
func (a *testApp) Auth() contracts.AuthInterface {
	a.authCalled = true
	return nil
}
func (a *testApp) Uploader() contracts.Uploader {
	a.uploaderCalled = true
	return nil
}

func (a *testApp) Singleflight() *singleflight.Group {
	a.sfCalled = true
	return nil
}
func (a *testApp) K8sClient() *contracts.K8sClient {
	a.k8sCalled = true
	return &contracts.K8sClient{}
}

type testdbManager struct {
	dbcalled bool
}

func (t *testdbManager) DB() *gorm.DB {
	t.dbcalled = true
	return nil
}

func (t *testdbManager) SetDB(db *gorm.DB) {
}

func (t *testdbManager) AutoMigrate(dst ...any) error {
	return nil
}

func (a *testApp) DBManager() contracts.DBManager {
	a.dbCalled = true
	return a.dbmanager
}
func (a *testApp) EventDispatcher() contracts.DispatcherInterface {
	a.eventCalled = true
	return nil
}
func (a *testApp) Config() *config.Config {
	a.configCalled = true
	return nil
}
func (a *testApp) Cache() contracts.CacheInterface {
	a.cacheCalled = true
	return nil
}
func (a *testApp) Metrics() contracts.Metrics {
	a.k8smetricsCalled = true
	return nil
}

func TestApp(t *testing.T) {
	a := &testApp{}
	instance.SetInstance(a)
	assert.Same(t, a, App())
}

func TestAuth(t *testing.T) {
	a := &testApp{}
	instance.SetInstance(a)
	Auth()
	assert.True(t, a.authCalled)
}

func TestCache(t *testing.T) {
	a := &testApp{}
	instance.SetInstance(a)
	Cache()
	assert.True(t, a.cacheCalled)
}

func TestConfig(t *testing.T) {
	a := &testApp{}
	instance.SetInstance(a)
	Config()
	assert.True(t, a.configCalled)
}

func TestDB(t *testing.T) {
	dbm := &testdbManager{}
	a := &testApp{dbmanager: dbm}
	instance.SetInstance(a)
	DB()
	assert.True(t, a.dbCalled)
	assert.True(t, dbm.dbcalled)
}

func TestEvent(t *testing.T) {
	a := &testApp{}
	instance.SetInstance(a)
	Event()
	assert.True(t, a.eventCalled)
}

func TestK8sClient(t *testing.T) {
	a := &testApp{}
	instance.SetInstance(a)
	K8sClient()
	assert.True(t, a.k8sCalled)
}

func TestK8sClientSet(t *testing.T) {
	a := &testApp{}
	instance.SetInstance(a)
	K8sClientSet()
	assert.True(t, a.k8sCalled)
}

func TestK8sMetrics(t *testing.T) {
	a := &testApp{}
	instance.SetInstance(a)
	K8sMetrics()
	assert.True(t, a.k8sCalled)
}

func TestMetrics(t *testing.T) {
	a := &testApp{}
	instance.SetInstance(a)
	Metrics()
	assert.True(t, a.k8smetricsCalled)
}

func TestOidc(t *testing.T) {
	a := &testApp{}
	instance.SetInstance(a)
	Oidc()
	assert.True(t, a.oidcCalled)
}

func TestSingleflight(t *testing.T) {
	a := &testApp{}
	instance.SetInstance(a)
	Singleflight()
	assert.True(t, a.sfCalled)
}

func TestUploader(t *testing.T) {
	a := &testApp{}
	instance.SetInstance(a)
	Uploader()
	assert.True(t, a.uploaderCalled)
}
