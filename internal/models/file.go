package models

import (
	"time"

	"github.com/dustin/go-humanize"
	"gorm.io/gorm"

	"github.com/duc-cnzj/mars/api/v4/types"
	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/utils/date"
)

type File struct {
	ID int `json:"id" gorm:"primaryKey;"`

	UploadType contracts.UploadType `json:"upload_type" gorm:"size:100;not null;default:'local'"`
	Path       string               `json:"path" gorm:"size:255;not null;comment:文件全路径"`
	Size       uint64               `json:"size" gorm:"not null;default:0;comment:文件大小"`
	Username   string               `json:"username" gorm:"size:255;not null;default:'';comment:用户名称"`

	Namespace     string `json:"namespace" gorm:"size:100;not null;default:'';"`
	Pod           string `json:"pod" gorm:"size:100;not null;default:'';"`
	Container     string `json:"container" gorm:"size:100;not null;default:'';"`
	ContainerPath string `json:"container_path" gorm:"size:255;not null;default:'';comment:容器中的文件路径"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

func (f *File) Uploader() (up contracts.Uploader) {
	switch f.UploadType {
	case app.Uploader().Type():
		up = app.Uploader()
	case app.LocalUploader().Type():
		up = app.LocalUploader()
	}
	return
}

func (f *File) DeleteFile() {
	if f.Path == "" {
		return
	}
	app.DB().Delete(f)
	if err := app.Uploader().Delete(f.Path); err != nil {
		mlog.Errorf("[File]: delete file err: '%s'", err.Error())
		return
	}
	mlog.Debugf("[File]: deleted '%s' ", f.Path)
}

func (f *File) ProtoTransform() *types.FileModel {
	return &types.FileModel{
		Id:             int64(f.ID),
		Path:           f.Path,
		Size:           int64(f.Size),
		Username:       f.Username,
		Namespace:      f.Namespace,
		Pod:            f.Pod,
		Container:      f.Container,
		Container_Path: f.ContainerPath,
		HumanizeSize:   humanize.Bytes(f.Size),
		CreatedAt:      date.ToRFC3339DatetimeString(&f.CreatedAt),
		UpdatedAt:      date.ToRFC3339DatetimeString(&f.UpdatedAt),
		DeletedAt:      date.ToRFC3339DatetimeString(&f.DeletedAt.Time),
	}
}
