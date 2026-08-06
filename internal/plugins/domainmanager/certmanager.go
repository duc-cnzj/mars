package domainmanager

import (
	"errors"
	"fmt"
	"strings"

	"github.com/duc-cnzj/mars/v6/internal/application"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/hasher"
)

// name 插件注册名。
var name = "cert-manager_domain_manager"

var _ application.DomainManager = (*certManager)(nil)

func init() {
	dr := &certManager{}
	application.RegisterPlugin(dr.Name(), dr)
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
	return name
}

// Initialize 从 args 读取 ns_prefix/cluster_issuer/wildcard_domain，校验必填项后保存。
func (d *certManager) Initialize(app application.PluginApp, args map[string]any) error {
	d.logger = app.Logger()

	if p, ok := args["ns_prefix"]; ok {
		s, ok := p.(string)
		if !ok {
			return errors.New("ns_prefix must be string")
		}
		d.nsPrefix = s
	}

	if issuer, ok := args["cluster_issuer"]; ok {
		s, ok := issuer.(string)
		if !ok {
			return errors.New("cluster_issuer must be string")
		}
		d.clusterIssuer = s
	}

	if wd, ok := args["wildcard_domain"]; ok {
		s, ok := wd.(string)
		if !ok {
			return errors.New("wildcard_domain must be string")
		}
		d.wildcardDomain = s
		d.domainSuffix = strings.TrimLeft(s, "*.")
	}

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
