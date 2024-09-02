package cron

//go:generate mockgen -destination ./mock_cron.go -package cron github.com/duc-cnzj/mars/v5/internal/cron Runner,Manager,Command
import (
	"github.com/google/wire"
)

var WireCron = wire.NewSet(NewManager, NewRobfigCronV3Runner)
