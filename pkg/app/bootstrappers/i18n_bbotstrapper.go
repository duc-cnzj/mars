package bootstrappers

import (
	"github.com/duc-cnzj/mars/pkg/contracts"
	t "github.com/duc-cnzj/mars/pkg/translator"
)

type I18nBootstrapper struct{}

func (i *I18nBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	t.Init()

	return nil
}
