package services

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/duc-cnzj/mars/api/v4/container"
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/dustin/go-humanize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
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
	return &containerSvc{eventRepo: eventRepo, k8sRepo: k8sRepo, fileRepo: fileRepo, logger: logger.WithModule("services/container")}
}

func (c *containerSvc) IsPodRunning(_ context.Context, request *container.IsPodRunningRequest) (*container.IsPodRunningResponse, error) {
	running, reason := c.k8sRepo.IsPodRunning(request.GetNamespace(), request.GetPod())

	return &container.IsPodRunningResponse{Running: running, Reason: reason}, nil
}

func (c *containerSvc) IsPodExists(ctx context.Context, request *container.IsPodExistsRequest) (*container.IsPodExistsResponse, error) {
	_, err := c.k8sRepo.GetPod(request.GetNamespace(), request.GetPod())
	if err != nil {
		c.logger.ErrorCtx(ctx, err)
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
		if request.ShowEvents {
			ret, _ := c.k8sRepo.ListEvents(request.Namespace)
			sort.Sort(sortEvents(ret))
			for _, event := range ret {
				if event.Regarding.Kind == "Pod" && event.Regarding.Name == request.Pod {
					logs = append(logs, event.Note)
				}
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
		c.logger.ErrorCtx(ctx, err)
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

	file, err := c.k8sRepo.CopyFileToPod(ctx, &repo.CopyFileToPodRequest{
		FileId:    request.FileId,
		Namespace: request.Namespace,
		Pod:       request.Pod,
		Container: request.Container,
	})
	if err != nil {
		return nil, err
	}

	c.eventRepo.FileAuditLog(
		types.EventActionType_Upload,
		MustGetUser(ctx).Name,
		fmt.Sprintf("上传文件到 pod: %s/%s/%s, 容器路径: '%s', 大小: %s。",
			request.Namespace,
			request.Pod,
			request.Container,
			file.ContainerPath,
			humanize.Bytes(file.Size),
		),
		file.ID,
	)

	return &container.CopyToPodResponse{
		PodFilePath: file.ContainerPath,
		FileName:    file.Path,
	}, err
}

func (c *containerSvc) StreamCopyToPod(server container.Container_StreamCopyToPodServer) error {
	var (
		ctx  = server.Context()
		user = MustGetUser(server.Context())
	)
	recv, err := server.Recv()
	if err != nil {
		return err
	}
	c.logger.DebugCtx(ctx, "StreamUploadFile", recv.Namespace, recv.Pod, recv.Container, recv.FileName)
	// 判断 pod 是否存在
	running, reason := c.k8sRepo.IsPodRunning(recv.Namespace, recv.Pod)
	if !running {
		return errors.New(reason)
	}
	// 如果没传入 container，使用默认的
	if recv.Container == "" {
		recv.Container, err = c.k8sRepo.FindDefaultContainer(ctx, recv.Namespace, recv.Pod)
		if err != nil {
			return err
		}
	}
	c.logger.DebugCtxf(ctx, "StreamUploadFile %s/%s/%s", recv.Namespace, recv.Pod, recv.Container)
	ch := make(chan []byte, 100)
	ch <- recv.GetData()
	go func() {
		defer close(ch)

		for {
			recv, err := server.Recv()
			if err != nil {
				if err == io.EOF {
					c.logger.DebugCtx(ctx, "StreamUploadFile: EOF")
					return
				}
				c.logger.ErrorCtx(ctx, "StreamUploadFile: error receiving data", err)
				return
			}
			ch <- recv.GetData()
		}
	}()

	file, err := c.fileRepo.StreamUploadFile(ctx, &repo.StreamUploadFileRequest{
		Namespace: recv.Namespace,
		Pod:       recv.Pod,
		Container: recv.Container,
		Username:  user.Name,
		FileName:  recv.FileName,
		FileData:  ch,
	})
	if err != nil {
		c.logger.ErrorCtx(ctx, err)
		return err
	}

	res, err := c.k8sRepo.CopyFileToPod(ctx, &repo.CopyFileToPodRequest{
		FileId:    int64(file.ID),
		Namespace: file.Namespace,
		Pod:       file.Pod,
		Container: file.Container,
	})
	if err != nil {
		c.logger.ErrorCtx(ctx, err)
		return err
	}

	c.eventRepo.FileAuditLog(
		types.EventActionType_Upload,
		MustGetUser(ctx).Name,
		fmt.Sprintf("[StreamUploadFile]: 上传文件到 pod: %s/%s/%s, 容器路径: '%s', 大小: %s。",
			file.Namespace,
			file.Pod,
			file.Container,
			file.ContainerPath,
			humanize.Bytes(file.Size),
		),
		file.ID,
	)

	return server.SendAndClose(&container.StreamCopyToPodResponse{
		Size:        int64(file.Size),
		PodFilePath: res.ContainerPath,
		Pod:         file.Pod,
		Namespace:   file.Namespace,
		Container:   file.Container,
		Filename:    res.Path,
	})
}

var tailLines int64 = 1000

func (c *containerSvc) StreamContainerLog(request *container.LogRequest, server container.Container_StreamContainerLogServer) error {
	c.logger.DebugCtxf(server.Context(), "StreamContainerLog: %v", request)
	podInfo, _ := c.k8sRepo.GetPod(request.Namespace, request.Pod)
	if podInfo == nil || (!request.ShowEvents && podInfo != nil && podInfo.Status.Phase == v1.PodPending) {
		return status.Error(codes.NotFound, "未找到日志")
	}

	if podInfo.Status.Phase == v1.PodSucceeded || podInfo.Status.Phase == v1.PodFailed || podInfo.Status.Phase == v1.PodPending {
		log, err := c.ContainerLog(server.Context(), request)
		if err != nil {
			c.logger.ErrorCtx(server.Context(), err)
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

	ch, err := c.k8sRepo.LogStream(server.Context(), request.Namespace, request.Pod, request.Container)
	if err != nil {
		c.logger.ErrorCtx(server.Context(), err)
		return err
	}

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				c.logger.Debug("[LogStream]: channel close")
				return nil
			}

			if err = server.Send(&container.LogResponse{
				Namespace:     request.Namespace,
				PodName:       request.Pod,
				ContainerName: request.Container,
				Log:           string(msg),
			}); err != nil {
				c.logger.ErrorCtx(server.Context(), err)
				return err
			}
		}
	}
}

func (c *containerSvc) Exec(server container.Container_ExecServer) error {
	var (
		pod       string
		namespace string
		co        string
		cmd       []string
		ctx       = server.Context()
	)
	recv, err := server.Recv()
	if err != nil {
		return err
	}
	co = recv.Container
	namespace = recv.Namespace
	pod = recv.Pod
	cmd = recv.Command
	// 判断 pod 是否存在
	running, reason := c.k8sRepo.IsPodRunning(recv.Namespace, recv.Pod)
	if !running {
		return errors.New(reason)
	}

	if co == "" {
		co, err = c.k8sRepo.FindDefaultContainer(ctx, namespace, pod)
		if err != nil {
			return err
		}
		c.logger.Debug("使用默认的容器: ", co)
	}

	r := c.fileRepo.NewRecorder(
		types.EventActionType_Exec,
		MustGetUser(ctx),
		&repo.Container{
			Namespace: namespace,
			Pod:       pod,
			Container: co,
		},
	)

	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	go func() {
		c.logger.DebugCtx(ctx, "Exec: ", recv.Message)
		writer.Write([]byte(recv.Message))
		for {
			request, err := server.Recv()
			if err != nil {
				c.logger.DebugCtx(ctx, err)
				return
			}

			c.logger.DebugCtx(ctx, "Exec: ", request.Message)
			writer.Write([]byte(request.Message))
		}
	}()

	pipe, pipeWriter := io.Pipe()
	defer pipe.Close()
	defer pipeWriter.Close()
	w := NewMultiWriterCloser(pipeWriter, r)
	go func() {
		defer c.logger.DebugCtx(ctx, "Exec close")
		defer func() {
			c.eventRepo.FileAuditLogWithDuration(
				types.EventActionType_Exec,
				r.User().Name,
				fmt.Sprintf("[Exec]: 用户进入容器执行命令，container: '%s', namespace: '%s', pod： '%s'", r.Container().Container, r.Container().Namespace, r.Container().Pod),
				r.File().ID,
				r.Duration(),
			)
		}()
		scanner := bufio.NewScanner(pipe)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			server.Send(&container.ExecResponse{
				Message: scanner.Text(),
			})
		}
	}()

	err = c.k8sRepo.ExecuteTTY(ctx, &repo.ExecuteTTYInput{
		Container: &repo.Container{
			Namespace: namespace,
			Pod:       pod,
			Container: co,
		},
		Reader:      reader,
		WriteCloser: w,
		Cmd:         cmd,
		TTY:         true,
	})
	if exitError, ok := err.(clientgoexec.ExitError); ok {
		server.Send(&container.ExecResponse{
			Error: &container.ExecError{
				Code:    int64(exitError.ExitStatus()),
				Message: exitError.Error(),
			},
		})
	}

	return err
}

func (c *containerSvc) ExecOnce(request *container.ExecOnceRequest, server container.Container_ExecOnceServer) error {
	var (
		err error
		ctx = server.Context()
	)
	running, reason := c.k8sRepo.IsPodRunning(request.Namespace, request.Pod)
	if !running {
		return errors.New(reason)
	}

	if request.Container == "" {
		request.Container, err = c.k8sRepo.FindDefaultContainer(ctx, request.Namespace, request.Pod)
		if err != nil {
			return err
		}
		c.logger.Debug("使用默认的容器: ", request.Container)
	}

	r := c.fileRepo.NewRecorder(types.EventActionType_Exec, auth.MustGetUser(ctx), &repo.Container{
		Namespace: request.Namespace,
		Pod:       request.Pod,
		Container: request.Container,
	})
	r.Write([]byte(fmt.Sprintf("mars@%s:/# %s", request.Container, strings.Join(request.Command, " "))))
	r.Write([]byte("\r\n"))

	pipe, pipeWriter := io.Pipe()
	defer pipe.Close()
	defer pipeWriter.Close()
	w := NewMultiWriterCloser(pipeWriter, r)
	go func() {
		defer func() {
			c.logger.DebugCtx(ctx, "ExecOnce exit")
			c.eventRepo.FileAuditLogWithDuration(
				types.EventActionType_Exec,
				r.User().Name,
				fmt.Sprintf("[ExecOnce]: 用户进入容器执行命令，container: '%s', namespace: '%s', pod： '%s'", r.Container().Container, r.Container().Namespace, r.Container().Pod),
				r.File().ID,
				r.Duration(),
			)
		}()
		scanner := bufio.NewScanner(pipe)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			server.Send(&container.ExecResponse{
				Message: scanner.Text(),
			})
		}
	}()

	err = c.k8sRepo.ExecuteTTY(ctx, &repo.ExecuteTTYInput{
		Container: &repo.Container{
			Namespace: request.Namespace,
			Pod:       request.Pod,
			Container: request.Container,
		},
		WriteCloser: w,
		Cmd:         request.Command,
		TTY:         false,
	})
	if exitError, ok := err.(clientgoexec.ExitError); ok {
		server.Send(&container.ExecResponse{
			Error: &container.ExecError{
				Code:    int64(exitError.ExitStatus()),
				Message: exitError.Error(),
			},
		})
	}

	return err
}

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

// MultiWriterCloser combines multiple io.Writer and io.Closer interfaces.
type MultiWriterCloser struct {
	writers []io.Writer
	closers []io.Closer
}

// Write writes data to all the underlying writers.
func (mwc *MultiWriterCloser) Write(p []byte) (n int, err error) {
	for _, w := range mwc.writers {
		n, err = w.Write(p)
		if err != nil {
			return
		}
	}
	return
}

// Close closes all the underlying closers.
func (mwc *MultiWriterCloser) Close() error {
	for _, c := range mwc.closers {
		err := c.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// NewMultiWriterCloser creates a new MultiWriterCloser.
func NewMultiWriterCloser(writers ...io.Writer) *MultiWriterCloser {
	mwc := &MultiWriterCloser{}
	for _, w := range writers {
		mwc.writers = append(mwc.writers, w)
		if c, ok := w.(io.Closer); ok {
			mwc.closers = append(mwc.closers, c)
		}
	}
	return mwc
}
