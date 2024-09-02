package uploader

import (
	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
)

func NewUploader(cfg *config.Config, logger mlog.Logger, data data.Data) (Uploader, error) {
	var (
		up  Uploader
		err error
	)

	logger = logger.WithModule("uploader/uploader")
	up, err = NewDiskUploader(cfg.UploadDir, logger)
	if err != nil {
		return nil, err
	}

	rootDir := "mars"
	if cfg.S3Enabled {
		up = NewS3(data.MinioCli(), cfg.S3Bucket, up, rootDir)
	}

	return up, nil
}
