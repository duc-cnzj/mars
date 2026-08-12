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

	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

// File 是上传器创建的底层文件句柄，在 *os.File 能力之上多了 io.StringWriter。
type File interface {
	io.ReadWriteCloser
	io.StringWriter
	Name() string
	Stat() (os.FileInfo, error)
	Seek(offset int64, whence int) (ret int64, err error)
}

// FileInfo 描述上传器里一个文件的元信息。
type FileInfo interface {
	Path() string
	Size() uint64
	LastModified() time.Time
}

// Uploader 是统一的文件上传/管理抽象，本地磁盘与 S3 各有一个实现。
// 对外的关键区别：S3 实现把磁盘实现当本地镜像，写操作先落盘再同步到对象存储。
type Uploader interface {
	// Disk 返回以 s 为子目录的独立 Uploader，路径基于当前 root 下拼 s。
	Disk(s string) Uploader
	// Type 返回上传器的后端类型（本地磁盘 / S3）。
	Type() schematype.UploadType
	// DeleteDir 删除目录及其下所有内容。
	DeleteDir(dir string) error
	// DirSize 返回当前 root 下所有文件的总字节数。
	DirSize() (int64, error)
	// Delete 删除单个文件。
	Delete(path string) error
	// Exists 判断 path 是否已存在。
	Exists(path string) bool
	// MkDir 创建目录；recursive 为 true 时递归创建父目录。
	MkDir(path string, recursive bool) error
	// AbsolutePath 返回 path 对应的完整路径。
	AbsolutePath(path string) string
	// Put 把 content 写入 path；path 已存在时返回错误，不覆盖。
	Put(path string, content io.Reader) (FileInfo, error)
	// Read 打开 path 并返回可读流。
	Read(path string) (io.ReadCloser, error)
	// Stat 返回 path 的文件信息。
	Stat(file string) (FileInfo, error)
	// AllDirectoryFiles 返回 dir 下全部文件的文件信息列表。
	AllDirectoryFiles(dir string) ([]FileInfo, error)
	// NewFile 在 path 创建新文件；path 已存在时返回错误，不覆盖。
	NewFile(path string) (File, error)
	// RemoveEmptyDir 清理目录树里的空目录（自最深目录起）。
	RemoveEmptyDir() error
	// LocalUploader 返回对应的本地磁盘上传器：本地实现返回自身，
	// S3 实现返回其底层本地镜像。
	LocalUploader() Uploader
}

// diskUploader 是基于本地文件系统的 Uploader 实现。
type diskUploader struct {
	rootDir string
	disk    string
	logger  mlog.Logger

	localUploader Uploader
}

var _ Uploader = (*diskUploader)(nil)

// NewDiskUploader 以 rootDir 为根创建本地磁盘上传器，rootDir 会先转成绝对路径。
func NewDiskUploader(rootDir string, logger mlog.Logger) (Uploader, error) {
	var err error

	if rootDir, err = filepath.Abs(rootDir); err != nil {
		return nil, err
	}

	up := &diskUploader{rootDir: rootDir, disk: "", logger: logger}
	up.localUploader = up
	return up, nil
}

// LocalUploader 本地实现返回自身。
func (u *diskUploader) LocalUploader() Uploader {
	return u.localUploader
}

// getPath 把入参 path 拼接到当前 root 下；path 已是 root 前缀时原样返回。
func (u *diskUploader) getPath(path string) string {
	if strings.HasPrefix(path, u.root()) {
		return path
	}
	return filepath.Join(u.root(), path)
}

// root 返回当前生效的根目录：未切子目录时是 rootDir，否则是 rootDir/disk。
func (u *diskUploader) root() string {
	if u.disk != "" {
		return filepath.Join(u.rootDir, u.disk)
	}

	return u.rootDir
}

// Type 返回本地磁盘类型。
func (u *diskUploader) Type() schematype.UploadType {
	return schematype.Local
}

// Disk 返回以 s 为子目录的新上传器，子目录基于当前 root。
func (u *diskUploader) Disk(s string) Uploader {
	return &diskUploader{
		rootDir:       u.root(),
		disk:          s,
		logger:        u.logger,
		localUploader: u.localUploader,
	}
}

