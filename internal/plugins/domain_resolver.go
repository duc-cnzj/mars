package plugins

import (
	"sync"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
)

var domainResolverOnce = sync.Once{}

type DomainResolverInterface interface {
	contracts.PluginInterface

	// GetDomainByIndex domainSuffix: test.com, project: mars, namespace: default index: 0,1,2..., preOccupiedLen: 预占用的长度
	GetDomainByIndex(domainSuffix, projectName, namespace string, index, preOccupiedLen int) string

	// GetDomain domainSuffix: test.com, project: mars, namespace: production, preOccupiedLen: 预占用的长度
	GetDomain(domainSuffix, projectName, namespace string, preOccupiedLen int) string
}

func GetDomainResolverPlugin() DomainResolverInterface {
	p := app.App().GetPluginByName(app.App().Config().DomainResolverPlugin.Name)
	domainResolverOnce.Do(func() {
		if err := p.Initialize(); err != nil {
			panic(err)
		}
		app.App().RegisterAfterShutdownFunc(func(applicationInterface contracts.ApplicationInterface) {
			p.Destroy()
		})
	})

	return p.(DomainResolverInterface)
}
