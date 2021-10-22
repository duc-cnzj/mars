package contracts

import (
	"context"
	"os"

	restclient "k8s.io/client-go/rest"

	"github.com/xanzy/go-gitlab"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/duc-cnzj/mars/internal/config"
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

type Option func(ApplicationInterface)

type ApplicationInterface interface {
	IsDebug() bool

	GitlabClient() *gitlab.Client
	SetGitlabClient(*gitlab.Client)

	SetMetrics(Metrics)
	Metrics() Metrics

	K8sClient() *K8sClient
	SetK8sClient(*K8sClient)

	Bootstrap() error
	Config() *config.Config

	DBManager() DBManager

	AddServer(Server)
	Run() chan os.Signal
	Shutdown()

	Done() <-chan struct{}

	BeforeServerRunHooks(Callback)
	RegisterBeforeShutdownFunc(Callback)
	RegisterAfterShutdownFunc(Callback)

	EventDispatcher() DispatcherInterface
	SetEventDispatcher(DispatcherInterface)

	SetPlugins(map[string]PluginInterface)
	GetPlugins() map[string]PluginInterface
	GetPluginByName(name string) PluginInterface
}
