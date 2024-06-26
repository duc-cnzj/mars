package services

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	eventsv1 "k8s.io/api/events/v1"

	"github.com/dustin/go-humanize"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	clientgoexec "k8s.io/client-go/util/exec"

	"github.com/duc-cnzj/mars/api/v4/container"
	"github.com/duc-cnzj/mars/api/v4/types"
	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/socket"
	"github.com/duc-cnzj/mars/v4/internal/utils"
	"github.com/duc-cnzj/mars/v4/internal/utils/executor"
	"github.com/duc-cnzj/mars/v4/internal/utils/recovery"
	"github.com/duc-cnzj/mars/v4/internal/utils/timer"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		ex := executor.NewDefaultRemoteExecutor()

		container.RegisterContainerServer(s, &containerSvc{
			NewRecorderFunc: socket.NewRecorder,
			Steamer:         &defaultStreamer{},
			Executor:        ex,
			PodFileCopier:   utils.NewFileCopier(ex, utils.NewDefaultArchiver()),
		})
	})
	RegisterEndpoint(container.RegisterContainerHandlerFromEndpoint)
}

type containerSvc struct {
	Steamer         Steamer
	Executor        contracts.RemoteExecutor
	PodFileCopier   contracts.PodFileCopier
	NewRecorderFunc func(types.EventActionType, contracts.UserInfo, timer.Timer, contracts.Container) contracts.RecorderInterface

	container.UnimplementedContainerServer
}

func (c *containerSvc) IsPodRunning(_ context.Context, request *container.IsPodRunningRequest) (*container.IsPodRunningResponse, error) {
	running, reason := utils.IsPodRunning(request.GetNamespace(), request.GetPod())

	return &container.IsPodRunningResponse{Running: running, Reason: reason}, nil
}

func (c *containerSvc) IsPodExists(_ context.Context, request *container.IsPodExistsRequest) (*container.IsPodExistsResponse, error) {
	_, err := app.K8sClient().PodLister.Pods(request.Namespace).Get(request.Pod)
	if err != nil && apierrors.IsNotFound(err) {
		return &container.IsPodExistsResponse{Exists: false}, nil
	}

	return &container.IsPodExistsResponse{Exists: true}, nil
}

type exitCodeStatus struct {
	message string
	code    int
}

func (c *containerSvc) Exec(request *container.ExecRequest, server container.Container_ExecServer) error {
	running, reason := utils.IsPodRunning(request.Namespace, request.Pod)
	if !running {
		return errors.New(reason)
	}

	if request.Container == "" {
		pod, _ := app.K8sClient().PodLister.Pods(request.Namespace).Get(request.Pod)
		request.Container = FindDefaultContainer(pod)
		mlog.Debug("使用默认的容器: ", request.Container)
	}

	var exitCode atomic.Value
	r := c.NewRecorderFunc(types.EventActionType_Exec, *auth.MustGetUser(server.Context()), timer.NewRealTimer(), contracts.Container{
		Namespace: request.Namespace,
		Pod:       request.Pod,
		Container: request.Container,
	})
	r.Write(fmt.Sprintf("mars@%s:/# %s", request.Container, strings.Join(request.Command, " ")))
	r.Write("\r\n")
	writer := newExecWriter(r)
	defer writer.Close()
	restConfig := app.K8sClient().RestConfig
	clientSet := app.K8sClient().Client

	go func() {
		defer writer.Close()
		defer recovery.HandlePanic("Exec")
		err := c.Executor.
			WithMethod("POST").
			WithContainer(request.Namespace, request.Pod, request.Container).
			WithCommand(request.Command).
			Execute(context.TODO(), clientSet, restConfig, nil, writer, writer, true, nil)
		if err != nil {
			if exitError, ok := err.(clientgoexec.ExitError); ok && exitError.Exited() {
				mlog.Debugf("[containerSvc]: exit %v, exit code: %d, err: %v", exitError.Exited(), exitError.ExitStatus(), exitError.Error())
				exitCode.Store(&exitCodeStatus{
					message: exitError.Error(),
					code:    exitError.ExitStatus(),
				})
			} else {
				mlog.Error(err)
				exitCode.Store(&exitCodeStatus{
					message: err.Error(),
					code:    1,
				})
			}
		}
	}()

	for {
		select {
		case msg, ok := <-writer.ch:
			if !ok {
				ec := exitCode.Load()
				if ec != nil {
					ecs := ec.(*exitCodeStatus)
					server.Send(&container.ExecResponse{
						Error: &container.ExecError{
							Code:    int64(ecs.code),
							Message: ecs.message,
						},
					})
				}
				return nil
			}
			if err := server.Send(&container.ExecResponse{
				Message: msg,
			}); err != nil {
				return err
			}
		case <-server.Context().Done():
			writer.Close()
			return server.Context().Err()
		}
	}
}

