package bootstrappers

import (
	"github.com/DuC-cnZj/mars/pkg/contracts"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type K8sClientBootstrapper struct{}

func (i *K8sClientBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	kubeconfig := "/Users/duc/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
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
