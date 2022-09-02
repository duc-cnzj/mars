package contracts

//go:generate mockgen -destination ../mock/mock_app.go -package mock github.com/duc-cnzj/mars/internal/contracts ApplicationInterface
//go:generate mockgen -destination ../mock/mock_tracer.go -package mock go.opentelemetry.io/otel/trace Tracer

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	eventsv1 "k8s.io/api/events/v1"

	"github.com/coreos/go-oidc/v3/oidc"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/duc-cnzj/mars/internal/config"
)

type Callback func(ApplicationInterface)

type Server interface {
	Run(context.Context) error
	Shutdown(context.Context) error
}

type Bootstrapper interface {
	Bootstrap(ApplicationInterface) error
	Tags() []string
}

type FanOutInterface[T runtime.Object] interface {
	RemoveListener(key string)
	AddListener(key string, ch chan<- T)
	Distribute(done <-chan struct{})
}

type K8sClient struct {
	Client        kubernetes.Interface
	MetricsClient versioned.Interface
	RestConfig    *restclient.Config

	PodInformer cache.SharedIndexInformer
	PodLister   v1.PodLister

	EventInformer cache.SharedIndexInformer

	EventFanOut FanOutInterface[*eventsv1.Event]
	PodFanOut   FanOutInterface[*corev1.Pod]
}

type OidcConfigItem struct {
	Provider           *oidc.Provider
	Config             oauth2.Config
	EndSessionEndpoint string
}
type OidcConfig map[string]OidcConfigItem

type ApplicationInterface interface {
	IsDebug() bool

	K8sClient() *K8sClient
	SetK8sClient(*K8sClient)

	Auth() AuthInterface
	SetAuth(AuthInterface)

	SetLocalUploader(Uploader)
	LocalUploader() Uploader

	SetUploader(Uploader)
	Uploader() Uploader

	Bootstrap() error
	Config() *config.Config

	DBManager() DBManager
	DB() *gorm.DB

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

	CacheLock() Locker
	SetCacheLock(Locker)

	SetTracer(trace.Tracer)
	GetTracer() trace.Tracer

	SetCronManager(CronManager)
	CronManager() CronManager
}
