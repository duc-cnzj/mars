package plugins

//go:generate mockgen -destination ../mock/mock_domain_manager.go -package mock github.com/duc-cnzj/mars/internal/plugins DomainManager

import (
	"sync"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
)

var domainManagerOnce = sync.Once{}

type DomainManager interface {
	contracts.PluginInterface

	// GetDomainByIndex domainSuffix: test.com, project: mars, namespace: default index: 0,1,2..., preOccupiedLen: 预占用的长度
	GetDomainByIndex(projectName, namespace string, index, preOccupiedLen int) string

	// GetDomain domainSuffix: test.com, project: mars, namespace: production, preOccupiedLen: 预占用的长度
	GetDomain(projectName, namespace string, preOccupiedLen int) string

	// GetCertSecretName 获取 HTTPS 证书对应的 secret
	GetCertSecretName(projectName string, index int) string

	// GetClusterIssuer CertManager 要用
	GetClusterIssuer() string

	// GetCerts 在 namespace 创建的时候注入证书信息
	GetCerts() (name, key, crt string)
}

func GetDomainManager() DomainManager {
	pcfg := app.Config().DomainManagerPlugin
	p := app.App().GetPluginByName(pcfg.Name)
	args := pcfg.GetArgs()

	domainManagerOnce.Do(func() {
		if err := p.Initialize(args); err != nil {
			panic(err)
		}
		app.App().RegisterAfterShutdownFunc(func(applicationInterface contracts.ApplicationInterface) {
			p.Destroy()
		})
	})

	return p.(DomainManager)
}
