package uploader

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewS3(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	assert.Implements(t, (*contracts.Uploader)(nil), NewS3(nil, "bkt", nil, "root"))
	assert.Equal(t, "root", NewS3(nil, "bkt", nil, "root").(*s3Uploader).rootDir)
	assert.Equal(t, "data", NewS3(nil, "bkt", nil, "").(*s3Uploader).rootDir)
}

func TestS3_AbsolutePath(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	m := gomock.NewController(t)
	defer m.Finish()
	localUp := mock.NewMockUploader(m)
	localUp.EXPECT().Disk(gomock.Any()).AnyTimes().Return(localUp)
	assert.Equal(t, "root/aaa", NewS3(nil, "bkt", localUp, "root").AbsolutePath("aaa"))
	assert.Equal(t, "root/a/b/aaa", NewS3(nil, "bkt", localUp, "root").Disk("a").Disk("b").AbsolutePath("aaa"))
}

func TestS3_AllDirectoryFiles(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up, _ := NewUploader("", "")
	s3Cli := NewS3(s3Client, testBucketName, up, "AllDirectoryFiles").Disk("sub")
	s3Cli.DeleteDir("")
	s3Cli.Put("aaa", strings.NewReader("aaa"))
	s3Cli.Put("bbb/bbb", strings.NewReader("bbb"))
	s3Cli.Put("ccc/ccc/ccc", strings.NewReader("ccc"))
	files, _ := s3Cli.AllDirectoryFiles("")
	assert.Len(t, files, 3)
	mm := make(map[string]struct{})
	for _, file := range files {
		mm[file.Path()] = struct{}{}
	}
	_, ok := mm["AllDirectoryFiles/sub/aaa"]
	assert.True(t, ok)
	_, ok = mm["AllDirectoryFiles/sub/bbb/bbb"]
	assert.True(t, ok)
	_, ok = mm["AllDirectoryFiles/sub/ccc/ccc/ccc"]
	assert.True(t, ok)
}

func TestS3_Delete(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up, _ := NewUploader("", "")
	s3Cli := NewS3(s3Client, testBucketName, up, "")
	s3Cli.DeleteDir("")
	s3Cli.Put("aaa", strings.NewReader("aaa"))
	assert.True(t, s3Cli.Exists("aaa"))
	s3Cli.Delete("aaa")
	assert.False(t, s3Cli.Exists("aaa"))
}

func TestS3_DeleteDir(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up, _ := NewUploader("", "")
	s3Cli := NewS3(s3Client, testBucketName, up, "")
	s3Cli.DeleteDir("")
	s3Cli.Put("cc/c.txt", strings.NewReader("aaa"))
	assert.True(t, s3Cli.Exists("cc/c.txt"))
	t.Log(s3Cli.DeleteDir(""))
	assert.False(t, s3Cli.Exists("cc/c.txt"))
}

func TestS3_DirSize(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up, _ := NewUploader("", "")
	s3Cli := NewS3(s3Client, testBucketName, up, "dirsize")
	s3Cli.DeleteDir("")
	s3Cli.Put("dirsize/cc/c.txt", strings.NewReader("aaa"))
	size, _ := s3Cli.DirSize()
	assert.Equal(t, int64(3), size)
}

func TestS3_Disk(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	m := gomock.NewController(t)
	defer m.Finish()
	localUp := mock.NewMockUploader(m)
	localUp.EXPECT().Disk("aa").Times(1).Return(localUp)
	localUp.EXPECT().Disk("bb").Times(1).Return(localUp)
	up := NewS3(nil, "bkt", localUp, "root")
	disk := up.Disk("aa")
	assert.Equal(t, "root/aa", disk.(*s3Uploader).root())
	c := disk.Disk("bb")
	assert.Equal(t, "root/aa/bb", c.(*s3Uploader).root())
}

func TestS3_Exists(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up, _ := NewUploader("", "")
	s3Cli := NewS3(s3Client, testBucketName, up, "data")
	s3Cli.DeleteDir("")
	s3Cli.Put("cc/c.txt", strings.NewReader("aaa"))
	assert.True(t, s3Cli.Exists("cc/c.txt"))
	assert.True(t, s3Cli.Exists("data/cc/c.txt"))
}

func TestS3_MkDir(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up := NewS3(nil, "bkt", nil, "root")

	assert.Nil(t, up.MkDir("", true))
}