func (c *containerSvc) CopyToPod(ctx context.Context, request *container.CopyToPodRequest) (*container.CopyToPodResponse, error) {
	if running, reason := utils.IsPodRunning(request.Namespace, request.Pod); !running {
		return nil, status.Error(codes.NotFound, reason)
	}

	var file models.File
	if err := app.DB().First(&file, request.FileId).Error; err != nil {
		return nil, err
	}
	res, err := c.PodFileCopier.Copy(request.Namespace, request.Pod, request.Container, file.Path, "", app.K8sClientSet(), app.K8sClient().RestConfig)
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

	return &container.CopyToPodResponse{
		PodFilePath: res.TargetDir,
		Output:      res.ErrOut,
		FileName:    res.FileName,
	}, err
}

func (c *containerSvc) StreamCopyToPod(server container.Container_StreamCopyToPodServer) error {
	var (
		fpath         string
		namespace     string
		pod           string
		containerName string
		user          = MustGetUser(server.Context())
		f             contracts.File
		disk          = "grpc_upload"
	)
	defer func() {
		f.Close()
	}()
	updisk := app.Uploader().Disk(disk)

	for {
		recv, err := server.Recv()
		if err != nil {
			if f != nil {
				stat, _ := f.Stat()
				f.Close()
				if err != io.EOF {
					return err
				}

				file := models.File{Path: f.Name(), Username: user.Name, Size: uint64(stat.Size()), UploadType: updisk.Type()}
				app.DB().Create(&file)
				res, err := c.CopyToPod(server.Context(), &container.CopyToPodRequest{
					FileId:    int64(file.ID),
					Namespace: namespace,
					Pod:       pod,
					Container: containerName,
				})
				if err != nil {
					return err
				}
				return server.SendAndClose(&container.StreamCopyToPodResponse{
					Size:        stat.Size(),
					PodFilePath: res.PodFilePath,
					Output:      res.Output,
					Pod:         pod,
					Namespace:   namespace,
					Container:   containerName,
					Filename:    res.FileName,
				})
			}

			return err
		}
		if fpath == "" {
			pod = recv.Pod
			namespace = recv.Namespace
			if recv.Container == "" {
				pod, err := app.K8sClient().PodLister.Pods(recv.Namespace).Get(recv.Pod)
				if err != nil {
					return err
				}

				recv.Container = FindDefaultContainer(pod)
				mlog.Debug("使用默认的容器: ", recv.Container)
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
			fpath = updisk.AbsolutePath(p)
			err := updisk.MkDir(filepath.Dir(p), true)
			if err != nil {
				mlog.Error(err)
				return err
			}
			f, err = app.Uploader().NewFile(fpath)
			if err != nil {
				mlog.Error(err)
				return err
			}
		}

		f.Write(recv.GetData())
	}
}

var tailLines int64 = 1000

type sortEvents []*eventsv1.Event

func (s sortEvents) Len() int {
	return len(s)
}

func (s sortEvents) Less(i, j int) bool {
	return s[i].ResourceVersion < s[j].ResourceVersion
}

func (s sortEvents) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (c *containerSvc) ContainerLog(ctx context.Context, request *container.LogRequest) (*container.LogResponse, error) {
	podInfo, _ := app.K8sClient().PodLister.Pods(request.Namespace).Get(request.Pod)
	if podInfo == nil || (!request.ShowEvents && podInfo != nil && podInfo.Status.Phase == v1.PodPending) {
		return nil, status.Error(codes.NotFound, "未找到日志")
	}

	if podInfo.Status.Phase == v1.PodPending {
		var logs []string
		ret, _ := app.K8sClient().EventLister.Events(request.Namespace).List(labels.Everything())
		sort.Sort(sortEvents(ret))
		for _, event := range ret {
			if event.Regarding.Kind == "Pod" && event.Regarding.Name == request.Pod {
				logs = append(logs, event.Note)
			}
		}
		return &container.LogResponse{
			Namespace:     request.Namespace,
			PodName:       request.Pod,
			ContainerName: request.Container,
			Log:           strings.Join(logs, "\n"),
		}, nil
	}

	var opt = &v1.PodLogOptions{
		Container: request.Container,
	}

	if podInfo.Status.Phase == v1.PodRunning {
		opt.TailLines = &tailLines
	}

	logs := app.K8sClientSet().CoreV1().Pods(request.Namespace).GetLogs(request.Pod, opt)
	do := logs.Do(context.Background())
	raw, err := do.Raw()
	if err != nil {
		return nil, err
	}

	return &container.LogResponse{
		Namespace:     request.Namespace,
		PodName:       request.Pod,
		ContainerName: request.Container,
		Log:           string(raw),
	}, nil
}

type Steamer interface {
	Stream(ctx context.Context, namespace, pod, container string) (io.ReadCloser, error)
}

type defaultStreamer struct{}

func (d *defaultStreamer) Stream(ctx context.Context, namespace, pod, container string) (io.ReadCloser, error) {
	logs := app.K8sClientSet().CoreV1().Pods(namespace).GetLogs(pod, &v1.PodLogOptions{
		Follow:    true,
		Container: container,
		TailLines: &tailLines,
	})

	return logs.Stream(ctx)
}

func scannerText(text string, fn func(s string)) error {
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		fn(scanner.Text())
	}
	return scanner.Err()
}

