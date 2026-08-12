package data

//go:generate go tool mockgen -destination ./mock_repo.go -package data github.com/duc-cnzj/mars/v6/internal/biz ProjectRepo,GitRepo,EventRepo,ChangelogRepo,K8sRepo,FileRepo,RepoRepo,NamespaceRepo,HelmerRepo,Recorder
//go:generate go tool mockgen -destination ./mock_exec_test.go -package data github.com/duc-cnzj/mars/v6/internal/data ExecutorManager,Executor
//go:generate go tool mockgen -destination ./mock_cache.go -package data github.com/duc-cnzj/mars/v6/internal/data Cache

import (
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/google/wire"
)

// WireDataSet 提供 data 层全部 repo 与生命周期依赖的装配集。
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
	NewFileRepo,
	NewGitRepo,
	NewAuthn,
)

// WireCache 提供 data.Cache 的装配集。
var WireCache = wire.NewSet(NewCacheImpl)
