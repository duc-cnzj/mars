package deploy

//go:generate go tool mockgen -destination ./mock_deploy.go -package deploy github.com/duc-cnzj/mars/v6/internal/deploy JobManager,Job,Percentable,ReleaseInstaller,DeployMsger,SafeWriteMessageChan
import (
	"github.com/google/wire"
)

// WireDeploy 提供部署流水线（JobManager + ReleaseInstaller）与定时休眠重部署器
// （NewHibernateDeployer 复用 ApplyProjectDeps 走共享编排 ApplyProject）。
var WireDeploy = wire.NewSet(
	NewJobManager,
	wire.Struct(new(JobManagerDeps), "*"),
	NewReleaseInstaller,
	NewHibernateDeployer,
	wire.Struct(new(ApplyProjectDeps), "*"),
)
