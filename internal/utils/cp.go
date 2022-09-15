package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"

	"github.com/duc-cnzj/mars/internal/contracts"

	"github.com/duc-cnzj/mars/internal/utils/recovery"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"

	"github.com/dustin/go-humanize"
	"github.com/mholt/archiver/v3"
)

type defaultArchiver struct{}

func NewDefaultArchiver() *defaultArchiver {
	return &defaultArchiver{}
}

func (m *defaultArchiver) Archive(sources []string, destination string) error {
	return archiver.Archive(sources, destination)
}

func (m *defaultArchiver) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (m *defaultArchiver) Remove(path string) error {
	return os.Remove(path)
}

type fileCopier struct {
	archiver contracts.Archiver
	executor contracts.RemoteExecutor
}

func NewFileCopier(executor contracts.RemoteExecutor, archiver contracts.Archiver) *fileCopier {
	return &fileCopier{executor: executor, archiver: archiver}
}

func (fc *fileCopier) Copy(namespace, pod, container, fpath, targetContainerDir string, clientSet kubernetes.Interface, config *restclient.Config) (*contracts.CopyFileToPodResult, error) {
	var (
		errbf, outbf      = bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
		reader, outStream = io.Pipe()
		uploader          = app.Uploader()
		localUploader     = app.LocalUploader()
	)
	if targetContainerDir == "" {
		targetContainerDir = "/tmp"
	}
	st, err := uploader.Stat(fpath)
	if err != nil {
		return nil, err
	}
	if st.Size() > app.Config().MaxUploadSize() {
		return nil, fmt.Errorf("最大不得超过 %s, 你上传的文件大小是 %s", humanize.Bytes(app.Config().MaxUploadSize()), humanize.Bytes(uint64(st.Size())))
	}

	baseName := filepath.Base(fpath)
	path := filepath.Join(filepath.Dir(fpath), baseName+".tar.gz")
	mlog.Debugf("[CopyFileToPod]: %v", path)
	var localPath string = fpath
	// 如果是非 local 类型的，需要远程下载到 local 进行打包，再上传到容器
	if uploader.Type() != contracts.Local {
		read, err := uploader.Read(fpath)
		if err != nil {
			return nil, err
		}
		defer read.Close()
		if localUploader.Exists(localPath) {
			localUploader.Delete(localPath)
		}
		put, err := localUploader.Put(localPath, read)
		if err != nil {
			return nil, err
		}
		localPath = put.Path()
		defer localUploader.Delete(localPath)
	}
	if err := fc.archiver.Archive([]string{localPath}, path); err != nil {
		return nil, err
	}
	defer fc.archiver.Remove(path)
	src, err := fc.archiver.Open(path)
	if err != nil {
		return nil, err
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()
	go func(reader *io.PipeReader, outStream *io.PipeWriter, src io.ReadCloser) {
		defer func() {
			reader.Close()
			outStream.Close()
			src.Close()
			wg.Done()
		}()
		defer recovery.HandlePanic("CopyFileToPod")

		if _, err := io.Copy(outStream, src); err != nil {
			mlog.Error(err)
		}
	}(reader, outStream, src)

	err = fc.executor.
		WithCommand([]string{"tar", "-zmxf", "-", "-C", targetContainerDir}).
		WithMethod("POST").
		WithContainer(namespace, pod, container).
		Execute(clientSet, config, reader, outbf, errbf, false, nil)

	return &contracts.CopyFileToPodResult{
		TargetDir:     targetContainerDir,
		ErrOut:        errbf.String(),
		StdOut:        outbf.String(),
		ContainerPath: filepath.Join(targetContainerDir, baseName),
		FileName:      baseName,
	}, err
}
