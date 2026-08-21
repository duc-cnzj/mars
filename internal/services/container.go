package services

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/duc-cnzj/mars/api/v6/proto/container"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/dustin/go-humanize"
)

var _ container.ContainerServer = (*containerSvc)(nil)

// containerSvc 是 container.ContainerServer 的 gRPC 实现：提供 Pod 状态查询、日志流式读取、
// 文件拷入拷出与 exec/exec-once 终端交互，经 access 校验访问权限，由 NewContainerSvc 构造。
type containerSvc struct {
	k8sBiz       biz.K8sBiz
	fileBiz      biz.FileBiz
	logger       mlog.Logger
	containerBiz biz.ContainerBiz
	accessBiz    biz.AccessBiz

	eventBiz biz.EventBiz

	container.UnimplementedContainerServer
}

// ContainerSvcDeps 收口 NewContainerSvc 的构造依赖，由 wire 按字段注入。
type ContainerSvcDeps struct {
	ContainerBiz biz.ContainerBiz
	EventBiz     biz.EventBiz
	K8sBiz       biz.K8sBiz
	FileBiz      biz.FileBiz
	AccessBiz    biz.AccessBiz
	Logger       mlog.Logger
}

// NewContainerSvc 收口容器服务的构造依赖，由 wire 按字段注入。
func NewContainerSvc(deps ContainerSvcDeps) container.ContainerServer {
	logger := deps.Logger.WithModule("services/container")
	return &containerSvc{
		containerBiz: deps.ContainerBiz,
		eventBiz:     deps.EventBiz,
		k8sBiz:       deps.K8sBiz,
		fileBiz:      deps.FileBiz,
		logger:       logger,
		accessBiz:    deps.AccessBiz,
	}
}

// IsPodRunning 查询指定 pod 是否处于 Running 状态（含原因），响应前做命名空间级访问控制。
func (c *containerSvc) IsPodRunning(ctx context.Context, request *container.IsPodRunningRequest) (*container.IsPodRunningResponse, error) {
	if _, err := c.accessBiz.RequireNamespaceAccessByName(ctx, request.GetNamespace()); err != nil {
		return nil, logError(ctx, c.logger, err)
	}
	running, reason := c.k8sBiz.IsPodRunning(request.GetNamespace(), request.GetPod())

	return &container.IsPodRunningResponse{Running: running, Reason: reason}, nil
}

// IsPodExists 查询指定 pod 是否存在：Get 失败即视为不存在（响应前做命名空间级访问控制）。
func (c *containerSvc) IsPodExists(ctx context.Context, request *container.IsPodExistsRequest) (*container.IsPodExistsResponse, error) {
	if _, err := c.accessBiz.RequireNamespaceAccessByName(ctx, request.GetNamespace()); err != nil {
		return nil, logError(ctx, c.logger, err)
	}
	_, err := c.k8sBiz.GetPod(request.GetNamespace(), request.GetPod())
	if err != nil {
		// 探活接口：pod 不存在是预期的否定结果（Exists:false），非真实错误，
		// 不记 Error 日志；仅 k8s API 故障等非 NotFound 错误才留痕。
		if !errs.IsNotFound(err) {
			c.logger.ErrorCtx(ctx, err)
		}
		return &container.IsPodExistsResponse{Exists: false}, nil
	}

	return &container.IsPodExistsResponse{Exists: true}, nil
}

// ContainerLog 返回一次性容器日志：守卫、Pending 事件聚合、尾部日志策略均下沉
// containerBiz.Log，transport 只做鉴权与 UTF-8 序列化卫生。
func (c *containerSvc) ContainerLog(ctx context.Context, request *container.LogRequest) (*container.LogResponse, error) {
	if _, err := c.accessBiz.RequireNamespaceAccessByName(ctx, request.Namespace); err != nil {
		return nil, logError(ctx, c.logger, err)
	}
	res, err := c.containerBiz.Log(ctx, &biz.LogInput{
		Namespace:  request.Namespace,
		Pod:        request.Pod,
		Container:  request.Container,
		ShowEvents: request.ShowEvents,
	})
	if err != nil {
		return nil, logError(ctx, c.logger, err)
	}

	return &container.LogResponse{
		Namespace:     request.Namespace,
		PodName:       request.Pod,
		ContainerName: request.Container,
		Log:           toValidUTF8String([]byte(res.Content)),
	}, nil
}

