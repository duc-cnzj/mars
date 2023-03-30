package contracts

//go:generate mockgen -destination ../mock/mock_app.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts ApplicationInterface
//go:generate mockgen -destination ../mock/mock_tracer.go -package mock go.opentelemetry.io/otel/trace Tracer

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type Callback func(ApplicationInterface)

// Server define booting server.
type Server interface {
	// Run server.
	Run(context.Context) error
	// Shutdown server.
	Shutdown(context.Context) error
}

// Bootstrapper boots.
type Bootstrapper interface {
	// Bootstrap when app start.
	Bootstrap(ApplicationInterface) error
	// Tags boot tags.
	Tags() []string
}

type OidcConfigItem struct {
	Provider           *oidc.Provider
	Config             oauth2.Config
	EndSessionEndpoint string
}

type OidcConfig map[string]OidcConfigItem

// ApplicationInterface app.
type ApplicationInterface interface {
	// IsDebug bool.
	IsDebug() bool

	// K8sClient return *K8sClient
	K8sClient() *K8sClient
	// SetK8sClient set *K8sClient
	SetK8sClient(*K8sClient)

	// Auth return AuthInterface.
	Auth() AuthInterface
	// SetAuth set AuthInterface.
	SetAuth(AuthInterface)

	// SetLocalUploader set local Uploader
	SetLocalUploader(Uploader)
	// LocalUploader get local Uploader
	LocalUploader() Uploader

	// SetUploader setter
	SetUploader(Uploader)
	// Uploader getter
	Uploader() Uploader

	// Bootstrap boots all.
	Bootstrap() error
	// Config app configuration.
	Config() *config.Config

	// DBManager db.
	DBManager() DBManager
	// DB instance.
	DB() *gorm.DB

	// Oidc sso cfg
	Oidc() OidcConfig
	// SetOidc setter
	SetOidc(OidcConfig)

	// AddServer add boot server
	AddServer(Server)
	// Run servers.
	Run() context.Context
	// Shutdown all servers.
	Shutdown()

	Done() <-chan struct{}

	BeforeServerRunHooks(Callback)
	RegisterBeforeShutdownFunc(Callback)
	RegisterAfterShutdownFunc(Callback)

	EventDispatcher() DispatcherInterface
	SetEventDispatcher(DispatcherInterface)

	SetPlugins(map[string]PluginInterface)
	GetPlugins() map[string]PluginInterface
	GetPluginByName(string) PluginInterface

	Singleflight() *singleflight.Group

	SetCache(CacheInterface)
	Cache() CacheInterface

	CacheLock() Locker
	SetCacheLock(Locker)

	SetTracer(trace.Tracer)
	GetTracer() trace.Tracer

	SetCronManager(CronManager)
	CronManager() CronManager
}
