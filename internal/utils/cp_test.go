package utils

import (
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewFileCopier(t *testing.T) {
	assert.Implements(t, (*contracts.PodFileCopier)(nil), NewFileCopier(nil, nil))
}

func Test_fileCopier_Copy(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	re := mock.NewMockRemoteExecutor(m)
	app := testutil.MockApp(m)
	up := mock.NewMockUploader(m)
	app.EXPECT().Uploader().Return(up)
	app.EXPECT().LocalUploader().Return(up)
	app.EXPECT().Config().Return(&config.Config{UploadMaxSize: "100M"})
	fi := mock.NewMockFileInfo(m)
	fi.EXPECT().Size().Return(uint64(10))
	up.EXPECT().Stat("a.txt").Return(fi, nil)
	up.EXPECT().Type().Return(contracts.S3)

	r := &readCloser{}
	up.EXPECT().Read("a.txt").Return(r, nil)
	up.EXPECT().Exists("a.txt").Return(true)
	up.EXPECT().Delete("a.txt")
	fin := mock.NewMockFileInfo(m)
	fin.EXPECT().Path().Return("a.txt")
	up.EXPECT().Delete("a.txt")
	up.EXPECT().Put("a.txt", r).Return(fin, nil)
	arch := mock.NewMockArchiver(m)
	arch.EXPECT().Archive([]string{"a.txt"}, "a.txt.tar.gz").Return(nil)
	arch.EXPECT().Open("a.txt.tar.gz").Return(r, nil)
	arch.EXPECT().Remove("a.txt.tar.gz")
	re.EXPECT().WithCommand([]string{"tar", "-zmxf", "-", "-C", "/abc"}).Return(re)
	re.EXPECT().WithMethod("POST").Return(re)
	re.EXPECT().WithContainer("ns", "pod", "app").Return(re)
	re.EXPECT().Execute(gomock.Any(), nil, nil, gomock.Any(), gomock.Any(), gomock.Any(), false, nil).Return(nil)

	_, err := NewFileCopier(re, arch).Copy("ns", "pod", "app", "a.txt", "/abc", nil, nil)
	assert.Nil(t, err)
}

func Test_fileCopier_Copy_DiffDir(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	re := mock.NewMockRemoteExecutor(m)
	app := testutil.MockApp(m)
	up := mock.NewMockUploader(m)
	app.EXPECT().Uploader().Return(up)
	app.EXPECT().LocalUploader().Return(up)
	app.EXPECT().Config().Return(&config.Config{UploadMaxSize: "100M"})
	fi := mock.NewMockFileInfo(m)
	fi.EXPECT().Size().Return(uint64(10))
	up.EXPECT().Stat("a.txt").Return(fi, nil)
	up.EXPECT().Type().Return(contracts.S3)

	r := &readCloser{}
	up.EXPECT().Read("a.txt").Return(r, nil)
	up.EXPECT().Exists("a.txt").Return(true)
	up.EXPECT().Delete("a.txt")
	fin := mock.NewMockFileInfo(m)
	fin.EXPECT().Path().Return("a.txt")
	up.EXPECT().Delete("a.txt")
	up.EXPECT().Put("a.txt", r).Return(fin, nil)
	arch := mock.NewMockArchiver(m)
	arch.EXPECT().Archive([]string{"a.txt"}, "a.txt.tar.gz").Return(nil)
	arch.EXPECT().Open("a.txt.tar.gz").Return(r, nil)
	arch.EXPECT().Remove("a.txt.tar.gz")
	re.EXPECT().WithCommand([]string{"tar", "-zmxf", "-", "-C", "/tmp"}).Return(re)
	re.EXPECT().WithMethod("POST").Return(re)
	re.EXPECT().WithContainer("ns", "pod", "app").Return(re)
	re.EXPECT().Execute(gomock.Any(), nil, nil, gomock.Any(), gomock.Any(), gomock.Any(), false, nil).Return(nil)

	res, err := NewFileCopier(re, arch).Copy("ns", "pod", "app", "a.txt", "", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "/tmp", res.TargetDir)
}

func Test_fileCopier_Copy_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	up := mock.NewMockUploader(m)
	app.EXPECT().Uploader().Return(up)
	app.EXPECT().LocalUploader().Return(up)
	up.EXPECT().Stat("a.txt").Return(nil, errors.New("xxx"))

	_, err := NewFileCopier(nil, nil).Copy("ns", "pod", "app", "a.txt", "/tmp", nil, nil)
	assert.Equal(t, "xxx", err.Error())
}

