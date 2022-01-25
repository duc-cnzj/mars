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

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
)

const DefaultRootDir = "/tmp/mars-uploads"

type Uploader struct {
	rootDir string
	disk    string
}

func NewUploader(rootDir string, disk string) (*Uploader, error) {
	var err error
	if rootDir == "" {
		rootDir = DefaultRootDir
		if !dirExists(rootDir) {
			if err := os.MkdirAll(rootDir, 0755); err != nil {
				return nil, err
			}
		}
		mlog.Warningf("rootDir not defined, use temp dir '%s'", rootDir)
	}

	if rootDir, err = filepath.Abs(rootDir); err != nil {
		return nil, err
	}

	return &Uploader{rootDir: rootDir, disk: disk}, nil
}

func (u *Uploader) getPath(path string) string {
	if strings.HasPrefix(path, u.root()) {
		return path
	}
	return filepath.Join(u.root(), path)
}

func (u *Uploader) root() string {
	if u.disk != "" {
		return filepath.Join(u.rootDir, u.disk)
	}

	return u.rootDir
}

func (u *Uploader) Disk(s string) contracts.Uploader {
	return &Uploader{
		rootDir: u.rootDir,
		disk:    s,
	}
}

func (u *Uploader) AbsolutePath(path string) string {
	return u.getPath(path)
}

func (u *Uploader) DeleteDir(dir string) error {
	dir = u.getPath(dir)
	if !u.DirExists(dir) {
		return fmt.Errorf("dir not exists : '%s'", dir)
	}

	return os.RemoveAll(dir)
}

func (u *Uploader) Delete(path string) error {
	return os.Remove(u.getPath(path))
}

func (u *Uploader) DirSize(dir string) (int64, error) {
	var size int64
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

func (u *Uploader) Exists(path string) bool {
	_, err := os.Stat(u.getPath(path))
	if err != nil {
		return false
	}
	return true
}

func (u *Uploader) MkDir(path string, recursive bool) error {
	dir := u.getPath(path)
	if recursive {
		return os.MkdirAll(dir, 0755)
	}

	return os.Mkdir(dir, 0755)
}

func (u *Uploader) DirExists(dir string) bool {
	return dirExists(u.getPath(dir))
}

func dirExists(dir string) bool {
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return true
	}
	return false
}

type fileInfo struct {
	path string
	size uint64
}

func (f *fileInfo) Path() string {
	return f.path
}

func (f *fileInfo) Size() uint64 {
	return f.size
}

func (u *Uploader) RemoveEmptyDir(dir string) error {
	var dirs []string
	filepath.WalkDir(u.getPath(dir), func(path string, d fs.DirEntry, err error) error {
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

func (u *Uploader) AllDirectoryFiles(dir string) ([]contracts.FileInfo, error) {
	var files []contracts.FileInfo
	err := filepath.Walk(u.getPath(dir),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, &fileInfo{
					path: path,
					size: uint64(info.Size()),
				})
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (u *Uploader) Put(path string, content io.Reader) (contracts.FileInfo, error) {
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

	return &fileInfo{
		path: create.Name(),
		size: uint64(stat.Size()),
	}, nil
}
