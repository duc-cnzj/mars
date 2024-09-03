package cmd

import (
	"fmt"

	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var testCmd = &cobra.Command{
	Use: "test",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Init(viper.GetString("config"))
		logger := mlog.NewForConfig(cfg)
		app, err := InitializeApp(cfg, logger, nil)
		if err != nil {
			logger.Fatal(err)
		}
		defer app.Shutdown()
		db := app.DB()
		for i := 0; i < 100; i++ {
			db.Repo.Create().
				SetName("name_" + fmt.Sprintf("%d", i)).
				SetGitProjectName("test" + fmt.Sprintf("%d", i)).
				SetGitProjectID(int32(i)).
				SaveX(cmd.Context())
		}
	},
}
