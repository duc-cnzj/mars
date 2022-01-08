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
	rootCmd.AddCommand(testCmd)
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
	rootCmd.PersistentFlags().StringP("app_port", "", "6000", "app port.")
	rootCmd.PersistentFlags().StringP("grpc_port", "", "50000", "grpc port.")
	rootCmd.PersistentFlags().StringP("kubeconfig", "", defaultConfig, "kubeconfig.")
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("app_port", rootCmd.PersistentFlags().Lookup("app_port"))
	viper.BindPFlag("kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig"))
	viper.BindPFlag("grpc_port", rootCmd.PersistentFlags().Lookup("grpc_port"))
}
