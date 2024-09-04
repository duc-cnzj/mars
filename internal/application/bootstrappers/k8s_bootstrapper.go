package bootstrappers

import (
	"os"

	"github.com/duc-cnzj/mars/v5/internal/application"
)

type K8sBootstrapper struct{}

func (d *K8sBootstrapper) Tags() []string {
	return []string{}
}

func (d *K8sBootstrapper) Bootstrap(appli application.App) error {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if appli.Config().KubeConfig != "" || (host != "" && port != "") {
		return appli.Data().InitK8s(appli.Done())
	}
	return nil
}