// CopyToPod 把已上传的文件写入指定 pod 容器路径：先校验 pod 运行态，落上传审计日志。
func (c *containerSvc) CopyToPod(ctx context.Context, request *container.CopyToPodRequest) (*container.CopyToPodResponse, error) {
	if _, err := c.accessBiz.RequireNamespaceAccessByName(ctx, request.Namespace); err != nil {
		return nil, logError(ctx, c.logger, err)
	}
	// 运行态前置校验与"空则找默认容器"统一走 biz：与 StreamCopyToPod/Exec 的
	// "pod 未运行 → 404"和 ResolveContainer 语义对齐。此前 CopyToPod 直接把
	// 未解析的 Container 透传给 data 层（CopyFileToPod 不兜底空 container），
	// 用户不传 container 时行为与流式入口不一致。
	if err := c.containerBiz.EnsurePodRunning(ctx, request.Namespace, request.Pod); err != nil {
		return nil, err
	}
	resolved, err := c.containerBiz.ResolveContainer(ctx, request.Namespace, request.Pod, request.Container)
	if err != nil {
		return nil, logError(ctx, c.logger, err)
	}
	request.Container = resolved

	file, err := c.k8sBiz.CopyFileToPod(ctx, &biz.CopyFileToPodInput{
		FileId:    request.FileId,
		Namespace: request.Namespace,
		Pod:       request.Pod,
		Container: request.Container,
	})
	if err != nil {
		return nil, logError(ctx, c.logger, err)
	}

	c.eventBiz.FileAuditLog(
		types.EventActionType_Upload,
		biz.MustGetUser(ctx).Name,
		biz.MustGetUser(ctx).Email,
		copyToPodAuditMsg("", request.Namespace, request.Pod, request.Container, file),
		file.ID,
	)

	return &container.CopyToPodResponse{
		PodFilePath: file.ContainerPath,
		FileName:    file.Path,
	}, nil
}

