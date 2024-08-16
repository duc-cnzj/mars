package cmd

import (
	"fmt"
	"log"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/application/bootstrappers"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/cron"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/event"
	"github.com/duc-cnzj/mars/v4/internal/locker"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/uploader"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/singleflight"
)

var serverBootstrappers = []application.Bootstrapper{
	&bootstrappers.EventBootstrapper{},
	&bootstrappers.DBBootstrapper{},
	&bootstrappers.K8sBootstrapper{},
	&bootstrappers.ApiGatewayBootstrapper{},
	&bootstrappers.PprofBootstrapper{},
	&bootstrappers.GrpcBootstrapper{},
	&bootstrappers.MetricsBootstrapper{},
	&bootstrappers.TracingBootstrapper{},
	&bootstrappers.CronBootstrapper{},
	&bootstrappers.PluginBootstrapper{},
	&bootstrappers.SSOBootstrapper{},
	&bootstrappers.S3Bootstrapper{},
	&bootstrappers.PluginBootstrapper{},
}

var apiGatewayCmd = &cobra.Command{
	Use:   "serve",
	Short: "start mars server use grpc.",
	PreRun: func(cmd *cobra.Command, args []string) {
		fmt.Println(logo)
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(log.Lshortfile)
		cfg := config.Init(viper.GetString("config"))
		logger := mlog.NewLogger(cfg)
		app, err := InitializeApp(cfg, logger, serverBootstrappers)
		if err != nil {
			logger.Fatal(err)
		}
		if err := app.Bootstrap(); err != nil {
			logger.Fatal(err)
		}
		<-app.Run().Done()
		app.Shutdown()
	},
}

func NewSingleflight() *singleflight.Group {
	return &singleflight.Group{}
}

func newApp(
	cfg *config.Config,
	data data.Data,
	cron cron.Manager,
	bootstrappers []application.Bootstrapper,
	logger mlog.Logger,
	uploader uploader.Uploader,
	auth auth.Auth,
	dispatcher event.Dispatcher,
	cache cache.Cache,
	cacheLock locker.Locker,
	sf *singleflight.Group,
	pm application.PluginManger,
	reg *application.GrpcRegistry,
	ws application.WsServer,
	pr *prometheus.Registry,
	// FIXME: 加载定时任务, 因为所有逻辑都统一在 repo 中, 所以把定时任务也定义成了一个 repo, 看看还有没有别的办法
	cronRepo repo.CronRepo,
) application.App {
	_ = cronRepo
	return application.NewApp(
		cfg,
		data,
		logger,
		uploader,
		auth,
		dispatcher,
		cron,
		cache,
		cacheLock,
		sf,
		pm,
		reg,
		ws,
		pr,
		application.WithBootstrappers(bootstrappers...),
	)
}

func init() {
	apiGatewayCmd.Flags().BoolP("debug", "", true, "debug mode.")
	apiGatewayCmd.Flags().StringP("metrics_port", "", "9091", "metrics port")
	apiGatewayCmd.Flags().StringP("app_port", "", "6000", "app port.")
	apiGatewayCmd.Flags().StringP("kubeconfig", "", "", "kubeconfig.")
	apiGatewayCmd.Flags().StringP("grpc_port", "", "", "grpc port.")
	apiGatewayCmd.Flags().StringP("exclude_server", "", "", "do not start these services(api/metrics/cron/profile), join with ','.")

	viper.BindPFlag("debug", apiGatewayCmd.Flags().Lookup("debug"))
	viper.BindPFlag("metrics_port", apiGatewayCmd.Flags().Lookup("metrics_port"))
	viper.BindPFlag("app_port", apiGatewayCmd.Flags().Lookup("app_port"))
	viper.BindPFlag("kubeconfig", apiGatewayCmd.Flags().Lookup("kubeconfig"))
	viper.BindPFlag("grpc_port", apiGatewayCmd.Flags().Lookup("grpc_port"))
	viper.BindPFlag("exclude_server", apiGatewayCmd.Flags().Lookup("exclude_server"))
}
