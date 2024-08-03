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

	"github.com/duc-cnzj/mars/api/v4/container"
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/uploader"
	"github.com/duc-cnzj/mars/v4/internal/util/closeable"
	"github.com/dustin/go-humanize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/rand"
	clientgoexec "k8s.io/client-go/util/exec"
)

var _ container.ContainerServer = (*containerSvc)(nil)

type containerSvc struct {
	k8sRepo  repo.K8sRepo
	fileRepo repo.FileRepo
	logger   mlog.Logger

	eventRepo repo.EventRepo

	container.UnimplementedContainerServer
}

func NewContainerSvc(eventRepo repo.EventRepo, k8sRepo repo.K8sRepo, fileRepo repo.FileRepo, logger mlog.Logger) container.ContainerServer {
	return &containerSvc{eventRepo: eventRepo, k8sRepo: k8sRepo, fileRepo: fileRepo, logger: logger}
}

func (c *containerSvc) IsPodRunning(_ context.Context, request *container.IsPodRunningRequest) (*container.IsPodRunningResponse, error) {
	running, reason := c.k8sRepo.IsPodRunning(request.GetNamespace(), request.GetPod())

	return &container.IsPodRunningResponse{Running: running, Reason: reason}, nil
}

func (c *containerSvc) IsPodExists(_ context.Context, request *container.IsPodExistsRequest) (*container.IsPodExistsResponse, error) {
	_, err := c.k8sRepo.GetPod(request.GetNamespace(), request.GetPod())
	if err != nil && apierrors.IsNotFound(err) {
		return &container.IsPodExistsResponse{Exists: false}, nil
	}

	return &container.IsPodExistsResponse{Exists: true}, nil
}

