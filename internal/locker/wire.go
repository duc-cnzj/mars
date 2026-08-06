package locker

//go:generate go tool mockgen -destination ./mock_locker.go -package locker github.com/duc-cnzj/mars/v6/internal/locker Locker

import "github.com/google/wire"

var WireLocker = wire.NewSet(NewLocker)
