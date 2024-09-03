package repo

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v5/internal/auth"

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
	repo := NewFileRepo(mlog.NewForConfig(nil), mockData, nil, nil, timer.NewRealTimer())
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
	repo := NewFileRepo(mlog.NewForConfig(nil), mockData, nil, nil, timer.NewRealTimer())
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
	repo := NewFileRepo(mlog.NewForConfig(nil), mockData, nil, nil, timer.NewRealTimer())

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
	repo := NewFileRepo(mlog.NewForConfig(nil), mockData, nil, mockUploader, timer.NewRealTimer())
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

	repo := NewFileRepo(mlog.NewForConfig(nil), mockData, nil, mockUploader, timer.NewRealTimer())

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
	repo := NewFileRepo(mlog.NewForConfig(nil), mockData, &cache.NoCache{}, mockUploader, timer.NewRealTimer())
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
	repo := NewFileRepo(mlog.NewForConfig(nil), mockData, nil, mockUploader, timer.NewRealTimer())
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

func TestFileRepo_List_HappyPath(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	db, _ := data.NewSqliteDB()
	defer db.Close()

	db.File.Create().SetPath("/a").SetUsername("a").SetDeletedAt(time.Now()).SaveX(context.TODO())
	db.File.Create().SetPath("/v").SetUsername("v").SaveX(context.TODO())
	db.File.Create().SetPath("/c").SetUsername("c").SaveX(context.TODO())

	mockData.EXPECT().DB().Return(db).AnyTimes()
	repo := &fileRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	input := &ListFileInput{
		Page:           1,
		PageSize:       10,
		OrderIDDesc:    nil,
		WithSoftDelete: false,
	}
	files, pagination, err := repo.List(context.TODO(), input)
	assert.Nil(t, err)
	assert.NotNil(t, files)
	assert.NotNil(t, pagination)
	assert.Len(t, files, 2)
}

func TestFileRepo_List_WithSoftDelete(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	db, _ := data.NewSqliteDB()
	defer db.Close()
	db.File.Create().SetPath("/a").SetUsername("a").SetDeletedAt(time.Now()).SaveX(context.TODO())
	db.File.Create().SetPath("/v").SetUsername("v").SaveX(context.TODO())
	db.File.Create().SetPath("/c").SetUsername("c").SaveX(context.TODO())

	mockData.EXPECT().DB().Return(db).AnyTimes()
	repo := &fileRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	input := &ListFileInput{
		Page:           1,
		PageSize:       10,
		OrderIDDesc:    nil,
		WithSoftDelete: true,
	}
	files, pagination, err := repo.List(context.TODO(), input)
	assert.Nil(t, err)
	assert.NotNil(t, files)
	assert.NotNil(t, pagination)
	assert.Len(t, files, 3)
}

func TestFileRepo_List_OrderIDDesc(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockData := data.NewMockData(m)
	db, _ := data.NewSqliteDB()
	defer db.Close()
	db.File.Create().SetPath("/a").SetUsername("a").SetDeletedAt(time.Now()).SaveX(context.TODO())
	db.File.Create().SetPath("/v").SetUsername("v").SaveX(context.TODO())
	db.File.Create().SetPath("/c").SetUsername("c").SaveX(context.TODO())

	mockData.EXPECT().DB().Return(db).AnyTimes()
	repo := &fileRepo{
		logger: mlog.NewForConfig(nil),
		data:   mockData,
	}
	orderIDDesc := true
	input := &ListFileInput{
		Page:           1,
		PageSize:       10,
		OrderIDDesc:    &orderIDDesc,
		WithSoftDelete: false,
	}
	files, pagination, err := repo.List(context.TODO(), input)
	assert.Nil(t, err)
	assert.NotNil(t, files)
	assert.NotNil(t, pagination)
	assert.Equal(t, "/c", files[0].Path)
	assert.Equal(t, "/v", files[1].Path)
}

func Test_fileRepo_MaxUploadSize(t *testing.T) {
	assert.Equal(t, uint64(111), (&fileRepo{maxUploadSize: 111}).MaxUploadSize())
}

func Test_fileRepo_NewRecorder(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockUploader.EXPECT().LocalUploader().Return(mockUploader)
	newRecorder := (&fileRepo{
		logger:   mlog.NewForConfig(nil),
		timer:    timer.NewRealTimer(),
		uploader: mockUploader,
	}).NewRecorder(&auth.UserInfo{}, &Container{})
	assert.NotNil(t, newRecorder)
	assert.NotNil(t, newRecorder.(*recorder).logger)
	assert.NotNil(t, newRecorder.(*recorder).timer)
	assert.NotNil(t, newRecorder.(*recorder).uploader)
	assert.NotNil(t, newRecorder.(*recorder).localUploader)
	assert.NotNil(t, newRecorder.(*recorder).container)
	assert.NotNil(t, newRecorder.(*recorder).user)
	assert.NotNil(t, newRecorder.(*recorder).fileRepo)
}

func Test_fileRepo_NewFile(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	repo := &fileRepo{
		uploader: mockUploader,
	}
	mockUploader.EXPECT().NewFile("aa")
	repo.NewFile("aa")
}

func Test_fileRepo_NewDisk(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	repo := &fileRepo{
		uploader: mockUploader,
	}
	mockUploader.EXPECT().Disk("aa")
	repo.NewDisk("aa")
}

func Test_recorder_Container(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockUploader.EXPECT().LocalUploader()
	repo := &fileRepo{
		uploader: mockUploader,
	}
	r := repo.NewRecorder(nil, &Container{})
	assert.NotNil(t, r.Container())
}

func Test_recorder_User(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockUploader.EXPECT().LocalUploader()
	repo := &fileRepo{
		uploader: mockUploader,
	}
	r := repo.NewRecorder(&auth.UserInfo{}, nil)
	assert.NotNil(t, r.User())
}

