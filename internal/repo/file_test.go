package repo

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/cache"
	"github.com/duc-cnzj/mars/v5/internal/uploader"

	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/schematype"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFileRepo_Create(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	db, _ := data.NewSqliteDB()
	defer db.Close()
	mockData.EXPECT().DB().Return(db).AnyTimes()
	mockData.EXPECT().Config().Return(&config.Config{UploadMaxSize: "10m"})
	repo := NewFileRepo(mlog.NewLogger(nil), mockData, nil, nil, timer.NewRealTimer())
	input := &CreateFileInput{
		Path:       "testPath",
		Username:   "testUser",
		Size:       100,
		UploadType: schematype.Local,
		Namespace:  "testNamespace",
		Pod:        "testPod",
		Container:  "testContainer",
	}
	file, err := repo.Create(context.TODO(), input)
	assert.Nil(t, err)
	assert.NotNil(t, file)
	assert.Equal(t, input.Path, file.Path)
	assert.Equal(t, input.Username, file.Username)
	assert.Equal(t, input.Size, file.Size)
	assert.Equal(t, input.UploadType, file.UploadType)
	assert.Equal(t, input.Namespace, file.Namespace)
	assert.Equal(t, input.Pod, file.Pod)
	assert.Equal(t, input.Container, file.Container)
}

func TestFileRepo_GetByID(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	db, _ := data.NewSqliteDB()
	defer db.Close()
	mockData.EXPECT().DB().Return(db).AnyTimes()
	mockData.EXPECT().Config().Return(&config.Config{UploadMaxSize: "10m"})
	repo := NewFileRepo(mlog.NewLogger(nil), mockData, nil, nil, timer.NewRealTimer())
	create, _ := repo.Create(context.TODO(), &CreateFileInput{
		Path:       "/",
		Username:   "aa",
		Size:       11,
		UploadType: "local",
		Namespace:  "ns",
		Pod:        "po",
		Container:  "c",
	})
	file, err := repo.GetByID(context.TODO(), create.ID)
	assert.Nil(t, err)
	assert.NotNil(t, file)
}

func TestFileRepo_Update(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	db, _ := data.NewSqliteDB()
	defer db.Close()
	mockData.EXPECT().DB().Return(db).AnyTimes()
	mockData.EXPECT().Config().Return(&config.Config{UploadMaxSize: "10m"})
	repo := NewFileRepo(mlog.NewLogger(nil), mockData, nil, nil, timer.NewRealTimer())

	create, _ := repo.Create(context.TODO(), &CreateFileInput{
		Path:       "/",
		Username:   "aa",
		Size:       11,
		UploadType: "local",
		Namespace:  "ns",
		Pod:        "po",
		Container:  "c",
	})

	input := &UpdateFileRequest{
		ID:            create.ID,
		ContainerPath: "testContainerPath",
		Namespace:     "testNamespace",
		Pod:           "testPod",
		Container:     "testContainer",
	}
	file, err := repo.Update(context.TODO(), input)
	assert.Nil(t, err)
	assert.NotNil(t, file)
	assert.Equal(t, input.ContainerPath, file.ContainerPath)
	assert.Equal(t, input.Namespace, file.Namespace)
	assert.Equal(t, input.Pod, file.Pod)
	assert.Equal(t, input.Container, file.Container)
}

func TestFileRepo_Delete(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	db, _ := data.NewSqliteDB()
	defer db.Close()
	mockData.EXPECT().DB().Return(db).AnyTimes()
	mockUploader := uploader.NewMockUploader(m)
	mockData.EXPECT().Config().Return(&config.Config{UploadMaxSize: "10m"})
	repo := NewFileRepo(mlog.NewLogger(nil), mockData, nil, mockUploader, timer.NewRealTimer())
	create, _ := repo.Create(context.TODO(), &CreateFileInput{
		Path:       "/",
		Username:   "aa",
		Size:       11,
		UploadType: "local",
		Namespace:  "ns",
		Pod:        "po",
		Container:  "c",
	})
	mockUploader.EXPECT().Delete("/")
	err := repo.Delete(context.TODO(), create.ID)
	assert.Nil(t, err)
}

