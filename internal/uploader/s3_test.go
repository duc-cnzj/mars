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

	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	pwd, _         = os.Getwd()
	testDir        = filepath.Join(pwd, "testdir")
	testBucketName = "testbucket"
)

func TestMain(m *testing.M) {
	_ = os.RemoveAll(testDir)
	_ = os.MkdirAll(testDir, os.ModePerm)
	exitCode := m.Run()
	_ = os.RemoveAll(testDir)
	os.Exit(exitCode)
}

// fakeMinio 是 minioAPI 的测试替身：每个方法对应一个函数字段，测试按需注入，
// 未注入的字段不会被调用（只调用被注入字段对应的方法）。
type fakeMinio struct {
	statFn   func(bucket, object string) (minio.ObjectInfo, error)
	removeFn func(bucket, object string) error
	getFn    func(bucket, object string) (io.ReadCloser, error)
	fputFn   func(bucket, object, localPath string) (minio.UploadInfo, error)
	listFn   func(bucket string) <-chan minio.ObjectInfo
}

func (f *fakeMinio) StatObject(_ context.Context, bucket, object string, _ minio.StatObjectOptions) (minio.ObjectInfo, error) {
	return f.statFn(bucket, object)
}

func (f *fakeMinio) RemoveObject(_ context.Context, bucket, object string, _ minio.RemoveObjectOptions) error {
	return f.removeFn(bucket, object)
}

func (f *fakeMinio) GetObject(_ context.Context, bucket, object string, _ minio.GetObjectOptions) (io.ReadCloser, error) {
	return f.getFn(bucket, object)
}

func (f *fakeMinio) FPutObject(_ context.Context, bucket, object, localPath string, _ minio.PutObjectOptions) (minio.UploadInfo, error) {
	return f.fputFn(bucket, object, localPath)
}

func (f *fakeMinio) ListObjects(_ context.Context, bucket string, _ minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	return f.listFn(bucket)
}

// listOf 把对象列表装进已填充并关闭的 channel，避免每次迭代都开 goroutine。
func listOf(objects ...minio.ObjectInfo) <-chan minio.ObjectInfo {
	ch := make(chan minio.ObjectInfo, len(objects))
	for _, object := range objects {
		ch <- object
	}
	close(ch)
	return ch
}

// newTestS3 构造 root 前缀的 s3Uploader，minio 侧用 fake，不需要真实磁盘。
func newTestS3(fake *fakeMinio, root string) *s3Uploader {
	return newS3(func() minioAPI { return fake }, testBucketName, nil, root).(*s3Uploader)
}

// newTestS3WithDisk 构造 root 前缀的 s3Uploader，本地镜像落到 testDir。
func newTestS3WithDisk(t *testing.T, fake *fakeMinio, root string) (*s3Uploader, *diskUploader) {
	t.Helper()
	up, err := NewDiskUploader(testDir, mlog.NewForConfig(nil))
	assert.NoError(t, err)
	return newS3(func() minioAPI { return fake }, testBucketName, up, root).(*s3Uploader), up.(*diskUploader)
}

func TestNewS3(t *testing.T) {
	up := newS3(nil, "bkt", nil, "root")
	assert.Implements(t, (*Uploader)(nil), up)
	assert.Equal(t, "root", up.(*s3Uploader).rootDir)

	up2 := newS3(nil, "bkt", nil, "")
	assert.Equal(t, "data", up2.(*s3Uploader).rootDir)
}

func TestS3_Type(t *testing.T) {
	up := newS3(nil, "bkt", nil, "root")

	assert.Equal(t, schematype.S3, up.Type())
}

func TestS3_Disk(t *testing.T) {
	up, err := NewDiskUploader(testDir, mlog.NewForConfig(nil))
	assert.NoError(t, err)
	s3u := newS3(nil, testBucketName, up, "data").(*s3Uploader)

	dd := s3u.Disk("one").(*s3Uploader)
	assert.Equal(t, "one", dd.disk)
	assert.Equal(t, "data/one", dd.root())
	assert.Equal(t, "one", dd.localUploader.(*diskUploader).disk)

	dd2 := dd.Disk("two").(*s3Uploader)
	assert.Equal(t, "two", dd2.disk)
	assert.Equal(t, "data/one/two", dd2.root())
}

