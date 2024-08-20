package application

import (
	"testing"

	auth2 "github.com/duc-cnzj/mars/v4/internal/auth"
	cache2 "github.com/duc-cnzj/mars/v4/internal/cache"
	config2 "github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/cron"
	data2 "github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/event"
	"github.com/duc-cnzj/mars/v4/internal/locker"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	uploader2 "github.com/duc-cnzj/mars/v4/internal/uploader"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/sync/singleflight"
)

type testBoot struct {
	tags []string
}

func (t *testBoot) Bootstrap(a App) error {
	return nil
}

func (t *testBoot) Tags() []string {
	return t.tags
}

func TestNewAppWithValidConfig(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	config := &config2.Config{
		Debug: true,
	}
	data := data2.NewMockData(m)
	logger := mlog.NewLogger(nil)
	uploader := uploader2.NewMockUploader(m)
	auth := auth2.NewMockAuth(m)
	dispatcher := event.NewMockDispatcher(m)
	cronManager := cron.NewMockManager(m)
	cache := cache2.NewMockCache(m)
	cacheLock := locker.NewMockLocker(m)
	sf := &singleflight.Group{}
	pm := NewMockPluginManger(m)
	reg := &GrpcRegistry{}
	ws := NewMockWsServer(m)
	pr := &prometheus.Registry{}

	b1 := &testBoot{
		tags: []string{"cron"},
	}
	appli := NewApp(
		config,
		data,
		logger,
		uploader,
		auth,
		dispatcher,
		cronManager,
		cache,
		cacheLock,
		sf,
		pm,
		reg,
		ws,
		pr,
		WithBootstrappers(b1, &testBoot{}),
		WithExcludeTags("cron"),
	)

	assert.NotNil(t, appli)
	assert.NotNil(t, appli.Data())
	assert.NotNil(t, appli.Logger())
	assert.NotNil(t, appli.Uploader())
	assert.NotNil(t, appli.Auth())
	assert.NotNil(t, appli.Dispatcher())
	assert.NotNil(t, appli.CronManager())
	assert.NotNil(t, appli.Cache())
	assert.NotNil(t, appli.Locker())
	assert.NotNil(t, appli.Singleflight())
	assert.NotNil(t, appli.PluginMgr())
	assert.NotNil(t, appli.GrpcRegistry())
	assert.NotNil(t, appli.WsServer())
	assert.NotNil(t, appli.PrometheusRegistry())
	assert.True(t, appli.IsDebug())

	assert.Len(t, appli.(*app).bootstrappers, 1)
	assert.Len(t, appli.(*app).excludeBoots, 1)
	assert.Equal(t, b1, appli.(*app).excludeBoots[0])
}

func TestWithBootstrappers(t *testing.T) {}

func TestWithExcludeTags(t *testing.T) {}

func TestWithMustBootedBootstrappers(t *testing.T) {}

func Test_app_AddServer(t *testing.T) {}

func Test_app_Auth(t *testing.T) {}

func Test_app_BeforeServerRunHooks(t *testing.T) {}

func Test_app_Bootstrap(t *testing.T) {}

func Test_app_Cache(t *testing.T) {}

func Test_app_Config(t *testing.T) {}

func Test_app_CronManager(t *testing.T) {}

func Test_app_DB(t *testing.T) {}

func Test_app_Data(t *testing.T) {}

func Test_app_Dispatcher(t *testing.T) {}

func Test_app_Done(t *testing.T) {}

func Test_app_GrpcRegistry(t *testing.T) {}

func Test_app_IsDebug(t *testing.T) {}

func Test_app_Locker(t *testing.T) {}

func Test_app_Logger(t *testing.T) {}

func Test_app_Oidc(t *testing.T) {}

func Test_app_PluginMgr(t *testing.T) {}

func Test_app_PrometheusRegistry(t *testing.T) {}

func Test_app_RegisterAfterShutdownFunc(t *testing.T) {}

func Test_app_RegisterBeforeShutdownFunc(t *testing.T) {}

func Test_app_Run(t *testing.T) {}

func Test_app_RunServerHooks(t *testing.T) {}

func Test_app_Shutdown(t *testing.T) {}

func Test_app_Singleflight(t *testing.T) {}

func Test_app_Uploader(t *testing.T) {}

func Test_app_WsServer(t *testing.T) {}

func Test_bootShortName(t *testing.T) {}

func Test_bootTags_has(t *testing.T) {}

func Test_excludeBootstrapperByTags(t *testing.T) {}

func Test_printConfig(t *testing.T) {}
