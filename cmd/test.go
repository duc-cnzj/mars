package cmd

import (
	"strings"

	"github.com/duc-cnzj/mars/internal/app"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/uploader"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/cobra"
)

type s3UploaderBootstraper struct{}

func (s *s3UploaderBootstraper) Bootstrap(app contracts.ApplicationInterface) error {
	var (
		endpoint        = app.Config().S3Endpoint
		accessKeyID     = app.Config().S3AccessKeyID
		secretAccessKey = app.Config().S3SecretAccessKey
		useSSL          = app.Config().S3UseSSL
	)
	if endpoint == "" || accessKeyID == "" || secretAccessKey == "" {
		return nil
	}

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return err
	}
	app.Config().UploadDir = "data"
	app.SetUploader(uploader.NewS3(minioClient, "mars", app.LocalUploader(), app.Config().UploadDir))
	return nil
}

func (s *s3UploaderBootstraper) Tags() []string {
	return []string{"s3", "uploader"}
}

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		app := app.NewApplication(
			config.Init(cfgFile),
			app.WithBootstrappers(&s3UploaderBootstraper{}),
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