func TestS3_DeleteDir(t *testing.T) {
	t.Run("remove objects under prefix", func(t *testing.T) {
		s3u, _ := newTestS3WithDisk(t, nil, "data")
		var removed []string
		s3u.getMinioAPI = func() minioAPI {
			return &fakeMinio{
				listFn: func(string) <-chan minio.ObjectInfo {
					return listOf(minio.ObjectInfo{Key: "data/a"}, minio.ObjectInfo{Key: "data/b"})
				},
				removeFn: func(_ string, object string) error {
					removed = append(removed, object)
					return nil
				},
			}
		}

		assert.NoError(t, s3u.DeleteDir(""))
		assert.Equal(t, []string{"data/a", "data/b"}, removed)
	})

	t.Run("list error", func(t *testing.T) {
		s3u, _ := newTestS3WithDisk(t, &fakeMinio{
			listFn: func(string) <-chan minio.ObjectInfo {
				return listOf(minio.ObjectInfo{Err: errors.New("list boom")})
			},
		}, "data")

		assert.Error(t, s3u.DeleteDir(""))
	})

	t.Run("remove error", func(t *testing.T) {
		s3u, _ := newTestS3WithDisk(t, &fakeMinio{
			listFn: func(string) <-chan minio.ObjectInfo {
				return listOf(minio.ObjectInfo{Key: "data/a"})
			},
			removeFn: func(string, string) error { return errors.New("remove boom") },
		}, "data")

		assert.Error(t, s3u.DeleteDir(""))
	})
}

func TestS3_DirSize(t *testing.T) {
	t.Run("sum sizes", func(t *testing.T) {
		s3u := newTestS3(&fakeMinio{
			listFn: func(string) <-chan minio.ObjectInfo {
				return listOf(
					minio.ObjectInfo{Key: "data/a", Size: 1},
					minio.ObjectInfo{Key: "data/b", Size: 2},
					minio.ObjectInfo{Key: "data/c", Size: 3},
				)
			},
		}, "data")

		size, err := s3u.DirSize()
		assert.NoError(t, err)
		assert.Equal(t, int64(6), size)
	})

	t.Run("list error", func(t *testing.T) {
		s3u := newTestS3(&fakeMinio{
			listFn: func(string) <-chan minio.ObjectInfo {
				return listOf(minio.ObjectInfo{Err: errors.New("boom")})
			},
		}, "data")

		size, err := s3u.DirSize()
		assert.Error(t, err)
		assert.Equal(t, int64(0), size)
	})
}

func TestS3_Delete(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		s3u, _ := newTestS3WithDisk(t, nil, "data")
		var removed string
		s3u.getMinioAPI = func() minioAPI {
			return &fakeMinio{removeFn: func(_ string, object string) error {
				removed = object
				return nil
			}}
		}

		assert.NoError(t, s3u.Delete("x"))
		assert.Equal(t, "data/x", removed)
	})

	t.Run("remove error", func(t *testing.T) {
		s3u, _ := newTestS3WithDisk(t, nil, "data")
		s3u.getMinioAPI = func() minioAPI { return &fakeMinio{removeFn: func(string, string) error { return errors.New("boom") }} }

		assert.Error(t, s3u.Delete("x"))
	})
}

func TestS3_Exists(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		s3u := newTestS3(&fakeMinio{statFn: func(string, string) (minio.ObjectInfo, error) {
			return minio.ObjectInfo{}, nil
		}}, "data")

		assert.True(t, s3u.Exists("x"))
	})

	t.Run("not exists", func(t *testing.T) {
		s3u := newTestS3(&fakeMinio{statFn: func(string, string) (minio.ObjectInfo, error) {
			return minio.ObjectInfo{}, errors.New("not found")
		}}, "data")

		assert.False(t, s3u.Exists("x"))
	})
}

func TestS3_MkDir(t *testing.T) {
	up := newS3(nil, "bkt", nil, "root")

	assert.NoError(t, up.MkDir("", true))
}

