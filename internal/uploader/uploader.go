package uploader

import (
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/minio/minio-go/v7"
)

// NewUploader 依据配置构造上传器：默认本地磁盘；S3Enabled 时把磁盘上传器当
// 本地镜像，外包一层 S3 上传器。getMinioCli 惰性返回 minio 客户端——客户端由
// S3Bootstrapper 在 bootstrap 阶段才初始化，wire 构建时拿到的还是 nil，必须用
// 取数函数延迟到首次 S3 操作时解析，否则 S3 上传器会永久持有 nil 客户端。
func NewUploader(cfg *config.Config, logger mlog.Logger, getMinioCli func() *minio.Client) (Uploader, error) {
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
		up = newS3(func() minioAPI {
			return &minioClient{Client: getMinioCli()}
		}, cfg.S3Bucket, up, rootDir)
	}

	return up, nil
}
