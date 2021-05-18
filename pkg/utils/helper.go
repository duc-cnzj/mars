package utils

import (
	"github.com/DuC-cnZj/mars/pkg/app/instance"
	"github.com/DuC-cnZj/mars/pkg/config"
	"github.com/DuC-cnZj/mars/pkg/contracts"
	"github.com/xanzy/go-gitlab"
	"gorm.io/gorm"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

func App() contracts.ApplicationInterface {
	return instance.App()
}

func Config() *config.Config {
	return App().Config()
}

func DB() *gorm.DB {
	return App().DBManager().DB()
}

func GitlabClient() *gitlab.Client {
	return App().GitlabClient()
}

func Event() contracts.DispatcherInterface {
	return App().EventDispatcher()
}

func K8sClient() *contracts.K8sClient {
	return App().K8sClient()
}

func K8s() *kubernetes.Clientset {
	return App().K8sClient().Client
}

func HelmClient() *contracts.HelmClient {
	return App().HelmConfig()
}

func HelmActionConfig() *action.Configuration {
	return App().HelmConfig().Config
}

func K8sMetrics() *versioned.Clientset {
	return App().K8sClient().MetricsClient
}