func TestS3_Read(t *testing.T) {
	t.Run("not exists", func(t *testing.T) {
		s3u := newTestS3(&fakeMinio{statFn: func(string, string) (minio.ObjectInfo, error) {
			return minio.ObjectInfo{}, errors.New("not found")
		}}, "data")

		rc, err := s3u.Read("x")
		assert.ErrorIs(t, err, os.ErrNotExist)
		assert.Nil(t, rc)
	})

	t.Run("read ok", func(t *testing.T) {
		s3u := newTestS3(&fakeMinio{
			statFn: func(string, string) (minio.ObjectInfo, error) { return minio.ObjectInfo{}, nil },
			getFn: func(string, string) (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader("aaa")), nil
			},
		}, "data")

		rc, err := s3u.Read("x")
		assert.NoError(t, err)
		defer rc.Close()
		all, err := io.ReadAll(rc)
		assert.NoError(t, err)
		assert.Equal(t, "aaa", string(all))
	})

	t.Run("get error", func(t *testing.T) {
		s3u := newTestS3(&fakeMinio{
			statFn: func(string, string) (minio.ObjectInfo, error) { return minio.ObjectInfo{}, nil },
			getFn:  func(string, string) (io.ReadCloser, error) { return nil, errors.New("boom") },
		}, "data")

		_, err := s3u.Read("x")
		assert.Error(t, err)
	})
}

func TestS3_AbsolutePath(t *testing.T) {
	up := newS3(nil, "bkt", nil, "data")

	assert.Equal(t, "data/aaa", up.AbsolutePath("aaa"))
}

func TestS3_Stat(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		tm := time.Now()
		s3u := newTestS3(&fakeMinio{statFn: func(string, object string) (minio.ObjectInfo, error) {
			return minio.ObjectInfo{Key: "data/x", Size: 3, LastModified: tm}, nil
		}}, "data")

		info, err := s3u.Stat("x")
		assert.NoError(t, err)
		assert.Equal(t, "data/x", info.Path())
		assert.Equal(t, uint64(3), info.Size())
		assert.Equal(t, tm, info.LastModified())
	})

	t.Run("error", func(t *testing.T) {
		s3u := newTestS3(&fakeMinio{statFn: func(string, string) (minio.ObjectInfo, error) {
			return minio.ObjectInfo{}, errors.New("not found")
		}}, "data")

		_, err := s3u.Stat("x")
		assert.Error(t, err)
	})
}

func TestS3_Put(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		s3u, up := newTestS3WithDisk(t, nil, "data")
		var uploadedLocal string
		s3u.getMinioAPI = func() minioAPI {
			return &fakeMinio{fputFn: func(_ string, object, localPath string) (minio.UploadInfo, error) {
				uploadedLocal = localPath
				return minio.UploadInfo{Key: object, Size: 3}, nil
			}}
		}

		info, err := s3u.Put("sub/a.txt", strings.NewReader("aaa"))
		assert.NoError(t, err)
		assert.Equal(t, "data/sub/a.txt", info.Path())
		assert.Equal(t, uint64(3), info.Size())
		assert.Equal(t, filepath.Join(testDir, "data", "sub", "a.txt"), uploadedLocal)
		assert.False(t, up.Exists("data/sub/a.txt")) // 本地镜像上传后已清
	})

	t.Run("local put error", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		localup := NewMockUploader(m)
		localup.EXPECT().Put("data/a.txt", gomock.Any()).Return(nil, errors.New("put boom"))
		s3u := &s3Uploader{localUploader: localup, bucket: testBucketName, rootDir: "data"}

		_, err := s3u.Put("a.txt", strings.NewReader("aaa"))
		assert.Error(t, err)
	})

	t.Run("upload error", func(t *testing.T) {
		s3u, up := newTestS3WithDisk(t, nil, "data")
		s3u.getMinioAPI = func() minioAPI {
			return &fakeMinio{fputFn: func(string, string, string) (minio.UploadInfo, error) {
				return minio.UploadInfo{}, errors.New("fput boom")
			}}
		}

		_, err := s3u.Put("b.txt", strings.NewReader("aaa"))
		assert.Error(t, err)
		assert.False(t, up.Exists("data/b.txt")) // 上传失败也要清本地镜像
	})
}

