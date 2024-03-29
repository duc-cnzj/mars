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

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

var DefaultRootDir = "/tmp/mars-uploads"

type diskUploader struct {
	rootDir string
	disk    string
}

func NewUploader(rootDir string, disk string) (contracts.Uploader, error) {
	var err error
	if rootDir == "" {
		rootDir = DefaultRootDir
		if !dirExists(rootDir) {
			if err := os.MkdirAll(rootDir, 0750); err != nil {
				return nil, err
			}
		}
		mlog.Warningf("rootDir not defined, use temp dir '%s'", rootDir)
	}

	if rootDir, err = filepath.Abs(rootDir); err != nil {
		return nil, err
	}

	return &diskUploader{rootDir: rootDir, disk: disk}, nil
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

func (u *diskUploader) Type() contracts.UploadType {
	return contracts.Local
}

func (u *diskUploader) Disk(s string) contracts.Uploader {
	return &diskUploader{
		rootDir: u.root(),
		disk:    s,
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

func NewFileInfo[T uint64 | int64 | int](path string, size T, lastModified time.Time) *fileInfo {
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

func (u *diskUploader) Stat(file string) (contracts.FileInfo, error) {
	fpath := u.getPath(file)
	stat, err := os.Stat(fpath)
	if err != nil {
		return nil, err
	}

	return NewFileInfo(fpath, stat.Size(), stat.ModTime()), nil
}

func (u *diskUploader) UnWrap() contracts.Uploader {
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
			mlog.Error(err)
			continue
		}
		if len(readDir) == 0 && root != u.getPath(dir) {
			os.Remove(root)
			mlog.Debug("rm: ", root)
		}
	}
	return nil
}

func (u *diskUploader) AllDirectoryFiles(dir string) ([]contracts.FileInfo, error) {
	var files []contracts.FileInfo
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

func (u *diskUploader) Put(path string, content io.Reader) (contracts.FileInfo, error) {
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

func (u *diskUploader) NewFile(path string) (contracts.File, error) {
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
