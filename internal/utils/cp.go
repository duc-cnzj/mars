package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/duc-cnzj/mars/internal/contracts"

	"github.com/duc-cnzj/mars/internal/utils/recovery"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"

	"github.com/dustin/go-humanize"
	"github.com/mholt/archiver/v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

type CopyFileToPodResult struct {
	TargetDir     string
	ErrOut        string
	StdOut        string
	ContainerPath string
	FileName      string
}

type CopyFileToPodFunc func(namespace, pod, container, fpath, targetContainerDir string) (*CopyFileToPodResult, error)

func CopyFileToPod(namespace, pod, container, fpath, targetContainerDir string) (*CopyFileToPodResult, error) {
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
	if err := archiver.Archive([]string{localPath}, path); err != nil {
		return nil, err
	}
	defer os.Remove(path)
	src, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	go func(reader *io.PipeReader, outStream *io.PipeWriter, src io.ReadCloser) {
		defer func() {
			reader.Close()
			outStream.Close()
			src.Close()
		}()
		defer recovery.HandlePanic("CopyFileToPod")

		if _, err := io.Copy(outStream, src); err != nil {
			mlog.Error(err)
		}
	}(reader, outStream, src)

	peo := &v1.PodExecOptions{
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		Container: container,
		Command:   []string{"tar", "-zmxf", "-", "-C", targetContainerDir},
	}

	req := app.K8sClientSet().CoreV1().
		RESTClient().
		Post().
		Namespace(namespace).
		Resource("pods").
		SubResource("exec").
		Name(pod)
	params := req.VersionedParams(peo, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(app.K8sClient().RestConfig, "POST", params.URL())
	if err != nil {
		mlog.Error(err)
		return nil, err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  reader,
		Stdout: outbf,
		Stderr: errbf,
	})

	return &CopyFileToPodResult{
		TargetDir:     targetContainerDir,
		ErrOut:        errbf.String(),
		StdOut:        outbf.String(),
		ContainerPath: filepath.Join(targetContainerDir, baseName),
		FileName:      baseName,
	}, err
}
