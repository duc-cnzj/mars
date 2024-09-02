package uploader

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v5/internal/ent/schema/schematype"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	pwd, _         = os.Getwd()
	testDir        = filepath.Join(pwd, "testdir")
	testBucketName = "testbucket"
	s3Client       *minio.Client
)

var (
	s3Endpoint string = os.Getenv("S3_ENDPOINT")
	s3KeyID    string = os.Getenv("S3_KEY_ID")
	s3SecretID string = os.Getenv("S3_SECRET_ID")
	skipS3     bool   = true
)

func TestMain(m *testing.M) {
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, os.ModePerm)
	if s3Endpoint == "" {
		s3Endpoint = "localhost:9000"
	}
	if s3KeyID == "" {
		s3KeyID = "root"
	}
	if s3SecretID == "" {
		s3SecretID = "root123456"
	}
	s3Client, _ = minio.New(s3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3KeyID, s3SecretID, ""),
		Secure: false,
	})
	exists, _ := s3Client.BucketExists(context.TODO(), testBucketName)
	if exists {
		s3Client.RemoveBucketWithOptions(context.TODO(), testBucketName, minio.RemoveBucketOptions{ForceDelete: true})
	}
	err := s3Client.MakeBucket(context.TODO(), testBucketName, minio.MakeBucketOptions{})
	if err == nil {
		skipS3 = false
	}
	exitCode := m.Run()
	os.RemoveAll(testDir)
	os.Exit(exitCode)
}

func TestNewS3(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	assert.Implements(t, (*Uploader)(nil), NewS3(nil, "bkt", nil, "root"))
	assert.Equal(t, "root", NewS3(nil, "bkt", nil, "root").(*s3Uploader).rootDir)
	assert.Equal(t, "data", NewS3(nil, "bkt", nil, "").(*s3Uploader).rootDir)
}

func TestS3_AllDirectoryFiles(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up, _ := NewDiskUploader(testDir, mlog.NewLogger(nil))
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
	up, _ := NewDiskUploader(testDir, mlog.NewLogger(nil))
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
	up, _ := NewDiskUploader(testDir, mlog.NewLogger(nil))
	s3Cli := NewS3(s3Client, testBucketName, up, "")
	s3Cli.DeleteDir("")
	s3Cli.Put("cc/c.txt", strings.NewReader("aaa"))
	assert.True(t, s3Cli.Exists("cc/c.txt"))
	s3Cli.DeleteDir("")
	assert.False(t, s3Cli.Exists("cc/c.txt"))
}

func TestS3_DirSize(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up, _ := NewDiskUploader(testDir, mlog.NewLogger(nil))
	s3Cli := NewS3(s3Client, testBucketName, up, "dirsize")
	s3Cli.DeleteDir("")
	s3Cli.Put("dirsize/cc/c.txt", strings.NewReader("aaa"))
	size, _ := s3Cli.DirSize()
	assert.Equal(t, int64(3), size)
}

func TestS3_Exists(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up, _ := NewDiskUploader(testDir, mlog.NewLogger(nil))
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
	up, _ := NewDiskUploader(testDir, mlog.NewLogger(nil))
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
	up, _ := NewDiskUploader(testDir, mlog.NewLogger(nil))
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
	up, _ := NewDiskUploader(testDir, mlog.NewLogger(nil))
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
	up, _ := NewDiskUploader(testDir, mlog.NewLogger(nil))
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
	up, _ := NewDiskUploader(testDir, mlog.NewLogger(nil))
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

	assert.Equal(t, schematype.S3, up.Type())
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

func Test_s3File_Name(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	s3f := &s3File{name: "aaa"}
	assert.Equal(t, "aaa", s3f.Name())
}

func Test_s3Uploader_NewFile(t *testing.T) {
	up := &s3Uploader{}
	assert.Same(t, up, up.UnWrap())
}

func Test_s3Uploader_LocalUploader(t *testing.T) {
	up := &s3Uploader{
		localUploader: &diskUploader{},
	}
	assert.IsType(t, &diskUploader{}, up.LocalUploader())
}

func Test_s3Uploader_NewFile1(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	localup := NewMockUploader(m)
	up := &s3Uploader{
		localUploader: localup,
	}
	localup.EXPECT().NewFile(gomock.Any()).Return(nil, errors.New("x"))
	_, err := up.NewFile("aaa")
	assert.Error(t, err)
}

func Test_s3File_Seek(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	file := NewMockFile(m)
	file.EXPECT().Seek(int64(1), 1).Return(int64(1), nil)
	s3f := &s3File{File: file}
	ret, err := s3f.Seek(int64(1), 1)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), ret)
}

func Test_s3Uploader_Put(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	localup := NewMockUploader(m)
	up := &s3Uploader{
		localUploader: localup,
	}
	localup.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil, errors.New("x"))
	_, err := up.Put("aaa", strings.NewReader("aaa"))
	assert.Error(t, err)
}

func Test_s3Uploader_AbsolutePath(t *testing.T) {
	up := &s3Uploader{
		rootDir: "/aaa",
		disk:    "bbb",
	}
	assert.Equal(t, "/aaa/bbb/aaa", up.AbsolutePath("aaa"))
}

func Test_s3File_Stat(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	file := NewMockFile(m)
	file.EXPECT().Stat().Return(nil, errors.New("x"))
	s3f := &s3File{File: file}
	_, err := s3f.Stat()
	assert.Error(t, err)
}

func Test_s3File_Close(t *testing.T) {
	if skipS3 {
		t.Skip()
	}
	up, _ := NewDiskUploader(testDir, mlog.NewLogger(nil))
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

func Test_s3File_Close1(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	file := NewMockFile(m)
	file.EXPECT().Close()
	file.EXPECT().Name().Return("aaa").Times(2)
	localup := NewMockUploader(m)
	localup.EXPECT().Read(gomock.Any()).Return(nil, errors.New("x"))
	localup.EXPECT().Delete(gomock.Any())
	s3f := &s3File{File: file, localUploader: localup}
	assert.Error(t, s3f.Close())
	assert.Nil(t, s3f.Close())
}
