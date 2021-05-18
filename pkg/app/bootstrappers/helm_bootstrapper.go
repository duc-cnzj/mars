package bootstrappers

import (
	"os"

	"github.com/DuC-cnZj/mars/pkg/contracts"
	"github.com/DuC-cnZj/mars/pkg/mlog"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

type HelmBootstrapper struct{}

func (h *HelmBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	settings := cli.New()

	settings.KubeConfig = "/Users/duc/.kube/config"

	settings.Debug = app.IsDebug()

	actionConfig := new(action.Configuration)

	mlog.Infof("%#v", settings)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), mlog.Debugf); err != nil {
		return err
	}

	app.SetHelmConfig(&contracts.HelmClient{
		Setting: settings,
		Config:  actionConfig,
	})
	return nil
}
