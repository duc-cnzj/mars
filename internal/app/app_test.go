package app

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/helm"

	"github.com/duc-cnzj/mars/v4/internal/app/bootstrappers"
	"github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/mock"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestApplication_AddServer(t *testing.T) {
	a := &application{}
	a.AddServer(nil)
	a.AddServer(nil)
	a.AddServer(nil)
	assert.Len(t, a.servers, 3)
}

func TestApplication_Auth(t *testing.T) {
	a := NewApplication(&config.Config{})
	m := gomock.NewController(t)
	defer m.Finish()
	authInterface := mock.NewMockAuthInterface(m)
	a.SetAuth(authInterface)
	assert.Same(t, authInterface, a.Auth())
}

func TestApplication_BeforeServerRunHooks(t *testing.T) {
	a := NewApplication(&config.Config{})
	called := false
	a.BeforeServerRunHooks(func(app contracts.ApplicationInterface) {
		called = true
	})
	assert.False(t, called)
	a.Run()
	assert.True(t, called)
}

type bootstrapper struct {
	err    error
	called bool
}

func (b *bootstrapper) Tags() []string {
	return nil
}
func (b *bootstrapper) Bootstrap(applicationInterface contracts.ApplicationInterface) error {
	b.called = true
	return b.err
}

func TestApplication_Bootstrap(t *testing.T) {
	b := &bootstrapper{called: false}
	a := NewApplication(&config.Config{}, WithBootstrappers(b))
	assert.False(t, b.called)
	a.Bootstrap()
	assert.True(t, b.called)

	ap := NewApplication(&config.Config{}, WithBootstrappers(&bootstrapper{err: errors.New("xxx")}))
	assert.Equal(t, "xxx", ap.Bootstrap().Error())
}

func TestApplication_Bootstrap1(t *testing.T) {
	e := errors.New("xxx")
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	l.EXPECT().Fatal(e).Times(1)
	NewApplication(&config.Config{}, WithMustBootedBootstrappers(&bootstrapper{err: e}))
}

func TestApplication_Cache(t *testing.T) {
	a := NewApplication(&config.Config{})
	m := gomock.NewController(t)
	defer m.Finish()
	c := mock.NewMockCacheInterface(m)
	a.SetCache(c)
	assert.IsType(t, (*cache.MetricsForCache)(nil), a.Cache())
	assert.Same(t, c, a.Cache().(*cache.MetricsForCache).Cache)

	a.SetCache(cache.NewMetricsForCache(c))
	assert.Same(t, c, a.Cache().(*cache.MetricsForCache).Cache)
}

func TestApplication_Config(t *testing.T) {
	cfg := &config.Config{}
	a := NewApplication(cfg)
	assert.Same(t, cfg, a.Config())
}

func TestApplication_DBManager(t *testing.T) {
	a := NewApplication(&config.Config{})
	assert.Implements(t, (*contracts.DBManager)(nil), a.DBManager())
}

func TestApplication_Done(t *testing.T) {
	a := NewApplication(&config.Config{})
	a.Shutdown()
	_, ok := <-a.Done()
	assert.False(t, ok)
}

func TestApplication_EventDispatcher(t *testing.T) {
	a := NewApplication(&config.Config{})
	m := gomock.NewController(t)
	defer m.Finish()
	e := mock.NewMockDispatcherInterface(m)
	a.SetEventDispatcher(e)
	assert.Same(t, e, a.EventDispatcher())
}

func TestApplication_GetPluginByName(t *testing.T) {
	a := NewApplication(&config.Config{})
	m := gomock.NewController(t)
	defer m.Finish()
	pluginInterface := mock.NewMockPluginInterface(m)
	a.SetPlugins(map[string]contracts.PluginInterface{
		"a": pluginInterface,
	})
	assert.Same(t, pluginInterface, a.GetPluginByName("a"))
}

func TestApplication_GetPlugins(t *testing.T) {
	a := NewApplication(&config.Config{})
	m := gomock.NewController(t)
	defer m.Finish()
	pluginInterface := mock.NewMockPluginInterface(m)
	a.SetPlugins(map[string]contracts.PluginInterface{
		"a": pluginInterface,
	})
	assert.Equal(t, map[string]contracts.PluginInterface{
		"a": pluginInterface,
	}, a.GetPlugins())
}

