package contracts

//go:generate mockgen -destination ../mock/mock_uploader.go -package mock github.com/duc-cnzj/mars/internal/contracts Uploader

import (
	"io"
	"os"
)

type FileInfo interface {
	Path() string
	Size() uint64
}

type Uploader interface {
	Disk(string) Uploader
	DeleteDir(dir string) error
	DirSize(dir string) (int64, error)
	Delete(path string) error
	Exists(path string) bool
	MkDir(path string, recursive bool) error
	AbsolutePath(path string) string
	Put(path string, content io.Reader) (FileInfo, error)
	AllDirectoryFiles(dir string) ([]FileInfo, error)
	NewFile(path string) (*os.File, error)
	RemoveEmptyDir(dir string) error
}
