package plugins

import (
	"sync"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
)

var domainResolverOnce = sync.Once{}

type DomainResolver interface {
	// GetDomainByIndex domainSuffix: test.com, project: mars, namespace: default index: 0,1,2..., preOccupiedLen: 预占用的长度
	GetDomainByIndex(domainSuffix, projectName, namespace string, index, preOccupiedLen int) string

	// GetDomain domainSuffix: test.com, project: mars, namespace: production, preOccupiedLen: 预占用的长度
	GetDomain(domainSuffix, projectName, namespace string, preOccupiedLen int) string
}

func GetDomainResolver() DomainResolver {
	pcfg := app.Config().DomainResolverPlugin
	p := app.App().GetPluginByName(pcfg.Name)
	args := pcfg.GetArgs()
	args["ns_prefix"] = app.Config().NsPrefix

	domainResolverOnce.Do(func() {
		if err := p.Initialize(args); err != nil {
			panic(err)
		}
		app.App().RegisterAfterShutdownFunc(func(applicationInterface contracts.ApplicationInterface) {
			p.Destroy()
		})
	})

	return p.(DomainResolver)
}
