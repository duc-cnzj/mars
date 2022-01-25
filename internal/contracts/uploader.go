package contracts

import (
	"io"
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
	RemoveEmptyDir(dir string) error
}
