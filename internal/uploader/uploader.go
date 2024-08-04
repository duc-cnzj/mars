package uploader

import (
	"github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

func NewUploader(cfg *config.Config, logger mlog.Logger, data data.Data, cache cache.Cache) (Uploader, error) {
	var (
		up  Uploader
		err error
	)
	up, err = NewDiskUploader(cfg, logger)
	if err != nil {
		return nil, err
	}

	rootDir := "mars"
	if cfg.S3Enabled {
		up = NewS3(data.MinioCli(), cfg.S3Bucket, up, rootDir)
	}

	return NewCacheUploader(up, logger, cache), nil
}
