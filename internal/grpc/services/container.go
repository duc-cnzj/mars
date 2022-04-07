package services

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/duc-cnzj/mars-client/v4/container"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/utils"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		container.RegisterContainerSvcServer(s, new(ContainerSvc))
	})
	RegisterEndpoint(container.RegisterContainerSvcHandlerFromEndpoint)
}

type ContainerSvc struct {
	container.UnsafeContainerSvcServer
}

func (c *ContainerSvc) IsPodRunning(_ context.Context, request *container.ContainerIsPodRunningRequest) (*container.ContainerIsPodRunningResponse, error) {
	running, reason := utils.IsPodRunning(request.GetNamespace(), request.GetPod())

	return &container.ContainerIsPodRunningResponse{Running: running, Reason: reason}, nil
}

func (c *ContainerSvc) IsPodExists(_ context.Context, request *container.ContainerIsPodExistsRequest) (*container.ContainerIsPodExistsResponse, error) {
	_, err := app.K8sClientSet().CoreV1().Pods(request.Namespace).Get(context.TODO(), request.Pod, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		return &container.ContainerIsPodExistsResponse{Exists: false}, nil
	}

	return &container.ContainerIsPodExistsResponse{Exists: true}, nil
}

func (c *ContainerSvc) Exec(request *container.ContainerExecRequest, server container.ContainerSvc_ExecServer) error {
	running, reason := utils.IsPodRunning(request.Namespace, request.Pod)
	if !running {
		return errors.New(reason)
	}

	if request.Container == "" {
		pod, _ := app.K8sClientSet().CoreV1().Pods(request.Namespace).Get(context.TODO(), request.Pod, metav1.GetOptions{})
		for _, co := range pod.Spec.Containers {
			request.Container = co.Name
			mlog.Debug("使用第一个容器: ", co.Name)
			break
		}
	}

	peo := &v1.PodExecOptions{
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
		Container: request.Container,
		Command:   request.Command,
	}

	req := app.K8sClient().Client.CoreV1().
		RESTClient().
		Post().
		Namespace(request.Namespace).
		Resource("pods").
		SubResource("exec").
		Name(request.Pod)

	params := req.VersionedParams(peo, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(app.K8sClient().RestConfig, "POST", params.URL())
	if err != nil {
		return err
	}

	rw := newExecReaderWriter(1024 * 1024)
	defer rw.Close()

	go func() {
		defer rw.Close()
		defer utils.HandlePanic("Exec")
		if err := exec.Stream(remotecommand.StreamOptions{
			Stdin:             rw,
			Stdout:            rw,
			Stderr:            rw,
			Tty:               false,
			TerminalSizeQueue: nil,
		}); err != nil {
			mlog.Error(err)
			return
		}
	}()

	for {
		select {
		case msg, ok := <-rw.ch:
			if !ok {
				return nil
			}
			if err := server.Send(&container.ContainerExecResponse{
				Data: msg,
			}); err != nil {
				return err
			}
		case <-server.Context().Done():
			return server.Context().Err()
		}
	}
}

func (c *ContainerSvc) CopyToPod(ctx context.Context, request *container.ContainerCopyToPodRequest) (*container.ContainerCopyToPodResponse, error) {
	if running, reason := utils.IsPodRunning(request.Namespace, request.Pod); !running {
		return nil, status.Error(codes.NotFound, reason)
	}

	var file models.File
	if err := app.DB().First(&file, request.FileId).Error; err != nil {
		return nil, err
	}
	res, err := utils.CopyFileToPod(request.Namespace, request.Pod, request.Container, file.Path, "")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	app.DB().Model(&file).Updates(map[string]any{
		"namespace":      request.Namespace,
		"pod":            request.Pod,
		"container":      request.Container,
		"container_path": res.ContainerPath,
	})

	FileAuditLog(MustGetUser(ctx).Name,
		fmt.Sprintf("上传文件到 pod: %s/%s/%s, 容器路径: '%s', 大小: %s。",
			request.Namespace,
			request.Pod,
			request.Container,
			res.ContainerPath,
			humanize.Bytes(file.Size),
		), file.ID)

	return &container.ContainerCopyToPodResponse{
		PodFilePath: res.TargetDir,
		Output:      res.ErrOut,
		FileName:    res.FileName,
	}, err
}

func (c *ContainerSvc) StreamCopyToPod(server container.ContainerSvc_StreamCopyToPodServer) error {
	var (
		fpath         string
		namespace     string
		pod           string
		containerName string
		user          = MustGetUser(server.Context())
		f             *os.File
		disk          = "grpc_upload"
	)
	defer f.Close()

	for {
		recv, err := server.Recv()
		if err != nil {
			if err == io.EOF && f != nil {
				stat, _ := f.Stat()
				f.Close()

				file := models.File{Path: f.Name(), Username: user.Name, Size: uint64(stat.Size())}
				app.DB().Create(&file)
				res, err := c.CopyToPod(server.Context(), &container.ContainerCopyToPodRequest{
					FileId:    int64(file.ID),
					Namespace: namespace,
					Pod:       pod,
					Container: containerName,
				})
				if err != nil {
					return err
				}
				return server.SendAndClose(&container.ContainerStreamCopyToPodResponse{
					Size:        stat.Size(),
					PodFilePath: res.PodFilePath,
					Output:      res.Output,
					Pod:         pod,
					Namespace:   namespace,
					Container:   containerName,
					Filename:    res.FileName,
				})
			}
			if f != nil {
				f.Close()
				app.Uploader().Disk(disk).Delete(f.Name())
			}
			return err
		}
		if fpath == "" {
			pod = recv.Pod
			namespace = recv.Namespace
			if recv.Container == "" {
				pod, err := app.K8sClientSet().CoreV1().Pods(recv.Namespace).Get(context.TODO(), recv.Pod, metav1.GetOptions{})
				if err != nil {
					return err
				}
				for _, co := range pod.Spec.Containers {
					recv.Container = co.Name
					mlog.Debug("使用第一个容器: ", co.Name)
					break
				}
			}
			containerName = recv.Container
			running, reason := utils.IsPodRunning(recv.Namespace, recv.Pod)
			if !running {
				return errors.New(reason)
			}

			// 某个用户/那天/时间/文件名称
			// duc/2006-01-02/15-03-04-random-str/xxx.tgz
			p := fmt.Sprintf("%s/%s/%s/%s",
				user.Name,
				time.Now().Format("2006-01-02"),
				fmt.Sprintf("%s-%s", time.Now().Format("15-04-05"), utils.RandomString(20)),
				filepath.Base(recv.GetFileName()))
			fpath = app.Uploader().Disk(disk).AbsolutePath(p)
			err := app.Uploader().Disk(disk).MkDir(filepath.Dir(p), true)
			if err != nil {
				mlog.Error(err)
				return err
			}
			f, err = os.Create(fpath)
			if err != nil {
				mlog.Error(err)
				return err
			}
		}

		f.Write(recv.GetData())
	}
}

func (c *ContainerSvc) ContainerLog(ctx context.Context, request *container.ContainerLogRequest) (*container.ContainerLogResponse, error) {
	if running, reason := utils.IsPodRunning(request.Namespace, request.Pod); !running {
		return nil, status.Errorf(codes.NotFound, reason)
	}

	var limit int64 = 2000
	logs := app.K8sClientSet().CoreV1().Pods(request.Namespace).GetLogs(request.Pod, &v1.PodLogOptions{
		Container: request.Container,
		TailLines: &limit,
	})
	do := logs.Do(context.Background())
	raw, err := do.Raw()
	if err != nil {
		return nil, err
	}

	return &container.ContainerLogResponse{
		Namespace:     request.Namespace,
		PodName:       request.Pod,
		ContainerName: request.Container,
		Log:           string(raw),
	}, nil
}

func (c *ContainerSvc) StreamContainerLog(request *container.ContainerLogRequest, server container.ContainerSvc_StreamContainerLogServer) error {
	if running, reason := utils.IsPodRunning(request.Namespace, request.Pod); !running {
		return status.Errorf(codes.NotFound, reason)
	}

	var limit int64 = 2000
	logs := app.K8sClientSet().CoreV1().Pods(request.Namespace).GetLogs(request.Pod, &v1.PodLogOptions{
		Follow:    true,
		Container: request.Container,
		TailLines: &limit,
	})
	stream, err := logs.Stream(context.TODO())
	if err != nil {
		return err
	}
	bf := bufio.NewReader(stream)

	ch := make(chan []byte)
	go func() {
		defer func() {
			mlog.Debug("[Stream]:  read exit!")
			close(ch)
		}()
		defer utils.HandlePanic("StreamContainerLog")

		for {
			bytes, err := bf.ReadBytes('\n')
			if err != nil {
				mlog.Debugf("[Stream]: %v", err)
				return
			}
			ch <- bytes
		}
	}()

	for {
		select {
		case <-app.App().Done():
			stream.Close()
			err := errors.New("server shutdown")
			mlog.Debug("[Stream]: client exit with: ", err)
			return err
		case <-server.Context().Done():
			stream.Close()
			mlog.Debug("[Stream]: client exit with: ", server.Context().Err())
			return server.Context().Err()
		case msg, ok := <-ch:
			if !ok {
				stream.Close()
				return errors.New("[Stream]: channel close")
			}

			if err := server.Send(&container.ContainerLogResponse{
				Namespace:     request.Namespace,
				PodName:       request.Pod,
				ContainerName: request.Container,
				Log:           string(msg),
			}); err != nil {
				stream.Close()
				return err
			}
		}
	}
}

type execReaderWriter struct {
	reader   io.Reader
	closedCh chan struct{}
	ch       chan string
}

func (rw *execReaderWriter) IsClosed() bool {
	select {
	case _, ok := <-rw.closedCh:
		if !ok {
			return true
		}
	default:
	}
	return false
}

func (rw *execReaderWriter) Close() error {
	if rw.IsClosed() {
		return nil
	}
	close(rw.ch)
	close(rw.closedCh)

	return nil
}

func newExecReaderWriter(size int) *execReaderWriter {
	return &execReaderWriter{
		reader:   bytes.NewReader(make([]byte, size)),
		closedCh: make(chan struct{}, 1),
		ch:       make(chan string, 100),
	}
}

func (rw *execReaderWriter) Read(p []byte) (int, error) {
	select {
	case <-rw.closedCh:
		return 0, errors.New("closed")
	default:
	}
	return len(p), nil
}

func (rw *execReaderWriter) Write(p []byte) (int, error) {
	select {
	case <-rw.closedCh:
		mlog.Warning("close")
		return 0, errors.New("closed")
	case rw.ch <- string(p):
		return len(p), nil
	}
}
