package uploader

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/duc-cnzj/mars/internal/utils"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/minio/minio-go/v7"
)

type S3 struct {
	localUploader contracts.Uploader
	client        *minio.Client
	bucket        string
	rootDir       string
	disk          string
}

func NewS3(client *minio.Client, bucket string, uploader contracts.Uploader, rootDir string) *S3 {
	return &S3{client: client, bucket: bucket, localUploader: uploader, rootDir: rootDir}
}

func (s *S3) Type() contracts.UploadType {
	return contracts.S3
}

func (s *S3) Disk(disk string) contracts.Uploader {
	return &S3{
		localUploader: s.localUploader.Disk(disk),
		client:        s.client,
		bucket:        s.bucket,
		rootDir:       s.root(),
		disk:          disk,
	}
}

func (s *S3) DeleteDir(dir string) error {
	dir = s.getPath(dir)
	s.localUploader.DeleteDir(dir)
	return s.Delete(dir)
}

func (s *S3) DirSize(dir string) (int64, error) {
	dir = s.getPath(dir)
	objects := s.client.ListObjects(context.TODO(), s.bucket, minio.ListObjectsOptions{
		Prefix:    dir,
		Recursive: true,
		MaxKeys:   0,
	})
	var size int64
	for object := range objects {
		size += object.Size
	}
	return size, nil
}

func (s *S3) Delete(path string) error {
	path = s.getPath(path)
	s.localUploader.Delete(path)
	return s.client.RemoveObject(context.TODO(), s.bucket, path, minio.RemoveObjectOptions{
		ForceDelete: true,
	})
}

func (s *S3) Exists(path string) bool {
	path = s.getPath(path)
	obj, err := s.client.GetObject(context.TODO(), s.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return false
	}
	_, err = obj.Stat()
	return err == nil
}

func (s *S3) MkDir(path string, recursive bool) error {
	return nil
}

func (s *S3) Read(path string) (io.ReadCloser, error) {
	if !s.Exists(path) {
		return nil, os.ErrNotExist
	}
	path = s.getPath(path)
	return s.client.GetObject(context.TODO(), s.bucket, path, minio.GetObjectOptions{})
}

func (s *S3) AbsolutePath(path string) string {
	return s.getPath(path)
}

func (s *S3) Stat(file string) (contracts.FileInfo, error) {
	path := s.getPath(file)
	object, err := s.client.StatObject(context.TODO(), s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	return &fileInfo{
		path: path,
		size: uint64(object.Size),
	}, nil
}

func (s *S3) Put(path string, content io.Reader) (contracts.FileInfo, error) {
	path = s.getPath(path)
	put, err := s.localUploader.Put(path, content)
	if err != nil {
		return nil, err
	}
	defer s.localUploader.Delete(put.Path())

	return s.put(path, put.Path())
}

func (s *S3) put(path string, localPath string) (contracts.FileInfo, error) {
	object, err := s.client.FPutObject(context.TODO(), s.bucket, path, localPath, minio.PutObjectOptions{})
	if err != nil {
		return nil, err
	}
	return &fileInfo{
		path: object.Key,
		size: uint64(object.Size),
	}, nil
}

func (s *S3) AllDirectoryFiles(dir string) ([]contracts.FileInfo, error) {
	dir = s.getPath(dir)
	objects := s.client.ListObjects(context.TODO(), s.bucket, minio.ListObjectsOptions{
		Prefix:    dir,
		Recursive: true,
	})
	var finfos []contracts.FileInfo
	for object := range objects {
		finfos = append(finfos, &fileInfo{
			path: object.Key,
			size: uint64(object.Size),
		})
	}
	return finfos, nil
}

func (s *S3) RemoveEmptyDir(dir string) error {
	return s.localUploader.RemoveEmptyDir(dir)
}

func (s *S3) NewFile(path string) (contracts.File, error) {
	file, err := s.localUploader.NewFile(s.getPath(path))
	if err != nil {
		return nil, err
	}
	return &s3File{
		localUploader: s.localUploader,
		s3:            s,
		name:          s.getPath(path),
		File:          file,
	}, err
}

type s3File struct {
	utils.Closeable
	localUploader contracts.Uploader
	s3            *S3
	contracts.File
	name string
}

func (s *s3File) Name() string {
	return s.name
}

func (s *s3File) Close() error {
	if s.Closeable.Close() {
		s.File.Close()
		defer s.localUploader.Delete(s.File.Name())
		open, err := s.localUploader.Read(s.File.Name())
		if err != nil {
			return err
		}
		defer open.Close()
		_, err = s.s3.put(s.name, s.File.Name())
		if err != nil {
			mlog.Error(err)
		}
		return err
	}
	return nil
}

func (s *S3) getPath(path string) string {
	if strings.HasPrefix(path, s.root()) {
		return path
	}
	return filepath.Join(s.root(), path)
}

func (s *S3) root() string {
	if s.disk != "" {
		return filepath.Join(s.rootDir, s.disk)
	}

	return s.rootDir
}
