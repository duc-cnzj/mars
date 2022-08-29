package bootstrappers

import (
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/uploader"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3UploaderBootstraper struct{}

func (s *S3UploaderBootstraper) Tags() []string {
	return []string{"s3", "uploader"}
}

func (s *S3UploaderBootstraper) Bootstrap(app contracts.ApplicationInterface) error {
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
