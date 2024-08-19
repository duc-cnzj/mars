package cron

//go:generate mockgen -destination ./mock_cron.go -package cron github.com/duc-cnzj/mars/v4/internal/cron Runner
import (
	"github.com/google/wire"
)

var WireCron = wire.NewSet(NewManager, NewRobfigCronV3Runner)