func Test_recorder_File(t *testing.T) {
	r := &recorder{
		file: &File{},
	}
	assert.NotNil(t, r.File())
}

func Test_max(t *testing.T) {
	assert.Equal(t, 2, max(1, 2))
	assert.Equal(t, 3, max(3, 2))
}

func Test_recorder_GetShell(t *testing.T) {
	r := &recorder{
		shell: "bash",
	}
	assert.Equal(t, "bash", r.GetShell())
	r.SetShell("sh")
	assert.Equal(t, "sh", r.GetShell())
}

func Test_recorder_Resize(t *testing.T) {
	r := &recorder{
		rows: 1,
		cols: 1,
	}
	r.Resize(2, 2)
	assert.Equal(t, uint16(2), r.rows)
	assert.Equal(t, uint16(2), r.cols)

	r.HeadLineColRow(1, 2)
	assert.Equal(t, uint16(2), r.rows)
	assert.Equal(t, uint16(2), r.cols)
}

func TestRecorderWrite_WhenCalled_WritesDataToBuffer(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockFile := uploader.NewMockFile(m)
	mockUploader.EXPECT().LocalUploader().Return(mockUploader)
	mockUploader.EXPECT().NewFile(gomock.Any()).Return(mockFile, nil)
	mockUploader.EXPECT().Disk(gomock.Any()).Return(mockUploader)
	mockFile.EXPECT().Write(gomock.Any())

	repo := &fileRepo{
		uploader: mockUploader,
		timer:    timer.NewRealTimer(),
	}
	r := repo.NewRecorder(&auth.UserInfo{}, &Container{})
	l, err := r.Write([]byte("testData"))
	assert.Nil(t, err)
	assert.Equal(t, 8, l)
	r.(*recorder).buffer.Flush()
}

func TestRecorderWrite_WhenWriteFails_ReturnsError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockUploader := uploader.NewMockUploader(m)
	mockUploader.EXPECT().LocalUploader().Return(mockUploader)
	mockUploader.EXPECT().Disk(gomock.Any()).Return(mockUploader)
	mockUploader.EXPECT().NewFile(gomock.Any()).Return(nil, errors.New("error"))

	repo := &fileRepo{
		uploader: mockUploader,
		timer:    timer.NewRealTimer(),
	}
	r := repo.NewRecorder(&auth.UserInfo{}, &Container{})
	_, err := r.Write([]byte("testData"))
	assert.NotNil(t, err)
}

type mockFileInfo struct {
	size int64
}

func (m *mockFileInfo) Name() string {
	return ""
}

func (m *mockFileInfo) Size() int64 {
	return m.size
}

func (m *mockFileInfo) Mode() fs.FileMode {
	return fs.FileMode(0644)
}

func (m *mockFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (m *mockFileInfo) IsDir() bool {
	return false
}

func (m *mockFileInfo) Sys() any {
	return nil
}

func TestRecorder_Close(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	db, _ := data.NewSqliteDB()
	defer db.Close()

	mockData := data.NewMockData(m)
	mockData.EXPECT().DB().Return(db).AnyTimes()

	up := uploader.NewMockUploader(m)
	localup := uploader.NewMockUploader(m)
	up.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().Type().Return(schematype.Local).AnyTimes()

	f := uploader.NewMockFile(m)
	r := &recorder{
		user:   &auth.UserInfo{Name: "duc"},
		timer:  timer.NewRealTimer(),
		logger: mlog.NewForConfig(nil),
		container: &Container{
			Namespace: "ns",
			Pod:       "po",
			Container: "c",
		},
		f:             f,
		buffer:        bufio.NewWriter(f),
		rows:          25,
		cols:          106,
		shell:         "bash-x",
		fileRepo:      &fileRepo{data: mockData},
		uploader:      up,
		localUploader: localup,
	}
	f.EXPECT().Stat().Times(0)
	r.Close()

	r.startTime = time.Now().Add(-2 * time.Second)
	up.EXPECT().Disk("shell").Return(up)
	upf := uploader.NewMockFile(m)
	up.EXPECT().NewFile(gomock.Any()).Return(upf, nil)
	f.EXPECT().Seek(int64(0), int(0)).Times(1)
	upf.EXPECT().WriteString(fmt.Sprintf("{\"version\": 2, \"width\": 106, \"height\": 120, \"timestamp\": %d, \"env\": {\"SHELL\": \"bash-x\", \"TERM\": \"xterm-256color\"}}\n", r.startTime.Unix()))
	upf.EXPECT().Write(gomock.Any()).AnyTimes()
	f.EXPECT().Read(gomock.Any()).Times(1).Return(0, io.EOF)
	f.EXPECT().Close().Times(1)
	f.EXPECT().Name().Return("xxx.txt")
	localup.EXPECT().Delete("xxx.txt").Times(1)
	upf.EXPECT().Stat().Times(1).Return(&mockFileInfo{size: 1}, nil)
	upf.EXPECT().Name().Return("aaa.txt")
	upf.EXPECT().Close().Times(1).Return(nil)

	r.Resize(90, 100)
	r.Resize(70, 10)
	r.Resize(70, 120)

	assert.Nil(t, r.Close())
	fmodel, err := db.File.Query().First(context.TODO())
	assert.Nil(t, err)

	assert.Equal(t, "duc", fmodel.Username)
	assert.Equal(t, "ns", fmodel.Namespace)
	assert.Equal(t, "aaa.txt", fmodel.Path)
	assert.Equal(t, uint64(1), fmodel.Size)
	assert.Equal(t, "c", fmodel.Container)
	assert.Equal(t, "po", fmodel.Pod)
}
