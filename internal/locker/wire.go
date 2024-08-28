package locker

//go:generate mockgen -destination ./mock_locker.go -package locker github.com/duc-cnzj/mars/v5/internal/locker Locker

import "github.com/google/wire"

var WireLocker = wire.NewSet(NewLocker)
