package locker

import "github.com/google/wire"

var WireLocker = wire.NewSet(NewLocker)
