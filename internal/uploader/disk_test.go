package uploader

import (
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/schematype"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
)

func TestNewUploader(t *testing.T) {
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewLogger(nil)
	up, err := NewUploader(cfg, logger, data.NewData(cfg, logger))
	assert.Nil(t, err)
	assert.Equal(t, testDir, up.(*diskUploader).rootDir)
	assert.Equal(t, "disk", up.Disk("disk").(*diskUploader).disk)
}

func TestFileInfo_Path(t *testing.T) {
	cfg := &config.Config{UploadDir: "/disk"}
	logger := mlog.NewLogger(nil)
	up, _ := NewUploader(cfg, logger, data.NewData(cfg, logger))
	assert.Equal(t, "/disk/aaa", up.(*diskUploader).getPath("aaa"))
	assert.Equal(t, "/disk/aaa", up.(*diskUploader).getPath("/disk/aaa"))
}

func TestUploader_AbsolutePath(t *testing.T) {
	cfg := &config.Config{UploadDir: "/disk"}
	logger := mlog.NewLogger(nil)
	up, _ := NewUploader(cfg, logger, data.NewData(cfg, logger))
	assert.Equal(t, "/disk/aaa", up.AbsolutePath("aaa"))
}

func TestUploader_Disk(t *testing.T) {
	cfg := &config.Config{UploadDir: "/disk"}
	logger := mlog.NewLogger(nil)
	up, _ := NewUploader(cfg, logger, data.NewData(cfg, logger))
	assert.Equal(t, "/disk/aa", up.Disk("/aa").AbsolutePath("/"))
	disk := up.Disk("1").Disk("2").Disk("3")
	d := disk.(*diskUploader)
	assert.Equal(t, "/disk/1/2", d.rootDir)
	assert.Equal(t, "3", d.disk)
}

func TestUploader_root(t *testing.T) {
	cfg := &config.Config{UploadDir: "/disk"}
	logger := mlog.NewLogger(nil)
	up, _ := NewUploader(cfg, logger, data.NewData(cfg, logger))
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
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewLogger(nil)
	up, _ := NewUploader(cfg, logger, data.NewData(cfg, logger))
	assert.Error(t, up.DeleteDir("aaa"))
	assert.Nil(t, up.MkDir("aaa", true))
	assert.Nil(t, up.DeleteDir("aaa"))
}

func TestUploader_Delete(t *testing.T) {
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewLogger(nil)
	up, _ := NewUploader(cfg, logger, data.NewData(cfg, logger))
	assert.Error(t, up.Delete("a.txt"))
	_, err := up.Put("a.txt", strings.NewReader("aaa"))
	assert.Nil(t, err)
	assert.Nil(t, up.Delete("a.txt"))
}

func TestUploader_DirSize(t *testing.T) {
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewLogger(nil)
	up, _ := NewUploader(cfg, logger, data.NewData(cfg, logger))
	size, _ := up.DirSize()
	assert.Equal(t, int64(0), size)

	up.MkDir("app", true)
	_, err := up.Put("/app/a.txt", strings.NewReader("xxx"))
	assert.Nil(t, err)
	_, err = up.Put("/app/ccc/a.txt", strings.NewReader("ccc"))
	assert.Nil(t, err)
	size, _ = up.DirSize()
	assert.Greater(t, size, int64(0))
}

func TestUploader_MkDir(t *testing.T) {
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewLogger(nil)
	up, _ := NewUploader(cfg, logger, data.NewData(cfg, logger))
	assert.Nil(t, up.MkDir("/b/c", true))
	assert.Nil(t, up.MkDir("/d", false))
}

func TestUploader_DirExists(t *testing.T) {
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewLogger(nil)
	up, _ := NewUploader(cfg, logger, data.NewData(cfg, logger))
	assert.Nil(t, up.MkDir("/b/c", true))

	assert.True(t, up.(*diskUploader).DirExists("/b/c"))
	assert.False(t, up.(*diskUploader).DirExists("/b/c/d"))

	up2, _ := NewUploader(cfg, logger, data.NewData(cfg, logger))
	assert.True(t, up2.(*diskUploader).DirExists(testDir+"/b/c"))
}

func TestUploader_RemoveEmptyDir(t *testing.T) {
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewLogger(nil)
	up, _ := NewUploader(cfg, logger, data.NewData(cfg, logger))
	assert.Nil(t, up.MkDir("/b/c", true))

	assert.Nil(t, up.RemoveEmptyDir())
	assert.False(t, up.(*diskUploader).DirExists("/b/c"))
	assert.False(t, up.(*diskUploader).DirExists("/b"))
	assert.True(t, up.(*diskUploader).DirExists(""))
}

func TestUploader_AllDirectoryFiles(t *testing.T) {
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewLogger(nil)
	up, _ := NewUploader(cfg, logger, data.NewData(cfg, logger))

	up.DeleteDir("/")

	up.Put("/a.txt", strings.NewReader("aa"))
	up.Put("/b/b.txt", strings.NewReader("b"))
	up.Put("/c/c/c.txt", strings.NewReader("c"))

	files, err := up.AllDirectoryFiles("")
	assert.Nil(t, err)
	assert.Len(t, files, 3)
}

func TestUploader_Put(t *testing.T) {
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewLogger(nil)
	up, _ := NewUploader(cfg, logger, data.NewData(cfg, logger))
	put, err := up.Put("/aa/bb/cc/c.txt", strings.NewReader("aaa"))
	assert.Nil(t, err)
	assert.Greater(t, put.Size(), uint64(0))
	assert.Equal(t, filepath.Join(testDir, "aa/bb/cc/c.txt"), put.Path())
}

func TestUploader_NewFile(t *testing.T) {
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewLogger(nil)
	up, _ := NewUploader(cfg, logger, data.NewData(cfg, logger))
	file, err := up.NewFile("/a/a/a/aaa.txt")
	assert.Nil(t, err)
	file.Close()
	assert.True(t, up.Exists("/a/a/a/aaa.txt"))
	_, err = up.NewFile("/a/a/a/aaa.txt")
	assert.Error(t, err)
}

func TestUploader_Type(t *testing.T) {
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewLogger(nil)
	up, _ := NewUploader(cfg, logger, data.NewData(cfg, logger))
	assert.Equal(t, schematype.Local, up.Type())
}

func TestUploader_Read(t *testing.T) {
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewLogger(nil)
	up, _ := NewUploader(cfg, logger, data.NewData(cfg, logger))
	put, err := up.Put("/aa/bb/cc/read.txt", strings.NewReader("aaa"))
	assert.Nil(t, err)
	defer up.Delete(put.Path())
	read, err := up.Read(put.Path())
	assert.Nil(t, err)
	defer read.Close()
	all, err := io.ReadAll(read)
	assert.Nil(t, err)
	assert.Equal(t, "aaa", string(all))
}

func TestUploader_Stat(t *testing.T) {
	cfg := &config.Config{UploadDir: testDir}
	logger := mlog.NewLogger(nil)
	up, _ := NewUploader(cfg, logger, data.NewData(cfg, logger))
	put, err := up.Put("/aa/bb/cc/stat.txt", strings.NewReader("aaa"))
	assert.Nil(t, err)
	stat, err := up.Stat(put.Path())
	assert.Nil(t, err)
	assert.Equal(t, uint64(3), stat.Size())
	assert.Equal(t, put.Path(), stat.Path())

	_, err = up.Stat("/aa/not-exist.file")
	assert.Error(t, err)

}

func Test_diskUploader_UnWrap(t *testing.T) {
	up := &diskUploader{}
	assert.Same(t, up, up.UnWrap())
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
