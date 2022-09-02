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

	"github.com/duc-cnzj/mars/internal/adapter"
	"github.com/duc-cnzj/mars/internal/app/bootstrappers"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/cache"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	mcron "github.com/duc-cnzj/mars/internal/cron"
	"github.com/duc-cnzj/mars/internal/database"
	"github.com/duc-cnzj/mars/internal/event"
	"github.com/duc-cnzj/mars/internal/metrics"
	"github.com/duc-cnzj/mars/internal/mlog"
)

type Hook string

const (
	BeforeRunHook  Hook = "before_run"
	BeforeDownHook Hook = "before_down"
	AfterDownHook  Hook = "after_down"
)

var _ contracts.ApplicationInterface = (*Application)(nil)

var MustBooted = []contracts.Bootstrapper{
	&bootstrappers.LogBootstrapper{},
}

type Application struct {
	done   context.Context
	config *config.Config

	doneFunc      func()
	servers       []contracts.Server
	bootstrappers []contracts.Bootstrapper
	mustBooted    []contracts.Bootstrapper
	excludeBoots  []contracts.Bootstrapper
	excludeTags   []string

	hooksMu sync.RWMutex
	hooks   map[Hook][]contracts.Callback

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

func (app *Application) CacheLock() contracts.Locker {
	return app.cacheLock
}

func (app *Application) SetCacheLock(l contracts.Locker) {
	app.cacheLock = l
}

func (app *Application) SetCache(c contracts.CacheInterface) {
	app.cache = c
}

func (app *Application) Cache() contracts.CacheInterface {
	if c, ok := app.cache.(*cache.MetricsForCache); ok {
		return c
	}

	return cache.NewMetricsForCache(app.cache)
}

func (app *Application) Auth() contracts.AuthInterface {
	return app.auth
}

func (app *Application) CronManager() contracts.CronManager {
	return app.cronManager
}
func (app *Application) SetCronManager(m contracts.CronManager) {
	app.cronManager = m
}

func (app *Application) SetAuth(auth contracts.AuthInterface) {
	app.auth = auth
}

func (app *Application) SetLocalUploader(uploader contracts.Uploader) {
	app.localUploader = uploader
}

func (app *Application) LocalUploader() contracts.Uploader {
	return app.localUploader
}
func (app *Application) SetUploader(uploader contracts.Uploader) {
	app.uploader = uploader
}

func (app *Application) Uploader() contracts.Uploader {
	return app.uploader
}

func (app *Application) Oidc() contracts.OidcConfig {
	return app.oidcProvider
}

func (app *Application) SetOidc(provider contracts.OidcConfig) {
	app.oidcProvider = provider
}

func (app *Application) GetPluginByName(name string) contracts.PluginInterface {
	return app.plugins[name]
}

func (app *Application) SetPlugins(plugins map[string]contracts.PluginInterface) {
	app.plugins = plugins
}

func (app *Application) GetPlugins() map[string]contracts.PluginInterface {
	return app.plugins
}

func (app *Application) Done() <-chan struct{} {
	return app.done.Done()
}

func (app *Application) K8sClient() *contracts.K8sClient {
	return app.clientSet
}

func (app *Application) SetK8sClient(client *contracts.K8sClient) {
	app.clientSet = client
}

func (app *Application) EventDispatcher() contracts.DispatcherInterface {
	return app.dispatcher
}

func (app *Application) Singleflight() *singleflight.Group {
	return app.sf
}

func (app *Application) SetEventDispatcher(dispatcher contracts.DispatcherInterface) {
	app.dispatcher = dispatcher
}

type Option func(*Application)

func WithBootstrappers(bootstrappers ...contracts.Bootstrapper) Option {
	return func(app *Application) {
		app.bootstrappers = bootstrappers
	}
}

func WithMustBootedBootstrappers(bootstrappers ...contracts.Bootstrapper) Option {
	return func(app *Application) {
		app.mustBooted = bootstrappers
	}
}

func WithExcludeTags(tags ...string) Option {
	return func(app *Application) {
		app.excludeTags = tags
	}
}

func NewApplication(config *config.Config, opts ...Option) contracts.ApplicationInterface {
	doneCtx, cancelFunc := context.WithCancel(context.Background())
	app := &Application{
		mustBooted:  MustBooted,
		config:      config,
		done:        doneCtx,
		doneFunc:    cancelFunc,
		hooks:       map[Hook][]contracts.Callback{},
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

type bootTags []string

func (bt bootTags) Has(tag string) bool {
	for _, t := range bt {
		if t == tag {
			return true
		}
	}
	return false
}

func excludeBootstrapperByTags(tags []string, boots []contracts.Bootstrapper) ([]contracts.Bootstrapper, []contracts.Bootstrapper) {
	var newBoots, excludeBoots []contracts.Bootstrapper
loop:
	for _, boot := range boots {
		for _, tag := range tags {
			if bootTags(boot.Tags()).Has(tag) {
				excludeBoots = append(excludeBoots, boot)
				continue loop
			}
		}

		newBoots = append(newBoots, boot)
	}
	return newBoots, excludeBoots
}

func printConfig(app *Application) {
	mlog.Debugf("imagepullsecrets %#v", app.Config().ImagePullSecrets)
	for _, boot := range app.excludeBoots {
		mlog.Warningf("[BOOT]: '%s' (%s) doesn't start because of exclude tags: '%s'", bootShortName(boot), strings.Join(boot.Tags(), ","), strings.Join(app.excludeTags, ","))
	}
}

func (app *Application) Bootstrap() error {
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

func (app *Application) Config() *config.Config {
	return app.config
}

func (app *Application) SetTracer(t trace.Tracer) {
	app.tracer = t
}

func (app *Application) GetTracer() trace.Tracer {
	return app.tracer
}

func (app *Application) DBManager() contracts.DBManager {
	return app.dbManager
}

func (app *Application) DB() *gorm.DB {
	return app.dbManager.DB()
}

func (app *Application) IsDebug() bool {
	return app.config.Debug
}

func (app *Application) AddServer(server contracts.Server) {
	app.servers = append(app.servers, server)
}

func (app *Application) Run() context.Context {
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

	app.RunServerHooks(BeforeRunHook)

	for _, server := range app.servers {
		if err := server.Run(context.Background()); err != nil {
			mlog.Fatal(err)
		}
	}

	return ch
}

func (app *Application) Shutdown() {
	app.doneFunc()
	app.RunServerHooks(BeforeDownHook)

	wg := &sync.WaitGroup{}
	for _, server := range app.servers {
		wg.Add(1)
		go func(server contracts.Server) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				mlog.Warningf("[Shutdown]: %s %s", reflect.TypeOf(server).String(), err)
			}
		}(server)
	}
	wg.Wait()

	app.RunServerHooks(AfterDownHook)

	mlog.Info("server graceful shutdown.")
}

func (app *Application) RegisterAfterShutdownFunc(fn contracts.Callback) {
	app.hooksMu.Lock()
	defer app.hooksMu.Unlock()
	app.hooks[AfterDownHook] = append(app.hooks[AfterDownHook], fn)
}

func (app *Application) RegisterBeforeShutdownFunc(fn contracts.Callback) {
	app.hooksMu.Lock()
	defer app.hooksMu.Unlock()
	app.hooks[BeforeDownHook] = append(app.hooks[BeforeDownHook], fn)
}

func (app *Application) RunServerHooks(hook Hook) {
	app.hooksMu.RLock()
	defer app.hooksMu.RUnlock()
	wg := sync.WaitGroup{}
	for _, cb := range app.hooks[hook] {
		wg.Add(1)
		go func(cb contracts.Callback) {
			defer wg.Done()
			cb(app)
		}(cb)
	}
	wg.Wait()
}

func (app *Application) BeforeServerRunHooks(cb contracts.Callback) {
	app.hooksMu.Lock()
	defer app.hooksMu.Unlock()
	app.hooks[BeforeRunHook] = append(app.hooks[BeforeRunHook], cb)
}

func bootShortName(boot contracts.Bootstrapper) string {
	if boot == nil {
		return ""
	}
	s := strings.Split(reflect.TypeOf(boot).String(), ".")
	return s[len(s)-1]
}
