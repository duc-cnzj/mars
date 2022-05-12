package socket

import (
	"io/fs"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	f := mock.NewMockFile(m)
	r := &Recorder{
		filepath: "/tmp/path",
		container: Container{
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
	}
	f.EXPECT().Stat().Times(0)
	r.Close()
	r.startTime = time.Now()
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	manager := mock.NewMockDBManager(m)
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s, _ := db.DB()
	defer s.Close()
	manager.EXPECT().DB().Return(db).AnyTimes()
	app.EXPECT().DBManager().Return(manager).AnyTimes()
	db.AutoMigrate(&models.File{}, &models.Event{})
	f.EXPECT().Stat().Times(1).Return(&mockFileInfo{size: 1}, nil)
	f.EXPECT().Close().Times(1)
	r.Close()
	fmodel := models.File{}
	db.First(&fmodel)
	emodel := models.Event{}
	db.First(&emodel)
	t.Log(emodel, fmodel)
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

	up := mock.NewMockUploader(m)
	up.EXPECT().Delete("aaa.txt").Times(1)
	app.EXPECT().Uploader().Return(up)
	f.EXPECT().Name().Return("aaa.txt")
	f.EXPECT().Stat().Return(&mockFileInfo{size: 0}, nil).Times(1)
	f.EXPECT().Close().Times(1)
	r.Close()
}

func TestRecorder_Write(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	f := mock.NewMockFile(m)
	r := &Recorder{
		filepath: "/tmp/path",
		container: Container{
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
	app := mock.NewMockApplicationInterface(m)
	instance.SetInstance(app)
	up := mock.NewMockUploader(m)
	up.EXPECT().Disk("shell").Times(1).Return(up)
	ff := mock.NewMockFile(m)
	ff.EXPECT().Name().Return("name")
	ff.EXPECT().Write(gomock.Any()).Times(1)
	up.EXPECT().NewFile(gomock.Any()).Times(1).Return(ff, nil)
	ff.EXPECT().WriteString(gomock.Any()).Times(2)
	app.EXPECT().Uploader().Return(up)
	r.Write("aaa")
	r.Write("bbb")
}
