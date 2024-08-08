package repo

import (
	"github.com/duc-cnzj/mars/v4/internal/util/timer"
	"github.com/google/wire"
)

var WireRepoSet = wire.NewSet(
	NewCronRepo,
	NewK8sRepo,
	NewDefaultHelmer,
	NewRepo,
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
	NewWsRepo,
	NewDomainRepo,
	NewAccessTokenRepo,
)
