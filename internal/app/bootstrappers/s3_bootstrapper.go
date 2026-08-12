package bootstrappers

import (
	"github.com/duc-cnzj/mars/v6/internal/app"
)

// S3Bootstrapper 初始化 s3 文件存储。
type S3Bootstrapper struct{}

// Tags 实现 Bootstrapper 接口的 Tags。
func (s *S3Bootstrapper) Tags() []string {
	return []string{}
}

// Bootstrap 实现 Bootstrapper 接口的 Bootstrap。
func (s *S3Bootstrapper) Bootstrap(deps app.BootstrapDeps) error {
	return deps.Data().InitS3()
}
