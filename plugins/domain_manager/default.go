package domain_manager

import (
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/plugins"
)

var _ plugins.DomainManager = (*DefaultDomainManager)(nil)

func init() {
	dr := &DefaultDomainManager{}
	plugins.RegisterPlugin(dr.Name(), dr)
}

type DefaultDomainManager struct{}

func (d *DefaultDomainManager) Name() string {
	return "default_domain_manager"
}

func (d *DefaultDomainManager) Initialize(args map[string]any) error {
	mlog.Info("[Plugin]: " + d.Name() + " plugin Initialize...")
	return nil
}

func (d *DefaultDomainManager) Destroy() error {
	mlog.Info("[Plugin]: " + d.Name() + " plugin Destroy...")
	return nil
}

func (d *DefaultDomainManager) GetDomainByIndex(projectName, namespace string, index, preOccupiedLen int) string {
	return Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        index,
		nsPrefix:     "devops",
		domainSuffix: "faker-domain.local",
	}.SubStr()
}

func (d *DefaultDomainManager) GetDomain(projectName, namespace string, preOccupiedLen int) string {
	return Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        -1,
		nsPrefix:     "devops",
		domainSuffix: "faker-domain.local",
	}.SubStr()
}

func (d *DefaultDomainManager) GetCertSecretName(projectName string, index int) string {
	return ""
}

func (d *DefaultDomainManager) GetClusterIssuer() string {
	return ""
}

func (d *DefaultDomainManager) GetCerts() (name, key, crt string) {
	return "", "", ""
}
