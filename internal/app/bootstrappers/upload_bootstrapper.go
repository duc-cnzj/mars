package bootstrappers

import (
	"fmt"
	"os"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/uploader"
)

type UploadBootstrapper struct{}

func (*UploadBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	if app.Config().UploadDir != "" {
		if info, err := os.Stat(app.Config().UploadDir); err != nil {
			if os.IsNotExist(err) {
				mlog.Infof("[UploadBootstrapper]: create upload dir %s", app.Config().UploadDir)
				if err := os.MkdirAll(app.Config().UploadDir, 0755); err != nil {
					return err
				}
			}
		} else if !info.IsDir() {
			return fmt.Errorf("upload_dir %s not dir", app.Config().UploadDir)
		}
	}

	up, err := uploader.NewUploader(app.Config().UploadDir, "")
	if err != nil {
		return err
	}
	app.SetUploader(up)

	return nil
}