func Test_fileCopier_Copy_Error2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	up := mock.NewMockUploader(m)
	app.EXPECT().Uploader().Return(up)
	app.EXPECT().LocalUploader().Return(up)
	app.EXPECT().Config().Return(&config.Config{UploadMaxSize: "1M"}).Times(2)
	fi := mock.NewMockFileInfo(m)
	up.EXPECT().Stat("a.txt").Return(fi, nil)
	fi.EXPECT().Size().Return(uint64(1024 * 1024 * 10)).Times(2)

	_, err := NewFileCopier(nil, nil).Copy("ns", "pod", "app", "a.txt", "/tmp", nil, nil)
	assert.Equal(t, "最大不得超过 1.0 MB, 你上传的文件大小是 10 MB", err.Error())
}

func Test_fileCopier_Copy_Error3(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	up := mock.NewMockUploader(m)
	app.EXPECT().Uploader().Return(up)
	app.EXPECT().LocalUploader().Return(up)
	app.EXPECT().Config().Return(&config.Config{UploadMaxSize: "1M"})
	fi := mock.NewMockFileInfo(m)
	up.EXPECT().Stat("a.txt").Return(fi, nil)
	fi.EXPECT().Size().Return(uint64(1))
	up.EXPECT().Type().Return(contracts.S3)
	up.EXPECT().Read("a.txt").Return(nil, errors.New("xxx"))

	_, err := NewFileCopier(nil, nil).Copy("ns", "pod", "app", "a.txt", "/tmp", nil, nil)
	assert.Equal(t, "xxx", err.Error())
}

func Test_fileCopier_Copy_Error4(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	up := mock.NewMockUploader(m)
	app.EXPECT().Uploader().Return(up)
	app.EXPECT().LocalUploader().Return(up)
	app.EXPECT().Config().Return(&config.Config{UploadMaxSize: "1M"})
	fi := mock.NewMockFileInfo(m)
	up.EXPECT().Stat("a.txt").Return(fi, nil)
	fi.EXPECT().Size().Return(uint64(1))
	up.EXPECT().Type().Return(contracts.S3)
	up.EXPECT().Exists("a.txt").Return(false)
	r := &readCloser{}
	up.EXPECT().Read("a.txt").Return(r, nil)
	up.EXPECT().Put("a.txt", r).Return(nil, errors.New("xxx"))

	_, err := NewFileCopier(nil, nil).Copy("ns", "pod", "app", "a.txt", "/tmp", nil, nil)
	assert.Equal(t, "xxx", err.Error())
}

func Test_fileCopier_Copy_Error5(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	up := mock.NewMockUploader(m)
	app.EXPECT().Uploader().Return(up)
	app.EXPECT().LocalUploader().Return(up)
	app.EXPECT().Config().Return(&config.Config{UploadMaxSize: "1M"})
	fi := mock.NewMockFileInfo(m)
	up.EXPECT().Stat("a.txt").Return(fi, nil)
	fi.EXPECT().Size().Return(uint64(1))
	up.EXPECT().Type().Return(contracts.Local)
	arch := mock.NewMockArchiver(m)
	arch.EXPECT().Archive(gomock.Any(), gomock.Any()).Return(errors.New("xxx"))

	_, err := NewFileCopier(nil, arch).Copy("ns", "pod", "app", "a.txt", "/tmp", nil, nil)
	assert.Equal(t, "xxx", err.Error())
}

func Test_fileCopier_Copy_Error6(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	up := mock.NewMockUploader(m)
	app.EXPECT().Uploader().Return(up)
	app.EXPECT().LocalUploader().Return(up)
	app.EXPECT().Config().Return(&config.Config{UploadMaxSize: "1M"})
	fi := mock.NewMockFileInfo(m)
	up.EXPECT().Stat("a.txt").Return(fi, nil)
	fi.EXPECT().Size().Return(uint64(1))
	up.EXPECT().Type().Return(contracts.Local)
	arch := mock.NewMockArchiver(m)
	arch.EXPECT().Archive(gomock.Any(), gomock.Any()).Return(nil)
	arch.EXPECT().Open("a.txt.tar.gz").Return(nil, errors.New("xxx"))
	arch.EXPECT().Remove("a.txt.tar.gz")

	_, err := NewFileCopier(nil, arch).Copy("ns", "pod", "app", "a.txt", "/tmp", nil, nil)
	assert.Equal(t, "xxx", err.Error())
}

type readCloser struct{}

func (r *readCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("closed")
}

func (r *readCloser) Close() error {
	return nil
}

func TestNewDefaultArchiver(t *testing.T) {
	assert.Implements(t, (*contracts.Archiver)(nil), NewDefaultArchiver())
}