func TestApplication_IsDebug(t *testing.T) {
	a := NewApplication(&config.Config{Debug: false})
	assert.False(t, a.IsDebug())
	a = NewApplication(&config.Config{Debug: true})
	assert.True(t, a.IsDebug())
}

func TestApplication_K8sClient(t *testing.T) {
	a := NewApplication(&config.Config{})
	c := &contracts.K8sClient{}
	a.SetK8sClient(c)
	assert.Same(t, c, a.K8sClient())
}

func TestApplication_Oidc(t *testing.T) {
	a := NewApplication(&config.Config{})
	assert.IsType(t, contracts.OidcConfig{}, a.Oidc())
}

type testServer struct {
	runErr         error
	shutdownErr    error
	beforeShutdown func(*testServer)
	ran            bool
	shutdown       bool
}

func (t *testServer) Run(ctx context.Context) error {
	t.ran = true
	return t.runErr
}

func (t *testServer) Shutdown(ctx context.Context) error {
	if t.beforeShutdown != nil {
		t.beforeShutdown(t)
	}
	t.shutdown = true
	return t.shutdownErr
}

func TestApplication_RegisterAfterShutdownFunc(t *testing.T) {
	a := NewApplication(&config.Config{})
	called := false
	a.RegisterAfterShutdownFunc(func(app contracts.ApplicationInterface) {
		called = true
	})
	ts := &testServer{
		beforeShutdown: func(server *testServer) {
			assert.False(t, called)
		},
	}
	a.AddServer(ts)
	a.Shutdown()
	assert.True(t, called)
}

func TestApplication_RegisterBeforeShutdownFunc(t *testing.T) {
	a := NewApplication(&config.Config{})
	called := false
	a.RegisterBeforeShutdownFunc(func(app contracts.ApplicationInterface) {
		called = true
	})
	ts := &testServer{
		beforeShutdown: func(server *testServer) {
			assert.True(t, called)
		},
	}
	a.AddServer(ts)
	a.Shutdown()
	assert.True(t, called)
}

func TestApplication_Run(t *testing.T) {
	a := NewApplication(&config.Config{})
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	e := errors.New("xxx")
	l.EXPECT().Fatal(e).Times(1)
	ts := &testServer{runErr: e}
	a.AddServer(ts)
	a.Run()
	assert.True(t, ts.ran)
}

func TestApplication_RunServerHooks(t *testing.T) {
	called := false
	a := &application{
		hooks: map[hook][]contracts.Callback{
			"aaa": {func(app contracts.ApplicationInterface) {
				called = true
			}},
		},
	}
	a.RunServerHooks("aaa")
	assert.True(t, called)
}

func TestApplication_Uploader(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	a := NewApplication(&config.Config{})
	up := mock.NewMockUploader(m)
	a.SetUploader(up)
	assert.Same(t, up, a.Uploader().UnWrap())
}

func TestApplication_LocalUploader(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	a := NewApplication(&config.Config{})
	up := mock.NewMockUploader(m)
	a.SetLocalUploader(up)
	assert.Same(t, up, a.LocalUploader().UnWrap())
}

func TestApplication_Singleflight(t *testing.T) {
	a := NewApplication(&config.Config{})
	assert.Same(t, a.Singleflight(), a.Singleflight())
}

func TestNewApplication(t *testing.T) {
	assert.Implements(t, (*contracts.ApplicationInterface)(nil), NewApplication(&config.Config{}))
}

func TestApplication_SetOidc(t *testing.T) {
	a := NewApplication(&config.Config{})
	cfg := contracts.OidcConfig{
		"a": contracts.OidcConfigItem{},
		"b": contracts.OidcConfigItem{},
	}
	a.SetOidc(cfg)
	assert.Equal(t, cfg, a.Oidc())
}

func TestApplication_SetTracer(t *testing.T) {
	a := NewApplication(&config.Config{})
	m := gomock.NewController(t)
	defer m.Finish()
	tracer := mock.NewMockTracer(m)
	a.SetTracer(tracer)
	assert.Same(t, tracer, a.GetTracer())
}

func TestApplication_Shutdown(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	l := mock.NewMockLoggerInterface(m)
	a := NewApplication(&config.Config{})
	mlog.SetLogger(l)
	defer mlog.SetLogger(logrus.New())
	e := errors.New("xxx")
	a.AddServer(&testServer{shutdownErr: e})
	l.EXPECT().Info(gomock.Any()).Times(1)
	l.EXPECT().Warningf(gomock.Any(), gomock.Any()).Times(1)
	a.Shutdown()
}

