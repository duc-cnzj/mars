package uploader

import (
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
)

func TestNewUploader_WithValidConfig_ReturnsUploader(t *testing.T) {
	cfg := &config.Config{}
	logger := mlog.NewForConfig(cfg)

	up, err := NewUploader(cfg, logger, nil)

	assert.NoError(t, err)
	assert.NotNil(t, up)

	assert.Same(t, up, up.(*diskUploader).localUploader)
}

func TestNewUploader_WithS3Enabled_ReturnsS3Uploader(t *testing.T) {
	cfg := &config.Config{S3Enabled: true, S3Bucket: "test-bucket"}
	logger := mlog.NewForConfig(cfg)

	up, err := NewUploader(cfg, logger, func() *minio.Client { return nil })

	assert.NoError(t, err)
	assert.IsType(t, &s3Uploader{}, up)

	assert.NotSame(t, up, up.(*s3Uploader).localUploader)
	assert.IsType(t, &diskUploader{}, up.(*s3Uploader).localUploader)
}

// TestNewUploader_WithS3EnabledAndClient_WrapsMinioClient 覆盖 NewUploader 的
// S3 惰性解析链路：传入的取数函数在 thunk 调用时才解析出 *minio.Client，
// 被包成 minioClient 适配器塞进 s3Uploader.getMinioAPI。
func TestNewUploader_WithS3EnabledAndClient_WrapsMinioClient(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{S3Enabled: true, S3Bucket: "test-bucket", UploadDir: testDir}
	logger := mlog.NewForConfig(cfg)

	cli, err := minio.New("127.0.0.1:1", &minio.Options{
		Creds:  credentials.NewStaticV4("k", "s", ""),
		Secure: false,
	})
	assert.NoError(t, err)

	up, err := NewUploader(cfg, logger, func() *minio.Client { return cli })
	assert.NoError(t, err)

	mc, ok := up.(*s3Uploader).getMinioAPI().(*minioClient)
	assert.True(t, ok)
	assert.Same(t, cli, mc.Client)
}
