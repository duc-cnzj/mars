package domainmanager

import (
	"errors"
	"fmt"
	"strings"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/hasher"
)

// certManagerName 插件注册名。
const certManagerName = "cert-manager_domain_manager"

var _ app.DomainManager = (*certManager)(nil)

func init() {
	dr := &certManager{}
	app.RegisterPlugin(dr.Name(), dr)
}

// certManager 基于 cert-manager + lets encrypt 签发证书；lets encrypt 对 subdomain 长度要求为 64。
type certManager struct {
	nsPrefix       string
	clusterIssuer  string
	wildcardDomain string
	domainSuffix   string

	logger mlog.Logger
}

// Name 返回插件名 cert-manager_domain_manager。
func (d *certManager) Name() string {
	return certManagerName
}

// Initialize 从 args 读取 ns_prefix/cluster_issuer/wildcard_domain，校验必填项后保存。
func (d *certManager) Initialize(pluginApp app.PluginApp, args map[string]any) error {
	d.logger = pluginApp.Logger()

	nsPrefix, err := stringArg(args, "ns_prefix")
	if err != nil {
		return err
	}
	d.nsPrefix = nsPrefix

	clusterIssuer, err := stringArg(args, "cluster_issuer")
	if err != nil {
		return err
	}
	d.clusterIssuer = clusterIssuer

	wildcardDomain, err := stringArg(args, "wildcard_domain")
	if err != nil {
		return err
	}
	d.wildcardDomain = wildcardDomain
	d.domainSuffix = strings.TrimLeft(wildcardDomain, "*.")

	if d.clusterIssuer == "" || d.wildcardDomain == "" {
		return errors.New("cluster_issuer, wildcard_domain required")
	}

	d.logger.Info("[Plugin]: " + d.Name() + " plugin Initialize...")
	return nil
}

// Destroy 输出销毁日志。
func (d *certManager) Destroy() error {
	d.logger.Info("[Plugin]: " + d.Name() + " plugin Destroy...")
	return nil
}

// GetCertSecretName 用 projectName+index 的哈希生成 cert-manager secret 名。
func (d *certManager) GetCertSecretName(projectName string, index int) string {
	return fmt.Sprintf("mars-tls-%s", hasher.Hash(fmt.Sprintf("%s-%d", projectName, index)))
}

// GetClusterIssuer 返回配置的 cert-manager clusterIssuer。
func (d *certManager) GetClusterIssuer() string {
	return d.clusterIssuer
}

// GetDomainByIndex 用通配域名生成带 index 的完整子域名。
func (d *certManager) GetDomainByIndex(projectName, namespace string, index, preOccupiedLen int) string {
	return Subdomain{
		maxLen:       maxDomainLength - preOccupiedLen,
		projectName:  projectName,
		namespace:    namespace,
		index:        index,
		nsPrefix:     d.nsPrefix,
		domainSuffix: d.domainSuffix,
	}.SubStr()
}

// GetCerts 证书由 cert-manager 管理，运行时不注入证书，返回空三元组。
func (d *certManager) GetCerts() (name, key, crt string) {
	return "", "", ""
}