// AbsolutePath 返回 path 对应的完整路径。
func (u *diskUploader) AbsolutePath(path string) string {
	return u.getPath(path)
}

// DeleteDir 递归删除 dir；dir 不存在时返回错误。
func (u *diskUploader) DeleteDir(dir string) error {
	dir = u.getPath(dir)
	if !u.DirExists(dir) {
		return fmt.Errorf("dir not exists : '%s'", dir)
	}

	return os.RemoveAll(dir)
}

// Delete 删除单个文件，文件不存在时返回 os 错误。
func (u *diskUploader) Delete(path string) error {
	return os.Remove(u.getPath(path))
}

// DirSize 统计当前 root 下所有文件的总字节数。
func (u *diskUploader) DirSize() (int64, error) {
	var size int64
	dir := u.root()
	// Walk 在目录不存在或子路径不可读时，会以 err != nil、info == nil 调用 walkFn；
	// 必须先判 err 再访问 info，否则 nil-deref 崩溃（实测复现）。
	err := filepath.Walk(u.getPath(dir), func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return size, nil
}

// Exists 判断 path 是否已存在。
func (u *diskUploader) Exists(path string) bool {
	_, err := os.Stat(u.getPath(path))
	return err == nil
}

// MkDir 创建目录；recursive 为 true 时用 MkdirAll 递归创建父目录。
func (u *diskUploader) MkDir(path string, recursive bool) error {
	dir := u.getPath(path)
	if recursive {
		return os.MkdirAll(dir, 0750)
	}

	return os.Mkdir(dir, 0750)
}

// DirExists 判断 dir 是否是一个已存在的目录。
func (u *diskUploader) DirExists(dir string) bool {
	return dirExists(u.getPath(dir))
}

// dirExists 判断给定路径是否是一个已存在的目录。
func dirExists(dir string) bool {
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return true
	}
	return false
}

// fileInfo 是 FileInfo 接口的本地实现。
type fileInfo struct {
	path         string
	size         uint64
	lastModified time.Time
}

// newFileInfo 构造 FileInfo；size 兼容 uint64/int64/int 三种整数类型。
func newFileInfo[T uint64 | int64 | int](path string, size T, lastModified time.Time) FileInfo {
	return &fileInfo{path: path, size: uint64(size), lastModified: lastModified}
}

// Path 返回文件完整路径。
func (f *fileInfo) Path() string {
	return f.path
}

// Size 返回文件字节数。
func (f *fileInfo) Size() uint64 {
	return f.size
}

// LastModified 返回文件最后修改时间。
func (f *fileInfo) LastModified() time.Time {
	return f.lastModified
}

// Read 打开 file 并返回可读流。
func (u *diskUploader) Read(file string) (io.ReadCloser, error) {
	return os.Open(u.getPath(file))
}

// Stat 返回 file 的文件信息，文件不存在时返回 os 错误。
func (u *diskUploader) Stat(file string) (FileInfo, error) {
	fpath := u.getPath(file)
	stat, err := os.Stat(fpath)
	if err != nil {
		return nil, err
	}

	return newFileInfo(fpath, stat.Size(), stat.ModTime()), nil
}

// RemoveEmptyDir 清理目录树里的空目录（自最深目录起，保留根目录本身）。
func (u *diskUploader) RemoveEmptyDir() error {
	var dirs []string
	dir := u.root()
	// best-effort 清理：遍历出错（如 root 目录不存在）时跳过不可达路径，
	// 只清理能正常访问到的空目录。WalkDir 回调同样要判 err 再访问 d，
	// 否则目录缺失时 d == nil 会 nil-deref 崩溃（实测复现）。
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
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
			_ = os.Remove(root)
			u.logger.Debug("rm: ", root)
		}
	}
	return nil
}

// AllDirectoryFiles 返回 dir 下全部文件的文件信息列表。
func (u *diskUploader) AllDirectoryFiles(dir string) ([]FileInfo, error) {
	var files []FileInfo
	err := filepath.Walk(u.getPath(dir),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, newFileInfo(path, info.Size(), info.ModTime()))
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// Put 把 content 写入 path；path 已存在时返回错误，不覆盖。
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

	return newFileInfo(create.Name(), stat.Size(), stat.ModTime()), nil
}

// NewFile 在 path 创建新文件；path 已存在时返回错误，不覆盖。
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
