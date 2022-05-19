package services

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/file"
	"github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFile_Authorize(t *testing.T) {
	ctx := context.TODO()
	ctx = auth.SetUser(ctx, &contracts.UserInfo{
		Roles: []string{"admin"},
	})
	_, err := new(File).Authorize(ctx, "")
	assert.Nil(t, err)
	ctx = auth.SetUser(ctx, &contracts.UserInfo{
		Roles: []string{},
	})
	_, err = new(File).Authorize(ctx, "")
	assert.Error(t, err)
}

func adminCtx() context.Context {
	ctx := context.TODO()
	ctx = auth.SetUser(ctx, &contracts.UserInfo{
		Roles: []string{"admin"},
	})
	return ctx
}
func TestFile_Delete(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	_, err := new(File).Delete(adminCtx(), &file.DeleteRequest{Id: 1})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, fromError.Code())
	db.AutoMigrate(&models.File{})
	_, err = new(File).Delete(adminCtx(), &file.DeleteRequest{Id: 1})
	fromError, _ = status.FromError(err)
	assert.Equal(t, codes.NotFound, fromError.Code())
	f := &models.File{
		Path:     "/tmp/aa.txt",
		Size:     100,
		Username: "duc",
	}
	db.Create(f)
	up := mock.NewMockUploader(m)
	up.EXPECT().Delete("/tmp/aa.txt").Times(1)
	app.EXPECT().Uploader().Return(up)
	assertAuditLogFired(m, app)
	_, err = new(File).Delete(adminCtx(), &file.DeleteRequest{Id: int64(f.ID)})
	assert.Nil(t, err)
}
func TestFile_DeleteUndocumentedFiles(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	app.EXPECT().Config().Return(&config.Config{UploadDir: "/tmp"}).AnyTimes()
	db.AutoMigrate(&models.File{})
	db.Create(&models.File{
		Path: "/tmp/path1",
	})
	up := mock.NewMockUploader(m)

	finfo1 := mock.NewMockFileInfo(m)
	finfo1.EXPECT().Size().Return(uint64(1)).AnyTimes()
	finfo1.EXPECT().Path().Return("/tmp/path1").AnyTimes()

	finfo2 := mock.NewMockFileInfo(m)
	finfo2.EXPECT().Size().Return(uint64(2)).AnyTimes()
	finfo2.EXPECT().Path().Return("/tmp/path2").AnyTimes()

	up.EXPECT().AllDirectoryFiles("/tmp").Return([]contracts.FileInfo{finfo1, finfo2}, nil)
	app.EXPECT().Uploader().Return(up).AnyTimes()
	up.EXPECT().Delete("/tmp/path2").Times(1)
	up.EXPECT().RemoveEmptyDir("/tmp")
	assertAuditLogFired(m, app)
	res, _ := new(File).DeleteUndocumentedFiles(adminCtx(), &file.DeleteUndocumentedFilesRequest{})
	assert.Len(t, res.Items, 1)
}

func TestFile_DiskInfo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{UploadDir: "/tmp"}).AnyTimes()
	up := mock.NewMockUploader(m)
	app.EXPECT().Uploader().Return(up).AnyTimes()
	up.EXPECT().DirSize(gomock.Any()).Return(int64(0), errors.New("")).Times(1)
	_, err := new(File).DiskInfo(adminCtx(), &file.DiskInfoRequest{})
	assert.Error(t, err)
	up.EXPECT().DirSize(gomock.Any()).Return(int64(100), nil).Times(1)
	res, _ := new(File).DiskInfo(adminCtx(), &file.DiskInfoRequest{})
	assert.Equal(t, "100 B", res.HumanizeUsage)
	assert.Equal(t, int64(100), res.Usage)
}

func TestFile_List(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, closeDB := testutil.SetGormDB(m, app)
	defer closeDB()
	app.EXPECT().Config().Return(&config.Config{UploadDir: "/tmp"}).AnyTimes()
	_, err := new(File).List(context.TODO(), &file.ListRequest{
		Page:           1,
		PageSize:       15,
		WithoutDeleted: true,
	})
	assert.Error(t, err)
	db.AutoMigrate(&models.File{})
	db.Create(&models.File{
		Path:          "/p",
		Size:          100,
		Username:      "duc",
		Namespace:     "ns",
		Pod:           "pod",
		Container:     "c",
		ContainerPath: "/tnt.txt",
	})
	f2 := &models.File{
		Path:          "/p2",
		Size:          100,
		Username:      "duc2",
		Namespace:     "ns2",
		Pod:           "pod2",
		Container:     "c2",
		ContainerPath: "/tnt2.txt",
	}
	db.Create(f2)
	db.Delete(f2)

	list, _ := new(File).List(context.TODO(), &file.ListRequest{
		Page:           1,
		PageSize:       15,
		WithoutDeleted: true,
	})
	assert.Equal(t, int64(1), list.Count)
	list, _ = new(File).List(context.TODO(), &file.ListRequest{
		Page:           1,
		PageSize:       15,
		WithoutDeleted: false,
	})
	assert.Equal(t, int64(2), list.Count)
}

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

func assertAuditLogFired(m *gomock.Controller, app *mock.MockApplicationInterface) *mock.MockDispatcherInterface {
	e := mock.NewMockDispatcherInterface(m)
	e.EXPECT().Dispatch(events.EventAuditLog, gomock.Any()).Times(1)
	app.EXPECT().EventDispatcher().Return(e).AnyTimes()

	return e
}
