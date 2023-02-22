package socket

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/internal/utils/timer"

	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

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
	app := testutil.MockApp(m)
	db, closeFn := testutil.SetGormDB(m, app)
	defer closeFn()
	db.AutoMigrate(&models.File{}, &models.Event{})
	up := mock.NewMockUploader(m)
	app.EXPECT().Uploader().Return(up).AnyTimes()
	app.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().Type().Return(contracts.Local).AnyTimes()

	f := mock.NewMockFile(m)
	r := &recorder{
		action: types.EventActionType_Shell,
		user:   contracts.UserInfo{Name: "duc"},
		timer:  timer.NewRealTimer(),
		container: contracts.Container{
			Namespace: "ns",
			Pod:       "po",
			Container: "c",
		},
		f:      f,
		buffer: bufio.NewWriter(f),
		rows:   25,
		cols:   106,
		shell:  "bash-x",
	}
	f.EXPECT().Stat().Times(0)
	r.Close()

	r.startTime = time.Now().Add(-2 * time.Second)
	up.EXPECT().Disk("shell").Return(up)
	upf := mock.NewMockFile(m)
	up.EXPECT().NewFile(gomock.Any()).Return(upf, nil)
	f.EXPECT().Seek(int64(0), int(0)).Times(1)
	upf.EXPECT().WriteString(fmt.Sprintf("{\"version\": 2, \"width\": 106, \"height\": 120, \"timestamp\": %d, \"env\": {\"SHELL\": \"bash-x\", \"TERM\": \"xterm-256color\"}}\n", r.startTime.Unix()))
	upf.EXPECT().Write(gomock.Any()).AnyTimes()
	f.EXPECT().Read(gomock.Any()).Times(1).Return(0, io.EOF)
	f.EXPECT().Close().Times(1)
	f.EXPECT().Name().Return("xxx.txt")
	up.EXPECT().Delete("xxx.txt").Times(1)
	upf.EXPECT().Stat().Times(1).Return(&mockFileInfo{size: 1}, nil)
	upf.EXPECT().Name().Return("aaa.txt")
	upf.EXPECT().Close().Times(1).Return(nil)

	r.Resize(90, 100)
	r.Resize(70, 10)
	r.Resize(70, 120)

	assert.Nil(t, r.Close())
	fmodel := models.File{}
	db.First(&fmodel)
	emodel := models.Event{}
	db.First(&emodel)

	findString := regexp.MustCompile(`\d+`).FindString(emodel.Duration)
	float, _ := strconv.ParseFloat(findString, 64)
	assert.GreaterOrEqual(t, float, float64(2))
	assert.Equal(t, "", emodel.New)
	assert.Equal(t, "duc", emodel.Username)
	assert.Equal(t, int(fmodel.ID), *emodel.FileID)
	assert.Equal(t, uint8(types.EventActionType_Shell), emodel.Action)

	assert.Equal(t, "duc", fmodel.Username)
	assert.Equal(t, "ns", fmodel.Namespace)
	assert.Equal(t, "aaa.txt", fmodel.Path)
	assert.Equal(t, uint64(1), fmodel.Size)
	assert.Equal(t, "c", fmodel.Container)
	assert.Equal(t, "po", fmodel.Pod)
}

func TestRecorder_Close2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeFn := testutil.SetGormDB(m, app)
	defer closeFn()
	db.AutoMigrate(&models.File{}, &models.Event{})
	up := mock.NewMockUploader(m)
	app.EXPECT().Uploader().Return(up).AnyTimes()
	app.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().Type().Return(contracts.Local).AnyTimes()

	f := mock.NewMockFile(m)
	r := &recorder{
		user:  contracts.UserInfo{Name: "duc"},
		timer: timer.NewRealTimer(),
		container: contracts.Container{
			Namespace: "ns",
			Pod:       "po",
			Container: "c",
		},
		f:      f,
		buffer: bufio.NewWriter(f),
		rows:   25,
		cols:   106,
		shell:  "bash-x",
	}
	f.EXPECT().Stat().Times(0)
	r.Close()

	r.startTime = time.Now().Add(-2 * time.Second)
	up.EXPECT().Disk("shell").Return(up)
	upf := mock.NewMockFile(m)
	up.EXPECT().NewFile(gomock.Any()).Return(upf, nil)
	f.EXPECT().Seek(int64(0), int(0)).Times(1)
	upf.EXPECT().WriteString(fmt.Sprintf("{\"version\": 2, \"width\": 106, \"height\": 25, \"timestamp\": %d, \"env\": {\"SHELL\": \"bash-x\", \"TERM\": \"xterm-256color\"}}\n", r.startTime.Unix()))
	upf.EXPECT().Write(gomock.Any()).AnyTimes()
	f.EXPECT().Read(gomock.Any()).Times(1).Return(0, io.EOF)
	f.EXPECT().Close().Times(1)
	f.EXPECT().Name().Return("xxx.txt")
	up.EXPECT().Delete("xxx.txt").Times(1)
	upf.EXPECT().Stat().Times(1).Return(&mockFileInfo{size: 0}, nil)
	upf.EXPECT().Name().Return("aaa.txt")
	upf.EXPECT().Close().Times(1).Return(nil)
	up.EXPECT().Delete("aaa.txt").Times(1)

	assert.Nil(t, r.Close())
}

