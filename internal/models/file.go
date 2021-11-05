package models

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	app "github.com/duc-cnzj/mars/internal/app/helper"

	"github.com/duc-cnzj/mars/internal/mlog"
	"gorm.io/gorm"
)

type File struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Path string `json:"path" gorm:"size:255;not null;comment:文件全路径"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

func (f *File) DeleteFile() {
	if f.Path == "" {
		return
	}
	dir := filepath.Dir(f.Path)
	if strings.HasPrefix(dir, "/tmp") {
		os.RemoveAll(dir)
		app.DB().Delete(f)
		mlog.Debug("[File]: remove all: " + dir)
	}
}
