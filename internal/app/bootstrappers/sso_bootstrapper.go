package bootstrappers

import (
	"github.com/duc-cnzj/mars/v6/internal/app"
)

// SSOBootstrapper 初始化 oidc sso provider。
type SSOBootstrapper struct{}

// Tags 实现 Bootstrapper 接口的 Tags。
func (s *SSOBootstrapper) Tags() []string {
	return []string{}
}

// Bootstrap 实现 Bootstrapper 接口的 Bootstrap。
func (s *SSOBootstrapper) Bootstrap(deps app.BootstrapDeps) error {
	deps.Data().InitOidcProvider()
	return nil
}
