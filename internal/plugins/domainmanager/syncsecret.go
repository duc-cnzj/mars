package domainmanager

import (
	"context"
	"errors"
	"strings"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	corev1 "k8s.io/api/core/v1"
)

var _ app.DomainManager = (*syncSecretDomainManager)(nil)

// SyncSecretSecretName 和 manual 方式保持名称一致，避免两种方式之间切换时需要手动部署才能生效的问题
const SyncSecretSecretName = ManualCertSecretName

// syncDeps 是 sync_secret 插件的依赖视图：只用 Logger 与 K8sRepo。
// K8sRepo 是单插件独有能力，不在 PluginApp 公共接口里，经 Resolve 断言取用。
type syncDeps interface {
	Logger() mlog.Logger
	K8sRepo() biz.K8sRepo
}

func init() {
	dr := &syncSecretDomainManager{}
	app.RegisterPlugin(dr.Name(), dr)
}

// syncSecretDomainManager 从 k8s secret 同步 TLS 证书，校验通过后作为证书来源。
type syncSecretDomainManager struct {
	nsPrefix       string
	wildcardDomain string
	domainSuffix   string

	secretNamespace string
	secretName      string

	k8sRepo biz.K8sRepo
	logger  mlog.Logger
}

// SyncSecretDomainManager 插件注册名。
const SyncSecretDomainManager = "sync_secret_domain_manager"

// Name 返回插件名 sync_secret_domain_manager。
func (d *syncSecretDomainManager) Name() string {
	return SyncSecretDomainManager
}

// Initialize 从 args 读取 ns_prefix/secret_namespace/secret_name/wildcard_domain，
// 拉取并校验 TLS secret：必须是 TLS 类型且 DNSNames 覆盖通配域名。
func (d *syncSecretDomainManager) Initialize(pluginApp app.PluginApp, args map[string]any) error {
	dep := app.Resolve[syncDeps](pluginApp)
	d.k8sRepo = dep.K8sRepo()
	d.logger = dep.Logger()

	nsPrefix, err := stringArg(args, "ns_prefix")
	if err != nil {
		return err
	}
	d.nsPrefix = nsPrefix

	secretNamespace, err := stringArg(args, "secret_namespace")
	if err != nil {
		return err
	}
	d.secretNamespace = secretNamespace

	secretName, err := stringArg(args, "secret_name")
	if err != nil {
		return err
	}
	d.secretName = secretName

	wildcardDomain, err := stringArg(args, "wildcard_domain")
	if err != nil {
		return err
	}
	d.wildcardDomain = wildcardDomain
	d.domainSuffix = strings.TrimLeft(wildcardDomain, "*.")

	if d.secretNamespace == "" || d.secretName == "" || d.wildcardDomain == "" {
		return errors.New("secret_namespace, secret_name, wildcard_domain required")
	}

	secret, err := d.k8sRepo.GetSecret(context.TODO(), d.secretNamespace, d.secretName)
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
	sec, err := d.k8sRepo.GetSecret(context.TODO(), d.secretNamespace, d.secretName)
	if err != nil {
		d.logger.Error("[TLS]: get secret error: ", err)
		return "", "", ""
	}
	return SyncSecretSecretName, string(sec.Data["tls.key"]), string(sec.Data["tls.crt"])
}
