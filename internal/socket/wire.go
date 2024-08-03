package socket

import "github.com/google/wire"

var WireSocket = wire.NewSet(NewWebsocketManager, NewJobManager)
