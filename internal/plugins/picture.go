package plugins

import (
	"sync"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
)

var pictureOnce = sync.Once{}

type Picture struct {
	Url       string
	Copyright string
}

type PictureInterface interface {
	contracts.PluginInterface

	Get(random bool) (*Picture, error)
}

func GetPicturePlugin() PictureInterface {
	pcfg := app.Config().PicturePlugin
	p := app.App().GetPluginByName(pcfg.Name)
	pictureOnce.Do(func() {
		if err := p.Initialize(pcfg.GetArgs()); err != nil {
			panic(err)
		}
		app.App().RegisterAfterShutdownFunc(func(applicationInterface contracts.ApplicationInterface) {
			p.Destroy()
		})
	})

	return p.(PictureInterface)
}
