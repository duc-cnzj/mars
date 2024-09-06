package services

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/duc-cnzj/mars/api/v5/container"
	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/dustin/go-humanize"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/client-go/tools/remotecommand"
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

	file, err := c.k8sRepo.CopyFileToPod(ctx, &repo.CopyFileToPodInput{
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

	res, err := c.k8sRepo.CopyFileToPod(ctx, &repo.CopyFileToPodInput{
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

	for msg := range ch {
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

	return nil
}

type sizeQueue struct {
	ch  chan *remotecommand.TerminalSize
	ctx context.Context
}

func (queue *sizeQueue) Next() *remotecommand.TerminalSize {
	select {
	case size, ok := <-queue.ch:
		if !ok {
			return nil
		}
		return size
	case <-queue.ctx.Done():
		return nil
	}
}

func (c *containerSvc) Exec(server container.Container_ExecServer) error {
	var (
		pod       string
		namespace string
		co        string
		cmd       []string
		once      sync.Once

		ctx, cancelFunc = context.WithCancel(server.Context())
	)
	defer cancelFunc()
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
		MustGetUser(ctx),
		&repo.Container{
			Namespace: namespace,
			Pod:       pod,
			Container: co,
		},
	)

	g, ctx := errgroup.WithContext(ctx)
	queue := &sizeQueue{
		ch:  make(chan *remotecommand.TerminalSize, 1),
		ctx: ctx,
	}

	reader, writer := io.Pipe()
	pipe, pipeWriter := io.Pipe()
	w := io.MultiWriter(pipeWriter, r)

	closeAll := func() {
		once.Do(func() {
			c.logger.DebugCtx(ctx, "closeAll")
			defer c.logger.DebugCtx(ctx, "closeAll done")
			reader.Close()
			writer.Close()
			pipe.Close()
			pipeWriter.Close()
			cancelFunc()
			r.Close()
			var fid int
			if r.File() != nil {
				fid = r.File().ID
			}
			c.eventRepo.FileAuditLogWithDuration(
				types.EventActionType_Exec,
				r.User().Name,
				fmt.Sprintf("[Exec]: 用户进入容器执行命令，container: '%s', namespace: '%s', pod： '%s'", r.Container().Container, r.Container().Namespace, r.Container().Pod),
				fid,
				r.Duration(),
			)
		})
	}

	g.Go(func() error {
		defer closeAll()
		if len(recv.Message) > 0 {
			writer.Write(recv.Message)
		}
		for {
			request, err := server.Recv()
			if err != nil {
				c.logger.DebugCtx(ctx, err)
				return err
			}

			if request.SizeQueue != nil {
				select {
				case queue.ch <- &remotecommand.TerminalSize{
					Width:  uint16(request.SizeQueue.Width),
					Height: uint16(request.SizeQueue.Height),
				}:
				default:
					c.logger.DebugCtx(ctx, "Exec: size queue full")
				}
			}

			c.logger.DebugCtxf(ctx, "Exec: %q", request.Message)
			if _, err := writer.Write(request.Message); err != nil {
				c.logger.DebugCtx(ctx, err)
			}
		}
	})

	g.Go(func() error {
		defer closeAll()
		rd := bufio.NewReader(pipe)
		for {
			b, err := rd.ReadByte()
			if err != nil {
				c.logger.DebugCtx(ctx, err)
				if err == bufio.ErrBufferFull {
					continue
				}
				return err
			}
			if err := server.Send(&container.ExecResponse{
				Message: []byte{b},
			}); err != nil {
				c.logger.ErrorCtx(ctx, err)
				return err
			}
		}
	})

	err = c.k8sRepo.Execute(ctx, &repo.Container{
		Namespace: namespace,
		Pod:       pod,
		Container: co,
	}, &repo.ExecuteInput{
		Stdin:             reader,
		Stdout:            w,
		Stderr:            w,
		TTY:               true,
		Cmd:               cmd,
		TerminalSizeQueue: queue,
	})
	if exitError, ok := err.(clientgoexec.ExitError); ok {
		server.Send(&container.ExecResponse{
			Error: &container.ExecError{
				Code:    int64(exitError.ExitStatus()),
				Message: exitError.Error(),
			},
		})
	}
	closeAll()
	c.logger.DebugCtx(ctx, "Exec: 等待彻底退出", err)
	go func() {
		err = g.Wait()
		c.logger.DebugCtx(ctx, "Exec: 彻底退出", err)
	}()
	return err
}

func (c *containerSvc) ExecOnce(request *container.ExecOnceRequest, server container.Container_ExecOnceServer) error {
	var (
		err  error
		ctx  = server.Context()
		once = sync.Once{}
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

	r := c.fileRepo.NewRecorder(MustGetUser(ctx), &repo.Container{
		Namespace: request.Namespace,
		Pod:       request.Pod,
		Container: request.Container,
	})
	r.Write([]byte(fmt.Sprintf("mars@%s:/# %s", request.Container, strings.Join(request.Command, " "))))
	r.Write([]byte("\r\n"))

	pipe, pipeWriter := io.Pipe()

	closeAll := func() {
		once.Do(func() {
			pipe.Close()
			pipeWriter.Close()
			r.Close()
			c.eventRepo.FileAuditLogWithDuration(
				types.EventActionType_Exec,
				r.User().Name,
				fmt.Sprintf("[ExecOnce]: 用户进入容器执行命令，container: '%s', namespace: '%s', pod： '%s'", r.Container().Container, r.Container().Namespace, r.Container().Pod),
				r.File().ID,
				r.Duration(),
			)
		})
	}
	w := io.MultiWriter(pipeWriter, r)
	go func() {
		defer closeAll()
		reader := bufio.NewReader(pipe)
		for {
			readByte, err := reader.ReadByte()
			if err != nil {
				if err == bufio.ErrBufferFull {
					continue
				}
				return
			}
			if err = server.Send(&container.ExecResponse{
				Message: []byte{readByte},
			}); err != nil {
				c.logger.DebugCtx(ctx, err)
			}
		}
	}()

	err = c.k8sRepo.Execute(ctx, &repo.Container{
		Namespace: request.Namespace,
		Pod:       request.Pod,
		Container: request.Container,
	}, &repo.ExecuteInput{
		Stdout: w,
		Stderr: w,
		TTY:    false,
		Cmd:    request.Command,
	})
	if exitError, ok := err.(clientgoexec.ExitError); ok {
		server.Send(&container.ExecResponse{
			Error: &container.ExecError{
				Code:    int64(exitError.ExitStatus()),
				Message: exitError.Error(),
			},
		})
	}

	closeAll()
	c.logger.DebugCtx(ctx, "ExecOnce: 彻底退出", err)
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
