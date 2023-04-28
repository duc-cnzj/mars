package app

import (
	"context"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"

	"github.com/duc-cnzj/mars/v4/internal/adapter"
	"github.com/duc-cnzj/mars/v4/internal/app/bootstrappers"
	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	mcron "github.com/duc-cnzj/mars/v4/internal/cron"
	"github.com/duc-cnzj/mars/v4/internal/database"
	"github.com/duc-cnzj/mars/v4/internal/event"
	"github.com/duc-cnzj/mars/v4/internal/metrics"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	muploader "github.com/duc-cnzj/mars/v4/internal/uploader"
	"github.com/duc-cnzj/mars/v4/internal/utils/recovery"
)

type hook string

const (
	beforeRunHook  hook = "before_run"
	beforeDownHook hook = "before_down"
	afterDownHook  hook = "after_down"
)

var _ contracts.ApplicationInterface = (*application)(nil)

var mustBooted = []contracts.Bootstrapper{
	&bootstrappers.LogBootstrapper{},
}

type application struct {
	done   context.Context
	config *config.Config

	doneFunc      func()
	servers       []contracts.Server
	bootstrappers []contracts.Bootstrapper
	mustBooted    []contracts.Bootstrapper
	excludeBoots  []contracts.Bootstrapper
	excludeTags   []string

	hooksMu sync.RWMutex
	hooks   map[hook][]contracts.Callback

	uploader      contracts.Uploader
	localUploader contracts.Uploader
	plugins       map[string]contracts.PluginInterface
	oidcProvider  contracts.OidcConfig
	auth          contracts.AuthInterface
	clientSet     *contracts.K8sClient
	dbManager     contracts.DBManager
	dispatcher    contracts.DispatcherInterface
	cronManager   contracts.CronManager
	cache         contracts.CacheInterface
	cacheLock     contracts.Locker
	tracer        trace.Tracer
	sf            *singleflight.Group
}

type Option func(*application)

// WithBootstrappers custom boots.
func WithBootstrappers(bootstrappers ...contracts.Bootstrapper) Option {
	return func(app *application) {
		app.bootstrappers = bootstrappers
	}
}

// WithMustBootedBootstrappers set mustBooted.
func WithMustBootedBootstrappers(bootstrappers ...contracts.Bootstrapper) Option {
	return func(app *application) {
		app.mustBooted = bootstrappers
	}
}

// WithExcludeTags set excludeTags.
func WithExcludeTags(tags ...string) Option {
	return func(app *application) {
		app.excludeTags = tags
	}
}

// NewApplication return contracts.ApplicationInterface.
func NewApplication(config *config.Config, opts ...Option) contracts.ApplicationInterface {
	doneCtx, cancelFunc := context.WithCancel(context.Background())
	app := &application{
		mustBooted:  mustBooted,
		config:      config,
		done:        doneCtx,
		doneFunc:    cancelFunc,
		hooks:       map[hook][]contracts.Callback{},
		servers:     []contracts.Server{},
		sf:          &singleflight.Group{},
		cache:       &cache.NoCache{},
		excludeTags: config.ExcludeServer.List(),
	}

	app.cronManager = mcron.NewManager(adapter.NewRobfigCronV3Runner(), app)
	app.dispatcher = event.NewDispatcher(app)
	app.dbManager = database.NewManager(app)

	for _, opt := range opts {
		opt(app)
	}

	var mustBootExcludeBoots, excludeBoots []contracts.Bootstrapper
	if len(app.excludeTags) > 0 {
		app.mustBooted, mustBootExcludeBoots = excludeBootstrapperByTags(app.excludeTags, app.mustBooted)
		app.bootstrappers, excludeBoots = excludeBootstrapperByTags(app.excludeTags, app.bootstrappers)
		app.excludeBoots = append(app.excludeBoots, mustBootExcludeBoots...)
		app.excludeBoots = append(app.excludeBoots, excludeBoots...)
	}

	instance.SetInstance(app)

	for _, bootstrapper := range app.mustBooted {
		func() {
			defer func(t time.Time) {
				metrics.BootstrapperStartMetrics.With(prometheus.Labels{"bootstrapper": bootShortName(bootstrapper)}).Set(time.Since(t).Seconds())
			}(time.Now())
			if err := bootstrapper.Bootstrap(app); err != nil {
				mlog.Fatal(err)
			}
		}()
	}

	if app.IsDebug() {
		printConfig(app)
	}

	return app
}

// CacheLock impl contracts.ApplicationInterface CacheLock.
func (app *application) CacheLock() contracts.Locker {
	return app.cacheLock
}

// SetCacheLock impl contracts.ApplicationInterface SetCacheLock.
func (app *application) SetCacheLock(l contracts.Locker) {
	app.cacheLock = l
}