func (c *containerSvc) StreamContainerLog(request *container.LogRequest, server container.Container_StreamContainerLogServer) error {
	podInfo, _ := app.K8sClient().PodLister.Pods(request.Namespace).Get(request.Pod)
	if podInfo == nil || (!request.ShowEvents && podInfo != nil && podInfo.Status.Phase == v1.PodPending) {
		return status.Error(codes.NotFound, "未找到日志")
	}

	if podInfo.Status.Phase == v1.PodSucceeded || podInfo.Status.Phase == v1.PodFailed || podInfo.Status.Phase == v1.PodPending {
		log, err := c.ContainerLog(server.Context(), request)
		if err != nil {
			return err
		}

		return scannerText(log.Log, func(s string) {
			server.Send(&container.LogResponse{
				Namespace:     request.Namespace,
				PodName:       request.Pod,
				ContainerName: request.Container,
				Log:           s,
			})
		})
	}

	stream, err := c.Steamer.Stream(context.TODO(), request.Namespace, request.Pod, request.Container)
	if err != nil {
		return err
	}
	bf := bufio.NewReader(stream)

	ch := make(chan []byte, 100)
	wg := sync.WaitGroup{}
	defer wg.Wait()
	wg.Add(1)
	go func() {
		defer func() {
			mlog.Debug("[Stream]:  read exit!")
			close(ch)
			wg.Done()
		}()
		defer recovery.HandlePanic("StreamContainerLog")

		for {
			bytes, err := bf.ReadBytes('\n')
			if err != nil {
				mlog.Debugf("[Stream]: %v", err)
				return
			}
			select {
			case ch <- bytes:
			default:
				mlog.Debug("[Stream]:  drop line!")
			}
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
			return nil
		case msg, ok := <-ch:
			if !ok {
				stream.Close()
				mlog.Debug("[Stream]: channel close")
				return nil
			}

			if err := server.Send(&container.LogResponse{
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

type execWriter struct {
	recorder  contracts.RecorderInterface
	closeable utils.Closeable
	ch        chan string
}

func (rw *execWriter) IsClosed() bool {
	return rw.closeable.IsClosed()
}

func (rw *execWriter) Close() error {
	if rw.closeable.Close() {
		rw.recorder.Close()
		close(rw.ch)
	}

	return nil
}

func newExecWriter(r contracts.RecorderInterface) *execWriter {
	return &execWriter{
		recorder: r,
		ch:       make(chan string, 100),
	}
}

func (rw *execWriter) Write(p []byte) (int, error) {
	if rw.closeable.IsClosed() {
		mlog.Warning("execWriter close")
		return 0, errors.New("closed")
	}
	rw.ch <- string(p)
	rw.recorder.Write(string(p))
	return len(p), nil
}

const defaultContainerAnnotationName = "kubectl.kubernetes.io/default-container"

func FindDefaultContainer(pod *v1.Pod) string {
	if name := pod.Annotations[defaultContainerAnnotationName]; len(name) > 0 {
		for _, co := range pod.Spec.Containers {
			if name == co.Name {
				return name
			}
		}
	}

	for _, co := range pod.Spec.Containers {
		return co.Name
	}

	return ""
}
