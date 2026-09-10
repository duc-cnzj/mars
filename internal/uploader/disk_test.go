package uploader

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
)

func TestNewUploader(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, err := NewUploader(cfg, logger, nil)
	assert.Nil(t, err)
	assert.Equal(t, testDir, up.(*diskUploader).rootDir)
	assert.Equal(t, "disk", up.Disk("disk").(*diskUploader).disk)
}

func TestFileInfo_Path(t *testing.T) {
	cfg := &config.Config{UploadDir: "/disk"}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)
	assert.Equal(t, "/disk/aaa", up.(*diskUploader).getPath("aaa"))
	assert.Equal(t, "/disk/aaa", up.(*diskUploader).getPath("/disk/aaa"))
}

func TestUploader_AbsolutePath(t *testing.T) {
	cfg := &config.Config{UploadDir: "/disk"}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)
	assert.Equal(t, "/disk/aaa", up.AbsolutePath("aaa"))
}

func TestUploader_Disk(t *testing.T) {
	cfg := &config.Config{UploadDir: "/disk"}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)
	assert.Equal(t, "/disk/aa", up.Disk("/aa").AbsolutePath("/"))
	disk := up.Disk("1").Disk("2").Disk("3")
	d := disk.(*diskUploader)
	assert.Equal(t, "/disk/1/2", d.rootDir)
	assert.Equal(t, "3", d.disk)
}

func TestUploader_root(t *testing.T) {
	cfg := &config.Config{UploadDir: "/disk"}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)
	assert.Equal(t, "/disk", up.(*diskUploader).rootDir)

	assert.Equal(t, "/tmp/xxx", (&diskUploader{rootDir: "/tmp/xxx"}).rootDir)
}

func TestFileInfo(t *testing.T) {
	assert.Equal(t, uint64(100), (&fileInfo{size: uint64(100)}).Size())
	assert.Equal(t, "/xxx", (&fileInfo{path: "/xxx"}).Path())
	n := time.Now()
	assert.Equal(t, n, (&fileInfo{lastModified: n}).LastModified())
}

func TestUploader_DeleteDir(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)
	assert.Error(t, up.DeleteDir("aaa"))
	assert.Nil(t, up.MkDir("aaa", true))
	assert.Nil(t, up.DeleteDir("aaa"))
}

func TestUploader_Delete(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)
	assert.Error(t, up.Delete("a.txt"))
	_, err := up.Put("a.txt", strings.NewReader("aaa"))
	assert.Nil(t, err)
	assert.Nil(t, up.Delete("a.txt"))
}

func TestUploader_DirSize(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)
	size, _ := up.DirSize()
	assert.Equal(t, int64(0), size)

	_ = up.MkDir("app", true)
	_, err := up.Put("/app/a.txt", strings.NewReader("xxx"))
	assert.Nil(t, err)
	_, err = up.Put("/app/ccc/a.txt", strings.NewReader("ccc"))
	assert.Nil(t, err)
	size, _ = up.DirSize()
	assert.Greater(t, size, int64(0))
}

func TestUploader_MkDir(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)
	assert.Nil(t, up.MkDir("/b/c", true))
	assert.Nil(t, up.MkDir("/d", false))
}

func TestUploader_DirExists(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)
	assert.Nil(t, up.MkDir("/b/c", true))

	assert.True(t, up.(*diskUploader).DirExists("/b/c"))
	assert.False(t, up.(*diskUploader).DirExists("/b/c/d"))

	up2, _ := NewUploader(cfg, logger, nil)
	assert.True(t, up2.(*diskUploader).DirExists(testDir+"/b/c"))
}

func TestUploader_RemoveEmptyDir(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)
	assert.Nil(t, up.MkDir("/b/c", true))

	assert.Nil(t, up.RemoveEmptyDir())
	assert.False(t, up.(*diskUploader).DirExists("/b/c"))
	assert.False(t, up.(*diskUploader).DirExists("/b"))
	assert.True(t, up.(*diskUploader).DirExists(""))
}

func TestUploader_AllDirectoryFiles(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)

	_ = up.DeleteDir("/")

	_, _ = up.Put("/a.txt", strings.NewReader("aa"))
	_, _ = up.Put("/b/b.txt", strings.NewReader("b"))
	_, _ = up.Put("/c/c/c.txt", strings.NewReader("c"))

	files, err := up.AllDirectoryFiles("")
	assert.Nil(t, err)
	assert.Len(t, files, 3)
}

func TestUploader_Put(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)
	put, err := up.Put("/aa/bb/cc/c.txt", strings.NewReader("aaa"))
	assert.Nil(t, err)
	assert.Greater(t, put.Size(), uint64(0))
	assert.Equal(t, filepath.Join(testDir, "aa/bb/cc/c.txt"), put.Path())
}

