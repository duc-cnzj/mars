package models

import (
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/testutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/utils/date"
	"github.com/dustin/go-humanize"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestFile_DeleteFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	app := mock.NewMockApplicationInterface(ctrl)
	dbManager := mock.NewMockDBManager(ctrl)
	defer ctrl.Finish()
	instance.SetInstance(app)
	sqlDB, _, _ := sqlmock.New()
	defer sqlDB.Close()
	gormDB, _ := gorm.Open(mysql.New(mysql.Config{SkipInitializeWithVersion: true, Conn: sqlDB}), &gorm.Config{})
	app.EXPECT().DBManager().Times(2).Return(dbManager)
	uploader := mock.NewMockUploader(ctrl)
	app.EXPECT().Uploader().Return(uploader).Times(2)
	uploader.EXPECT().Delete(gomock.Any()).Times(1)
	dbManager.EXPECT().DB().Return(gormDB).Times(2)
	m := File{
		ID:            1,
		Path:          "/filepath",
		Size:          1000,
		Username:      "duc",
		Namespace:     "devops",
		Pod:           "pod",
		Container:     "container",
		ContainerPath: "path",
		CreatedAt:     time.Time{},
		UpdatedAt:     time.Time{},
		DeletedAt:     gorm.DeletedAt{},
	}
	m.DeleteFile()

	(&File{}).DeleteFile()
	uploader.EXPECT().Delete(gomock.Any()).Times(1).Return(errors.New("xxx"))
	(&File{Path: "xxx", ID: 9999}).DeleteFile()
}

func TestFile_ProtoTransform(t *testing.T) {
	m := File{
		ID:            1,
		Path:          "/filepath",
		Size:          1000,
		Username:      "duc",
		Namespace:     "devops",
		Pod:           "pod",
		Container:     "container",
		ContainerPath: "path",
		CreatedAt:     time.Time{},
		UpdatedAt:     time.Time{},
		DeletedAt:     gorm.DeletedAt{},
	}
	assert.Equal(t, &types.FileModel{
		Id:             int64(m.ID),
		Path:           m.Path,
		Size:           int64(m.Size),
		Username:       m.Username,
		Namespace:      m.Namespace,
		Pod:            m.Pod,
		Container:      m.Container,
		Container_Path: m.ContainerPath,
		HumanizeSize:   humanize.Bytes(m.Size),
		CreatedAt:      date.ToRFC3339DatetimeString(&m.CreatedAt),
		UpdatedAt:      date.ToRFC3339DatetimeString(&m.UpdatedAt),
		DeletedAt:      date.ToRFC3339DatetimeString(&m.DeletedAt.Time),
	}, m.ProtoTransform())
}

func TestFile_Uploader(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	local := mock.NewMockUploader(m)
	up := mock.NewMockUploader(m)
	up.EXPECT().Type().Return(contracts.S3).AnyTimes()
	local.EXPECT().Type().Return(contracts.Local).AnyTimes()
	app.EXPECT().LocalUploader().Return(local).AnyTimes()
	app.EXPECT().Uploader().Return(up).AnyTimes()
	var tests = []struct {
		up contracts.Uploader
		f  File
	}{
		{
			up: local,
			f: File{
				UploadType: contracts.Local,
			},
		},
		{
			up: up,
			f: File{
				UploadType: contracts.S3,
			},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run("", func(t *testing.T) {
			assert.Same(t, tt.up, tt.f.Uploader())
		})
	}
}

func TestFile_Uploader1(t *testing.T) {
	type fields struct {
		ID            int
		UploadType    contracts.UploadType
		Path          string
		Size          uint64
		Username      string
		Namespace     string
		Pod           string
		Container     string
		ContainerPath string
		CreatedAt     time.Time
		UpdatedAt     time.Time
		DeletedAt     gorm.DeletedAt
	}
	tests := []struct {
		name   string
		fields fields
		wantUp contracts.Uploader
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &File{
				ID:            tt.fields.ID,
				UploadType:    tt.fields.UploadType,
				Path:          tt.fields.Path,
				Size:          tt.fields.Size,
				Username:      tt.fields.Username,
				Namespace:     tt.fields.Namespace,
				Pod:           tt.fields.Pod,
				Container:     tt.fields.Container,
				ContainerPath: tt.fields.ContainerPath,
				CreatedAt:     tt.fields.CreatedAt,
				UpdatedAt:     tt.fields.UpdatedAt,
				DeletedAt:     tt.fields.DeletedAt,
			}
			assert.Equalf(t, tt.wantUp, f.Uploader(), "Uploader()")
		})
	}
}
