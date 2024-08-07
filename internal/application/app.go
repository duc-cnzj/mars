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

	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/config"
	mcron "github.com/duc-cnzj/mars/v4/internal/cron"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/event"
	"github.com/duc-cnzj/mars/v4/internal/locker"
	"github.com/duc-cnzj/mars/v4/internal/metrics"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/uploader"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
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
	mustBooted    []Bootstrapper
	excludeBoots  []Bootstrapper
	excludeTags   []string

	hooksMu sync.RWMutex
	hooks   map[hook][]Callback

	config        *config.Config
	logger        mlog.Logger
	uploader      uploader.Uploader
	localUploader uploader.Uploader
	auth          auth.Auth
	dispatcher    event.Dispatcher
	cronManager   mcron.Manager
	cache         cache.Cache
	cacheLock     locker.Locker
	tracer        trace.Tracer
	sf            *singleflight.Group
	data          data.Data
	pluginManager PluginManger
	reg           *GrpcRegistry
	ws            WsServer
}

func (app *app) WsServer() WsServer {
	return app.ws
}

type Option func(*app)

// WithBootstrappers custom boots.
func WithBootstrappers(bootstrappers ...Bootstrapper) Option {
	return func(app *app) {
		app.bootstrappers = bootstrappers
	}
}

// WithMustBootedBootstrappers set mustBooted.
func WithMustBootedBootstrappers(bootstrappers ...Bootstrapper) Option {
	return func(app *app) {
		app.mustBooted = bootstrappers
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
	//tracer trace.Tracer,
	sf *singleflight.Group,
	pm PluginManger,
	reg *GrpcRegistry,
	ws WsServer,
	opts ...Option,
) App {
	doneCtx, cancelFunc := context.WithCancel(context.TODO())
	appli := &app{
		done:          doneCtx,
		doneFunc:      cancelFunc,
		servers:       []Server{},
		excludeTags:   config.ExcludeServer.List(),
		hooksMu:       sync.RWMutex{},
		hooks:         map[hook][]Callback{},
		config:        config,
		logger:        logger,
		uploader:      uploader,
		localUploader: uploader.LocalUploader(),
		auth:          auth,
		dispatcher:    dispatcher,
		cronManager:   cronManager,
		cache:         cache,
		cacheLock:     cacheLock,
		tracer:        nil,
		//tracer:        tracer,
		sf:            sf,
		data:          data,
		pluginManager: pm,
		reg:           reg,
		ws:            ws,
	}

	for _, opt := range opts {
		opt(appli)
	}

	var mustBootExcludeBoots, excludeBoots []Bootstrapper
	if len(appli.excludeTags) > 0 {
		appli.mustBooted, mustBootExcludeBoots = excludeBootstrapperByTags(appli.excludeTags, appli.mustBooted)
		appli.bootstrappers, excludeBoots = excludeBootstrapperByTags(appli.excludeTags, appli.bootstrappers)
		appli.excludeBoots = append(appli.excludeBoots, mustBootExcludeBoots...)
		appli.excludeBoots = append(appli.excludeBoots, excludeBoots...)
	}

	for _, bootstrapper := range appli.mustBooted {
		func() {
			defer func(t time.Time) {
				metrics.BootstrapperStartMetrics.With(prometheus.Labels{"bootstrapper": bootShortName(bootstrapper)}).Set(time.Since(t).Seconds())
			}(time.Now())
			if err := bootstrapper.Bootstrap(appli); err != nil {
				appli.logger.Fatal(err)
			}
		}()
	}

	if appli.IsDebug() {
		printConfig(appli)
	}

	return appli
}

// Oidc impl App Oidc.
func (app *app) Oidc() data.OidcConfig {
	return app.data.OidcConfig()
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

// printConfig print config.
func printConfig(app *app) {
	app.logger.Debugf("imagepullsecrets %#v", app.Config().ImagePullSecrets)
	for _, boot := range app.excludeBoots {
		app.logger.Warningf("[BOOT]: '%s' (%s) doesn't start because of exclude tags: '%s'", bootShortName(boot), strings.Join(boot.Tags(), ","), strings.Join(app.excludeTags, ","))
	}
}

// Bootstrap impl App Bootstrap.
func (app *app) Bootstrap() error {
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

// Config impl App Config.
func (app *app) Config() *config.Config {
	return app.config
}

// GetTracer impl App GetTracer.
func (app *app) GetTracer() trace.Tracer {
	return app.tracer
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
		if err := server.Run(context.TODO()); err != nil {
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
			defer app.logger.HandlePanic("[Stop]: " + serverName)

			if err := server.Shutdown(ctx); err != nil {
				app.logger.Warningf("[Stop]: %s %s", serverName, err)
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
