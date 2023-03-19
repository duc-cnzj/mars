package cmd

import (
	"strings"

	"github.com/duc-cnzj/mars/v4/internal/app"
	"github.com/duc-cnzj/mars/v4/internal/app/bootstrappers"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"

	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		app := app.NewApplication(
			config.Init(cfgFile),
			app.WithBootstrappers(
				&bootstrappers.UploadBootstrapper{},
				&bootstrappers.S3UploaderBootstraper{},
			),
		)
		if err := app.Bootstrap(); err != nil {
			mlog.Fatal(err)
		}
		up := app.Uploader()
		disk := up.Disk("god")

		p1 := "2022/duc.txt"
		p2 := "2022/duc1.txt"
		p3 := "2022/admin/admin.txt"
		//
		put(disk, p1)
		put(disk, p2)
		put(disk, p3)

		up.AllDirectoryFiles("")
		//mlog.Warning(d6.DirSize("/"))
		//mlog.Warning(disk.DeleteDir("2022/admin"))
		//
		//put(disk, p1)
		//put(disk, p2)
		//put(disk, p3)

		//disk.Delete(p1)
		//disk.Delete(p2)
		//disk.Delete(p3)
	},
}

func put(up contracts.Uploader, path string) {
	up.Put(path, strings.NewReader("aaa"))
}
