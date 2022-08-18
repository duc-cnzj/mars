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

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	var defaultConfig string
	if home := homedir.HomeDir(); home != "" {
		defaultConfig = filepath.Join(home, ".kube", "config")
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $DIR/config.yaml)")
	rootCmd.PersistentFlags().BoolP("debug", "", true, "debug mode.")

	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("app_port", rootCmd.PersistentFlags().Lookup("app_port"))
	viper.BindPFlag("kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig"))
	viper.BindPFlag("grpc_port", rootCmd.PersistentFlags().Lookup("grpc_port"))
	viper.BindPFlag("metrics_port", rootCmd.PersistentFlags().Lookup("metrics_port"))

	apiCronCmd.Flags().StringP("metrics_port", "", "9091", "metrics port")

	apiGatewayCmd.Flags().StringP("metrics_port", "", "9091", "metrics port")
	apiGatewayCmd.Flags().StringP("kubeconfig", "", defaultConfig, "kubeconfig.")
	apiGatewayCmd.Flags().StringP("app_port", "", "6000", "app port.")
	apiGatewayCmd.Flags().StringP("grpc_port", "", "", "grpc port.")
}