func TestWithExcludeTags(t *testing.T) {
	tags := []string{"a", "b", "c"}
	assert.Equal(t, tags, NewApplication(&config.Config{}, WithExcludeTags(tags...)).(*application).excludeTags)
}

type boota struct{}

func (b *boota) Bootstrap(applicationInterface contracts.ApplicationInterface) error {
	return nil
}

func (b *boota) Tags() []string {
	return []string{"a", "aa", "aaa"}
}

type bootb struct{}

func (b *bootb) Bootstrap(applicationInterface contracts.ApplicationInterface) error {
	return nil
}

func (b *bootb) Tags() []string {
	return []string{"b", "bb", "bbb"}
}

type bootc struct{}

func (b *bootc) Bootstrap(applicationInterface contracts.ApplicationInterface) error {
	return nil
}

func (b *bootc) Tags() []string {
	return []string{"c", "cc", "ccc"}
}

func Test_excludeBootstrapperByTags(t *testing.T) {
	var cases = []struct {
		tags  []string
		boots []contracts.Bootstrapper
		wants []contracts.Bootstrapper
	}{
		{
			tags: []string{"a", "b"},
			boots: []contracts.Bootstrapper{
				&boota{},
				&bootb{},
				&bootc{},
			},
			wants: []contracts.Bootstrapper{
				&bootc{},
			},
		},
		{
			tags: []string{"c", "d"},
			boots: []contracts.Bootstrapper{
				&boota{},
				&bootb{},
				&bootc{},
			},
			wants: []contracts.Bootstrapper{
				&boota{},
				&bootb{},
			},
		},
		{
			tags: []string{},
			boots: []contracts.Bootstrapper{
				&boota{},
				&bootb{},
				&bootc{},
			},
			wants: []contracts.Bootstrapper{
				&boota{},
				&bootb{},
				&bootc{},
			},
		},
		{
			tags: []string{"a", "aa"},
			boots: []contracts.Bootstrapper{
				&boota{},
				&boota{},
				&boota{},
				&bootb{},
				&bootc{},
			},
			wants: []contracts.Bootstrapper{
				&bootb{},
				&bootc{},
			},
		},
	}
	for _, ca := range cases {
		res, _ := excludeBootstrapperByTags(ca.tags, ca.boots)
		assert.Equal(t, ca.wants, res)
	}
}

type CustomBoot struct{}

func (c CustomBoot) Bootstrap(applicationInterface contracts.ApplicationInterface) error {
	return nil
}

func (c CustomBoot) Tags() []string {
	return nil
}

func Test_bootShortName(t *testing.T) {
	assert.Empty(t, bootShortName(nil))
	assert.Equal(t, "EventBootstrapper", bootShortName(&bootstrappers.EventBootstrapper{}))
	assert.Equal(t, "CustomBoot", bootShortName(CustomBoot{}))
}

type cacheLock struct {
	contracts.Locker
}

func TestApplication_CacheLock(t *testing.T) {
	a := NewApplication(&config.Config{})
	assert.Nil(t, a.CacheLock())
	cl := &cacheLock{}
	a.SetCacheLock(cl)
	assert.NotNil(t, a.CacheLock())
}

type cm struct {
	contracts.CronManager
}

func TestApplication_CronManager(t *testing.T) {
	a := NewApplication(&config.Config{})
	assert.NotNil(t, a.CronManager())
	c := &cm{}
	a.SetCronManager(c)
	assert.Same(t, c, a.CronManager())
}

func TestApplication_DB(t *testing.T) {
	a := NewApplication(&config.Config{}).(*application)
	assert.Same(t, a.dbManager.DB(), a.DB())
}

func Test_printConfig(t *testing.T) {
	assert.NotPanics(t, func() {
		printConfig(&application{config: &config.Config{}, excludeBoots: []contracts.Bootstrapper{&boota{}}})
	})
}

func Test_application_lazyCache(t *testing.T) {
	c := cache.NewMetricsForCache(nil)
	assert.Same(t, c, (&application{cache: c}).lazyCache())
}

func Test_application_Helmer(t *testing.T) {
	a := NewApplication(&config.Config{}).(*application)
	assert.NotNil(t, a.helmer)
	assert.IsType(t, &helm.DefaultHelmer{}, a.helmer)
}
