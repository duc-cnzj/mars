package bootstrappers

import (
	"github.com/duc-cnzj/mars/internal/contracts"
	t "github.com/duc-cnzj/mars/internal/translator"
)

type I18nBootstrapper struct{}

func (i *I18nBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	t.Init()

	return nil
}
