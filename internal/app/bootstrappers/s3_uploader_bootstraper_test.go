package bootstrappers

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestS3UploaderBootstraper_Bootstrap(t *testing.T) {
	assert.Equal(t, []string{"s3", "uploader"}, (&S3UploaderBootstrapper{}).Tags())
}

func TestS3UploaderBootstraper_Boot1(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{
		S3Endpoint:        "",
		S3UseSSL:          false,
		S3SecretAccessKey: "",
		S3AccessKeyID:     "",
	}).AnyTimes()
	app.EXPECT().SetUploader(gomock.Any()).Times(0)
	(&S3UploaderBootstrapper{}).Bootstrap(app)
}

func TestS3UploaderBootstraper_Boot2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{
		S3Endpoint:        "xxx",
		S3UseSSL:          false,
		S3SecretAccessKey: "xxx",
		S3AccessKeyID:     "xxx",
	}).AnyTimes()
	app.EXPECT().SetUploader(gomock.Any()).Times(1)
	app.EXPECT().LocalUploader().Return(nil).Times(1)
	(&S3UploaderBootstrapper{}).Bootstrap(app)
}
