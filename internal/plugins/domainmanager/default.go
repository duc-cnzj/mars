package domainmanager

import (
	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
)

var _ application.DomainManager = (*defaultDomainManager)(nil)

func init() {
	dr := &defaultDomainManager{}
	application.RegisterPlugin(dr.Name(), dr)
}

type defaultDomainManager struct {
	logger mlog.Logger
}

func NewDefaultDomainManager() application.DomainManager {
	return &defaultDomainManager{}
}

func (d *defaultDomainManager) Name() string {
	return "default_domain_manager"
}

func (d *defaultDomainManager) Initialize(app application.App, args map[string]any) error {
	d.logger = app.Logger()
	d.logger.Info("[Plugin]: " + d.Name() + " plugin Initialize...")
	return nil
}

func (d *defaultDomainManager) Destroy() error {
	d.logger.Info("[Plugin]: " + d.Name() + " plugin Destroy...")
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
