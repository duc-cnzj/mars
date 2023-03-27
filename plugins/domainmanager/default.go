package domainmanager

import (
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/plugins"
)

var _ plugins.DomainManager = (*defaultDomainManager)(nil)

func init() {
	dr := &defaultDomainManager{}
	plugins.RegisterPlugin(dr.Name(), dr)
}

type defaultDomainManager struct{}

func NewDefaultDomainManager() plugins.DomainManager {
	return &defaultDomainManager{}
}

func (d *defaultDomainManager) Name() string {
	return "default_domain_manager"
}

func (d *defaultDomainManager) Initialize(args map[string]any) error {
	mlog.Info("[Plugin]: " + d.Name() + " plugin Initialize...")
	return nil
}

func (d *defaultDomainManager) Destroy() error {
	mlog.Info("[Plugin]: " + d.Name() + " plugin Destroy...")
	return nil
}

func (d *defaultDomainManager) GetDomainByIndex(projectName, namespace string, index, preOccupiedLen int) string {
	return Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        index,
		nsPrefix:     "devops",
		domainSuffix: "faker-domain.local",
	}.SubStr()
}

func (d *defaultDomainManager) GetDomain(projectName, namespace string, preOccupiedLen int) string {
	return Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        -1,
		nsPrefix:     "devops",
		domainSuffix: "faker-domain.local",
	}.SubStr()
}

func (d *defaultDomainManager) GetCertSecretName(projectName string, index int) string {
	return ""
}

func (d *defaultDomainManager) GetClusterIssuer() string {
	return ""
}

func (d *defaultDomainManager) GetCerts() (name, key, crt string) {
	return "", "", ""
}