// SetCache impl contracts.ApplicationInterface SetCache.
func (app *application) SetCache(c contracts.CacheInterface) {
	app.cache = c
}

// Cache impl contracts.ApplicationInterface Cache.
func (app *application) Cache() contracts.CacheInterface {
	if c, ok := app.cache.(*cache.MetricsForCache); ok {
		return c
	}

	return cache.NewMetricsForCache(app.cache)
}

// Auth impl contracts.ApplicationInterface Auth.
func (app *application) Auth() contracts.AuthInterface {
	return app.auth
}

// CronManager impl contracts.ApplicationInterface CronManager.
func (app *application) CronManager() contracts.CronManager {
	return app.cronManager
}

// SetCronManager impl contracts.ApplicationInterface SetCronManager.
func (app *application) SetCronManager(m contracts.CronManager) {
	app.cronManager = m
}

// SetAuth impl contracts.ApplicationInterface SetAuth.
func (app *application) SetAuth(auth contracts.AuthInterface) {
	app.auth = auth
}

// SetLocalUploader impl contracts.ApplicationInterface SetLocalUploader.
func (app *application) SetLocalUploader(uploader contracts.Uploader) {
	app.localUploader = muploader.NewCacheUploader(uploader, app.Cache())
}

// LocalUploader impl contracts.ApplicationInterface LocalUploader.
func (app *application) LocalUploader() contracts.Uploader {
	return app.localUploader
}

// SetUploader impl contracts.ApplicationInterface SetUploader.
func (app *application) SetUploader(uploader contracts.Uploader) {
	app.uploader = muploader.NewCacheUploader(uploader, app.Cache())
}

// Uploader impl contracts.ApplicationInterface Uploader.
func (app *application) Uploader() contracts.Uploader {
	return app.uploader
}

// Oidc impl contracts.ApplicationInterface Oidc.
func (app *application) Oidc() contracts.OidcConfig {
	return app.oidcProvider
}

// SetOidc impl contracts.ApplicationInterface SetOidc.
func (app *application) SetOidc(provider contracts.OidcConfig) {
	app.oidcProvider = provider
}

// GetPluginByName impl contracts.ApplicationInterface GetPluginByName.
func (app *application) GetPluginByName(name string) contracts.PluginInterface {
	return app.plugins[name]
}

// SetPlugins impl contracts.ApplicationInterface SetPlugins.
func (app *application) SetPlugins(plugins map[string]contracts.PluginInterface) {
	app.plugins = plugins
}

// GetPlugins impl contracts.ApplicationInterface GetPlugins.
func (app *application) GetPlugins() map[string]contracts.PluginInterface {
	return app.plugins
}

// Done impl contracts.ApplicationInterface Done.
func (app *application) Done() <-chan struct{} {
	return app.done.Done()
}

// K8sClient impl contracts.ApplicationInterface K8sClient.
func (app *application) K8sClient() *contracts.K8sClient {
	return app.clientSet
}

// SetK8sClient impl contracts.ApplicationInterface SetK8sClient.
func (app *application) SetK8sClient(client *contracts.K8sClient) {
	app.clientSet = client
}

// EventDispatcher impl contracts.ApplicationInterface EventDispatcher.
func (app *application) EventDispatcher() contracts.DispatcherInterface {
	return app.dispatcher
}

// Singleflight impl contracts.ApplicationInterface Singleflight.
func (app *application) Singleflight() *singleflight.Group {
	return app.sf
}

// SetEventDispatcher impl contracts.ApplicationInterface SetEventDispatcher.
func (app *application) SetEventDispatcher(dispatcher contracts.DispatcherInterface) {
	app.dispatcher = dispatcher
}

type bootTags []string

func (bt bootTags) has(tag string) bool {
	for _, t := range bt {
		if t == tag {
			return true
		}
	}
	return false
}

// excludeBootstrapperByTags exclude tags.
func excludeBootstrapperByTags(tags []string, boots []contracts.Bootstrapper) ([]contracts.Bootstrapper, []contracts.Bootstrapper) {
	var newBoots, excludeBoots []contracts.Bootstrapper
loop:
	for _, boot := range boots {
		for _, tag := range tags {
			if bootTags(boot.Tags()).has(tag) {
				excludeBoots = append(excludeBoots, boot)
				continue loop
			}
		}

		newBoots = append(newBoots, boot)
	}
	return newBoots, excludeBoots
}

// { impl contracts.ApplicationInterface {.
func printConfig(app *application) {
	mlog.Debugf("imagepullsecrets %#v", app.Config().ImagePullSecrets)
	for _, boot := range app.excludeBoots {
		mlog.Warningf("[BOOT]: '%s' (%s) doesn't start because of exclude tags: '%s'", bootShortName(boot), strings.Join(boot.Tags(), ","), strings.Join(app.excludeTags, ","))
	}
}

