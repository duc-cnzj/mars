package domainmanager

import (
	"errors"
	"strings"

	"github.com/duc-cnzj/mars/v6/internal/application"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	corev1 "k8s.io/api/core/v1"
)

var _ application.DomainManager = (*syncSecretDomainManager)(nil)

// SyncSecretSecretName 和 manual 方式保持名称一致，避免两种方式之间切换时需要手动部署才能生效的问题
const SyncSecretSecretName = ManualCertSecretName

func init() {
	dr := &syncSecretDomainManager{}
	application.RegisterPlugin(dr.Name(), dr)
}

// syncSecretDomainManager 从 k8s secret 同步 TLS 证书，校验通过后作为证书来源。
type syncSecretDomainManager struct {
	nsPrefix       string
	wildcardDomain string
	domainSuffix   string

	secretNamespace string
	secretName      string

	data   data.Data
	logger mlog.Logger
}

// SyncSecretDomainManager 插件注册名。
const SyncSecretDomainManager = "sync_secret_domain_manager"

// Name 返回插件名 sync_secret_domain_manager。
func (d *syncSecretDomainManager) Name() string {
	return SyncSecretDomainManager
}

// Initialize 从 args 读取 ns_prefix/secret_namespace/secret_name/wildcard_domain，
// 拉取并校验 TLS secret：必须是 TLS 类型且 DNSNames 覆盖通配域名。
func (d *syncSecretDomainManager) Initialize(app application.PluginApp, args map[string]any) error {
	d.data = app.Data()
	d.logger = app.Logger()

	if p, ok := args["ns_prefix"]; ok {
		s, ok := p.(string)
		if !ok {
			return errors.New("ns_prefix must be string")
		}
		d.nsPrefix = s
	}

	if p, ok := args["secret_namespace"]; ok {
		s, ok := p.(string)
		if !ok {
			return errors.New("secret_namespace must be string")
		}
		d.secretNamespace = s
	}

	if p, ok := args["secret_name"]; ok {
		s, ok := p.(string)
		if !ok {
			return errors.New("secret_name must be string")
		}
		d.secretName = s
	}

	if wd, ok := args["wildcard_domain"]; ok {
		s, ok := wd.(string)
		if !ok {
			return errors.New("wildcard_domain must be string")
		}
		d.wildcardDomain = s
		d.domainSuffix = strings.TrimLeft(s, "*.")
	}

	if d.secretNamespace == "" || d.secretName == "" || d.wildcardDomain == "" {
		return errors.New("secret_namespace, secret_name, wildcard_domain required")
	}

	secret, err := d.data.K8sClient().SecretLister.Secrets(d.secretNamespace).Get(d.secretName)
	if err != nil {
		return err
	}

	if secret.Type != corev1.SecretTypeTLS {
		return errors.New("secret not verified")
	}

	var (
		tlsKey = secret.Data["tls.key"]
		tlsCrt = secret.Data["tls.crt"]
	)
	err = validateTLSWildcardDomain(tlsKey, tlsCrt, d.wildcardDomain)
	if err != nil {
		return err
	}

	d.logger.Info("[Plugin]: " + d.Name() + " plugin Initialize...")

	return nil
}

// Destroy 输出销毁日志。
func (d *syncSecretDomainManager) Destroy() error {
	d.logger.Info("[Plugin]: " + d.Name() + " plugin Destroy...")
	return nil
}

// GetDomainByIndex 用同步 secret 的通配域名生成带 index 的完整子域名。
func (d *syncSecretDomainManager) GetDomainByIndex(projectName, namespace string, index, preOccupiedLen int) string {
	return Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        index,
		nsPrefix:     d.nsPrefix,
		domainSuffix: d.domainSuffix,
	}.SubStr()
}

// GetCertSecretName 返回同步证书对应的 secret 名。
func (d *syncSecretDomainManager) GetCertSecretName(projectName string, index int) string {
	return SyncSecretSecretName
}

// GetClusterIssuer 同步 secret 不使用 cert-manager，返回空串。
func (d *syncSecretDomainManager) GetClusterIssuer() string {
	return ""
}

// GetCerts 从 k8s secret 实时读取 tls.key/tls.crt；读取失败时返回空三元组并记录错误。
func (d *syncSecretDomainManager) GetCerts() (name, key, crt string) {
	sec, err := d.data.K8sClient().SecretLister.Secrets(d.secretNamespace).Get(d.secretName)
	if err != nil {
		d.logger.Error("[TLS]: get secret error: ", err)
		return "", "", ""
	}
	return SyncSecretSecretName, string(sec.Data["tls.key"]), string(sec.Data["tls.crt"])
}
