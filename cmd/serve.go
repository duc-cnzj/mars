package cmd

import (
	"path/filepath"

	"github.com/duc-cnzj/mars/v4/internal/app"
	"github.com/duc-cnzj/mars/v4/internal/app/bootstrappers"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
)

var serverBootstrappers = []contracts.Bootstrapper{
	&bootstrappers.EventBootstrapper{},
	&bootstrappers.PluginsBootstrapper{},
	&bootstrappers.AuthBootstrapper{},
	&bootstrappers.UploadBootstrapper{},
	&bootstrappers.CacheBootstrapper{},
	&bootstrappers.K8sClientBootstrapper{},
	&bootstrappers.DBBootstrapper{},
	&bootstrappers.ApiGatewayBootstrapper{},
	&bootstrappers.PprofBootstrapper{},
	&bootstrappers.GrpcBootstrapper{},
	&bootstrappers.MetricsBootstrapper{},
	&bootstrappers.OidcBootstrapper{},
	&bootstrappers.TracingBootstrapper{},
	&bootstrappers.CronBootstrapper{},
	&bootstrappers.AppBootstrapper{},
	&bootstrappers.S3UploaderBootstraper{},
}

var apiGatewayCmd = &cobra.Command{
	Use:   "serve",
	Short: "start mars server use grpc.",
	Run: func(cmd *cobra.Command, args []string) {
		app := app.NewApplication(
			config.Init(cfgFile),
			app.WithBootstrappers(serverBootstrappers...),
		)
		if err := app.Bootstrap(); err != nil {
			mlog.Fatal(err)
		}
		<-app.Run().Done()
		app.Shutdown()
	},
}

func init() {
	var defaultConfig string
	if home := homedir.HomeDir(); home != "" {
		defaultConfig = filepath.Join(home, ".kube", "config")
	}

	apiGatewayCmd.Flags().StringVar(&cfgFile, "config", "", "config file (default is $DIR/config.yaml)")
	apiGatewayCmd.Flags().BoolP("debug", "", true, "debug mode.")
	apiGatewayCmd.Flags().StringP("metrics_port", "", "9091", "metrics port")
	apiGatewayCmd.Flags().StringP("app_port", "", "6000", "app port.")
	apiGatewayCmd.Flags().StringP("kubeconfig", "", defaultConfig, "kubeconfig.")
	apiGatewayCmd.Flags().StringP("grpc_port", "", "", "grpc port.")
	apiGatewayCmd.Flags().StringP("exclude_server", "", "", "do not start these services(api/metrics/cron/profile), join with ','.")

	viper.BindPFlag("config", apiGatewayCmd.Flags().Lookup("config"))
	viper.BindPFlag("debug", apiGatewayCmd.Flags().Lookup("debug"))
	viper.BindPFlag("metrics_port", apiGatewayCmd.Flags().Lookup("metrics_port"))
	viper.BindPFlag("app_port", apiGatewayCmd.Flags().Lookup("app_port"))
	viper.BindPFlag("kubeconfig", apiGatewayCmd.Flags().Lookup("kubeconfig"))
	viper.BindPFlag("grpc_port", apiGatewayCmd.Flags().Lookup("grpc_port"))
	viper.BindPFlag("exclude_server", apiGatewayCmd.Flags().Lookup("exclude_server"))
}
