package domainmanager

import (
	"errors"
	"strings"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/plugins"
)

const ManualCertSecretName = "mars-external-tls-secret"

var _ plugins.DomainManager = (*manualDomainManager)(nil)

func init() {
	dr := &manualDomainManager{}
	plugins.RegisterPlugin(dr.Name(), dr)
}

type manualDomainManager struct {
	nsPrefix       string
	wildcardDomain string
	domainSuffix   string

	tlsCrt string
	tlsKey string
}

func (m *manualDomainManager) Name() string {
	return "manual_domain_manager"
}

func (m *manualDomainManager) Initialize(args map[string]any) error {
	if p, ok := args["ns_prefix"]; ok {
		m.nsPrefix = p.(string)
	}

	if p, ok := args["tls_crt"]; ok {
		m.tlsCrt = p.(string)
	}

	if p, ok := args["tls_key"]; ok {
		m.tlsKey = p.(string)
	}
	if wd, ok := args["wildcard_domain"]; ok {
		m.wildcardDomain = wd.(string)
		m.domainSuffix = strings.TrimLeft(m.wildcardDomain, "*.")
	}
	if m.tlsKey == "" || m.tlsCrt == "" || m.wildcardDomain == "" {
		return errors.New("tls_crt, tls_key, wildcard_domain required")
	}
	if err := validateTelsWildcardDomain(m.tlsKey, m.tlsCrt, m.wildcardDomain); err != nil {
		return err
	}
	mlog.Info("[Plugin]: " + m.Name() + " plugin Initialize...")
	return nil
}

func (m *manualDomainManager) Destroy() error {
	mlog.Info("[Plugin]: " + m.Name() + " plugin Destroy...")
	return nil
}

func (m *manualDomainManager) GetDomainByIndex(projectName, namespace string, index, preOccupiedLen int) string {
	return Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        index,
		nsPrefix:     m.nsPrefix,
		domainSuffix: m.domainSuffix,
	}.SubStr()
}

func (m *manualDomainManager) GetDomain(projectName, namespace string, preOccupiedLen int) string {
	return Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        -1,
		nsPrefix:     m.nsPrefix,
		domainSuffix: m.domainSuffix,
	}.SubStr()
}

func (m *manualDomainManager) GetCertSecretName(projectName string, index int) string {
	return ManualCertSecretName
}

func (m *manualDomainManager) GetClusterIssuer() string {
	return ""
}

func (m *manualDomainManager) GetCerts() (name, key, crt string) {
	return ManualCertSecretName, m.tlsKey, m.tlsCrt
}
