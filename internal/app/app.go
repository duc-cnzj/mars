package app

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	app "github.com/duc-cnzj/mars/internal/app/helper"

	"github.com/xanzy/go-gitlab"

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
	&bootstrappers.K8sClientBootstrapper{},
	&bootstrappers.GitlabBootstrapper{},
	&bootstrappers.ValidatorBootstrapper{},
	&bootstrappers.I18nBootstrapper{},
	&bootstrappers.DBBootstrapper{},
	&bootstrappers.WebBootstrapper{},
	&bootstrappers.RouterBootstrapper{},
}

type Application struct {
	config *config.Config

	bootstrappers []contracts.Bootstrapper

	dbManager contracts.DBManager

	clientSet *contracts.K8sClient

	gitlabClient *gitlab.Client

	httpHandler http.Handler
	httpServer  *http.Server

	done     context.Context
	doneFunc func()

	hooks map[Hook][]contracts.Callback

	dispatcher contracts.DispatcherInterface

	plugins map[string]contracts.PluginInterface
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

func (app *Application) GitlabClient() *gitlab.Client {
	return app.gitlabClient
}

func (app *Application) SetGitlabClient(client *gitlab.Client) {
	app.gitlabClient = client
}

func (app *Application) K8sClient() *contracts.K8sClient {
	return app.clientSet
}

func (app *Application) SetK8sClient(client *contracts.K8sClient) {
	app.clientSet = client
}

func (app *Application) HttpHandler() http.Handler {
	return app.httpHandler
}

func (app *Application) SetHttpHandler(handler http.Handler) {
	app.httpHandler = handler
	app.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%s", app.Config().AppPort),
		Handler: handler,
	}
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
	mlog.Warningf("imagepullsecrets %#v", app.App().Config().ImagePullSecrets)
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

func (app *Application) Run() chan os.Signal {
	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	app.RunServerHooks(BeforeRunHook)

	go func() {
		mlog.Infof("server running at %s.", app.httpServer.Addr)
		if err := app.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			mlog.Fatal(err)
		}
	}()

	if app.Config().ProfileEnabled {
		go func() {
			mlog.Info("Starting pprof server on localhost:6060.")
			if err := http.ListenAndServe("localhost:6060", nil); err != nil && err != http.ErrServerClosed {
				mlog.Error(err)
			}
		}()
	}

	return done
}

func (app *Application) Shutdown() {
	var err error

	app.doneFunc()

	app.RunServerHooks(BeforeDownHook)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = app.httpServer.Shutdown(ctx)
	if err != nil {
		mlog.Error(err)
	}

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
	for _, cb := range app.hooks[hook] {
		cb(app)
	}
}

func (app *Application) BeforeServerRunHooks(cb contracts.Callback) {
	app.hooks[BeforeRunHook] = append(app.hooks[BeforeRunHook], cb)
}