func TestUploader_NewFile(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)
	file, err := up.NewFile("/a/a/a/aaa.txt")
	assert.Nil(t, err)
	file.Close()
	assert.True(t, up.Exists("/a/a/a/aaa.txt"))
	_, err = up.NewFile("/a/a/a/aaa.txt")
	assert.Error(t, err)
}

func TestUploader_Type(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)
	assert.Equal(t, schematype.Local, up.Type())
}

func TestUploader_Read(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)
	put, err := up.Put("/aa/bb/cc/read.txt", strings.NewReader("aaa"))
	assert.Nil(t, err)
	defer func() { _ = up.Delete(put.Path()) }()
	read, err := up.Read(put.Path())
	assert.Nil(t, err)
	defer read.Close()
	all, err := io.ReadAll(read)
	assert.Nil(t, err)
	assert.Equal(t, "aaa", string(all))
}

func TestUploader_Stat(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)
	put, err := up.Put("/aa/bb/cc/stat.txt", strings.NewReader("aaa"))
	assert.Nil(t, err)
	stat, err := up.Stat(put.Path())
	assert.Nil(t, err)
	assert.Equal(t, uint64(3), stat.Size())
	assert.Equal(t, put.Path(), stat.Path())

	_, err = up.Stat("/aa/not-exist.file")
	assert.Error(t, err)

}

func Test_diskUploader_LocalUploader(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	localup := NewMockUploader(m)
	up := &diskUploader{
		localUploader: localup,
	}
	assert.Same(t, localup, up.LocalUploader())
}

// Test_diskUploader_DirSize_MissingRoot 回归：root 目录不存在时 DirSize 必须返回错误而非 panic。
// 修复前 filepath.Walk 以 err!=nil、info==nil 调用 walkFn，info.IsDir() 直接 nil-deref 崩溃。
func Test_diskUploader_DirSize_MissingRoot(t *testing.T) {
	up, err := NewDiskUploader("/nonexistent-dir-xyz", mlog.NewForConfig(nil))
	assert.NoError(t, err)

	size, err := up.DirSize()
	assert.Error(t, err)
	assert.Equal(t, int64(0), size)
}

// Test_diskUploader_RemoveEmptyDir_MissingRoot 回归：root 目录不存在时 RemoveEmptyDir 必须静默返回 nil。
// 修复前 WalkDir 回调 d==nil 时 d.IsDir() nil-deref 崩溃。
func Test_diskUploader_RemoveEmptyDir_MissingRoot(t *testing.T) {
	up, err := NewDiskUploader("/nonexistent-dir-xyz", mlog.NewForConfig(nil))
	assert.NoError(t, err)

	assert.NoError(t, up.RemoveEmptyDir())
}

// Test_diskUploader_AllDirectoryFiles_MissingRoot 回归：root 目录不存在时 AllDirectoryFiles 返回错误而非 panic。
func Test_diskUploader_AllDirectoryFiles_MissingRoot(t *testing.T) {
	up, err := NewDiskUploader("/nonexistent-dir-xyz", mlog.NewForConfig(nil))
	assert.NoError(t, err)

	files, err := up.AllDirectoryFiles("")
	assert.Error(t, err)
	assert.Nil(t, files)
}

// TestUploader_Delete_MissingFile 覆盖 Delete 对不存在文件返回错误的分支。
func TestUploader_Delete_MissingFile(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)
	assert.Error(t, up.Delete("/definitely-not-exist.txt"))
}

// TestUploader_Put_AlreadyExists 覆盖 Put 对已存在文件拒绝覆盖的分支。
func TestUploader_Put_AlreadyExists(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)

	_, err := up.Put("/exists.txt", strings.NewReader("a"))
	assert.NoError(t, err)
	_, err = up.Put("/exists.txt", strings.NewReader("b"))
	assert.Error(t, err)
}

// errorReader 是一个读到一半就报错的 io.Reader，用于覆盖 Put 的 io.Copy 失败分支。
type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("read boom") }

// TestUploader_Put_CopyError 覆盖 Put 中 io.Copy 失败的错误分支。
func TestUploader_Put_CopyError(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)

	_, err := up.Put("/copy-error.txt", errorReader{})
	assert.Error(t, err)
}

// TestUploader_MkDir_NonRecursive_ExistingDir 覆盖 MkDir 非递归模式对已存在目录返回错误的分支。
func TestUploader_MkDir_NonRecursive_ExistingDir(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)

	assert.NoError(t, up.MkDir("/dir-for-mkdir", false))
	assert.Error(t, up.MkDir("/dir-for-mkdir", false))
}

