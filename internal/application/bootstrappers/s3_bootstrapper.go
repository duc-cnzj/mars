package bootstrappers

import (
	"github.com/duc-cnzj/mars/v5/internal/application"
)

type S3Bootstrapper struct{}

func (d *S3Bootstrapper) Tags() []string {
	return []string{}
}

func (d *S3Bootstrapper) Bootstrap(app application.App) error {
	return app.Data().InitS3()
}
