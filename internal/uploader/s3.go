package uploader

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/util/closeable"
	"github.com/minio/minio-go/v7"
)

// minioAPI 抽象 minio.Client 的 S3 操作子集，s3Uploader 通过它访问对象存储。
// 抽象成接口后，S3 层可以脱离真实服务器，用测试替身单测，覆盖不依赖外部服务。
type minioAPI interface {
	StatObject(ctx context.Context, bucketName, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error)
	RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error
	// GetObject 返回 io.ReadCloser 而非 *minio.Object：后者没有导出构造器、字段全私有，
	// 测试里无法伪造，收窄返回类型后 *minio.Client 由 minioClient 适配进接口。
	GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error)
	FPutObject(ctx context.Context, bucketName, objectName, filePath string, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo
}

var _ minioAPI = (*minioClient)(nil)

// minioClient 把 *minio.Client 适配成 minioAPI：嵌入的 *minio.Client 自动提供
// StatObject/RemoveObject/FPutObject/ListObjects，仅 GetObject 需要显式收窄返回类型。
type minioClient struct {
	*minio.Client
}

// GetObject 转发到底层 *minio.Client，把 *minio.Object 返回值收窄为 io.ReadCloser。
func (m *minioClient) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error) {
	return m.Client.GetObject(ctx, bucketName, objectName, opts)
}

// s3Uploader 是基于对象存储（S3/MinIO）的 Uploader 实现。
// 它把本地磁盘上传器当临时镜像：写操作先落盘，再同步到对象存储。
// getMinioAPI 是惰性取数函数而非客户端实例：minio 客户端由 S3Bootstrapper 在
// bootstrap 阶段才初始化，wire 构建时拿到的还是 nil，必须在每次 S3 操作时
// 实时解析，否则 S3 上传器会永久持有 nil 客户端。
type s3Uploader struct {
	localUploader Uploader
	bucket        string
	rootDir       string
	disk          string
	getMinioAPI   func() minioAPI
}

// newS3 构造 S3 上传器。getMinioAPI 是惰性取数函数，每次 S3 操作时调用，
// 以拿到 bootstrap 之后才初始化完成的 minio 客户端。
func newS3(getMinioAPI func() minioAPI, bucket string, uploader Uploader, rootDir string) Uploader {
	if rootDir == "" {
		rootDir = "data"
	}
	return &s3Uploader{
		getMinioAPI:   getMinioAPI,
		bucket:        bucket,
		localUploader: uploader,
		rootDir:       rootDir,
	}
}

// Type 返回对象存储类型。
func (s *s3Uploader) Type() schematype.UploadType {
	return schematype.S3
}

// Disk 返回以 disk 为子目录的新上传器，本地镜像同样切到对应子目录。
func (s *s3Uploader) Disk(disk string) Uploader {
	return &s3Uploader{
		localUploader: s.localUploader.Disk(disk),
		getMinioAPI:   s.getMinioAPI,
		bucket:        s.bucket,
		rootDir:       s.root(),
		disk:          disk,
	}
}

// DeleteDir 递归删除 dir 下全部对象（按前缀列出再逐个删），并清理本地镜像。
func (s *s3Uploader) DeleteDir(dir string) error {
	dir = s.getPath(dir)
	// 本地镜像 best-effort：本地目录不存在/删除失败不阻塞对象存储侧删除。
	_ = s.localUploader.DeleteDir(dir)
	cli := s.getMinioAPI()
	objects := cli.ListObjects(context.TODO(), s.bucket, minio.ListObjectsOptions{
		Prefix:    dir,
		Recursive: true,
	})
	for object := range objects {
		if object.Err != nil {
			return object.Err
		}
		if err := cli.RemoveObject(context.TODO(), s.bucket, object.Key, minio.RemoveObjectOptions{ForceDelete: true}); err != nil {
			return err
		}
	}
	return nil
}

// DirSize 统计当前 root 前缀下全部对象的字节数之和。
func (s *s3Uploader) DirSize() (int64, error) {
	dir := s.root()
	cli := s.getMinioAPI()
	objects := cli.ListObjects(context.TODO(), s.bucket, minio.ListObjectsOptions{
		Prefix:    dir,
		Recursive: true,
	})
	var size int64
	for object := range objects {
		if object.Err != nil {
			return 0, object.Err
		}
		size += object.Size
	}
	return size, nil
}

// Delete 删除单个对象，并清理本地镜像。
func (s *s3Uploader) Delete(path string) error {
	path = s.getPath(path)
	// 本地镜像 best-effort：本地删除失败不阻塞对象存储侧删除。
	_ = s.localUploader.Delete(path)
	return s.getMinioAPI().RemoveObject(context.TODO(), s.bucket, path, minio.RemoveObjectOptions{
		ForceDelete: true,
	})
}

// Exists 判断对象是否存在（对不存在的对象 StatObject 返回错误）。
func (s *s3Uploader) Exists(path string) bool {
	path = s.getPath(path)
	_, err := s.getMinioAPI().StatObject(context.TODO(), s.bucket, path, minio.StatObjectOptions{})
	return err == nil
}

// MkDir S3 无目录概念，直接返回 nil。
func (s *s3Uploader) MkDir(path string, recursive bool) error {
	// S3 does not require directories to be created explicitly
	return nil
}

