package uploader

import (
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
)

func TestNewUploader_WithValidConfig_ReturnsUploader(t *testing.T) {
	cfg := &config.Config{}
	logger := mlog.NewForConfig(cfg)
	data := data.NewData(cfg, logger)

	up, err := NewUploader(cfg, logger, data)

	assert.NoError(t, err)
	assert.NotNil(t, up)

	assert.Same(t, up, up.(*diskUploader).localUploader)
}

func TestNewUploader_WithS3Enabled_ReturnsS3Uploader(t *testing.T) {
	cfg := &config.Config{S3Enabled: true, S3Bucket: "test-bucket"}
	logger := mlog.NewForConfig(cfg)
	data := data.NewData(cfg, logger)

	up, err := NewUploader(cfg, logger, data)

	assert.NoError(t, err)
	assert.IsType(t, &s3Uploader{}, up)

	assert.NotSame(t, up, up.(*s3Uploader).localUploader)
	assert.IsType(t, &diskUploader{}, up.(*s3Uploader).localUploader)
}
