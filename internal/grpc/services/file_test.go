package services

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/file"
	"github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
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

	_, err = new(File).Authorize(context.TODO(), "MaxUploadSize")
	assert.Nil(t, err)
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
	testutil.AssertAuditLogFired(m, app)
	_, err = new(File).Delete(adminCtx(), &file.DeleteRequest{Id: int64(f.ID)})
	assert.Nil(t, err)
}

func TestFile_DiskInfo(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{UploadDir: "/tmp"}).AnyTimes()
	up := mock.NewMockUploader(m)
	app.EXPECT().Uploader().Return(up).AnyTimes()
	up.EXPECT().DirSize().Return(int64(0), errors.New("")).Times(1)
	_, err := new(File).DiskInfo(adminCtx(), &file.DiskInfoRequest{})
	assert.Error(t, err)
	up.EXPECT().DirSize().Return(int64(100), nil).Times(1)
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

func TestFile_MaxUploadSize(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{UploadMaxSize: "10M"}).AnyTimes()
	size, err := new(File).MaxUploadSize(context.TODO(), &file.MaxUploadSizeRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "10 MB", size.HumanizeSize)
	assert.Equal(t, uint64(10*1000*1000), size.Bytes)
}
