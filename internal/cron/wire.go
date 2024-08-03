package cron

import (
	"github.com/google/wire"
)

var WireCron = wire.NewSet(NewManager, NewRobfigCronV3Runner)
