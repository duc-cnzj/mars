package bootstrappers

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/contracts"
	mevent "github.com/duc-cnzj/mars/internal/event/events"

	"github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/cache"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	key, _ := rsa.GenerateKey(rand.Reader, 1024)
	privateKey, _ := x509.MarshalPKCS8PrivateKey(key)
	bf := bytes.Buffer{}
	pem.Encode(&bf, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKey})
	app.EXPECT().Config().Return(&config.Config{
		PrivateKey: bf.String(),
	}).Times(1)
	app.EXPECT().SetAuth(auth.NewAuth(key, key.Public().(*rsa.PublicKey))).Times(1)
	(&AuthBootstrapper{}).Bootstrap(app)
}

type cacheMatcher struct {
	wants any
	t     *testing.T
}

func (c *cacheMatcher) Matches(x any) bool {
	assert.IsType(c.t, c.wants, x)
	return true
}

func (c *cacheMatcher) String() string {
	return ""
}

func TestCacheBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().Config().Return(&config.Config{
		DBDriver:    "sqlite",
		CacheDriver: "db",
	}).Times(1)
	app.EXPECT().Singleflight().Times(1)
	app.EXPECT().SetCache(&cacheMatcher{
		wants: (*cache.Cache)(nil),
		t:     t,
	})
	assert.Nil(t, (&CacheBootstrapper{}).Bootstrap(app))
	app.EXPECT().Config().Return(&config.Config{
		DBDriver:    "mysql",
		CacheDriver: "db",
	}).Times(1)
	app.EXPECT().SetCache(&cacheMatcher{
		wants: (*cache.DBCache)(nil),
		t:     t,
	})
	assert.Nil(t, (&CacheBootstrapper{}).Bootstrap(app))
	app.EXPECT().Config().Return(&config.Config{
		CacheDriver: "xxxx",
	}).Times(1)
	assert.Error(t, (&CacheBootstrapper{}).Bootstrap(app))
}

func TestMetricsBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().SetMetrics(gomock.Any()).Times(1)
	app.EXPECT().AddServer(gomock.Any()).Times(1)
	(&MetricsBootstrapper{}).Bootstrap(app)
}

func TestOidcBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().Config().Times(1).Return(&config.Config{})
	app.EXPECT().SetOidc(gomock.Any()).Times(1)
	(&OidcBootstrapper{}).Bootstrap(app)
}

func TestPluginsBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().SetPlugins(gomock.Any()).Times(1)
	(&PluginsBootstrapper{}).Bootstrap(app)
}

func TestGrpcBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().AddServer(gomock.Any()).Times(1)
	app.EXPECT().Config().Times(1).Return(&config.Config{})
	(&GrpcBootstrapper{}).Bootstrap(app)
}

func TestPprofBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().AddServer(gomock.Any()).Times(0)
	app.EXPECT().Config().Times(1).Return(&config.Config{
		ProfileEnabled: false,
	})
	(&PprofBootstrapper{}).Bootstrap(app)
	app.EXPECT().AddServer(gomock.Any()).Times(1)
	app.EXPECT().Config().Times(1).Return(&config.Config{
		ProfileEnabled: true,
	})
	(&PprofBootstrapper{}).Bootstrap(app)
}

func TestEventBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	mevent.Register("xx", func(a any, event contracts.Event) error {
		return nil
	})
	app := mock.NewMockApplicationInterface(controller)
	d := mock.NewMockDispatcherInterface(controller)
	app.EXPECT().EventDispatcher().Return(d).AnyTimes()
	assert.Greater(t, len(mevent.RegisteredEvents()), 0)
	d.EXPECT().Listen(gomock.Any(), gomock.Any()).Times(len(mevent.RegisteredEvents()))
	(&EventBootstrapper{}).Bootstrap(app)
}

func TestLogBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().IsDebug().Return(false).AnyTimes()
	app.EXPECT().Config().Return(&config.Config{LogChannel: "xxx"}).Times(2)
	err := (&LogBootstrapper{}).Bootstrap(app)
	assert.Error(t, err)

	app.EXPECT().Config().Return(&config.Config{LogChannel: "logrus"}).Times(1)
	err = (&LogBootstrapper{}).Bootstrap(app)
	assert.Nil(t, err)
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).Times(1)
	app.EXPECT().Config().Return(&config.Config{LogChannel: "zap"}).Times(1)
	err = (&LogBootstrapper{}).Bootstrap(app)
	assert.Nil(t, err)
}

func TestDBBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().Config().Return(&config.Config{DBDriver: "sqlite", DBDatabase: "file::memory:?cache=shared"}).Times(1)

	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).Times(1)
	app.EXPECT().IsDebug().Return(false).AnyTimes()
	dbm := mock.NewMockDBManager(controller)
	dbm.EXPECT().SetDB(gomock.Any()).Times(1)
	dbm.EXPECT().AutoMigrate(gomock.Any()).Times(1)
	app.EXPECT().DBManager().Return(dbm).AnyTimes()

	assert.Nil(t, (&DBBootstrapper{}).Bootstrap(app))

	app.EXPECT().Config().Return(&config.Config{DBDriver: "xxx"}).Times(1)
	assert.Equal(t, "db_driver must in ['sqlite', 'mysql']", (&DBBootstrapper{}).Bootstrap(app).Error())
}

func TestAppBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	instance.SetInstance(app)
	app.EXPECT().BeforeServerRunHooks(gomock.Any()).Times(1)

	app.EXPECT().Config().Return(&config.Config{
		GitServerPlugin:     config.Plugin{Name: "test_git_server"},
		PicturePlugin:       config.Plugin{Name: "test_picture"},
		DomainManagerPlugin: config.Plugin{Name: "test_domain"},
		WsSenderPlugin:      config.Plugin{Name: "test_wssender"},
	}).AnyTimes()

	gits := mock.NewMockGitServer(controller)
	app.EXPECT().GetPluginByName("test_git_server").Return(gits).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	gits.EXPECT().Initialize(gomock.All()).AnyTimes()

	pictrure := mock.NewMockPictureInterface(controller)
	app.EXPECT().GetPluginByName("test_picture").Return(pictrure).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	pictrure.EXPECT().Initialize(gomock.All()).AnyTimes()

	d := mock.NewMockDomainManager(controller)
	app.EXPECT().GetPluginByName("test_domain").Return(d).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	d.EXPECT().Initialize(gomock.All()).AnyTimes()

	ws := mock.NewMockWsSender(controller)
	app.EXPECT().GetPluginByName("test_wssender").Return(ws).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.All()).AnyTimes()
	ws.EXPECT().Initialize(gomock.All()).AnyTimes()

	assert.Nil(t, (&AppBootstrapper{}).Bootstrap(app))
}

func TestApiGatewayBootstrapper_Bootstrap(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	app := mock.NewMockApplicationInterface(controller)
	app.EXPECT().Config().Return(&config.Config{GrpcPort: "50000"})
	app.EXPECT().AddServer(&apiGateway{endpoint: fmt.Sprintf("localhost:%s", "50000")}).Times(1)
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).Times(1)
	assert.Nil(t, (&ApiGatewayBootstrapper{}).Bootstrap(app))
}
