package application

import (
	"context"
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
	ws := NewMockWsHttpServer(m)
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

func TestWithBootstrappers(t *testing.T) {
	a := &app{}
	WithBootstrappers(&testBoot{})(a)
	assert.Len(t, a.bootstrappers, 1)
}

func TestWithExcludeTags(t *testing.T) {
	a := &app{}
	WithExcludeTags("cron")(a)
	assert.Len(t, a.excludeTags, 1)
}

func TestWithMustBootedBootstrappers(t *testing.T) {
	a := &app{}
	WithMustBootedBootstrappers(&testBoot{})(a)
	assert.Len(t, a.mustBooted, 1)
}

type testServer struct {
	Server
}

func Test_app_AddServer(t *testing.T) {
	a := &app{}
	a.AddServer(&testServer{})
	assert.Len(t, a.servers, 1)
}

func Test_app_Auth(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	auth := auth2.NewMockAuth(m)
	a := &app{
		auth: auth,
	}
	assert.NotNil(t, a.Auth())
}

func Test_app_BeforeServerRunHooks(t *testing.T) {
	a := &app{hooks: map[hook][]Callback{}}
	a.BeforeServerRunHooks(func(App) {})
	assert.Len(t, a.hooks, 1)
}

func Test_app_Bootstrap(t *testing.T) {
	a := &app{bootstrappers: []Bootstrapper{&testBoot{}}}
	assert.Nil(t, a.Bootstrap())
}

func Test_app_Cache(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	cache := cache2.NewMockCache(m)
	a := &app{
		cache: cache,
	}
	assert.NotNil(t, a.Cache())
}

func Test_app_Config(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	config := &config2.Config{}
	a := &app{
		config: config,
	}
	assert.NotNil(t, a.Config())
}

func Test_app_CronManager(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	cronManager := cron.NewMockManager(m)
	a := &app{
		cronManager: cronManager,
	}
	assert.NotNil(t, a.CronManager())
}

func Test_app_DB(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	data := data2.NewMockData(m)
	data.EXPECT().DB()
	a := &app{data: data}
	assert.Nil(t, a.DB())
}

func Test_app_Data(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	data := data2.NewMockData(m)
	a := &app{data: data}
	assert.NotNil(t, a.Data())
}

func Test_app_Dispatcher(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	dispatcher := event.NewMockDispatcher(m)
	a := &app{
		dispatcher: dispatcher,
	}
	assert.NotNil(t, a.Dispatcher())
}

func Test_app_Done(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()
	a := &app{done: ctx}
	assert.NotNil(t, a.Done())
}

func Test_app_GrpcRegistry(t *testing.T) {
	a := &app{reg: &GrpcRegistry{}}
	assert.NotNil(t, a.GrpcRegistry())
}

func Test_app_IsDebug(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	config := &config2.Config{
		Debug: true,
	}
	a := &app{
		config: config,
	}
	assert.True(t, a.IsDebug())
}

func Test_app_Locker(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	locker := locker.NewMockLocker(m)
	a := &app{
		cacheLock: locker,
	}
	assert.NotNil(t, a.Locker())
}

func Test_app_Logger(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	logger := mlog.NewLogger(nil)
	a := &app{
		logger: logger,
	}
	assert.NotNil(t, a.Logger())
}

func Test_app_Oidc(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	data := data2.NewMockData(m)
	data.EXPECT().OidcConfig()
	a := &app{data: data}
	assert.Nil(t, a.Oidc())
}

func Test_app_PluginMgr(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	pm := NewMockPluginManger(m)
	a := &app{pluginManager: pm}
	assert.NotNil(t, a.PluginMgr())
}

func Test_app_PrometheusRegistry(t *testing.T) {
	a := &app{prometheusRegistry: &prometheus.Registry{}}
	assert.NotNil(t, a.PrometheusRegistry())
}

func Test_app_RegisterAfterShutdownFunc(t *testing.T) {
	a := &app{hooks: map[hook][]Callback{}}
	a.RegisterAfterShutdownFunc(func(App) {})
	assert.Len(t, a.hooks[afterDownHook], 1)
}

func Test_app_RegisterBeforeShutdownFunc(t *testing.T) {
	a := &app{hooks: map[hook][]Callback{}}
	a.RegisterBeforeShutdownFunc(func(App) {})
	assert.Len(t, a.hooks[beforeDownHook], 1)
}

func Test_app_Run(t *testing.T) {
	//a := &app{}
	//assert.NotNil(t, a.Run())
}

func Test_app_RunServerHooks(t *testing.T) {
	a := &app{hooks: map[hook][]Callback{}}
	a.RunServerHooks(afterDownHook)
}

func Test_app_Shutdown(t *testing.T) {
	called := false
	a := &app{hooks: map[hook][]Callback{}, logger: mlog.NewLogger(nil), doneFunc: func() {
		called = true
	}}
	a.Shutdown()
	assert.True(t, called)
}

func Test_app_Singleflight(t *testing.T) {
	a := &app{sf: &singleflight.Group{}}
	assert.NotNil(t, a.Singleflight())
}

func Test_app_Uploader(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	uploader := uploader2.NewMockUploader(m)
	a := &app{
		uploader: uploader,
	}
	assert.NotNil(t, a.Uploader())
}

func Test_app_WsServer(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ws := NewMockWsHttpServer(m)
	a := &app{
		ws: ws,
	}
	assert.NotNil(t, a.WsServer())
}

func Test_bootShortName(t *testing.T) {
	assert.Equal(t, "testBoot", bootShortName(&testBoot{}))
}

func Test_bootTags_has(t *testing.T) {
	assert.True(t, bootTags{"test"}.has("test"))
	assert.False(t, bootTags{"test"}.has("test1"))
}

func Test_excludeBootstrapperByTags(t *testing.T) {
	boots := []Bootstrapper{&testBoot{tags: []string{"test"}}, &testBoot{tags: []string{"test1"}}}
	b1, b2 := excludeBootstrapperByTags([]string{"test"}, boots)
	assert.Len(t, b1, 1)
	assert.Len(t, b2, 1)
	assert.Equal(t, "test1", b1[0].Tags()[0])
}

func Test_printConfig(t *testing.T) {
	//printConfig(nil)
}