// TestUploader_Put_MkDirError 覆盖 Put 中 MkdirAll 失败的错误分支：
// 目标目录的父级是普通文件时，MkdirAll 报 ENOTDIR。
func TestUploader_Put_MkDirError(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)

	_, err := up.Put("/block.txt", strings.NewReader("x"))
	assert.NoError(t, err)
	_, err = up.Put("/block.txt/sub/a.txt", strings.NewReader("y"))
	assert.Error(t, err)
}

// TestUploader_Put_CreateError 覆盖 Put 中 os.Create 失败的错误分支：
// 文件名超过单组件长度上限时，Create 报 ENAMETOOLONG（且父目录已存在，
// 不会先命中 MkdirAll 分支）。
func TestUploader_Put_CreateError(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)

	assert.NoError(t, up.MkDir("/long", true))
	_, err := up.Put("/long/"+strings.Repeat("a", 300), strings.NewReader("x"))
	assert.Error(t, err)
}

// TestUploader_NewFile_MkDirError 覆盖 NewFile 中 MkdirAll 失败的错误分支：
// 目标目录的父级是普通文件时，MkdirAll 报 ENOTDIR。
func TestUploader_NewFile_MkDirError(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)

	_, err := up.Put("/block-new.txt", strings.NewReader("x"))
	assert.NoError(t, err)
	_, err = up.NewFile("/block-new.txt/sub/a.txt")
	assert.Error(t, err)
}

// TestUploader_RemoveEmptyDir_ReadDirError 覆盖 RemoveEmptyDir 里
// os.ReadDir 失败的错误分支：目录无读权限时 ReadDir 报 EACCES，
// 该目录被记录后由错误分支跳过（不 panic、不删除）。
func TestUploader_RemoveEmptyDir_ReadDirError(t *testing.T) {
	resetTestDir(t)
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewForConfig(nil)
	up, _ := NewUploader(cfg, logger, nil)

	assert.NoError(t, up.MkDir("/locked", true))
	locked := filepath.Join(testDir, "locked")
	assert.NoError(t, os.Chmod(locked, 0))
	defer func() { _ = os.Chmod(locked, 0755) }()

	assert.NoError(t, up.RemoveEmptyDir())
}

// TestNewDiskUploader_AbsError 覆盖 NewDiskUploader 中 filepath.Abs 失败的
// 错误分支（并透传到 NewUploader）：把 cwd 逐级下钻到绝对路径超过 getcwd 缓冲上限，
// getcwd 报 ENAMETOOLONG。chdir 按组件解析、不受全路径长度限制，故能走到深处；
// getcwd 必须缓冲完整路径，路径一超限必然失败。
//
// 取值约束（须同时满足，否则平台差异会放跑错误）：
//   - 路径须 > 4096：macOS PATH_MAX=1024、Linux PATH_MAX=4096，取较大的 4096。
//     每层目录名 20 字符（含分隔符 21），400 层 ≈8400 字符，双平台都压爆 getcwd。
//   - 深度须 > 342：os.Getwd 的 '..' 兜底重建算法在 len(parent)>=1024（约 342 层）
//     直接返回 ENAMETOOLONG，400 层远超，兜底也必然失败。
//     （此前 800 层单字符名仅 ≈1600 字符，压不爆 Linux 的 4096，CI 上红过；macOS 的
//     mkdir/chdir/RemoveAll 耗时随深度近似平方增长，故用长目录名压低深度。）
func TestNewDiskUploader_AbsError(t *testing.T) {
	orig, err := os.Getwd()
	assert.NoError(t, err)
	tmp, err := os.MkdirTemp("", "deep-cwd")
	assert.NoError(t, err)
	defer os.RemoveAll(tmp) // 深树删除：Go RemoveAll 用 fd 相对遍历，不受 PATH_MAX 限制
	assert.NoError(t, os.Chdir(tmp))
	defer func() { _ = os.Chdir(orig) }() // 先注册，确保任何失败路径都能恢复 cwd

	logger := mlog.NewForConfig(nil) // 提前构造 logger，避免在 cwd 失效时初始化

	// 20 字符目录名 × 400 层 → cwd 路径 ≈8400 字符，超过 Linux getcwd 缓冲 4096。
	deepDir := strings.Repeat("d", 20)
	for i := 0; i < 400; i++ {
		assert.NoError(t, os.Mkdir(deepDir, 0755))
		assert.NoError(t, os.Chdir(deepDir))
	}

	_, err = NewDiskUploader("relative-dir", nil)
	assert.Error(t, err)

	// NewUploader 把 NewDiskUploader 的失败原样透传。
	_, err = NewUploader(&config.Config{UploadDir: "relative-upload"}, logger, nil)
	assert.Error(t, err)
}
