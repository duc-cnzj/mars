package event

import "github.com/google/wire"

var WireEvent = wire.NewSet(NewDispatcher)