// Bootstrap impl contracts.ApplicationInterface Bootstrap.
func (app *application) Bootstrap() error {
	for _, bootstrapper := range app.bootstrappers {
		err := func() error {
			defer func(t time.Time) {
				metrics.BootstrapperStartMetrics.With(prometheus.Labels{"bootstrapper": bootShortName(bootstrapper)}).Set(time.Since(t).Seconds())
			}(time.Now())
			return bootstrapper.Bootstrap(app)
		}()
		if err != nil {
			return err
		}
	}

	return nil
}

// Config impl contracts.ApplicationInterface Config.
func (app *application) Config() *config.Config {
	return app.config
}

// SetTracer impl contracts.ApplicationInterface SetTracer.
func (app *application) SetTracer(t trace.Tracer) {
	app.tracer = t
}

// GetTracer impl contracts.ApplicationInterface GetTracer.
func (app *application) GetTracer() trace.Tracer {
	return app.tracer
}

// DBManager impl contracts.ApplicationInterface DBManager.
func (app *application) DBManager() contracts.DBManager {
	return app.dbManager
}

// DB impl contracts.ApplicationInterface DB.
func (app *application) DB() *gorm.DB {
	return app.dbManager.DB()
}

// IsDebug impl contracts.ApplicationInterface IsDebug.
func (app *application) IsDebug() bool {
	return app.config.Debug
}

// AddServer impl contracts.ApplicationInterface AddServer.
func (app *application) AddServer(server contracts.Server) {
	app.servers = append(app.servers, server)
}

// Run impl contracts.ApplicationInterface Run.
func (app *application) Run() context.Context {
	sig := make(chan os.Signal, 2)
	ch, cancel := context.WithCancel(context.TODO())
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	go func() {
		s1 := <-sig
		cancel()
		mlog.Warningf("收到系统信号 %v, 再次执行 ctrl+c 强制退出!", s1.String())
		s2 := <-sig
		mlog.Warningf("收到 %v 信号，执行强制退出!", s2.String())
		os.Exit(1)
	}()

	app.RunServerHooks(beforeRunHook)

	for _, server := range app.servers {
		if err := server.Run(context.Background()); err != nil {
			mlog.Fatal(err)
		}
	}

	return ch
}

// Shutdown impl contracts.ApplicationInterface Shutdown.
func (app *application) Shutdown() {
	app.doneFunc()
	app.RunServerHooks(beforeDownHook)

	wg := &sync.WaitGroup{}
	for _, server := range app.servers {
		wg.Add(1)
		go func(server contracts.Server) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()
			serverName := reflect.TypeOf(server).String()
			defer recovery.HandlePanic("[Shutdown]: " + serverName)

			if err := server.Shutdown(ctx); err != nil {
				mlog.Warningf("[Shutdown]: %s %s", serverName, err)
			}
		}(server)
	}
	wg.Wait()

	app.RunServerHooks(afterDownHook)

	mlog.Info("server graceful shutdown.")
}

// RegisterAfterShutdownFunc impl contracts.ApplicationInterface RegisterAfterShutdownFunc.
func (app *application) RegisterAfterShutdownFunc(fn contracts.Callback) {
	app.hooksMu.Lock()
	defer app.hooksMu.Unlock()
	app.hooks[afterDownHook] = append(app.hooks[afterDownHook], fn)
}

// RegisterBeforeShutdownFunc impl contracts.ApplicationInterface RegisterBeforeShutdownFunc.
func (app *application) RegisterBeforeShutdownFunc(fn contracts.Callback) {
	app.hooksMu.Lock()
	defer app.hooksMu.Unlock()
	app.hooks[beforeDownHook] = append(app.hooks[beforeDownHook], fn)
}

// RunServerHooks impl contracts.ApplicationInterface RunServerHooks.
func (app *application) RunServerHooks(hook hook) {
	app.hooksMu.RLock()
	defer app.hooksMu.RUnlock()
	wg := sync.WaitGroup{}
	for _, cb := range app.hooks[hook] {
		wg.Add(1)
		go func(cb contracts.Callback) {
			defer wg.Done()
			defer recovery.HandlePanic("[RunServerHooks]: " + string(hook))
			cb(app)
		}(cb)
	}
	wg.Wait()
}

// BeforeServerRunHooks impl contracts.ApplicationInterface BeforeServerRunHooks.
func (app *application) BeforeServerRunHooks(cb contracts.Callback) {
	app.hooksMu.Lock()
	defer app.hooksMu.Unlock()
	app.hooks[beforeRunHook] = append(app.hooks[beforeRunHook], cb)
}

// bootShortName get bootstrapper basename.
func bootShortName(boot contracts.Bootstrapper) string {
	if boot == nil {
		return ""
	}
	s := strings.Split(reflect.TypeOf(boot).String(), ".")
	return s[len(s)-1]
}
