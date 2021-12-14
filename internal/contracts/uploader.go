package contracts

import (
	"io"
	"os"
)

type Uploader interface {
	Disk(string) Uploader
	DeleteDir(dir string) error
	Delete(path string) error
	Exists(path string) bool
	MkDir(path string, recursive bool) error
	AbsolutePath(path string) string
	Put(path string, content io.Reader) (*os.File, error)
}
