package bootstrappers

import "github.com/duc-cnzj/mars/v4/internal/application"

type K8sBootstrapper struct{}

func (d *K8sBootstrapper) Tags() []string {
	return []string{}
}

func (d *K8sBootstrapper) Bootstrap(appli application.App) error {
	appli.Data().K8sClient.Start(appli.Done())
	return nil
}
