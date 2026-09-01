package data

//go:generate go tool mockgen -destination ./mock_repo.go -package data github.com/duc-cnzj/mars/v6/internal/biz ProjectRepo,GitRepo,EventRepo,ChangelogRepo,K8sRepo,FileRepo,RepoRepo,NamespaceRepo,HelmerRepo,Recorder,UserRepo
//go:generate go tool mockgen -destination ./mock_exec_test.go -package data github.com/duc-cnzj/mars/v6/internal/data ExecutorManager,Executor
//go:generate go tool mockgen -destination ./mock_cache.go -package data github.com/duc-cnzj/mars/v6/internal/data Cache

import (
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/google/wire"
)

// WireDataSet 提供 data 层全部 repo 与生命周期依赖的装配集。
// userRepo 同时是 biz.UserRepo（用户管理）与 biz.EffectiveRolesProvider（生效角色解析）
// 的实现，构造器返回具体类型 *userRepo，按 wire 的类型同一性匹配须显式 Bind 两个接口，
// 使同一实例注入 UserBiz 与 AuthBiz。
var WireDataSet = wire.NewSet(
	NewK8sRepo,
	NewDefaultHelmer,
	NewRepo,
	NewDefaultArchiver,
	NewExecutorManager,
	timer.NewReal,
	NewNamespaceRepo,
	NewProjectRepo,
	NewChangelogRepo,
	NewAccessTokenRepo,
	NewEventRepo,
	NewUserRepo,
	NewFileRepo,
	NewGitRepo,
	NewAuthn,
	wire.Bind(new(biz.UserRepo), new(*userRepo)),
	wire.Bind(new(biz.EffectiveRolesProvider), new(*userRepo)),
)

// WireCache 提供 data.Cache 的装配集。
var WireCache = wire.NewSet(NewCacheImpl)
