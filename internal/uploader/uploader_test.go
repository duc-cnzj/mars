package uploader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var pwd, _ = os.Getwd()
var testDir = filepath.Join(pwd, "testdir")

func TestMain(m *testing.M) {
	os.RemoveAll(testDir)
	d := DefaultRootDir
	DefaultRootDir = testDir
	exitCode := m.Run()
	DefaultRootDir = d
	os.RemoveAll(testDir)
	os.Exit(exitCode)
}

func TestNewUploader(t *testing.T) {
	uploader, err := NewUploader("/", "disk")
	assert.Nil(t, err)
	assert.Equal(t, "/", uploader.rootDir)
	assert.Equal(t, "disk", uploader.disk)

	_, err = os.Stat(DefaultRootDir)
	assert.True(t, os.IsNotExist(err))
	_, err = NewUploader("", "aaa")
	assert.Nil(t, err)
}

func TestFileInfo_Path(t *testing.T) {
	uploader, _ := NewUploader("/", "disk")
	assert.Equal(t, "/disk/aaa", uploader.getPath("aaa"))
	assert.Equal(t, "/disk/aaa", uploader.getPath("/disk/aaa"))
}

func TestUploader_AbsolutePath(t *testing.T) {
	uploader, _ := NewUploader("/", "disk")
	assert.Equal(t, "/disk/aaa", uploader.AbsolutePath("aaa"))
}

func TestUploader_Disk(t *testing.T) {
	uploader, _ := NewUploader("/", "disk")
	assert.Equal(t, "/disk/aa", uploader.Disk("/aa").AbsolutePath("/"))
	disk := uploader.Disk("1").Disk("2").Disk("3")
	d := disk.(*Uploader)
	assert.Equal(t, "/disk/1/2", d.rootDir)
	assert.Equal(t, "3", d.disk)
}

func TestUploader_root(t *testing.T) {
	uploader, _ := NewUploader("/", "disk")
	assert.Equal(t, "/disk", uploader.root())

	assert.Equal(t, "/tmp/xxx", (&Uploader{rootDir: "/tmp/xxx"}).root())
}

func TestFileInfo(t *testing.T) {
	assert.Equal(t, uint64(100), (&fileInfo{size: uint64(100)}).Size())
	assert.Equal(t, "/xxx", (&fileInfo{path: "/xxx"}).Path())
}

func TestUploader_DeleteDir(t *testing.T) {
	up, _ := NewUploader("", "aaa")
	assert.Error(t, up.DeleteDir("aaa"))
	assert.Nil(t, up.MkDir("aaa", true))
	assert.Nil(t, up.DeleteDir("aaa"))
}

func TestUploader_Delete(t *testing.T) {
	up, _ := NewUploader("", "aaa")
	assert.Error(t, up.Delete("a.txt"))
	_, err := up.Put("a.txt", strings.NewReader("aaa"))
	assert.Nil(t, err)
	assert.Nil(t, up.Delete("a.txt"))
}

func TestUploader_DirSize(t *testing.T) {
	up, _ := NewUploader("", "aaa")
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
	up, _ := NewUploader("", "aaa")
	assert.Nil(t, up.MkDir("/b/c", true))
	assert.Nil(t, up.MkDir("/d", false))
}

func TestUploader_DirExists(t *testing.T) {
	up, _ := NewUploader("", "aaa")
	assert.Nil(t, up.MkDir("/b/c", true))

	assert.True(t, up.DirExists("/b/c"))
	assert.False(t, up.DirExists("/b/c/d"))

	up2, _ := NewUploader("", "")
	assert.True(t, up2.DirExists("/aaa/b/c"))
}

func TestUploader_RemoveEmptyDir(t *testing.T) {
	up, _ := NewUploader("", "aaa")
	assert.Nil(t, up.MkDir("/b/c", true))

	assert.Nil(t, up.RemoveEmptyDir())
	assert.False(t, up.DirExists("/b/c"))
	assert.False(t, up.DirExists("/b"))
	assert.True(t, up.DirExists(""))
}

func TestUploader_AllDirectoryFiles(t *testing.T) {
	up, _ := NewUploader("", "ccc")

	up.Put("/a.txt", strings.NewReader("aa"))
	up.Put("/b/b.txt", strings.NewReader("b"))
	up.Put("/c/c/c.txt", strings.NewReader("c"))

	files, err := up.AllDirectoryFiles("")
	assert.Nil(t, err)
	assert.Len(t, files, 3)
}

func TestUploader_Put(t *testing.T) {
	up, _ := NewUploader("", "aaa")
	put, err := up.Put("/aa/bb/cc/c.txt", strings.NewReader("aaa"))
	assert.Nil(t, err)
	assert.Greater(t, put.Size(), uint64(0))
	assert.Equal(t, filepath.Join(DefaultRootDir, "aaa", "aa/bb/cc/c.txt"), put.Path())
}

func TestUploader_NewFile(t *testing.T) {
	up, _ := NewUploader("", "new_file")
	file, err := up.NewFile("/a/a/a/aaa.txt")
	assert.Nil(t, err)
	file.Close()
	assert.True(t, up.Exists("/a/a/a/aaa.txt"))
}
