package bootstrappers

import (
	"github.com/duc-cnzj/mars/internal/contracts"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type K8sClientBootstrapper struct{}

func (i *K8sClientBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	var (
		config *restclient.Config
		err    error
	)

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
