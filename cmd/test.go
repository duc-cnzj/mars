package cmd

import (
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use: "test",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
