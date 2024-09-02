package uploader

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/duc-cnzj/mars/v5/internal/ent/schema/schematype"
	"github.com/duc-cnzj/mars/v5/internal/util/closeable"
	"github.com/minio/minio-go/v7"
)

type s3Uploader struct {
	localUploader Uploader
	client        *minio.Client
	bucket        string
	rootDir       string
	disk          string
}

// NewS3 creates a new S3 uploader with the specified parameters.
func NewS3(client *minio.Client, bucket string, uploader Uploader, rootDir string) Uploader {
	if rootDir == "" {
		rootDir = "data"
	}
	return &s3Uploader{
		client:        client,
		bucket:        bucket,
		localUploader: uploader,
		rootDir:       rootDir,
	}
}

func (s *s3Uploader) Type() schematype.UploadType {
	return schematype.S3
}

func (s *s3Uploader) Disk(disk string) Uploader {
	return &s3Uploader{
		localUploader: s.localUploader.Disk(disk),
		client:        s.client,
		bucket:        s.bucket,
		rootDir:       s.root(),
		disk:          disk,
	}
}

func (s *s3Uploader) DeleteDir(dir string) error {
	dir = s.getPath(dir)
	s.localUploader.DeleteDir(dir)
	return s.Delete(dir)
}

func (s *s3Uploader) DirSize() (int64, error) {
	dir := s.root()
	objects := s.client.ListObjects(context.TODO(), s.bucket, minio.ListObjectsOptions{
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

func (s *s3Uploader) Delete(path string) error {
	path = s.getPath(path)
	s.localUploader.Delete(path)
	return s.client.RemoveObject(context.TODO(), s.bucket, path, minio.RemoveObjectOptions{
		ForceDelete: true,
	})
}

func (s *s3Uploader) Exists(path string) bool {
	path = s.getPath(path)
	_, err := s.client.StatObject(context.TODO(), s.bucket, path, minio.GetObjectOptions{})
	return err == nil
}

func (s *s3Uploader) MkDir(path string, recursive bool) error {
	// S3 does not require directories to be created explicitly
	return nil
}

func (s *s3Uploader) Read(path string) (io.ReadCloser, error) {
	if !s.Exists(path) {
		return nil, os.ErrNotExist
	}
	path = s.getPath(path)
	return s.client.GetObject(context.TODO(), s.bucket, path, minio.GetObjectOptions{})
}

func (s *s3Uploader) AbsolutePath(path string) string {
	return s.getPath(path)
}

func (s *s3Uploader) Stat(file string) (FileInfo, error) {
	path := s.getPath(file)
	object, err := s.client.StatObject(context.TODO(), s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	return NewFileInfo(path, uint64(object.Size), object.LastModified), nil
}

func (s *s3Uploader) Put(path string, content io.Reader) (FileInfo, error) {
	path = s.getPath(path)
	put, err := s.localUploader.Put(path, content)
	if err != nil {
		return nil, err
	}
	defer s.localUploader.Delete(put.Path())
	return s.uploadToS3(path, put.Path())
}

func (s *s3Uploader) uploadToS3(path, localPath string) (FileInfo, error) {
	object, err := s.client.FPutObject(context.TODO(), s.bucket, path, localPath, minio.PutObjectOptions{})
	if err != nil {
		return nil, err
	}
	return NewFileInfo(object.Key, uint64(object.Size), object.LastModified), nil
}

func (s *s3Uploader) AllDirectoryFiles(dir string) ([]FileInfo, error) {
	dir = s.getPath(dir)
	objects := s.client.ListObjects(context.TODO(), s.bucket, minio.ListObjectsOptions{
		Prefix:    dir,
		Recursive: true,
	})
	var files []FileInfo
	for object := range objects {
		if object.Err != nil {
			return nil, object.Err
		}
		files = append(files, NewFileInfo(object.Key, uint64(object.Size), object.LastModified))
	}
	return files, nil
}

func (s *s3Uploader) RemoveEmptyDir() error {
	return s.localUploader.RemoveEmptyDir()
}

func (s *s3Uploader) UnWrap() Uploader {
	return s
}

func (s *s3Uploader) LocalUploader() Uploader {
	return s.localUploader
}

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

type s3File struct {
	closeable.Closeable
	File
	localUploader Uploader
	s3            *s3Uploader
	name          string
}

func (s *s3File) Name() string {
	return s.name
}

type s3OsFileInfo struct {
	name string
	os.FileInfo
}

func (s *s3OsFileInfo) Name() string {
	return s.name
}

func (s *s3File) Seek(offset int64, whence int) (ret int64, err error) {
	return s.File.Seek(offset, whence)
}

func (s *s3File) Stat() (os.FileInfo, error) {
	stat, err := s.File.Stat()
	if err != nil {
		return nil, err
	}
	return &s3OsFileInfo{name: s.name, FileInfo: stat}, nil
}

func (s *s3File) Close() error {
	if s.Closeable.Close() {
		if err := s.File.Close(); err != nil {
			return err
		}
		defer s.localUploader.Delete(s.File.Name())
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

func (s *s3Uploader) getPath(path string) string {
	if strings.HasPrefix(path, s.root()) {
		return path
	}
	return filepath.Join(s.root(), path)
}

func (s *s3Uploader) root() string {
	if s.disk != "" {
		return filepath.Join(s.rootDir, s.disk)
	}
	return s.rootDir
}
