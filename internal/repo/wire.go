package repo

//go:generate mockgen -destination ./mock_repo.go -package repo github.com/duc-cnzj/mars/v5/internal/repo ProjectRepo,GitRepo,AccessTokenRepo,EventRepo,AuthRepo,ChangelogRepo,K8sRepo,EndpointRepo,FileRepo,RepoRepo,PictureRepo,NamespaceRepo,HelmerRepo,Recorder,ExecutorManager,Executor

import (
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
	"github.com/google/wire"
)

var WireRepoSet = wire.NewSet(
	NewCronRepo,
	NewAuthRepo,
	NewK8sRepo,
	NewDefaultHelmer,
	NewRepo,
	NewDefaultArchiver,
	NewExecutorManager,
	NewFileRepo,
	NewEndpointRepo,
	timer.NewReal,
	NewNamespaceRepo,
	NewEventRepo,
	NewPictureRepo,
	NewProjectRepo,
	NewGitRepo,
	NewChangelogRepo,
	NewAccessTokenRepo,
)
