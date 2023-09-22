package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"

	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
)

type testApp struct {
	authCalled          bool
	cacheCalled         bool
	configCalled        bool
	dbCalled            bool
	eventCalled         bool
	k8sCalled           bool
	dbmanager           contracts.DBManager
	oidcCalled          bool
	helmCalled          bool
	uploaderCalled      bool
	sfCalled            bool
	tracerCalled        bool
	cronManagerCalled   bool
	cacheLockCalled     bool
	localUploaderCalled bool

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

func (a *testApp) LocalUploader() contracts.Uploader {
	a.localUploaderCalled = true
	return nil
}

func (a *testApp) Singleflight() *singleflight.Group {
	a.sfCalled = true
	return nil
}

func (a *testApp) GetTracer() trace.Tracer {
	a.tracerCalled = true
	return nil
}

func (a *testApp) CronManager() contracts.CronManager {
	a.cronManagerCalled = true
	return nil
}

func (a *testApp) CacheLock() contracts.Locker {
	a.cacheLockCalled = true
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

func (a *testApp) Helmer() contracts.Helmer {
	a.helmCalled = true
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

func TestLocalUploader(t *testing.T) {
	a := &testApp{}
	instance.SetInstance(a)
	LocalUploader()
	assert.True(t, a.localUploaderCalled)
}

func TestTracer(t *testing.T) {
	a := &testApp{}
	instance.SetInstance(a)
	Tracer()
	assert.True(t, a.tracerCalled)
}

func TestCacheLock(t *testing.T) {
	a := &testApp{}
	instance.SetInstance(a)
	CacheLock()
	assert.True(t, a.cacheLockCalled)
}

func TestCronManager(t *testing.T) {
	a := &testApp{}
	instance.SetInstance(a)
	CronManager()
	assert.True(t, a.cronManagerCalled)
}

func TestHelmer(t *testing.T) {
	a := &testApp{}
	instance.SetInstance(a)
	Helmer()
	assert.True(t, a.helmCalled)
}
