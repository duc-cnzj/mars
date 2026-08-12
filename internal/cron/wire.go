package cron

//go:generate go tool mockgen -destination ./mock_cron.go -package cron github.com/duc-cnzj/mars/v6/internal/cron Runner,Manager
import (
	"github.com/google/wire"
)

// WireCron 提供 cron.Manager 与 cron.Runner 的装配集。
var WireCron = wire.NewSet(NewManager, NewRobfigCronV3Runner)
