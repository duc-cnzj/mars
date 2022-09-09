package socket

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"sync"
	"testing"
	"time"

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
	up.EXPECT().Type().Return(contracts.Local).AnyTimes()

	f := mock.NewMockFile(m)
	r := &Recorder{
		timer:    realTimer{},
		filepath: "/tmp/path",
		container: contracts.Container{
			Namespace: "ns",
			Pod:       "po",
			Container: "c",
		},
		f:      f,
		buffer: bufio.NewWriter(f),
		t: &MyPtyHandler{
			conn: &WsConn{user: contracts.UserInfo{
				OpenIDClaims: contracts.OpenIDClaims{
					Name: "duc",
				},
			}},
		},
	}
	f.EXPECT().Stat().Times(0)
	r.Close()
	r.startTime = time.Now().Add(-2 * time.Second)
	r.currentStartTime = currentStart{t: r.startTime}

	f.EXPECT().Stat().Times(1).Return(&mockFileInfo{size: 1}, nil)
	f.EXPECT().Close().Times(1)
	r.timer = &fakeTimer{t: time.Now()}
	f.EXPECT().Write(gomock.Any()).Times(0)
	r.Resize(100, 100)
	r.timer = &fakeTimer{t: time.Now().Add(4 * time.Second)}
	f.EXPECT().Write(gomock.Any()).Times(1)
	r.Resize(100, 100)
	r.Close()
	fmodel := models.File{}
	db.First(&fmodel)
	emodel := models.Event{}
	db.First(&emodel)

	duration, _ := time.ParseDuration(emodel.Duration)
	assert.GreaterOrEqual(t, duration.Seconds(), float64(2))
	assert.Equal(t, "", emodel.New)
	assert.Equal(t, "duc", emodel.Username)
	assert.Equal(t, int(fmodel.ID), *emodel.FileID)
	assert.Equal(t, uint8(types.EventActionType_Shell), emodel.Action)

	assert.Equal(t, "duc", fmodel.Username)
	assert.Equal(t, "ns", fmodel.Namespace)
	assert.Equal(t, "/tmp/path", fmodel.Path)
	assert.Equal(t, uint64(1), fmodel.Size)
	assert.Equal(t, "c", fmodel.Container)
	assert.Equal(t, "po", fmodel.Pod)

	up.EXPECT().Delete("aaa.txt").Times(1)
	f.EXPECT().Name().Return("aaa.txt")
	f.EXPECT().Stat().Return(&mockFileInfo{size: 0}, nil).Times(1)
	f.EXPECT().Close().Times(1)
	r.Close()

	f.EXPECT().Stat().Return(nil, errors.New("xxx")).Times(1)
	assert.Equal(t, "xxx", r.Close().Error())
}

func TestRecorder_Write(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	f := mock.NewMockFile(m)
	r := &Recorder{
		timer:    realTimer{},
		filepath: "/tmp/path",
		container: contracts.Container{
			Namespace: "ns",
			Pod:       "po",
			Container: "c",
		},
		f:         f,
		startTime: time.Time{},
		t: &MyPtyHandler{
			conn: &WsConn{user: contracts.UserInfo{
				OpenIDClaims: contracts.OpenIDClaims{
					Name: "duc",
				},
			}},
		},
		once: sync.Once{},
	}
	app := testutil.MockApp(m)
	up := mock.NewMockUploader(m)
	up.EXPECT().Disk("shell").Times(1).Return(up)
	ff := mock.NewMockFile(m)
	ff.EXPECT().Name().Return("name")
	ff.EXPECT().Write(gomock.Any()).Times(0)
	up.EXPECT().NewFile(gomock.Any()).Times(1).Return(ff, nil)
	ff.EXPECT().WriteString(gomock.Any()).Times(0)
	app.EXPECT().Uploader().Return(up)
	r.Write("aaa")
	r.Write("bbb")
}

func TestRecorder_Write_Error(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	f := mock.NewMockFile(m)
	r := &Recorder{
		timer:    realTimer{},
		filepath: "/tmp/path",
		container: contracts.Container{
			Namespace: "ns",
			Pod:       "po",
			Container: "c",
		},
		f:         f,
		startTime: time.Time{},
		t: &MyPtyHandler{
			conn: &WsConn{user: contracts.UserInfo{
				OpenIDClaims: contracts.OpenIDClaims{
					Name: "duc",
				},
			}},
		},
		once: sync.Once{},
	}
	up := mock.NewMockUploader(m)
	up.EXPECT().Disk("shell").Times(1).Return(up)
	up.EXPECT().NewFile(gomock.Any()).Times(1).Return(nil, errors.New("xxx"))
	app := testutil.MockApp(m)
	app.EXPECT().Uploader().Return(up)

	err := r.Write("bbb")
	assert.Equal(t, "xxx", err.Error())
}

func Test_currentStart(t *testing.T) {
	var c = currentStart{}
	t1 := time.Now()
	c.Set(t1)
	assert.Equal(t, t1.String(), c.Get().String())
}

func TestRecorder_Resize(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	bf := &bytes.Buffer{}
	writer := bufio.NewWriter(bf)
	tnow := time.Date(2022, 9, 1, 1, 0, 0, 0, time.Local)
	r := &Recorder{
		startTime: tnow,
		timer:     &fakeTimer{t: tnow},
		shell:     "mock_bash",
		buffer:    writer,
	}
	t0 := tnow.Add(1 * time.Second)
	r.timer = &fakeTimer{t: t0}
	assert.Nil(t, r.Resize(100, 100))
	writer.Flush()
	assert.Equal(t, fmt.Sprintf(startLine, 100, 100, t0.Unix(), r.shell), bf.String())
	assert.NotZero(t, r.currentStartTime.Get())
	t1 := t0.Add(1 * time.Second)
	r.timer = &fakeTimer{t: t1}
	assert.Equal(t, ErrResizeTooFrequently, r.Resize(200, 200))
	t3 := t0.Add(4 * time.Second)
	r.timer = &fakeTimer{t: t3}
	assert.Equal(t, ErrResizeTooFrequently, r.Resize(200, 200))
	assert.True(t, t0.Equal(r.currentStartTime.Get()))
	t4 := t0.Add(6 * time.Second)
	r.timer = &fakeTimer{t: t4}
	assert.Nil(t, r.Resize(200, 200))
}

type fakeTimer struct {
	t time.Time
}

func (f *fakeTimer) Now() time.Time {
	return f.t
}

func Test_realTimer_Now(t *testing.T) {
	assert.Equal(t, time.Now().Format("2006-01-02 15:04"), realTimer{}.Now().Format("2006-01-02 15:04"))
}
