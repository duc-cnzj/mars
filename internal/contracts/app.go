package contracts

//go:generate mockgen -destination ../mock/mock_app.go -package mock github.com/duc-cnzj/mars/internal/contracts ApplicationInterface
//go:generate mockgen -destination ../mock/mock_metrics.go -package mock github.com/duc-cnzj/mars/internal/contracts Metrics

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/utils/singleflight"
)

type Callback func(ApplicationInterface)

type Metrics interface {
	IncWebsocketConn()
	DecWebsocketConn()
}

type Server interface {
	Run(context.Context) error
	Shutdown(context.Context) error
}

type Bootstrapper interface {
	Bootstrap(ApplicationInterface) error
}

type K8sClient struct {
	Client        kubernetes.Interface
	MetricsClient versioned.Interface
	RestConfig    *restclient.Config
}

type OidcConfigItem struct {
	Provider           *oidc.Provider
	Config             oauth2.Config
	EndSessionEndpoint string
}
type OidcConfig map[string]OidcConfigItem

type ApplicationInterface interface {
	IsDebug() bool

	SetMetrics(Metrics)
	Metrics() Metrics

	K8sClient() *K8sClient
	SetK8sClient(*K8sClient)

	Auth() AuthInterface
	SetAuth(AuthInterface)

	SetUploader(Uploader)
	Uploader() Uploader

	Bootstrap() error
	Config() *config.Config

	DBManager() DBManager

	Oidc() OidcConfig
	SetOidc(OidcConfig)

	AddServer(Server)
	Run() context.Context
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
}
