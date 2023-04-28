package commands

import (
	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/cron"
)

func init() {
	cron.Register(func(manager contracts.CronManager, app contracts.ApplicationInterface) {
		// uploader.DirSizeCacheSeconds
		manager.NewCommand("disk_info", diskInfo).EveryFifteenMinutes()
	})
}

func diskInfo() error {
	_, err := app.Uploader().DirSize()
	return err
}
