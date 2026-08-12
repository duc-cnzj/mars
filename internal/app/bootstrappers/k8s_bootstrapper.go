package bootstrappers

import (
	"github.com/duc-cnzj/mars/v6/internal/app"
)

// K8sBootstrapper 在 k8s 环境运行时启动 k8s client。
type K8sBootstrapper struct{}

// Tags 实现 Bootstrapper 接口的 Tags。
func (k *K8sBootstrapper) Tags() []string {
	return []string{}
}

// Bootstrap 实现 Bootstrapper 接口的 Bootstrap。
func (k *K8sBootstrapper) Bootstrap(deps app.BootstrapDeps) error {
	if deps.Config().IsK8sEnv() {
		return deps.Data().InitK8s(deps.Done())
	}
	return nil
}
