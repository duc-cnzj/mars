package cmd

import (
	"github.com/duc-cnzj/mars/internal/app"
	"github.com/duc-cnzj/mars/internal/app/bootstrappers"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"

	"github.com/spf13/cobra"
)

var CronBootstrappers = []contracts.Bootstrapper{
	&bootstrappers.EventBootstrapper{},
	&bootstrappers.CronBootstrapper{},
	&bootstrappers.PluginsBootstrapper{},
	&bootstrappers.CacheBootstrapper{},
	&bootstrappers.DBBootstrapper{},
	&bootstrappers.PprofBootstrapper{},
	&bootstrappers.MetricsBootstrapper{},
	&bootstrappers.TracingBootstrapper{},
}

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "start mars cronjob.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Init(cfgFile)
		cfg.StartCron = true
		app := app.NewApplication(cfg, app.WithBootstrappers(CronBootstrappers...))
		if err := app.Bootstrap(); err != nil {
			mlog.Fatal(err)
		}
		<-app.Run().Done()
		app.Shutdown()
	},
}
