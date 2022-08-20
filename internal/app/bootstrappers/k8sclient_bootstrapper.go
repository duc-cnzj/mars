package bootstrappers

import (
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type K8sClientBootstrapper struct{}

func (k *K8sClientBootstrapper) Tags() []string {
	return []string{}
}

func (k *K8sClientBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	var (
		config *restclient.Config
		err    error
	)

	runtime.ErrorHandlers = []func(err error){
		func(err error) {
			mlog.Warning(err)
		},
	}

	if app.Config().KubeConfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", app.Config().KubeConfig)
		if err != nil {
			return err
		}
	} else {
		config, err = restclient.InClusterConfig()
		if err != nil {
			return err
		}
	}

	// 客户端不限速，有可能会把集群打死。
	config.QPS = -1

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	metrics, err := metricsv.NewForConfig(config)
	if err != nil {
		return err
	}

	app.SetK8sClient(&contracts.K8sClient{
		Client:        clientset,
		MetricsClient: metrics,
		RestConfig:    config,
	})

	return nil
}
