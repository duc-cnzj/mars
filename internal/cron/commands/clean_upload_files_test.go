package commands

import (
	"errors"

	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/testutil"
	"github.com/duc-cnzj/mars/v4/internal/uploader"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"testing"
	"time"
)

func Test_listFiles_PrettyYaml(t *testing.T) {
	lf := listFiles{
		{
			Path:         "/tmp/2.txt",
			HumanizeSize: "10 MB",
		},
		{
			Path:         "/tmp/1.txt",
			HumanizeSize: "1 B",
		},
	}
	assert.Equal(t, `- name: /tmp/2.txt
  size: 10 MB
- name: /tmp/1.txt
  size: 1 B
`, lf.PrettyYaml())
}

func TestCleanUploadFiles(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	app.EXPECT().Config().Return(&config.Config{UploadDir: "/tmp"}).AnyTimes()
	db.AutoMigrate(&models.File{})
	var files = []models.File{
		{
			UploadType: contracts.Local,
			Path:       "/tmp/path1",
			CreatedAt:  time.Now().Add(-24 * time.Hour),
		},
		{
			UploadType: contracts.S3,
			Path:       "/tmp/path2",
			CreatedAt:  time.Now().Add(-24 * time.Hour),
		},
		{
			UploadType: contracts.S3,
			Path:       "/tmp/path3",
			CreatedAt:  time.Now(),
		},
		{
			UploadType: contracts.Local,
			Path:       "/tmp/path4",
			CreatedAt:  time.Now().Add(-48 * time.Hour),
		},
	}
	for _, f := range files {
		db.Create(&f)
	}
	up := mock.NewMockUploader(m)
	localUp := mock.NewMockUploader(m)
	app.EXPECT().Uploader().Return(up).AnyTimes()
	app.EXPECT().LocalUploader().Return(localUp).AnyTimes()

	up.EXPECT().Type().Return(contracts.S3).AnyTimes()
	localUp.EXPECT().Type().Return(contracts.Local).AnyTimes()
	localUp.EXPECT().Exists("/tmp/path1").Return(true)
	up.EXPECT().Exists("/tmp/path2").Return(false)
	localUp.EXPECT().RemoveEmptyDir()

	db.Create(&models.File{
		UploadType: contracts.Local,
		Path:       "/tmp/local/path4",
		CreatedAt:  time.Now().Add(-24 * time.Hour),
	})
	up.EXPECT().AllDirectoryFiles(gomock.Any()).Return([]contracts.FileInfo{
		uploader.NewFileInfo("/tmp/up/path1", 100, time.Now().Add(-24*time.Hour)),
		uploader.NewFileInfo("/tmp/up/path2", 100, time.Now()),
		uploader.NewFileInfo("/tmp/up/path3", 100, time.Now().Add(-48*time.Hour)),
		uploader.NewFileInfo("/tmp/up/path4", 100, time.Now().Add(-24*time.Hour)),
	}, nil)
	up.EXPECT().Delete("/tmp/up/path4").Times(1)
	up.EXPECT().Delete("/tmp/up/path1").Times(1).Return(errors.New("xxx"))
	up.EXPECT().Delete("/tmp/up/path2").Times(0)
	up.EXPECT().Delete("/tmp/up/path3").Times(0)

	localUp.EXPECT().Exists("/tmp/local/path4").Return(true).Times(1)
	localUp.EXPECT().AllDirectoryFiles(gomock.Any()).Return([]contracts.FileInfo{
		uploader.NewFileInfo("/tmp/local/path1", 100, time.Now().Add(-24*time.Hour)),
		uploader.NewFileInfo("/tmp/local/path2", 100, time.Now()),
		uploader.NewFileInfo("/tmp/local/path3", 100, time.Now().Add(-48*time.Hour)),
		uploader.NewFileInfo("/tmp/local/path4", 100, time.Now().Add(-24*time.Hour)),
	}, nil)
	localUp.EXPECT().Delete("/tmp/local/path1").Times(1)
	testutil.AssertAuditLogFiredWithMsg(m, app, "删除未被记录的文件")
	assert.Nil(t, cleanUploadFiles())
}
