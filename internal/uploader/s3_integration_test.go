package uploader

import (
	"context"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// s3RealOnce 保证整包只建立一次真实 minio 连接：测试 bucket 的清理/重建只做一遍。
var s3RealOnce sync.Once

// s3RealAPI 是连上真实 minio 的客户端（经 minioClient 适配进 minioAPI）；连不上时为 nil。
var s3RealAPI minioAPI

// s3RealSkip 为 true 表示真实 minio 不可用，S3 集成测试整体跳过。
var s3RealSkip = true

// realMinio 返回可用的真实 minioAPI，连不上则 t.Skip。
// 首次调用建立连接：读 S3_ENDPOINT 环境变量（缺省 localhost:9000），凭据取
// S3_KEY_ID/S3_SECRET_ID（缺省 minioadmin，与 CI minio 容器一致），重建测试 bucket。
// MakeBucket 失败视为服务不可用，带重试以容忍 CI 里 minio 容器的启动延迟。
func realMinio(t *testing.T) minioAPI {
	t.Helper()
	s3RealOnce.Do(func() {
		endpoint := os.Getenv("S3_ENDPOINT")
		if endpoint == "" {
			endpoint = "localhost:9000"
		}
		keyID := os.Getenv("S3_KEY_ID")
		if keyID == "" {
			keyID = "minioadmin"
		}
		secret := os.Getenv("S3_SECRET_ID")
		if secret == "" {
			secret = "minioadmin"
		}
		// 端口探测带 3 次、间隔 500ms 的重试：minio 未部署时连接立即被拒、快速跳过；
		// CI 里容器刚启动、端口未就绪时，重试能覆盖启动窗口而不误判。
		var (
			conn net.Conn
			err  error
		)
		for i := 0; i < 3; i++ {
			conn, err = net.DialTimeout("tcp", endpoint, 500*time.Millisecond)
			if err == nil {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		if err != nil {
			return
		}
		_ = conn.Close()
		cli, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(keyID, secret, ""),
			Secure: false,
		})
		if err != nil {
			return
		}
		// 端口在监听但 HTTP 接口可能仍在就绪：重试 3 次、间隔 1s 覆盖 CI 启动窗口。
		for i := 0; i < 3; i++ {
			ok, err := cli.BucketExists(context.TODO(), testBucketName)
			if err == nil && ok {
				_ = cli.RemoveBucketWithOptions(context.TODO(), testBucketName, minio.RemoveBucketOptions{ForceDelete: true})
			}
			if err := cli.MakeBucket(context.TODO(), testBucketName, minio.MakeBucketOptions{}); err == nil {
				s3RealAPI = &minioClient{Client: cli}
				s3RealSkip = false
				return
			}
			time.Sleep(time.Second)
		}
	})
	if s3RealSkip {
		t.Skip("minio 不可用，跳过 S3 集成测试")
	}
	return s3RealAPI
}

// newRealS3 构造连真实 minio 的 s3Uploader：本地镜像落在独立临时目录，root 前缀隔离各测试。
// 构造后先按 root 前缀清空对象，保证同一前缀的断言与运行次数无关（-count 重复跑不挂）。
func newRealS3(t *testing.T, root string) *s3Uploader {
	t.Helper()
	up, err := NewDiskUploader(t.TempDir(), mlog.NewForConfig(nil))
	require.NoError(t, err)
	s3u := newS3(func() minioAPI { return realMinio(t) }, testBucketName, up, root).(*s3Uploader)
	require.NoError(t, s3u.DeleteDir(""))
	return s3u
}

// TestS3Integration_PutReadDelete 走真实 S3 的写入→读取→删除全链路，验证对象内容不丢失。
func TestS3Integration_PutReadDelete(t *testing.T) {
	s3u := newRealS3(t, "it/putread")

	info, err := s3u.Put("a.txt", strings.NewReader("hello s3"))
	require.NoError(t, err)
	assert.Equal(t, "it/putread/a.txt", info.Path())
	assert.True(t, s3u.Exists("a.txt"))

	rc, err := s3u.Read("a.txt")
	require.NoError(t, err)
	defer rc.Close()
	all, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, "hello s3", string(all))

	// 真实 StatObject 校验：Stat 返回对象真实字节数与最近修改时间，而非假值。
	stat, err := s3u.Stat("a.txt")
	require.NoError(t, err)
	assert.Equal(t, uint64(8), stat.Size())
	assert.False(t, stat.LastModified().IsZero())

	require.NoError(t, s3u.Delete("a.txt"))
	assert.False(t, s3u.Exists("a.txt"))
}

// TestS3Integration_DirSizeAndList 验证真实 ListObjects 对前缀目录的统计与列举语义。
func TestS3Integration_DirSizeAndList(t *testing.T) {
	s3u := newRealS3(t, "it/dirlist")

	_, err := s3u.Put("a.txt", strings.NewReader("aaa"))
	require.NoError(t, err)
	_, err = s3u.Put("b/c.txt", strings.NewReader("bbbb"))
	require.NoError(t, err)

	size, err := s3u.DirSize()
	require.NoError(t, err)
	assert.Equal(t, int64(7), size)

	files, err := s3u.AllDirectoryFiles("")
	require.NoError(t, err)
	assert.Len(t, files, 2)
}

// TestS3Integration_DeleteDir 验证按前缀递归删除对象的真实行为。
func TestS3Integration_DeleteDir(t *testing.T) {
	s3u := newRealS3(t, "it/deletedir")

	_, err := s3u.Put("x.txt", strings.NewReader("x"))
	require.NoError(t, err)
	_, err = s3u.Put("sub/y.txt", strings.NewReader("y"))
	require.NoError(t, err)

	require.NoError(t, s3u.DeleteDir(""))
	files, err := s3u.AllDirectoryFiles("")
	require.NoError(t, err)
	assert.Len(t, files, 0)
}

// TestS3Integration_NewFileCloseUploads 验证 NewFile 写入并 Close 后对象真实落到 S3。
func TestS3Integration_NewFileCloseUploads(t *testing.T) {
	s3u := newRealS3(t, "it/newfile")

	file, err := s3u.NewFile("f.txt")
	require.NoError(t, err)
	_, err = file.WriteString("newfile-content")
	require.NoError(t, err)
	require.NoError(t, file.Close())

	assert.True(t, s3u.Exists("f.txt"))
	// 真实读回校验：Close 经 FPutObject 上传的内容与本地写入一致，而非只存在键。
	rc, err := s3u.Read("f.txt")
	require.NoError(t, err)
	defer rc.Close()
	all, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, "newfile-content", string(all))
}
