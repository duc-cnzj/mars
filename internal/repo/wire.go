package repo

import (
	"github.com/duc-cnzj/mars/v4/internal/utils/timer"
	"github.com/google/wire"
)

var WireRepoSet = wire.NewSet(
	NewCronRepo,
	NewK8sRepo,
	NewDefaultHelmer,
	NewToolRepo,
	NewDefaultArchiver,
	NewExecutorManager,
	NewFileRepo,
	NewEndpointRepo,
	timer.NewRealTimer,
	NewNamespaceRepo,
	NewEventRepo,
	NewPictureRepo,
	NewProjectRepo,
	NewGitRepo,
	NewChangelogRepo,
	NewGitProjectRepo,
	NewWsRepo,
	NewDomainRepo,
	NewAccessTokenRepo,
)
