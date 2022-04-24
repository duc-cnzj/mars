package models

import (
	"time"

	app "github.com/duc-cnzj/mars/internal/app/helper"

	"github.com/duc-cnzj/mars/internal/mlog"
	"gorm.io/gorm"
)

type File struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Path     string `json:"path" gorm:"size:255;not null;comment:文件全路径"`
	Size     uint64 `json:"size" gorm:"not null;default:0;comment:文件大小"`
	Username string `json:"username" gorm:"size:255;not null;default:'';comment:用户名称"`

	Namespace     string `json:"namespace" gorm:"size:100;not null;default:'';"`
	Pod           string `json:"pod" gorm:"size:100;not null;default:'';"`
	Container     string `json:"container" gorm:"size:100;not null;default:'';"`
	ContainerPath string `json:"container_path" gorm:"size:255;not null;default:'';comment:容器中的文件路径"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

func (f *File) DeleteFile() {
	if f.Path == "" {
		return
	}
	app.DB().Delete(f)
	if err := app.Uploader().Delete(f.Path); err != nil {
		mlog.Errorf("[File]: delete file err: '%s'", err.Error())
	}
	mlog.Debugf("[File]: deleted '%s' ", f.Path)
}