func TestS3_uploadToS3(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		var gotBucket, gotObject, gotLocal string
		s3u := newTestS3(&fakeMinio{fputFn: func(bucket, object, localPath string) (minio.UploadInfo, error) {
			gotBucket, gotObject, gotLocal = bucket, object, localPath
			return minio.UploadInfo{Key: object, Size: 3}, nil
		}}, "data")

		info, err := s3u.uploadToS3("data/x", "/local/x")
		assert.NoError(t, err)
		assert.Equal(t, testBucketName, gotBucket)
		assert.Equal(t, "data/x", gotObject)
		assert.Equal(t, "/local/x", gotLocal)
		assert.Equal(t, "data/x", info.Path())
		assert.Equal(t, uint64(3), info.Size())
	})

	t.Run("error", func(t *testing.T) {
		s3u := newTestS3(&fakeMinio{fputFn: func(string, string, string) (minio.UploadInfo, error) {
			return minio.UploadInfo{}, errors.New("boom")
		}}, "data")

		_, err := s3u.uploadToS3("data/x", "/local/x")
		assert.Error(t, err)
	})
}

func TestS3_AllDirectoryFiles(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		s3u := newTestS3(&fakeMinio{listFn: func(string) <-chan minio.ObjectInfo {
			return listOf(
				minio.ObjectInfo{Key: "data/a", Size: 1},
				minio.ObjectInfo{Key: "data/b/c", Size: 2},
			)
		}}, "data")

		files, err := s3u.AllDirectoryFiles("")
		assert.NoError(t, err)
		assert.Len(t, files, 2)
		assert.Equal(t, "data/a", files[0].Path())
		assert.Equal(t, "data/b/c", files[1].Path())
	})

	t.Run("list error", func(t *testing.T) {
		s3u := newTestS3(&fakeMinio{listFn: func(string) <-chan minio.ObjectInfo {
			return listOf(minio.ObjectInfo{Err: errors.New("boom")})
		}}, "data")

		_, err := s3u.AllDirectoryFiles("")
		assert.Error(t, err)
	})
}

func TestS3_RemoveEmptyDir(t *testing.T) {
	s3u, _ := newTestS3WithDisk(t, nil, "data")

	assert.NoError(t, s3u.RemoveEmptyDir())
}

func TestS3_LocalUploader(t *testing.T) {
	s3u, up := newTestS3WithDisk(t, nil, "data")

	assert.Same(t, up, s3u.LocalUploader())
}

func TestS3_NewFile(t *testing.T) {
	t.Run("local error", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		localup := NewMockUploader(m)
		localup.EXPECT().NewFile("data/a.txt").Return(nil, errors.New("boom"))
		s3u := &s3Uploader{localUploader: localup, rootDir: "data"}

		_, err := s3u.NewFile("a.txt")
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		s3u, up := newTestS3WithDisk(t, nil, "data")

		file, err := s3u.NewFile("a.txt")
		assert.NoError(t, err)
		s3f := file.(*s3File)
		assert.Equal(t, "data/a.txt", s3f.name)
		assert.NoError(t, s3f.File.Close()) // 直接关底层文件，不触发 S3 上传
		assert.True(t, up.Exists("data/a.txt"))
	})
}

func TestS3_getPath(t *testing.T) {
	s3u := &s3Uploader{rootDir: "data"}

	assert.Equal(t, "data/a", s3u.getPath("a"))
	assert.Equal(t, "data/a", s3u.getPath("data/a"))
}

func TestS3_root(t *testing.T) {
	s3u := &s3Uploader{rootDir: "data"}
	assert.Equal(t, "data", s3u.root())

	s3uDisk := &s3Uploader{rootDir: "data", disk: "sub"}
	assert.Equal(t, "data/sub", s3uDisk.root())
}

func Test_s3File_Name(t *testing.T) {
	s3f := &s3File{name: "aaa"}

	assert.Equal(t, "aaa", s3f.Name())
}

func Test_s3OsFileInfo_Name(t *testing.T) {
	info := &s3OsFileInfo{name: "aaa"}

	assert.Equal(t, "aaa", info.Name())
}

func Test_s3File_Seek(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	file := NewMockFile(m)
	file.EXPECT().Seek(int64(1), 1).Return(int64(1), nil)
	s3f := &s3File{File: file}

	ret, err := s3f.Seek(int64(1), 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), ret)
}