func TestRecorder_Close3(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeFn := testutil.SetGormDB(m, app)
	defer closeFn()
	db.AutoMigrate(&models.File{}, &models.Event{})
	up := mock.NewMockUploader(m)
	app.EXPECT().Uploader().Return(up).AnyTimes()
	app.EXPECT().LocalUploader().Return(up).AnyTimes()
	up.EXPECT().Type().Return(contracts.Local).AnyTimes()

	f := mock.NewMockFile(m)
	r := &recorder{
		user:  contracts.UserInfo{Name: "duc"},
		timer: timer.NewRealTimer(),
		container: contracts.Container{
			Namespace: "ns",
			Pod:       "po",
			Container: "c",
		},
		f:      f,
		buffer: bufio.NewWriter(f),
		rows:   25,
		cols:   106,
		shell:  "bash-x",
	}
	f.EXPECT().Stat().Times(0)
	r.Close()

	r.startTime = time.Now().Add(-2 * time.Second)
	up.EXPECT().Disk("shell").Return(up)
	upf := mock.NewMockFile(m)
	up.EXPECT().NewFile(gomock.Any()).Return(upf, nil)
	f.EXPECT().Seek(int64(0), int(0)).Times(1)
	upf.EXPECT().WriteString(fmt.Sprintf("{\"version\": 2, \"width\": 106, \"height\": 25, \"timestamp\": %d, \"env\": {\"SHELL\": \"bash-x\", \"TERM\": \"xterm-256color\"}}\n", r.startTime.Unix()))
	upf.EXPECT().Write(gomock.Any()).AnyTimes()
	f.EXPECT().Read(gomock.Any()).Times(1).Return(0, io.EOF)
	f.EXPECT().Close().Times(1)
	f.EXPECT().Name().Return("xxx.txt")
	up.EXPECT().Delete("xxx.txt").Times(1)
	upf.EXPECT().Stat().Times(1).Return(nil, errors.New("xxx"))
	upf.EXPECT().Name().Return("aaa.txt")
	upf.EXPECT().Close().Times(1).Return(nil)
	up.EXPECT().Delete("aaa.txt").Times(1)

	assert.Equal(t, "xxx", r.Close().Error())
}

func TestRecorder_Write(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	f := mock.NewMockFile(m)
	r := &recorder{
		user:  contracts.UserInfo{Name: "duc"},
		timer: timer.NewRealTimer(),
		container: contracts.Container{
			Namespace: "ns",
			Pod:       "po",
			Container: "c",
		},
		f:         f,
		startTime: time.Time{},
		once:      sync.Once{},
	}
	app := testutil.MockApp(m)
	up := mock.NewMockUploader(m)
	up.EXPECT().Disk("tmp").Times(1).Return(up)
	ff := mock.NewMockFile(m)
	ff.EXPECT().Write(gomock.Any()).Times(0)
	up.EXPECT().NewFile(gomock.Any()).Times(1).Return(ff, nil)
	ff.EXPECT().WriteString(gomock.Any()).Times(0)
	app.EXPECT().LocalUploader().Return(up)
	r.Write("aaa")
	r.Write("bbb")
}

func TestRecorder_Write_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	f := mock.NewMockFile(m)
	r := &recorder{
		user:  contracts.UserInfo{Name: "duc"},
		timer: timer.NewRealTimer(),
		container: contracts.Container{
			Namespace: "ns",
			Pod:       "po",
			Container: "c",
		},
		f:         f,
		startTime: time.Time{},
		once:      sync.Once{},
	}
	up := mock.NewMockUploader(m)
	up.EXPECT().Disk("tmp").Times(1).Return(up)
	up.EXPECT().NewFile(gomock.Any()).Times(1).Return(nil, errors.New("xxx"))
	app := testutil.MockApp(m)
	app.EXPECT().LocalUploader().Return(up)

	err := r.Write("bbb")
	assert.Equal(t, "xxx", err.Error())
}

func TestRecorder_Resize(t *testing.T) {
	r := &recorder{}
	r.Resize(10, 20)
	r.Resize(20, 10)
	assert.Equal(t, uint16(20), r.rows)
	assert.Equal(t, uint16(20), r.cols)
}

func Test_realTimer_Now(t *testing.T) {
	assert.Equal(t, time.Now().Format("2006-01-02 15:04"), timer.NewRealTimer().Now().Format("2006-01-02 15:04"))
}

func TestRecorder_HeadLineColRow(t *testing.T) {
	r := &recorder{}
	r.HeadLineColRow(10, 20)
	r.HeadLineColRow(20, 10)
	assert.Equal(t, uint16(20), r.rows)
	assert.Equal(t, uint16(20), r.cols)
}

func Test_max(t *testing.T) {
	assert.Equal(t, 2, max(1, 2))
	assert.Equal(t, 3, max(3, 2))
	assert.Equal(t, 2, max(2, 2))
}

func TestNewRecorder(t *testing.T) {
	c := contracts.Container{
		Namespace: "ns",
		Pod:       "p",
		Container: "c",
	}
	r := NewRecorder(types.EventActionType_Exec, contracts.UserInfo{Name: "duc"}, timer.NewRealTimer(), c)
	assert.Equal(t, types.EventActionType_Exec, r.(*recorder).action)
	assert.Equal(t, contracts.UserInfo{Name: "duc"}, r.(*recorder).user)
	assert.NotNil(t, r.(*recorder).timer)
	assert.Equal(t, c, r.(*recorder).container)
}
