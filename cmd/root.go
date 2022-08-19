package cmd

import (
	"os"
	"path/filepath"

	"github.com/duc-cnzj/mars/version"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:     "app",
		Short:   "mars app.",
		Version: version.GetVersion().String(),
	}

	configExampleFile []byte
)

func Execute(configFile []byte) {
	configExampleFile = configFile
	if !version.GetVersion().HasBuildInfo() {
		rootCmd.AddCommand(testCmd)
	}
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(apiGatewayCmd)
	apiGatewayCmd.AddCommand(cronCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	var defaultConfig string
	if home := homedir.HomeDir(); home != "" {
		defaultConfig = filepath.Join(home, ".kube", "config")
	}

	apiGatewayCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $DIR/config.yaml)")
	apiGatewayCmd.PersistentFlags().BoolP("debug", "", true, "debug mode.")
	apiGatewayCmd.PersistentFlags().StringP("metrics_port", "", "9091", "metrics port")
	apiGatewayCmd.Flags().StringP("app_port", "", "6000", "app port.")
	apiGatewayCmd.Flags().StringP("kubeconfig", "", defaultConfig, "kubeconfig.")
	apiGatewayCmd.Flags().StringP("grpc_port", "", "", "grpc port.")
	apiGatewayCmd.Flags().BoolP("start_cron", "", false, "start cronjob.")

	viper.BindPFlag("debug", apiGatewayCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("metrics_port", apiGatewayCmd.PersistentFlags().Lookup("metrics_port"))
	viper.BindPFlag("app_port", apiGatewayCmd.Flags().Lookup("app_port"))
	viper.BindPFlag("kubeconfig", apiGatewayCmd.Flags().Lookup("kubeconfig"))
	viper.BindPFlag("grpc_port", apiGatewayCmd.Flags().Lookup("grpc_port"))
	viper.BindPFlag("start_cron", apiGatewayCmd.Flags().Lookup("start_cron"))
}