func Test_s3File_Stat(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		file := NewMockFile(m)
		file.EXPECT().Stat().Return(nil, errors.New("boom"))
		s3f := &s3File{File: file}

		_, err := s3f.Stat()
		assert.Error(t, err)
	})

	t.Run("ok", func(t *testing.T) {
		s3u, _ := newTestS3WithDisk(t, nil, "data")
		file, err := s3u.NewFile("stat.txt")
		assert.NoError(t, err)
		s3f := file.(*s3File)
		_, err = s3f.WriteString("aaa")
		assert.NoError(t, err)

		stat, err := s3f.Stat()
		assert.NoError(t, err)
		assert.Equal(t, "data/stat.txt", stat.Name())
		assert.Equal(t, int64(3), stat.Size())
		assert.NoError(t, s3f.File.Close())
	})
}

func Test_s3File_Close(t *testing.T) {
	t.Run("file close error", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		file := NewMockFile(m)
		file.EXPECT().Close().Return(errors.New("close boom"))
		s3f := &s3File{File: file}

		assert.Error(t, s3f.Close())
		assert.NoError(t, s3f.Close()) // 二次 Close 幂等
	})

	t.Run("local read error", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		file := NewMockFile(m)
		file.EXPECT().Close().Return(nil)
		file.EXPECT().Name().Return("local").Times(2)
		localup := NewMockUploader(m)
		localup.EXPECT().Read("local").Return(nil, errors.New("read boom"))
		localup.EXPECT().Delete("local")
		s3f := &s3File{File: file, localUploader: localup}

		assert.Error(t, s3f.Close())
		assert.NoError(t, s3f.Close())
	})

	t.Run("upload error", func(t *testing.T) {
		m := gomock.NewController(t)
		defer m.Finish()
		file := NewMockFile(m)
		file.EXPECT().Close().Return(nil)
		file.EXPECT().Name().Return("local").Times(3)
		localup := NewMockUploader(m)
		localup.EXPECT().Read("local").Return(io.NopCloser(strings.NewReader("")), nil)
		localup.EXPECT().Delete("local")
		fake := &fakeMinio{fputFn: func(string, string, string) (minio.UploadInfo, error) {
			return minio.UploadInfo{}, errors.New("fput boom")
		}}
		s3f := &s3File{
			File:          file,
			localUploader: localup,
			s3:            &s3Uploader{getMinioAPI: func() minioAPI { return fake }, bucket: testBucketName},
			name:          "data/x",
		}

		assert.Error(t, s3f.Close())
		assert.NoError(t, s3f.Close())
	})

	t.Run("upload success and idempotent", func(t *testing.T) {
		s3u, up := newTestS3WithDisk(t, nil, "data")
		var uploadedLocal string
		s3u.getMinioAPI = func() minioAPI {
			return &fakeMinio{fputFn: func(_ string, object, localPath string) (minio.UploadInfo, error) {
				uploadedLocal = localPath
				return minio.UploadInfo{Key: object, Size: 3}, nil
			}}
		}

		file, err := s3u.NewFile("close.txt")
		assert.NoError(t, err)
		n, err := file.WriteString("aaa")
		assert.NoError(t, err)
		assert.Equal(t, 3, n)

		assert.NoError(t, file.Close())
		assert.NoError(t, file.Close()) // 二次 Close 幂等
		assert.Equal(t, filepath.Join(testDir, "data", "close.txt"), uploadedLocal)
		assert.False(t, up.Exists("data/close.txt")) // 上传后本地镜像已清
	})
}

// Test_minioClient_GetObject 覆盖 minioClient 适配器的转发路径：连接必拒的端点
// 必然返回错误，无需真实 minio 服务，即可验证 *minio.Client 的 GetObject 被正确包装。
func Test_minioClient_GetObject(t *testing.T) {
	cli, err := minio.New("127.0.0.1:1", &minio.Options{
		Creds:  credentials.NewStaticV4("k", "s", ""),
		Secure: false,
	})
	assert.NoError(t, err)

	rc, err := (&minioClient{Client: cli}).GetObject(context.TODO(), "b", "o", minio.GetObjectOptions{})
	assert.Error(t, err)
	assert.Nil(t, rc)
}
