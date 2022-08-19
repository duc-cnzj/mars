package cmd

import (
	"github.com/duc-cnzj/mars/internal/app"
	"github.com/duc-cnzj/mars/internal/app/bootstrappers"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"

	"github.com/spf13/cobra"
)

var ServerBootstrappers = []contracts.Bootstrapper{
	&bootstrappers.EventBootstrapper{},
	&bootstrappers.PluginsBootstrapper{},
	&bootstrappers.AuthBootstrapper{},
	&bootstrappers.UploadBootstrapper{},
	&bootstrappers.CacheBootstrapper{},
	&bootstrappers.K8sClientBootstrapper{},
	&bootstrappers.DBBootstrapper{},
	&bootstrappers.DistributedLocksBootstrapper{},
	&bootstrappers.ApiGatewayBootstrapper{},
	&bootstrappers.PprofBootstrapper{},
	&bootstrappers.GrpcBootstrapper{},
	&bootstrappers.MetricsBootstrapper{},
	&bootstrappers.OidcBootstrapper{},
	&bootstrappers.TracingBootstrapper{},
	&bootstrappers.CronBootstrapper{},
	&bootstrappers.AppBootstrapper{},
}

var apiGatewayCmd = &cobra.Command{
	Use:   "serve",
	Short: "start mars server use grpc.",
	Run: func(cmd *cobra.Command, args []string) {
		app := app.NewApplication(config.Init(cfgFile), app.WithBootstrappers(ServerBootstrappers...))
		if err := app.Bootstrap(); err != nil {
			mlog.Fatal(err)
		}
		<-app.Run().Done()
		app.Shutdown()
	},
}
