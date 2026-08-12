package bootstrappers

import (
	"github.com/duc-cnzj/mars/v6/internal/app"
)

// DBBootstrapper 启动数据库：初始化连接 + 可选自动迁移。
type DBBootstrapper struct{}

// Tags 实现 Bootstrapper 接口的 Tags。
func (d *DBBootstrapper) Tags() []string {
	return []string{}
}

// Bootstrap 实现 Bootstrapper 接口的 Bootstrap。
func (d *DBBootstrapper) Bootstrap(deps app.BootstrapDeps) error {
	closeFunc, err := deps.Data().InitDB()
	if err != nil {
		return err
	}
	deps.RegisterAfterShutdownFunc(func() {
		closeFunc()
	})
	deps.Logger().Info("[DB]: auto migrate database")
	if deps.Config().DBAutoMigrate {
		return deps.Data().Migrate()
	}
	return nil
}