func TestFileRepo_ShowRecords(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	db, _ := data.NewSqliteDB()
	defer db.Close()
	mockData.EXPECT().DB().Return(db).AnyTimes()
	mockData.EXPECT().Config().Return(&config.Config{UploadMaxSize: "10m"})
	mockUploader := uploader.NewMockUploader(m)

	repo := NewFileRepo(mlog.NewLogger(nil), mockData, nil, mockUploader, timer.NewRealTimer())

	create, _ := repo.Create(context.TODO(), &CreateFileInput{
		Path:       "/",
		Username:   "aa",
		Size:       11,
		UploadType: "local",
		Namespace:  "ns",
		Pod:        "po",
		Container:  "c",
	})
	localMockUploader := uploader.NewMockUploader(m)
	mockUploader.EXPECT().LocalUploader().Return(localMockUploader).Times(2)
	localMockUploader.EXPECT().Type().Return(schematype.Local)

	localMockUploader.EXPECT().Read(gomock.Any()).Return(&testReadCloser{}, nil)
	_, err := repo.ShowRecords(context.TODO(), create.ID)
	assert.Nil(t, err)
}

type testReadCloser struct {
	io.ReadCloser
}

func TestFileRepo_DiskInfo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	db, _ := data.NewSqliteDB()
	defer db.Close()
	mockData.EXPECT().DB().Return(db).AnyTimes()
	mockUploader := uploader.NewMockUploader(m)

	mockData.EXPECT().Config().Return(&config.Config{UploadMaxSize: "10m"})
	repo := NewFileRepo(mlog.NewLogger(nil), mockData, &cache.NoCache{}, mockUploader, timer.NewRealTimer())
	mockUploader.EXPECT().DirSize().Return(int64(100), nil)
	_, err := repo.DiskInfo(true)
	assert.Nil(t, err)
}

func TestFileRepo_StreamUploadFile(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	db, _ := data.NewSqliteDB()
	defer db.Close()
	mockData.EXPECT().DB().Return(db).AnyTimes()
	mockData.EXPECT().Config().Return(&config.Config{UploadMaxSize: "10m"})
	mockUploader := uploader.NewMockUploader(m)
	mockUploader.EXPECT().Disk(StreamUploadFileDisk).Return(mockUploader)
	mockUploader.EXPECT().AbsolutePath(gomock.Any()).Return("/abs-dir/aaa")
	mockUploader.EXPECT().MkDir("/abs-dir", true)
	mockFile := uploader.NewMockFile(m)
	mockUploader.EXPECT().NewFile("/abs-dir/aaa").Return(mockFile, nil)
	mockFile.EXPECT().Close()
	mockFile.EXPECT().Write(gomock.Any())
	mockFile.EXPECT().Stat().Return(&fakeOsFileInfo{}, nil)
	mockFile.EXPECT().Name().Return("testFile")
	mockUploader.EXPECT().Type().Return(schematype.UploadType("aa"))
	repo := NewFileRepo(mlog.NewLogger(nil), mockData, nil, mockUploader, timer.NewRealTimer())
	input := &StreamUploadFileRequest{
		Namespace: "testNamespace",
		Pod:       "testPod",
		Container: "testContainer",
		Username:  "testUser",
		FileName:  "testFile",
		FileData:  make(chan []byte),
	}
	go func() {
		input.FileData <- []byte("testData")
		close(input.FileData)
	}()
	file, err := repo.StreamUploadFile(context.TODO(), input)
	assert.Nil(t, err)
	assert.NotNil(t, file)
	assert.Equal(t, "aa", string(file.UploadType))
	assert.Equal(t, uint64(111), file.Size)
	assert.Equal(t, "testUser", file.Username)
	assert.Equal(t, "testNamespace", file.Namespace)
	assert.Equal(t, "testPod", file.Pod)
	assert.Equal(t, "testContainer", file.Container)
}

type fakeOsFileInfo struct {
	os.FileInfo
}

func (f *fakeOsFileInfo) Size() int64 {
	return 111
}
