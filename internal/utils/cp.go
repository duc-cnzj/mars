package utils

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/mholt/archiver/v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

type CopyFileToPodResult struct {
	TargetDir string
	Output    string
}

func CopyFileToPod(namespace, pod, container, fpath, targetContainerDir string) (*CopyFileToPodResult, error) {
	var (
		bf                = bytes.NewBuffer([]byte{})
		reader, outStream = io.Pipe()
	)
	if targetContainerDir == "" {
		targetContainerDir = "/tmp"
	}
	_, err := os.Stat(fpath)
	if err != nil {
		mlog.Error(err)
		return nil, err
	}
	base := filepath.Base(fpath)
	dir := filepath.Dir(fpath)
	path := filepath.Join(dir, base+".tar.gz")
	mlog.Debugf("[CopyFileToPod]: %v", path)
	if err := archiver.Archive([]string{fpath}, path); err != nil {
		return nil, err
	}
	src, err := os.Open(path)
	if err != nil {
		mlog.Error(err)
		return nil, err
	}
	go func() {
		defer func() {
			reader.Close()
			outStream.Close()
			src.Close()
		}()

		if _, err := io.Copy(outStream, src); err != nil {
			mlog.Error(err)
		}
	}()
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
		Stdout: bf,
		Stderr: bf,
	})

	return &CopyFileToPodResult{
		TargetDir: targetContainerDir,
		Output:    bf.String(),
	}, err
}
