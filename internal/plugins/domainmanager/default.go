package domainmanager

import (
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

var _ app.DomainManager = (*defaultDomainManager)(nil)

func init() {
	dr := &defaultDomainManager{}
	app.RegisterPlugin(dr.Name(), dr)
}

type defaultDomainManager struct {
	logger mlog.Logger
}

// Name 返回插件名 default_domain_manager。
func (d *defaultDomainManager) Name() string {
	return "default_domain_manager"
}

// Initialize 保存 pluginApp.Logger 并输出初始化日志。
func (d *defaultDomainManager) Initialize(pluginApp app.PluginApp, args map[string]any) error {
	d.logger = pluginApp.Logger()
	d.logger.Info("[Plugin]: " + d.Name() + " plugin Initialize...")
	return nil
}

// Destroy 输出销毁日志，无需额外清理。
func (d *defaultDomainManager) Destroy() error {
	d.logger.Info("[Plugin]: " + d.Name() + " plugin Destroy...")
	return nil
}

// GetDomainByIndex 用占位域名生成带 index 的完整子域名。
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

// GetCertSecretName 默认实现不产出证书，返回空串。
func (d *defaultDomainManager) GetCertSecretName(projectName string, index int) string {
	return ""
}

// GetClusterIssuer 默认实现不产出证书，返回空串。
func (d *defaultDomainManager) GetClusterIssuer() string {
	return ""
}

// GetCerts 默认实现不注入证书，返回空三元组。
func (d *defaultDomainManager) GetCerts() (name, key, crt string) {
	return "", "", ""
}
