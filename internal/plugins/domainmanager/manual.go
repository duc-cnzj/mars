package domainmanager

import (
	"errors"
	"strings"

	"github.com/duc-cnzj/mars/v6/internal/application"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

// ManualCertSecretName 手动证书方式使用的 secret 名。
const ManualCertSecretName = "mars-external-tls-secret" // #nosec G101

var _ application.DomainManager = (*manualDomainManager)(nil)

func init() {
	dr := &manualDomainManager{}
	application.RegisterPlugin(dr.Name(), dr)
}

// manualDomainManager 使用用户手动配置的证书与通配域名。
type manualDomainManager struct {
	nsPrefix       string
	wildcardDomain string
	domainSuffix   string

	tlsCrt string
	tlsKey string
	logger mlog.Logger
}

// Name 返回插件名 manual_domain_manager。
func (m *manualDomainManager) Name() string {
	return "manual_domain_manager"
}

// Initialize 从 args 读取 ns_prefix/tls_crt/tls_key/wildcard_domain，
// 校验三项必填参数并通过 validateTLSWildcardDomain 验证证书域名匹配。
func (m *manualDomainManager) Initialize(app application.PluginApp, args map[string]any) error {
	m.logger = app.Logger()

	nsPrefix, err := stringArg(args, "ns_prefix")
	if err != nil {
		return err
	}
	m.nsPrefix = nsPrefix

	tlsCrt, err := stringArg(args, "tls_crt")
	if err != nil {
		return err
	}
	m.tlsCrt = tlsCrt

	tlsKey, err := stringArg(args, "tls_key")
	if err != nil {
		return err
	}
	m.tlsKey = tlsKey

	wildcardDomain, err := stringArg(args, "wildcard_domain")
	if err != nil {
		return err
	}
	m.wildcardDomain = wildcardDomain
	m.domainSuffix = strings.TrimLeft(wildcardDomain, "*.")
	if m.tlsKey == "" || m.tlsCrt == "" || m.wildcardDomain == "" {
		return errors.New("tls_crt, tls_key, wildcard_domain required")
	}
	if err := validateTLSWildcardDomain(m.tlsKey, m.tlsCrt, m.wildcardDomain); err != nil {
		return err
	}

	m.logger.Info("[Plugin]: " + m.Name() + " plugin Initialize...")
	return nil
}

// Destroy 输出销毁日志。
func (m *manualDomainManager) Destroy() error {
	m.logger.Info("[Plugin]: " + m.Name() + " plugin Destroy...")
	return nil
}

// GetDomainByIndex 用配置的通配域名生成带 index 的完整子域名。
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

// GetCertSecretName 返回手动证书对应的 secret 名。
func (m *manualDomainManager) GetCertSecretName(projectName string, index int) string {
	return ManualCertSecretName
}

// GetClusterIssuer 手动证书不使用 cert-manager，返回空串。
func (m *manualDomainManager) GetClusterIssuer() string {
	return ""
}

// GetCerts 返回手动配置的 secret 名、tlsKey 与 tlsCrt。
func (m *manualDomainManager) GetCerts() (name, key, crt string) {
	return ManualCertSecretName, m.tlsKey, m.tlsCrt
}