func TestS3_NewFile(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up, _ := NewUploader("", "")
	s3Cli := NewS3(s3Client, testBucketName, up, "")
	s3Cli.DeleteDir("")
	_, err := s3Cli.Put("aaa", strings.NewReader("aaa"))
	assert.Nil(t, err)
	file, err := s3Cli.NewFile("bb/cc.txt")
	assert.Nil(t, err)
	_, err = file.WriteString("666")
	assert.Nil(t, err)
	assert.Nil(t, file.Close())
	assert.True(t, s3Cli.Exists("bb/cc.txt"))
	read, _ := s3Cli.Read("bb/cc.txt")
	all, _ := io.ReadAll(read)
	assert.Equal(t, "666", string(all))
}

func TestS3_Put(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up, _ := NewUploader("", "")
	s3Cli := NewS3(s3Client, testBucketName, up, "")
	s3Cli.DeleteDir("")
	put, err := s3Cli.Put("aaa", strings.NewReader("aaa"))
	assert.Nil(t, err)
	assert.Equal(t, s3Cli.(*s3Uploader).getPath("aaa"), put.Path())

	diskOne := s3Cli.Disk("one")
	info, _ := diskOne.Put("one.txt", strings.NewReader("one"))
	assert.Equal(t, "data/one/one.txt", info.Path())

	diskOneTwo := diskOne.Disk("two")
	two, _ := diskOneTwo.Put("two.txt", strings.NewReader("two"))
	assert.Equal(t, "data/one/two/two.txt", two.Path())
}

func TestS3_Read(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up, _ := NewUploader("", "")
	s3Cli := NewS3(s3Client, testBucketName, up, "")
	s3Cli.DeleteDir("")
	_, err := s3Cli.Read("aaa")
	assert.Error(t, err)
	_, err = s3Cli.Put("aaa", strings.NewReader("aaa"))
	assert.Nil(t, err)
	read, err := s3Cli.Read("aaa")
	assert.Nil(t, err)
	defer read.Close()
	all, _ := io.ReadAll(read)
	assert.Equal(t, "aaa", string(all))
}

func TestS3_RemoveEmptyDir(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up, _ := NewUploader("", "")
	s3Cli := NewS3(s3Client, testBucketName, up, "")
	s3Cli.DeleteDir("")
	_, err := s3Cli.Put("aaa", strings.NewReader("aaa"))
	assert.Nil(t, err)
	assert.Nil(t, s3Cli.RemoveEmptyDir())
}

func TestS3_Stat(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up, _ := NewUploader("", "")
	s3Cli := NewS3(s3Client, testBucketName, up, "")
	s3Cli.DeleteDir("")
	_, err := s3Cli.Put("aaa", strings.NewReader("aaa"))
	assert.Nil(t, err)
	stat, err := s3Cli.Stat("aaa")
	assert.Nil(t, err)
	assert.Equal(t, uint64(3), stat.Size())
	assert.Equal(t, "data/aaa", stat.Path())
	assert.Equal(t, time.Now().Format("2006-01-02"), stat.LastModified().Format("2006-01-02"))
}

func TestS3_Type(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up := NewS3(nil, "bkt", nil, "root")

	assert.Equal(t, contracts.S3, up.Type())
}

func TestS3_getPath(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up := NewS3(nil, "bkt", nil, "root")

	assert.Equal(t, "root/666", up.(*s3Uploader).getPath("666"))
}

func TestS3_root(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up := NewS3(nil, "bkt", nil, "root")

	assert.Equal(t, "root", up.(*s3Uploader).root())
}

func Test_s3File_Close(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up, _ := NewUploader("", "")
	s3Cli := NewS3(s3Client, testBucketName, up, "data")
	s3Cli.DeleteDir("")
	file, err := s3Cli.NewFile("s3file_close.txt")
	assert.Nil(t, err)
	file.WriteString("aaa")
	stat, err := file.Stat()
	file.Close()
	assert.Nil(t, err)
	assert.Equal(t, s3Cli.(*s3Uploader).getPath("s3file_close.txt"), stat.Name())
	assert.Equal(t, s3Cli.(*s3Uploader).getPath("s3file_close.txt"), file.Name())

	assert.Equal(t, int64(3), stat.Size())
	assert.True(t, s3Cli.Exists(s3Cli.(*s3Uploader).getPath("s3file_close.txt")))
}

func Test_s3File_Name(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	s3f := &s3File{name: "aaa"}
	assert.Equal(t, "aaa", s3f.Name())
}

func Test_s3File_Seek(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	m := gomock.NewController(t)
	defer m.Finish()
	f := mock.NewMockFile(m)
	s := &s3File{
		File: f,
	}
	f.EXPECT().Seek(int64(10), 1).Times(1).Return(int64(1), nil)
	ret, err := s.Seek(int64(10), 1)
	assert.Equal(t, int64(1), ret)
	assert.Nil(t, err)
}

func Test_s3Uploader_NewFile(t *testing.T) {
	up := &s3Uploader{}
	assert.Same(t, up, up.UnWrap())
}