func (c *containerSvc) ContainerLog(ctx context.Context, request *container.LogRequest) (*container.LogResponse, error) {
	podInfo, _ := c.k8sRepo.GetPod(request.Namespace, request.Pod)
	if podInfo == nil || (!request.ShowEvents && podInfo != nil && podInfo.Status.Phase == v1.PodPending) {
		return nil, status.Error(codes.NotFound, "未找到日志")
	}

	if podInfo.Status.Phase == v1.PodPending {
		var logs []string
		ret, _ := c.k8sRepo.ListEvents(request.Namespace)
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

	logs, err := c.k8sRepo.GetPodLogs(request.Namespace, request.Pod, opt)
	if err != nil {
		return nil, err
	}

	return &container.LogResponse{
		Namespace:     request.Namespace,
		PodName:       request.Pod,
		ContainerName: request.Container,
		Log:           logs,
	}, nil
}

func (c *containerSvc) CopyToPod(ctx context.Context, request *container.CopyToPodRequest) (*container.CopyToPodResponse, error) {
	if running, reason := c.k8sRepo.IsPodRunning(request.Namespace, request.Pod); !running {
		return nil, status.Error(codes.NotFound, reason)
	}

	file, err := c.fileRepo.GetByID(ctx, int(request.FileId))
	if err != nil {
		return nil, err
	}

	result, err := c.k8sRepo.Copy(ctx, request.Namespace, request.Pod, request.Container, file.Path, "")
	if err != nil {
		return nil, err
	}

	file.Update().
		SetNamespace(request.Namespace).
		SetPod(request.Pod).
		SetContainer(request.Container).
		SetContainerPath(result.ContainerPath).
		Save(ctx)

	c.eventRepo.FileAuditLog(
		types.EventActionType_Upload,
		MustGetUser(ctx).Name,
		fmt.Sprintf("上传文件到 pod: %s/%s/%s, 容器路径: '%s', 大小: %s。",
			request.Namespace,
			request.Pod,
			request.Container,
			result.ContainerPath,
			humanize.Bytes(file.Size),
		), file.ID)

	return &container.CopyToPodResponse{
		PodFilePath: result.TargetDir,
		Output:      result.ErrOut,
		FileName:    result.FileName,
	}, err
}

func (c *containerSvc) StreamContainerLog(request *container.LogRequest, server container.Container_StreamContainerLogServer) error {
	podInfo, _ := c.k8sRepo.GetPod(request.Namespace, request.Pod)
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

	stream, err := c.k8sRepo.LogStream(context.TODO(), request.Namespace, request.Pod, request.Container)
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
			c.logger.Debug("[LogStream]:  read exit!")
			close(ch)
			wg.Done()
		}()
		defer c.logger.HandlePanic("StreamContainerLog")

		for {
			bytes, err := bf.ReadBytes('\n')
			if err != nil {
				c.logger.Debugf("[LogStream]: %v", err)
				return
			}
			select {
			case ch <- bytes:
			default:
				c.logger.Debug("[LogStream]:  drop line!")
			}
		}
	}()

	for {
		select {
		// TODO
		//case <-c.app.Done():
		//	stream.Close()
		//	err := errors.New("server shutdown")
		//	c.logger.Debug("[LogStream]: client exit with: ", err)
		//	return err
		case <-server.Context().Done():
			stream.Close()
			c.logger.Debug("[LogStream]: client exit with: ", server.Context().Err())
			return nil
		case msg, ok := <-ch:
			if !ok {
				stream.Close()
				c.logger.Debug("[LogStream]: channel close")
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

func (c *containerSvc) StreamCopyToPod(server container.Container_StreamCopyToPodServer) error {
	var (
		fpath         string
		namespace     string
		pod           string
		containerName string
		user          = MustGetUser(server.Context())
		f             uploader.File

		disk = "grpc_upload"
	)
	defer func() {
		f.Close()
	}()
	updisk := c.fileRepo.NewDisk(disk)

	for {
		recv, err := server.Recv()
		if err != nil {
			if f != nil {
				stat, _ := f.Stat()
				f.Close()
				if err != io.EOF {
					return err
				}

				file, _ := c.fileRepo.Create(context.TODO(), &repo.CreateFileInput{
					Path:       f.Name(),
					Username:   user.Name,
					Size:       uint64(stat.Size()),
					UploadType: updisk.Type(),
				})

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
				pod, err := c.k8sRepo.GetPod(recv.Namespace, recv.Pod)
				if err != nil {
					return err
				}

				recv.Container = c.k8sRepo.FindDefaultContainer(pod)
				c.logger.Debug("使用默认的容器: ", recv.Container)
			}
			containerName = recv.Container
			running, reason := c.k8sRepo.IsPodRunning(recv.Namespace, recv.Pod)
			if !running {
				return errors.New(reason)
			}

			// 某个用户/那天/时间/文件名称
			// duc/2006-01-02/15-03-04-random-str/xxx.tgz
			p := fmt.Sprintf("%s/%s/%s/%s",
				user.Name,
				time.Now().Format("2006-01-02"),
				fmt.Sprintf("%s-%s", time.Now().Format("15-04-05"), rand.String(20)),
				filepath.Base(recv.GetFileName()))
			fpath = updisk.AbsolutePath(p)
			err := updisk.MkDir(filepath.Dir(p), true)
			if err != nil {
				c.logger.Error(err)
				return err
			}
			f, err = c.fileRepo.NewFile(fpath)
			if err != nil {
				c.logger.Error(err)
				return err
			}
		}

		f.Write(recv.GetData())
	}
}

type exitCodeStatus struct {
	message string
	code    int
}

func (c *containerSvc) Exec(request *container.ExecRequest, server container.Container_ExecServer) error {
	running, reason := c.k8sRepo.IsPodRunning(request.Namespace, request.Pod)
	if !running {
		return errors.New(reason)
	}

	if request.Container == "" {
		pod, _ := c.k8sRepo.GetPod(request.Namespace, request.Pod)
		request.Container = c.k8sRepo.FindDefaultContainer(pod)
		c.logger.Debug("使用默认的容器: ", request.Container)
	}

	var exitCode atomic.Value
	r := c.fileRepo.NewRecorder(types.EventActionType_Exec, auth.MustGetUser(server.Context()), &repo.Container{
		Namespace: request.Namespace,
		Pod:       request.Pod,
		Container: request.Container,
	})
	r.Write(fmt.Sprintf("mars@%s:/# %s", request.Container, strings.Join(request.Command, " ")))
	r.Write("\r\n")
	writer := newExecWriter(r, c.logger)
	defer writer.Close()

	go func() {
		defer writer.Close()
		defer c.logger.HandlePanic("Exec")
		err := c.k8sRepo.Execute(server.Context(), &repo.Container{
			Namespace: request.Namespace,
			Pod:       request.Pod,
			Container: request.Container,
		}, &repo.ExecuteInput{
			Stdout: writer,
			Stderr: writer,
			TTY:    true,
		})
		if err != nil {
			if exitError, ok := err.(clientgoexec.ExitError); ok && exitError.Exited() {
				c.logger.Debugf("[containerSvc]: exit %v, exit code: %d, err: %v", exitError.Exited(), exitError.ExitStatus(), exitError.Error())
				exitCode.Store(&exitCodeStatus{
					message: exitError.Error(),
					code:    exitError.ExitStatus(),
				})
			} else {
				c.logger.Error(err)
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

func scannerText(text string, fn func(s string)) error {
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		fn(scanner.Text())
	}
	return scanner.Err()
}

type execWriter struct {
	logger    mlog.Logger
	recorder  contracts.RecorderInterface
	closeable closeable.Closeable
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

func newExecWriter(r repo.Recorder, logger mlog.Logger) *execWriter {
	return &execWriter{
		recorder: r,
		logger:   logger,
		ch:       make(chan string, 100),
	}
}

func (rw *execWriter) Write(p []byte) (int, error) {
	if rw.closeable.IsClosed() {
		rw.logger.Warning("execWriter close")
		return 0, errors.New("closed")
	}
	rw.ch <- string(p)
	rw.recorder.Write(string(p))
	return len(p), nil
}
