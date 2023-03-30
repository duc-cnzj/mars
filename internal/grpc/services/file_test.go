package services

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/file"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

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
	_, err := new(fileSvc).Authorize(ctx, "")
	assert.Nil(t, err)
	ctx = auth.SetUser(ctx, &contracts.UserInfo{
		Roles: []string{},
	})
	_, err = new(fileSvc).Authorize(ctx, "")
	assert.Error(t, err)

	_, err = new(fileSvc).Authorize(context.TODO(), "MaxUploadSize")
	assert.Nil(t, err)

	_, err = new(fileSvc).Authorize(context.TODO(), "/file.fileSvc/MaxUploadSize")
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
	_, err := new(fileSvc).Delete(adminCtx(), &file.DeleteRequest{Id: 1})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, fromError.Code())
	db.AutoMigrate(&models.File{})
	_, err = new(fileSvc).Delete(adminCtx(), &file.DeleteRequest{Id: 1})
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
	_, err = new(fileSvc).Delete(adminCtx(), &file.DeleteRequest{Id: int64(f.ID)})
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
	_, err := new(fileSvc).DiskInfo(adminCtx(), &file.DiskInfoRequest{})
	assert.Error(t, err)
	up.EXPECT().DirSize().Return(int64(100), nil).Times(1)
	res, _ := new(fileSvc).DiskInfo(adminCtx(), &file.DiskInfoRequest{})
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
	_, err := new(fileSvc).List(context.TODO(), &file.ListRequest{
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

	list, _ := new(fileSvc).List(context.TODO(), &file.ListRequest{
		Page:           1,
		PageSize:       15,
		WithoutDeleted: true,
	})
	assert.Equal(t, int64(1), list.Count)
	list, _ = new(fileSvc).List(context.TODO(), &file.ListRequest{
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
	size, err := new(fileSvc).MaxUploadSize(context.TODO(), &file.MaxUploadSizeRequest{})
	assert.Nil(t, err)
	assert.Equal(t, "10 MB", size.HumanizeSize)
	assert.Equal(t, uint64(10*1000*1000), size.Bytes)
}

func TestFile_Show(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	db, fn := testutil.SetGormDB(m, app)
	defer fn()
	_, err := new(fileSvc).ShowRecords(context.TODO(), &file.ShowRecordsRequest{
		Id: 1,
	})
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, fromError.Code())
	db.AutoMigrate(&models.File{})
	_, err = new(fileSvc).ShowRecords(context.TODO(), &file.ShowRecordsRequest{
		Id: 1000,
	})
	fromError, _ = status.FromError(err)
	assert.Equal(t, codes.NotFound, fromError.Code())
	f := &models.File{
		UploadType: "local",
		Path:       "/tmp/x.txt",
		Size:       10,
	}
	db.Create(f)
	up := mock.NewMockUploader(m)
	localUp := mock.NewMockUploader(m)
	up.EXPECT().Type().Return(contracts.S3)
	localUp.EXPECT().Type().Return(contracts.Local)
	app.EXPECT().Uploader().Return(up).Times(1)
	app.EXPECT().LocalUploader().Return(localUp).Times(2)
	localUp.EXPECT().Read("/tmp/x.txt").Return(nil, errors.New("xxx"))
	_, err = new(fileSvc).ShowRecords(context.TODO(), &file.ShowRecordsRequest{
		Id: int64(f.ID),
	})
	assert.Equal(t, "xxx", err.Error())

	up.EXPECT().Type().Return(contracts.S3)
	localUp.EXPECT().Type().Return(contracts.Local)
	app.EXPECT().Uploader().Return(up).Times(1)
	app.EXPECT().LocalUploader().Return(localUp).Times(2)
	localUp.EXPECT().Read("/tmp/x.txt").Return(&rc{Reader: strings.NewReader("abc")}, nil)

	res, err := new(fileSvc).ShowRecords(context.TODO(), &file.ShowRecordsRequest{
		Id: int64(f.ID),
	})
	assert.Nil(t, err)
	assert.Equal(t, []string{"abc"}, res.Items)
}

type rc struct {
	io.Reader
}

func (r *rc) Read(p []byte) (n int, err error) {
	return r.Reader.Read(p)
}

func (r *rc) Close() error {
	return nil
}

func Test_transformToRecords(t *testing.T) {
	var tests = []struct {
		reader io.Reader
		wants  []string
	}{
		{
			reader: strings.NewReader(`aaaaa`),
			wants:  []string{"aaaaa"},
		},
		{
			reader: strings.NewReader(strings.TrimSpace(`
{"version": 2, }
[aaa]
[bbb]
{"version": 2, }
[ccc]
[ddd]
`)),
			wants: []string{"{\"version\": 2, }\n[aaa]\n[bbb]", "{\"version\": 2, }\n[ccc]\n[ddd]"},
		},
		{
			reader: strings.NewReader(strings.TrimSpace(`
{"version": 2, }
[aaa]
[bbb]{"version": 2, }
[ccc]
[ddd]
`)),
			wants: []string{"{\"version\": 2, }\n[aaa]\n[bbb]{\"version\": 2, }\n[ccc]\n[ddd]"},
		},
		{
			reader: strings.NewReader(strings.TrimSpace(`
{"version": 2, }
[aaa]
[bbb\n]{"version": 2, }
[ccc]
[ddd]
`)),
			wants: []string{"{\"version\": 2, }\n[aaa]\n[bbb\\n]{\"version\": 2, }\n[ccc]\n[ddd]"},
		},
		{
			reader: strings.NewReader("{\"version\": 2, }\n\n\n\n\n[aaa]"),
			wants:  []string{"{\"version\": 2, }\n[aaa]"},
		},
		{
			reader: strings.NewReader("{\"version\": 2, }\na\na\nb\n\n[aaa]"),
			wants:  []string{"{\"version\": 2, }\na\na\nb\n[aaa]"},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run("", func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wants, transformToRecords(tt.reader))
		})
	}
}
