package cmd

import (
	"os"

	"github.com/duc-cnzj/mars/version"

	"github.com/spf13/cobra"
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
