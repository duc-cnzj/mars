package contracts

import (
	"net/http"
	"os"

	"github.com/xanzy/go-gitlab"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/DuC-cnZj/mars/pkg/config"
)

type ShutdownFunc func(ApplicationInterface)

type Bootstrapper interface {
	Bootstrap(ApplicationInterface) error
}

type K8sClient struct {
	Client        *kubernetes.Clientset
	MetricsClient *versioned.Clientset
}

type HelmClient struct {
	Setting *cli.EnvSettings
	Config  *action.Configuration
}

type Option func(ApplicationInterface)

type ApplicationInterface interface {
	IsDebug() bool

	GitlabClient() *gitlab.Client
	SetGitlabClient(*gitlab.Client)

	K8sClient() *K8sClient
	SetK8sClient(*K8sClient)

	HelmConfig() *HelmClient
	SetHelmConfig(*HelmClient)

	Bootstrap() error
	Config() *config.Config

	DBManager() DBManager

	Run() chan os.Signal
	Shutdown()

	RegisterBeforeShutdownFunc(ShutdownFunc)
	RegisterAfterShutdownFunc(ShutdownFunc)

	EventDispatcher() DispatcherInterface
	SetEventDispatcher(DispatcherInterface)

	HttpHandler() http.Handler
	SetHttpHandler(http.Handler)
}
