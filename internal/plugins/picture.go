package plugins

//go:generate mockgen -destination ../mock/mock_plugin_picture.go -package mock github.com/duc-cnzj/mars/v4/internal/plugins PictureInterface

import (
	"context"
	"sync"

	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
)

var pictureOnce = sync.Once{}

type PictureInterface interface {
	contracts.PluginInterface

	Get(ctx context.Context, random bool) (*contracts.Picture, error)
}

func GetPicture() PictureInterface {
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
