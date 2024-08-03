package application

import "github.com/google/wire"

var WireApp = wire.NewSet(NewPluginManager)