// Read 返回对象的可读流。GetObject 返回的是惰性句柄，对象不存在要到首次读取才报错，
// 因此先用 StatObject 快查，让"不存在"在调用处立即暴露而非延迟到读取时。
func (s *s3Uploader) Read(path string) (io.ReadCloser, error) {
	if !s.Exists(path) {
		return nil, os.ErrNotExist
	}
	path = s.getPath(path)
	return s.getMinioAPI().GetObject(context.TODO(), s.bucket, path, minio.GetObjectOptions{})
}

// AbsolutePath 返回 path 对应的完整对象 key。
func (s *s3Uploader) AbsolutePath(path string) string {
	return s.getPath(path)
}

// Stat 返回对象的元信息，对象不存在时返回错误。
func (s *s3Uploader) Stat(file string) (FileInfo, error) {
	path := s.getPath(file)
	object, err := s.getMinioAPI().StatObject(context.TODO(), s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	return newFileInfo(path, uint64(object.Size), object.LastModified), nil
}

// Put 先把 content 写入本地镜像，再上传到对象存储；上传后清掉本地临时文件。
func (s *s3Uploader) Put(path string, content io.Reader) (FileInfo, error) {
	path = s.getPath(path)
	put, err := s.localUploader.Put(path, content)
	if err != nil {
		return nil, err
	}
	// 无论上传成败，本地临时镜像都要清掉。
	defer func() { _ = s.localUploader.Delete(put.Path()) }()
	return s.uploadToS3(path, put.Path())
}

// uploadToS3 把本地文件 localPath 上传为对象 key path，返回对象侧文件信息。
func (s *s3Uploader) uploadToS3(path, localPath string) (FileInfo, error) {
	object, err := s.getMinioAPI().FPutObject(context.TODO(), s.bucket, path, localPath, minio.PutObjectOptions{})
	if err != nil {
		return nil, err
	}
	return newFileInfo(object.Key, uint64(object.Size), object.LastModified), nil
}

// AllDirectoryFiles 返回 dir 前缀下全部对象的文件信息列表。
func (s *s3Uploader) AllDirectoryFiles(dir string) ([]FileInfo, error) {
	dir = s.getPath(dir)
	cli := s.getMinioAPI()
	objects := cli.ListObjects(context.TODO(), s.bucket, minio.ListObjectsOptions{
		Prefix:    dir,
		Recursive: true,
	})
	var files []FileInfo
	for object := range objects {
		if object.Err != nil {
			return nil, object.Err
		}
		files = append(files, newFileInfo(object.Key, uint64(object.Size), object.LastModified))
	}
	return files, nil
}

// RemoveEmptyDir 清理本地镜像的空目录（对象存储无目录概念）。
func (s *s3Uploader) RemoveEmptyDir() error {
	return s.localUploader.RemoveEmptyDir()
}

// LocalUploader 返回底层本地镜像上传器。
func (s *s3Uploader) LocalUploader() Uploader {
	return s.localUploader
}

// NewFile 在本地镜像创建文件，包装成关闭时自动上传到对象存储的 s3File。
func (s *s3Uploader) NewFile(path string) (File, error) {
	file, err := s.localUploader.NewFile(s.getPath(path))
	if err != nil {
		return nil, err
	}
	return &s3File{
		localUploader: s.localUploader,
		s3:            s,
		name:          s.getPath(path),
		File:          file,
	}, nil
}

// s3File 是带对象存储上传能力的文件包装：Close 时把本地内容同步到 S3 再删除本地镜像。
type s3File struct {
	closeable.Closeable
	File
	localUploader Uploader
	s3            *s3Uploader
	name          string
}

// Name 返回对象侧 key（而非本地临时路径）。
func (s *s3File) Name() string {
	return s.name
}

// s3OsFileInfo 把本地文件信息伪装成对象侧文件名，让外部看到的是对象 key。
type s3OsFileInfo struct {
	name string
	os.FileInfo
}

// Name 返回对象侧文件名。
func (s *s3OsFileInfo) Name() string {
	return s.name
}

// Seek 转发到底层文件。
func (s *s3File) Seek(offset int64, whence int) (ret int64, err error) {
	return s.File.Seek(offset, whence)
}

// Stat 返回底层文件信息，但文件名替换为对象侧 key。
func (s *s3File) Stat() (os.FileInfo, error) {
	stat, err := s.File.Stat()
	if err != nil {
		return nil, err
	}
	return &s3OsFileInfo{name: s.name, FileInfo: stat}, nil
}

// Close 关闭并上传：首次 Close 把本地内容上传到对象存储后删除本地镜像，
// 再次 Close 是幂等空操作（由 closeable.Closeable 保证）。
func (s *s3File) Close() error {
	if s.Closeable.Close() {
		if err := s.File.Close(); err != nil {
			return err
		}
		defer func() { _ = s.localUploader.Delete(s.File.Name()) }()
		open, err := s.localUploader.Read(s.File.Name())
		if err != nil {
			return err
		}
		defer open.Close()
		_, err = s.s3.uploadToS3(s.name, s.File.Name())
		return err
	}
	return nil
}

// getPath 把入参 path 拼接到当前 root 前缀下；path 已是 root 前缀时原样返回。
func (s *s3Uploader) getPath(path string) string {
	if strings.HasPrefix(path, s.root()) {
		return path
	}
	return filepath.Join(s.root(), path)
}

// root 返回当前生效的根前缀：未切子目录时是 rootDir，否则是 rootDir/disk。
func (s *s3Uploader) root() string {
	if s.disk != "" {
		return filepath.Join(s.rootDir, s.disk)
	}
	return s.rootDir
}
