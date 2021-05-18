package cmd

import (
	"github.com/DuC-cnZj/mars/pkg/app"
	"github.com/DuC-cnZj/mars/pkg/config"
	"github.com/DuC-cnZj/mars/pkg/mlog"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start mars server.",
	Run: func(cmd *cobra.Command, args []string) {
		app := app.NewApplication(config.Init(cfgFile))
		if err := app.Bootstrap(); err != nil {
			mlog.Fatal(err)
		}
		<-app.Run()
		app.Shutdown()
	},
}
