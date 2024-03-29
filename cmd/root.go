package cmd

import (
	"os"

	"github.com/duc-cnzj/mars/v4/version"

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
	rootCmd.AddCommand(showCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
