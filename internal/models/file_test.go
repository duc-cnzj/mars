package models

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/utils/date"
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
