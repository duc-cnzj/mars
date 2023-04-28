package contracts

//go:generate mockgen -destination ../mock/mock_uploader.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts Uploader
//go:generate mockgen -destination ../mock/mock_uploader_file.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts File
//go:generate mockgen -destination ../mock/mock_uploader_fileinfo.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts FileInfo

import (
	"io"
	"os"
	"time"
)

type UploadType string

const (
	Local UploadType = "local"
	S3    UploadType = "s3"
)

type File interface {
	io.ReadWriteCloser
	io.StringWriter
	Name() string
	Stat() (os.FileInfo, error)
	Seek(offset int64, whence int) (ret int64, err error)
}

type FileInfo interface {
	Path() string
	Size() uint64
	LastModified() time.Time
}

type Uploader interface {
	Disk(string) Uploader
	Type() UploadType
	DeleteDir(dir string) error
	DirSize() (int64, error)
	Delete(path string) error
	Exists(path string) bool
	MkDir(path string, recursive bool) error
	AbsolutePath(path string) string
	Put(path string, content io.Reader) (FileInfo, error)
	Read(string string) (io.ReadCloser, error)
	Stat(file string) (FileInfo, error)
	AllDirectoryFiles(dir string) ([]FileInfo, error)
	NewFile(path string) (File, error)
	RemoveEmptyDir() error
	UnWrap() Uploader
}
