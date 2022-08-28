package bootstrappers

import (
	"fmt"
	"os"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/uploader"
)

type UploadBootstrapper struct{}

func (*UploadBootstrapper) Tags() []string {
	return []string{}
}

func (*UploadBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	cfg := app.Config()
	if cfg.UploadDir != "" {
		if info, err := os.Stat(cfg.UploadDir); err != nil {
			if os.IsNotExist(err) {
				mlog.Infof("[UploadBootstrapper]: create upload dir %s", cfg.UploadDir)
				if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
					return err
				}
			}
		} else if !info.IsDir() {
			return fmt.Errorf("upload_dir %s not dir", cfg.UploadDir)
		}
	}

	up, err := uploader.NewUploader(cfg.UploadDir, "")
	if err != nil {
		return err
	}
	app.SetLocalUploader(up)
	app.SetUploader(up)

	return nil
}
