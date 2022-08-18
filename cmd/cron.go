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
	&bootstrappers.DistributedLocksBootstrapper{},
	&bootstrappers.PprofBootstrapper{},
	&bootstrappers.MetricsBootstrapper{},
	&bootstrappers.TracingBootstrapper{},
}

var apiCronCmd = &cobra.Command{
	Use:   "cron",
	Short: "start mars cronjob.",
	Run: func(cmd *cobra.Command, args []string) {
		app := app.NewApplication(config.Init(cfgFile), app.WithBootstrappers(CronBootstrappers...))
		if err := app.Bootstrap(); err != nil {
			mlog.Fatal(err)
		}
		<-app.Run().Done()
		app.Shutdown()
	},
}
