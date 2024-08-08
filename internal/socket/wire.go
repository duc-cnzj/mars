package socket

import (
	"github.com/duc-cnzj/mars/v4/internal/util/counter"
	"github.com/google/wire"
)

var WireSocket = wire.NewSet(NewWebsocketManager, NewJobManager, counter.NewCounter, NewReleaseInstaller)
