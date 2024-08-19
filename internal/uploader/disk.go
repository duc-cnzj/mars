package uploader

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/ent/schema/schematype"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
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
	Type() schematype.UploadType
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
	LocalUploader() Uploader
}

type diskUploader struct {
	rootDir string
	disk    string
	logger  mlog.Logger

	localUploader Uploader
}

var _ Uploader = (*diskUploader)(nil)

func NewDiskUploader(rootDir string, logger mlog.Logger) (Uploader, error) {
	var err error

	if rootDir, err = filepath.Abs(rootDir); err != nil {
		return nil, err
	}

	up := &diskUploader{rootDir: rootDir, disk: "", logger: logger}
	up.localUploader = up
	return up, nil
}

func (u *diskUploader) LocalUploader() Uploader {
	return u.localUploader
}

func (u *diskUploader) getPath(path string) string {
	if strings.HasPrefix(path, u.root()) {
		return path
	}
	return filepath.Join(u.root(), path)
}

func (u *diskUploader) root() string {
	if u.disk != "" {
		return filepath.Join(u.rootDir, u.disk)
	}

	return u.rootDir
}

func (u *diskUploader) Type() schematype.UploadType {
	return schematype.Local
}

func (u *diskUploader) Disk(s string) Uploader {
	return &diskUploader{
		rootDir:       u.root(),
		disk:          s,
		logger:        u.logger,
		localUploader: u.localUploader,
	}
}

func (u *diskUploader) AbsolutePath(path string) string {
	return u.getPath(path)
}

func (u *diskUploader) DeleteDir(dir string) error {
	dir = u.getPath(dir)
	if !u.DirExists(dir) {
		return fmt.Errorf("dir not exists : '%s'", dir)
	}

	return os.RemoveAll(dir)
}

func (u *diskUploader) Delete(path string) error {
	return os.Remove(u.getPath(path))
}

func (u *diskUploader) DirSize() (int64, error) {
	var size int64
	dir := u.root()
	if err := filepath.Walk(u.getPath(dir), func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	}); err != nil {
		return 0, err
	}

	return size, nil
}

func (u *diskUploader) Exists(path string) bool {
	_, err := os.Stat(u.getPath(path))
	return err == nil
}

func (u *diskUploader) MkDir(path string, recursive bool) error {
	dir := u.getPath(path)
	if recursive {
		return os.MkdirAll(dir, 0750)
	}

	return os.Mkdir(dir, 0750)
}

func (u *diskUploader) DirExists(dir string) bool {
	return dirExists(u.getPath(dir))
}

func dirExists(dir string) bool {
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return true
	}
	return false
}

type fileInfo struct {
	path         string
	size         uint64
	lastModified time.Time
}

func NewFileInfo[T uint64 | int64 | int](path string, size T, lastModified time.Time) FileInfo {
	return &fileInfo{path: path, size: uint64(size), lastModified: lastModified}
}

func (f *fileInfo) Path() string {
	return f.path
}

func (f *fileInfo) Size() uint64 {
	return f.size
}

func (f *fileInfo) LastModified() time.Time {
	return f.lastModified
}

func (u *diskUploader) Read(file string) (io.ReadCloser, error) {
	return os.Open(u.getPath(file))
}

func (u *diskUploader) Stat(file string) (FileInfo, error) {
	fpath := u.getPath(file)
	stat, err := os.Stat(fpath)
	if err != nil {
		return nil, err
	}

	return NewFileInfo(fpath, stat.Size(), stat.ModTime()), nil
}

func (u *diskUploader) UnWrap() Uploader {
	return u
}

func (u *diskUploader) RemoveEmptyDir() error {
	var dirs []string
	dir := u.root()
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})
	sort.Sort(sort.Reverse(sort.StringSlice(dirs)))
	for _, root := range dirs {
		readDir, err := os.ReadDir(root)
		if err != nil {
			u.logger.Error(err)
			continue
		}
		if len(readDir) == 0 && root != u.getPath(dir) {
			os.Remove(root)
			u.logger.Debug("rm: ", root)
		}
	}
	return nil
}

func (u *diskUploader) AllDirectoryFiles(dir string) ([]FileInfo, error) {
	var files []FileInfo
	err := filepath.Walk(u.getPath(dir),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, NewFileInfo(path, info.Size(), info.ModTime()))
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (u *diskUploader) Put(path string, content io.Reader) (FileInfo, error) {
	fullpath := u.getPath(path)

	if u.Exists(fullpath) {
		return nil, fmt.Errorf("file already exist: '%s'", fullpath)
	}

	dir := filepath.Dir(fullpath)
	if !u.DirExists(dir) {
		if err := u.MkDir(dir, true); err != nil {
			return nil, err
		}
	}
	create, err := os.Create(fullpath)
	if err != nil {
		return nil, err
	}
	defer create.Close()
	if _, err := io.Copy(create, bufio.NewReaderSize(content, 4*1024*1024)); err != nil {
		return nil, err
	}
	stat, _ := create.Stat()

	return NewFileInfo(create.Name(), stat.Size(), stat.ModTime()), nil
}

func (u *diskUploader) NewFile(path string) (File, error) {
	fullpath := u.getPath(path)

	if u.Exists(fullpath) {
		return nil, fmt.Errorf("file already exist: '%s'", fullpath)
	}

	dir := filepath.Dir(fullpath)
	if !u.DirExists(dir) {
		if err := u.MkDir(dir, true); err != nil {
			return nil, err
		}
	}

	return os.Create(fullpath)
}
