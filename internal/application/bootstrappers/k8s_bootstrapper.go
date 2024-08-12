package bootstrappers

import "github.com/duc-cnzj/mars/v4/internal/application"

type K8sBootstrapper struct{}

func (d *K8sBootstrapper) Tags() []string {
	return []string{}
}

func (d *K8sBootstrapper) Bootstrap(appli application.App) error {
	if appli.Config().KubeConfig != "" {
		return appli.Data().InitK8s(appli.Done())
	}
	return nil
}