// StreamCopyToPod 流式上传文件到 pod：接收分帧数据落盘后拷贝进容器，落上传审计日志。
func (c *containerSvc) StreamCopyToPod(server container.Container_StreamCopyToPodServer) error {
	var (
		ctx  = server.Context()
		user = biz.MustGetUser(server.Context())
	)
	recv, err := server.Recv()
	if err != nil {
		return err
	}
	if _, err := c.accessBiz.RequireNamespaceAccessByName(ctx, recv.Namespace); err != nil {
		return logError(ctx, c.logger, err)
	}
	c.logger.DebugCtx(ctx, "StreamUploadFile", recv.Namespace, recv.Pod, recv.Container, recv.FileName)
	// 运行态前置校验统一走 biz.EnsurePodRunning，与 CopyToPod/Exec 共用 404 映射。
	if err := c.containerBiz.EnsurePodRunning(ctx, recv.Namespace, recv.Pod); err != nil {
		return err
	}
	// 如果没传入 container，使用默认的
	recv.Container, err = c.containerBiz.ResolveContainer(ctx, recv.Namespace, recv.Pod, recv.Container)
	if err != nil {
		return logError(ctx, c.logger, err)
	}
	c.logger.DebugCtxf(ctx, "StreamUploadFile %s/%s/%s", recv.Namespace, recv.Pod, recv.Container)
	ch := make(chan []byte, 100)
	ch <- recv.GetData()
	go func() {
		defer close(ch)
		defer c.logger.HandlePanic("StreamUploadFile: recv loop")

		for {
			recv, err := server.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					c.logger.DebugCtx(ctx, "StreamUploadFile: EOF")
					return
				}
				c.logger.ErrorCtx(ctx, "StreamUploadFile: error receiving data", err)
				return
			}
			select {
			case ch <- recv.GetData():
			case <-ctx.Done():
				return
			}
		}
	}()

	file, err := c.fileBiz.StreamUploadFile(ctx, &biz.StreamUploadFileRequest{
		Namespace: recv.Namespace,
		Pod:       recv.Pod,
		Container: recv.Container,
		Username:  user.Name,
		FileName:  recv.FileName,
		FileData:  ch,
	})
	if err != nil {
		return logError(ctx, c.logger, err)
	}

	res, err := c.k8sBiz.CopyFileToPod(ctx, &biz.CopyFileToPodInput{
		FileId:    int64(file.ID),
		Namespace: file.Namespace,
		Pod:       file.Pod,
		Container: file.Container,
	})
	if err != nil {
		return logError(ctx, c.logger, err)
	}

	c.eventBiz.FileAuditLog(
		types.EventActionType_Upload,
		biz.MustGetUser(ctx).Name,
		biz.MustGetUser(ctx).Email,
		copyToPodAuditMsg("[StreamUploadFile]: ", file.Namespace, file.Pod, file.Container, file),
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

// StreamContainerLog 流式返回容器日志：实时流/一次性文本的策略选择下沉 containerBiz.LogStream，
// transport 只做鉴权与帧下发（Running 逐帧转发，其余阶段逐行切分）。
func (c *containerSvc) StreamContainerLog(request *container.LogRequest, server container.Container_StreamContainerLogServer) error {
	if _, err := c.accessBiz.RequireNamespaceAccessByName(server.Context(), request.Namespace); err != nil {
		return logError(server.Context(), c.logger, err)
	}
	c.logger.DebugCtxf(server.Context(), "StreamContainerLog: %v", request)
	res, err := c.containerBiz.LogStream(server.Context(), &biz.LogInput{
		Namespace:  request.Namespace,
		Pod:        request.Pod,
		Container:  request.Container,
		ShowEvents: request.ShowEvents,
	})
	if err != nil {
		return logError(server.Context(), c.logger, err)
	}

	if res.Source == biz.LogSourceLive {
		for msg := range res.Stream {
			if err = server.Send(&container.LogResponse{
				Namespace:     request.Namespace,
				PodName:       request.Pod,
				ContainerName: request.Container,
				Log:           toValidUTF8String(msg),
			}); err != nil {
				return logError(server.Context(), c.logger, err)
			}
		}
		return nil
	}

	// 一次性文本同样先做 UTF-8 卫生再逐行下发，与 ContainerLog 对齐。
	return scannerText(toValidUTF8String([]byte(res.Content)), func(s string) {
		if err := server.Send(&container.LogResponse{
			Namespace:     request.Namespace,
			PodName:       request.Pod,
			ContainerName: request.Container,
			Log:           s,
		}); err != nil {
			// 客户端断连时 Send 失败是常态，记录日志并继续消费剩余行。
			c.logger.DebugCtx(server.Context(), err)
		}
	})
}

// Exec 建立双向流交互式终端会话（首帧携带命令/首屏输入/窗口尺寸），透传给 containerBiz.Exec。
func (c *containerSvc) Exec(server container.Container_ExecServer) error {
	recv, err := server.Recv()
	if err != nil {
		return err
	}
	if _, err := c.accessBiz.RequireNamespaceAccessByName(server.Context(), recv.Namespace); err != nil {
		return logError(server.Context(), c.logger, err)
	}
	// 首帧携带的 Message/SizeQueue 一并传给 biz：Message 作为终端首屏输入，
	// SizeQueue 作为会话初始终端窗口（在 recorder 上预先应用）。
	return c.containerBiz.Exec(server.Context(), server, biz.MustGetUser(server.Context()), &biz.ExecInput{
		Namespace:    recv.Namespace,
		Pod:          recv.Pod,
		Container:    recv.Container,
		Command:      recv.Command,
		FirstMessage: recv.Message,
		InitialSize:  recv.SizeQueue,
	})
}

// ExecOnce 执行单次命令并把结果流回客户端，透传给 containerBiz.ExecOnce。
func (c *containerSvc) ExecOnce(request *container.ExecOnceRequest, server container.Container_ExecOnceServer) error {
	if _, err := c.accessBiz.RequireNamespaceAccessByName(server.Context(), request.Namespace); err != nil {
		return logError(server.Context(), c.logger, err)
	}
	return c.containerBiz.ExecOnce(server.Context(), server, biz.MustGetUser(server.Context()), &biz.ExecOnceInput{
		Namespace: request.Namespace,
		Pod:       request.Pod,
		Container: request.Container,
		Command:   request.Command,
	})
}

// ForceDeletePod 强制删除指定 pod：先做命名空间级访问控制，再以指定宽限期执行
// 强制删除（gracePeriodSeconds=0 即不等优雅终止）。成功与失败均落审计日志
// （EventActionType_ForceDeletePod），错误统一经 logError 打印。
func (c *containerSvc) ForceDeletePod(ctx context.Context, request *container.ForceDeletePodRequest) (*container.ForceDeletePodResponse, error) {
	if _, err := c.accessBiz.RequireNamespaceAccessByName(ctx, request.GetNamespace()); err != nil {
		return nil, logError(ctx, c.logger, err)
	}
	user := biz.MustGetUser(ctx)
	msg := fmt.Sprintf("强制删除 pod: %s/%s", request.GetNamespace(), request.GetPod())
	if err := c.k8sBiz.ForceDeletePod(ctx, request.GetNamespace(), request.GetPod(), request.GetGracePeriodSeconds()); err != nil {
		c.eventBiz.AuditLog(types.EventActionType_ForceDeletePod, user.Name, user.Email, msg+", 失败: "+err.Error())
		return nil, logError(ctx, c.logger, err)
	}
	c.eventBiz.AuditLog(types.EventActionType_ForceDeletePod, user.Name, user.Email, msg)
	return &container.ForceDeletePodResponse{
		Deleted:   true,
		Namespace: request.GetNamespace(),
		Pod:       request.GetPod(),
		Message:   fmt.Sprintf("pod %s/%s 已强制删除", request.GetNamespace(), request.GetPod()),
	}, nil
}

// scannerText 按行切分文本并回调 fn；放大 Scanner 缓冲以容纳超长日志行。
func scannerText(text string, fn func(s string)) error {
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Split(bufio.ScanLines)
	// 默认 max token 仅 64KB，容器日志单行常见超长（栈回溯、超大 JSON、base64），
	// 不调大会让整段日志因 bufio.ErrTooLong 而流式失败。
	scanner.Buffer(make([]byte, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		fn(scanner.Text())
	}
	return scanner.Err()
}

// toValidUTF8String 过滤无效 UTF-8 字符，防止 gRPC 序列化失败
func toValidUTF8String(b []byte) string {
	if utf8.Valid(b) {
		return string(b)
	}
	return strings.ToValidUTF8(string(b), "")
}

// copyToPodAuditMsg 构造"上传文件到 pod"审计消息：CopyToPod/StreamCopyToPod 两入口
// 共用同一消息结构，仅 prefix 区分入口（流式带 [StreamUploadFile] 标识），避免文案
// 与 humanize 单位在两处各写一遍导致漂移。
func copyToPodAuditMsg(prefix, namespace, pod, container string, fil *biz.File) string {
	return fmt.Sprintf("%s上传文件到 pod: %s/%s/%s, 容器路径: '%s', 大小: %s。",
		prefix, namespace, pod, container, fil.ContainerPath, humanize.Bytes(fil.Size))
}
