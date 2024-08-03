package auth

import "github.com/google/wire"

var WireAuth = wire.NewSet(NewAuthn)
