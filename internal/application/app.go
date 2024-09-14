package application

import (
	"context"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/cache"
	"github.com/duc-cnzj/mars/v5/internal/config"
	mcron "github.com/duc-cnzj/mars/v5/internal/cron"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/ent"
	"github.com/duc-cnzj/mars/v5/internal/event"
	"github.com/duc-cnzj/mars/v5/internal/locker"
	"github.com/duc-cnzj/mars/v5/internal/metrics"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/uploader"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/singleflight"
)

type hook string

const (
	beforeRunHook  hook = "before_run"
	beforeDownHook hook = "before_down"
	afterDownHook  hook = "after_down"
)

var _ App = (*app)(nil)

type app struct {
	done context.Context

	doneFunc      func()
	servers       []Server
	bootstrappers []Bootstrapper
	excludeBoots  []Bootstrapper
	excludeTags   []string

	hooksMu sync.RWMutex
	hooks   map[hook][]Callback

	timer              timer.Timer
	config             *config.Config
	logger             mlog.Logger
	uploader           uploader.Uploader
	auth               auth.Auth
	dispatcher         event.Dispatcher
	cronManager        mcron.Manager
	cache              cache.Cache
	cacheLock          locker.Locker
	sf                 *singleflight.Group
	data               data.Data
	pluginManager      PluginManger
	reg                *GrpcRegistry
	prometheusRegistry *prometheus.Registry
	httpHandler        HttpHandler
}

type Option func(*app)

// WithBootstrappers custom boots.
func WithBootstrappers(bootstrappers ...Bootstrapper) Option {
	return func(app *app) {
		app.bootstrappers = bootstrappers
	}
}

// WithExcludeTags set excludeTags.
func WithExcludeTags(tags ...string) Option {
	return func(app *app) {
		app.excludeTags = tags
	}
}

// NewApp return App.
func NewApp(
	config *config.Config,
	data data.Data,
	logger mlog.Logger,
	uploader uploader.Uploader,
	auth auth.Auth,
	dispatcher event.Dispatcher,
	cronManager mcron.Manager,
	cache cache.Cache,
	cacheLock locker.Locker,
	sf *singleflight.Group,
	pm PluginManger,
	reg *GrpcRegistry,
	pr *prometheus.Registry,
	httpHandler HttpHandler,
	timer timer.Timer,
	opts ...Option,
) App {
	doneCtx, cancelFunc := context.WithCancel(context.TODO())
	appli := &app{
		timer:              timer,
		httpHandler:        httpHandler,
		done:               doneCtx,
		prometheusRegistry: pr,
		doneFunc:           cancelFunc,
		servers:            []Server{},
		excludeTags:        config.ExcludeServer.List(),
		hooksMu:            sync.RWMutex{},
		hooks:              map[hook][]Callback{},
		config:             config,
		logger:             logger.WithModule("app/app"),
		uploader:           uploader,
		auth:               auth,
		dispatcher:         dispatcher,
		cronManager:        cronManager,
		cache:              cache,
		cacheLock:          cacheLock,
		sf:                 sf,
		data:               data,
		pluginManager:      pm,
		reg:                reg,
	}

	for _, opt := range opts {
		opt(appli)
	}

	var excludeBoots []Bootstrapper
	if len(appli.excludeTags) > 0 {
		appli.bootstrappers, excludeBoots = excludeBootstrapperByTags(appli.excludeTags, appli.bootstrappers)
		appli.excludeBoots = append(appli.excludeBoots, excludeBoots...)
	}

	return appli
}

// Oidc impl App Oidc.
func (app *app) Oidc() data.OidcConfig {
	return app.data.OidcConfig()
}

func (app *app) PrometheusRegistry() *prometheus.Registry {
	return app.prometheusRegistry
}

// PluginMgr impl App PluginMgr.
func (app *app) PluginMgr() PluginManger {
	return app.pluginManager
}

// Data impl App Data.
func (app *app) Data() data.Data {
	return app.data
}

func (app *app) Logger() mlog.Logger {
	return app.logger
}

// Locker impl App CacheLock.
func (app *app) Locker() locker.Locker {
	return app.cacheLock
}

// Cache impl App Cache.
func (app *app) Cache() cache.Cache {
	return app.cache
}

func (app *app) HttpHandler() HttpHandler {
	return app.httpHandler
}

// Auth impl App Auth.
func (app *app) Auth() auth.Auth {
	return app.auth
}

// GrpcRegistry impl App GrpcRegistry.
func (app *app) GrpcRegistry() *GrpcRegistry {
	return app.reg
}

// CronManager impl App CronManager.
func (app *app) CronManager() mcron.Manager {
	return app.cronManager
}

