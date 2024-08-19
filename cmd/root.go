package cmd

import (
	"os"

	"github.com/duc-cnzj/mars/v4/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:     "app",
		Short:   "mars app.",
		Version: version.GetVersion().String(),
	}

	configExampleFile []byte

	logo string
)

// Execute root cmd.
func Execute(configFile []byte, logoStr string) {
	configExampleFile = configFile
	logo = logoStr
	if !version.GetVersion().HasBuildInfo() {
		rootCmd.AddCommand(testCmd)
	}
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(apiGatewayCmd)
	rootCmd.AddCommand(inspect)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $DIR/config.yaml)")
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.SetEnvPrefix("MARS")
	viper.BindEnv("config")
}
