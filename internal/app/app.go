package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	app "github.com/duc-cnzj/mars/internal/app/helper"

	"github.com/duc-cnzj/mars/internal/app/bootstrappers"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/database"
	"github.com/duc-cnzj/mars/internal/mlog"
)

type Hook string

const (
	BeforeRunHook  Hook = "before_run"
	BeforeDownHook      = "before_down"
	AfterDownHook       = "after_down"
)

var _ contracts.ApplicationInterface = (*Application)(nil)

var DefaultBootstrappers = []contracts.Bootstrapper{
	&bootstrappers.PluginsBootstrapper{},
	&bootstrappers.AuthBootstrapper{},
	&bootstrappers.UploadBootstrapper{},
	&bootstrappers.K8sClientBootstrapper{},
	&bootstrappers.DBBootstrapper{},
	&bootstrappers.ApiGatewayBootstrapper{},
	&bootstrappers.PprofBootstrapper{},
	&bootstrappers.GrpcBootstrapper{},
	&bootstrappers.MetricsBootstrapper{},
	&bootstrappers.OidcBootstrapper{},
	&bootstrappers.TracingBootstrapper{},
}

type emptyMetrics struct{}

func (e *emptyMetrics) IncWebsocketConn() {
	return
}

func (e *emptyMetrics) DecWebsocketConn() {
	return
}

type Application struct {
	done          context.Context
	doneFunc      func()
	config        *config.Config
	clientSet     *contracts.K8sClient
	dbManager     contracts.DBManager
	dispatcher    contracts.DispatcherInterface
	metrics       contracts.Metrics
	servers       []contracts.Server
	bootstrappers []contracts.Bootstrapper
	hooks         map[Hook][]contracts.Callback
	plugins       map[string]contracts.PluginInterface
	oidcProvider  contracts.OidcConfig
	uploader      contracts.Uploader
	auth          contracts.AuthInterface
}

func (app *Application) Auth() contracts.AuthInterface {
	return app.auth
}

func (app *Application) SetAuth(auth contracts.AuthInterface) {
	app.auth = auth
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

func (app *Application) SetMetrics(metrics contracts.Metrics) {
	app.metrics = metrics
}

func (app *Application) Metrics() contracts.Metrics {
	return app.metrics
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

func (app *Application) SetEventDispatcher(dispatcher contracts.DispatcherInterface) {
	app.dispatcher = dispatcher
}

func NewApplication(config *config.Config, opts ...contracts.Option) contracts.ApplicationInterface {
	var mustBooted = []contracts.Bootstrapper{
		&bootstrappers.LogBootstrapper{},
		&bootstrappers.EventBootstrapper{},
	}

	doneCtx, cancelFunc := context.WithCancel(context.Background())
	app := &Application{
		bootstrappers: DefaultBootstrappers,
		config:        config,
		done:          doneCtx,
		doneFunc:      cancelFunc,
		hooks:         map[Hook][]contracts.Callback{},
		servers:       []contracts.Server{},
		metrics:       &emptyMetrics{},
	}

	app.dbManager = database.NewManager(app)

	for _, opt := range opts {
		opt(app)
	}

	instance.SetInstance(app)

	for _, bootstrapper := range mustBooted {
		if err := bootstrapper.Bootstrap(app); err != nil {
			mlog.Fatal(err)
		}
	}

	if app.IsDebug() {
		printConfig()
	}

	return app
}

func printConfig() {
	mlog.Debugf("imagepullsecrets %#v", app.App().Config().ImagePullSecrets)
}

func (app *Application) Bootstrap() error {
	for _, bootstrapper := range app.bootstrappers {
		if err := bootstrapper.Bootstrap(app); err != nil {
			return err
		}
	}

	return nil
}

func (app *Application) Config() *config.Config {
	return app.config
}

func (app *Application) DBManager() contracts.DBManager {
	return app.dbManager
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
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

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
				mlog.Error(err)
			}
		}(server)
	}
	wg.Wait()

	app.RunServerHooks(AfterDownHook)

	mlog.Info("server graceful shutdown.")
}

func (app *Application) RegisterAfterShutdownFunc(fn contracts.Callback) {
	app.hooks[AfterDownHook] = append(app.hooks[AfterDownHook], fn)
}

func (app *Application) RegisterBeforeShutdownFunc(fn contracts.Callback) {
	app.hooks[BeforeDownHook] = append(app.hooks[BeforeDownHook], fn)
}

func (app *Application) RunServerHooks(hook Hook) {
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
	app.hooks[BeforeRunHook] = append(app.hooks[BeforeRunHook], cb)
}