// Uploader impl App Uploader.
func (app *app) Uploader() uploader.Uploader {
	return app.uploader
}

// Done impl App Done.
func (app *app) Done() <-chan struct{} {
	return app.done.Done()
}

// Dispatcher impl event.Dispatcher.
func (app *app) Dispatcher() event.Dispatcher {
	return app.dispatcher
}

// Singleflight impl App Singleflight.
func (app *app) Singleflight() *singleflight.Group {
	return app.sf
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
func excludeBootstrapperByTags(tags []string, boots []Bootstrapper) ([]Bootstrapper, []Bootstrapper) {
	var newBoots, excludeBoots []Bootstrapper
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

// Bootstrap impl App Bootstrap.
func (app *app) Bootstrap() error {
	for _, bootstrapper := range app.bootstrappers {
		err := func() error {
			defer func(t time.Time) {
				metrics.BootstrapperStartMetrics.With(prometheus.Labels{"bootstrapper": bootShortName(bootstrapper)}).Set(app.timer.Since(t).Seconds())
			}(app.timer.Now())
			return bootstrapper.Bootstrap(app)
		}()
		if err != nil {
			return err
		}
	}

	return nil
}

// Config impl App Config.
func (app *app) Config() *config.Config {
	return app.config
}

// DB impl App DB.
func (app *app) DB() *ent.Client {
	return app.data.DB()
}

// IsDebug impl App IsDebug.
func (app *app) IsDebug() bool {
	return app.config.Debug
}

// AddServer impl App AddServer.
func (app *app) AddServer(server Server) {
	app.servers = append(app.servers, server)
}

// Run impl App Run.
func (app *app) Run() context.Context {
	sig := make(chan os.Signal, 2)
	ch, cancel := context.WithCancel(context.TODO())
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	go func() {
		s1 := <-sig
		cancel()
		app.logger.Warningf("收到系统信号 %v, 再次执行 ctrl+c 强制退出!", s1.String())
		s2 := <-sig
		app.logger.Warningf("收到 %v 信号，执行强制退出!", s2.String())
		os.Exit(1)
	}()

	app.RunServerHooks(beforeRunHook)

	for _, server := range app.servers {
		if err := server.Run(app.done); err != nil {
			app.logger.Fatal(err)
		}
	}

	return ch
}

// Shutdown impl App Shutdown.
func (app *app) Shutdown() {
	app.doneFunc()
	app.RunServerHooks(beforeDownHook)

	wg := &sync.WaitGroup{}
	for _, server := range app.servers {
		wg.Add(1)
		go func(server Server) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()
			serverName := reflect.TypeOf(server).String()
			defer app.logger.HandlePanic("[Remove]: " + serverName)

			if err := server.Shutdown(ctx); err != nil {
				app.logger.Warningf("[Remove]: %s %s", serverName, err)
			}
		}(server)
	}
	wg.Wait()

	app.RunServerHooks(afterDownHook)

	app.logger.Info("server graceful shutdown.")
}

// RegisterAfterShutdownFunc impl App RegisterAfterShutdownFunc.
func (app *app) RegisterAfterShutdownFunc(fn Callback) {
	app.hooksMu.Lock()
	defer app.hooksMu.Unlock()
	app.hooks[afterDownHook] = append(app.hooks[afterDownHook], fn)
}

// RegisterBeforeShutdownFunc impl App RegisterBeforeShutdownFunc.
func (app *app) RegisterBeforeShutdownFunc(fn Callback) {
	app.hooksMu.Lock()
	defer app.hooksMu.Unlock()
	app.hooks[beforeDownHook] = append(app.hooks[beforeDownHook], fn)
}

// RunServerHooks impl App RunServerHooks.
func (app *app) RunServerHooks(hook hook) {
	app.hooksMu.RLock()
	defer app.hooksMu.RUnlock()
	wg := sync.WaitGroup{}
	for _, cb := range app.hooks[hook] {
		wg.Add(1)
		go func(cb Callback) {
			defer wg.Done()
			defer app.logger.HandlePanic("[RunServerHooks]: " + string(hook))
			cb(app)
		}(cb)
	}
	wg.Wait()
}

// BeforeServerRunHooks impl App BeforeServerRunHooks.
func (app *app) BeforeServerRunHooks(cb Callback) {
	app.hooksMu.Lock()
	defer app.hooksMu.Unlock()
	app.hooks[beforeRunHook] = append(app.hooks[beforeRunHook], cb)
}

// bootShortName get bootstrapper basename.
func bootShortName(boot Bootstrapper) string {
	if boot == nil {
		return ""
	}
	s := strings.Split(reflect.TypeOf(boot).String(), ".")
	return s[len(s)-1]
}
