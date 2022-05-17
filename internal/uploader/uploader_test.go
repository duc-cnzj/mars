package uploader

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUploader(t *testing.T) {
	uploader, err := NewUploader("/", "disk")
	assert.Nil(t, err)
	assert.Equal(t, "/", uploader.rootDir)
	assert.Equal(t, "disk", uploader.disk)
}

func TestFileInfo_Path(t *testing.T) {
	uploader, _ := NewUploader("/", "disk")
	assert.Equal(t, "/disk/aaa", uploader.getPath("aaa"))
}

func TestUploader_AbsolutePath(t *testing.T) {
	uploader, _ := NewUploader("/", "disk")
	assert.Equal(t, "/disk/aaa", uploader.AbsolutePath("aaa"))
}

func TestUploader_Disk(t *testing.T) {
	uploader, _ := NewUploader("/", "disk")
	assert.Equal(t, "/aa", uploader.Disk("aa").AbsolutePath("/"))
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
