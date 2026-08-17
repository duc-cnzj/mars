package biz

// container.go 定义容器终端用例层 ContainerBiz：把 Exec/ExecOnce 的双向流复用
// （io.Pipe 接线、errgroup 双 goroutine、size queue 桥接、退出码映射）与命令输出
// 审计捕获收进 biz，transport 只做鉴权、首帧解析与流透传。
//
// 流经端口接口注入：ExecStream/ExecOnceStream 由 transport 侧具体的 gRPC server
// 直接满足（grpc.BidiStreamingServer / ServerStreamingServer 本就带 Recv/Send/Context），
// biz 不依赖任何 gRPC server 具体类型，便于在 biz 层用假流单测。

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/duc-cnzj/mars/api/v6/proto/container"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
)

// ExecStream 是 Exec 用例所需的双向流端口：transport 把 container.Container_ExecServer
// 直接传给 biz（grpc.BidiStreamingServer 天然实现该接口），biz 只依赖端口。
type ExecStream interface {
	// Recv 接收下一个请求帧（流结束或错误时返回 error）。
	Recv() (*container.ExecRequest, error)
	// Send 发送一个响应帧。
	Send(*container.ExecResponse) error
	// Context 返回流的 context（取消即中断用例）。
	Context() context.Context
}

// ExecOnceStream 是 ExecOnce 用例所需的发送流端口：请求已由 transport 解析为
// ExecOnceInput，biz 只需流式回传输出。
type ExecOnceStream interface {
	// Send 发送一个响应帧。
	Send(*container.ExecResponse) error
	// Context 返回流的 context（取消即中断用例）。
	Context() context.Context
}

// ExecInput 是 Exec 用例的输入：由 transport 从首个 ExecRequest 帧解析后传入。
// FirstMessage 是首帧携带的终端输入（在 recv 循环启动前写入 stdin），
// InitialSize 是首帧携带的初始终端窗口（在会话启动前应用到 recorder）。
type ExecInput struct {
	Namespace    string
	Pod          string
	Container    string
	Command      []string
	FirstMessage []byte
	InitialSize  *container.TerminalSize
}

// ExecOnceInput 是 ExecOnce 用例的输入：由 transport 从 ExecOnceRequest 解析后传入。
type ExecOnceInput struct {
	Namespace string
	Pod       string
	Container string
	Command   []string
}

// ContainerBiz 封装容器终端用例编排：交互式会话的流复用、一次性命令的有界输出审计、退出码映射。
type ContainerBiz interface {
	// Exec 建立交互式终端会话：input 来自首个请求帧，后续帧经 stream 流式注入，
	// 输出逐字节回传；会话结束时经 recorder 落审计日志。
	Exec(ctx context.Context, stream ExecStream, user *UserInfo, input *ExecInput) error
	// ExecOnce 执行一次性命令并流式回传输出，输出被有界捕获后落入审计日志。
	ExecOnce(ctx context.Context, stream ExecOnceStream, user *UserInfo, input *ExecOnceInput) error
	// Log 返回一次性容器日志：Pending+ShowEvents 时聚合 pod 相关事件，其余阶段返回尾部日志。
	// pod 不存在或（非 ShowEvents 且 Pending）返回 NotFound。
	Log(ctx context.Context, input *LogInput) (*LogResult, error)
	// LogStream 返回流式日志源：Running pod 返回实时日志流，其余阶段返回一次性文本（与 Log 同源）。
	LogStream(ctx context.Context, input *LogInput) (*LogStreamResult, error)
	// EnsurePodRunning 校验目标 pod 处于 Running 状态；未运行返回 NotFound（reason 为 k8s 返回的原因）。
	// 统一 Exec/CopyToPod/StreamCopyToPod 三处"运行态前置检查"，保证"pod 不存在/未就绪 → 404"映射只此一处。
	EnsurePodRunning(ctx context.Context, namespace, pod string) error
	// ResolveContainer 解析目标容器：未指定时回落到 pod 默认容器。
	// 与 EnsurePodRunning 同为容器操作前置解析，统一 Exec/ExecOnce/CopyToPod/
	// StreamCopyToPod/copyFromPod 五处"空则找默认"语义，依赖收进 receiver 使签名只留业务参数。
	ResolveContainer(ctx context.Context, namespace, pod, container string) (string, error)
}

// LogInput 是容器日志用例的输入：由 transport 从 LogRequest 解析后传入。
type LogInput struct {
	Namespace  string
	Pod        string
	Container  string
	ShowEvents bool
}

// LogSource 是容器日志的读取策略：一次性文本还是实时流。
type LogSource uint8

const (
	// LogSourceContent 表示 Result 携带完整日志文本（事件聚合或尾部），transport 一次性下发。
	LogSourceContent LogSource = iota
	// LogSourceLive 表示 Result 携带实时流频道，transport 逐帧转发。
	LogSourceLive
)

// LogResult 是一次性容器日志用例的输出：Content 为事件聚合或尾部日志文本。
type LogResult struct {
	Content string
}

// LogStreamResult 是流式容器日志用例的输出：Live 时逐帧发送 Stream，否则把 Content 逐行切分发送。
type LogStreamResult struct {
	Source  LogSource
	Content string
	Stream  <-chan []byte
}

// containerBiz 是 ContainerBiz 的实现：依赖 k8s 执行、文件录音器、事件审计与计时器。
type containerBiz struct {
	logger   mlog.Logger
	k8sBiz   K8sBiz
	fileBiz  FileBiz
	eventBiz EventBiz
	timer    timer.Timer
}

// NewContainerBiz 构造容器终端用例实现。
func NewContainerBiz(logger mlog.Logger, k8sBiz K8sBiz, fileBiz FileBiz, eventBiz EventBiz, timer timer.Timer) ContainerBiz {
	return &containerBiz{logger: logger, k8sBiz: k8sBiz, fileBiz: fileBiz, eventBiz: eventBiz, timer: timer}
}

// EnsurePodRunning 实现"目标 pod 必须处于 Running"的前置校验：未运行返回
// gRPC NotFound（reason 为 k8s 返回的原因）。Exec/CopyToPod/StreamCopyToPod
// 三个执行/拷贝入口共用，保证"pod 不存在/未就绪 → 404"映射只此一处。
func (cb *containerBiz) EnsurePodRunning(ctx context.Context, namespace, pod string) error {
	running, reason := cb.k8sBiz.IsPodRunning(namespace, pod)
	if !running {
		cb.logger.DebugCtx(ctx, "EnsurePodRunning: pod not running", namespace, pod, reason)
		return errs.NotFound(reason)
	}
	return nil
}

// ResolveContainer 解析目标容器：未指定时回落到 k8s pod 的默认容器。
// 五个执行/拷贝入口共用（Exec/ExecOnce 与 transport CopyToPod/StreamCopyToPod/copyFromPod），
// 统一"空则找默认"语义。本方法不打印日志：FindDefaultContainer 失败错误由三个
// transport 调用点（CopyToPod/StreamCopyToPod/copyFromPod）用 logError 统一打印。
func (cb *containerBiz) ResolveContainer(ctx context.Context, namespace, pod, container string) (string, error) {
	if container != "" {
		return container, nil
	}
	defaultContainer, err := cb.k8sBiz.FindDefaultContainer(ctx, namespace, pod)
	if err != nil {
		return "", err
	}
	cb.logger.DebugCtx(ctx, "使用默认的容器: ", defaultContainer)
	return defaultContainer, nil
}

// Exec 实现交互式终端会话编排，逻辑自 transport containerSvc.Exec 下沉，行为不变。
// 关键约束：gRPC ServerStream.SendMsg 不保证并发安全，send loop 与退出错误帧的发送
// 必须经 sendMu 串行化。
func (cb *containerBiz) Exec(ctx context.Context, stream ExecStream, user *UserInfo, input *ExecInput) error {
	var (
		once   sync.Once
		sendMu sync.Mutex
	)
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	// 串行化所有向客户端的发送：send loop goroutine 与主流程的退出错误帧可能并发。
	sendMsg := func(resp *container.ExecResponse) error {
		sendMu.Lock()
		defer sendMu.Unlock()
		return stream.Send(resp)
	}

	if err := cb.EnsurePodRunning(ctx, input.Namespace, input.Pod); err != nil {
		return err
	}

	co, err := cb.ResolveContainer(ctx, input.Namespace, input.Pod, input.Container)
	if err != nil {
		return err
	}

	r := cb.fileBiz.NewRecorder(user, &Container{Namespace: input.Namespace, Pod: input.Pod, Container: co})
	if input.InitialSize != nil {
		r.Resize(uint16(input.InitialSize.Width), uint16(input.InitialSize.Height))
	}

	g, ctx := errgroup.WithContext(ctx)
	sizeCh := make(chan *TerminalSize, 1)
	queue := newExecSizeQueue(ctx, sizeCh, r)

	reader, writer := io.Pipe()
	pipe, pipeWriter := io.Pipe()
	w := io.MultiWriter(pipeWriter, r)

	closeAll := func() {
		once.Do(func() {
			cb.logger.DebugCtx(ctx, "closeAll")
			defer cb.logger.DebugCtx(ctx, "closeAll done")
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
			cb.eventBiz.FileAuditLogWithDuration(
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
		defer cb.logger.HandlePanic("biz.ContainerBiz.Exec: recv loop")
		if len(input.FirstMessage) > 0 {
			if _, err := writer.Write(input.FirstMessage); err != nil {
				cb.logger.DebugCtx(ctx, err)
			}
		}
		for {
			request, err := stream.Recv()
			if err != nil {
				cb.logger.DebugCtx(ctx, err)
				return err
			}

			if request.SizeQueue != nil {
				cb.logger.DebugCtxf(ctx, "Exec: resize w: %d, h: %d", request.SizeQueue.Width, request.SizeQueue.Height)
				select {
				case sizeCh <- &TerminalSize{
					Width:  uint16(request.SizeQueue.Width),
					Height: uint16(request.SizeQueue.Height),
				}:
				default:
					cb.logger.DebugCtx(ctx, "Exec: size queue full")
				}
			}

			cb.logger.DebugCtxf(ctx, "Exec: %q", request.Message)
			if _, err := writer.Write(request.Message); err != nil {
				cb.logger.DebugCtx(ctx, err)
			}
		}
	})

	g.Go(func() error {
		defer closeAll()
		defer cb.logger.HandlePanic("biz.ContainerBiz.Exec: send loop")
		rd := bufio.NewReader(pipe)
		for {
			b, err := rd.ReadByte()
			if err != nil {
				cb.logger.DebugCtx(ctx, err)
				return err
			}
			if err := sendMsg(&container.ExecResponse{
				Message: []byte{b},
			}); err != nil {
				cb.logger.ErrorCtx(ctx, err)
				return err
			}
		}
	})

	err = cb.k8sBiz.Execute(ctx, &Container{
		Namespace: input.Namespace,
		Pod:       input.Pod,
		Container: co,
	}, &ExecuteInput{
		Stdin:             reader,
		Stdout:            w,
		Stderr:            w,
		TTY:               true,
		Cmd:               input.Command,
		TerminalSizeQueue: queue,
	})
	var exitError *ExecExitError
	if errors.As(err, &exitError) {
		if sendErr := sendMsg(&container.ExecResponse{
			Error: &container.ExecError{
				Code:    int64(exitError.Code),
				Message: exitError.Message,
			},
		}); sendErr != nil {
			cb.logger.DebugCtx(ctx, "Exec: send exit error failed", sendErr)
		}
	}
	closeAll()
	cb.logger.DebugCtx(ctx, "Exec: 等待彻底退出", err)
	go func() {
		defer cb.logger.HandlePanic("biz.ContainerBiz.Exec: errgroup wait")
		if egErr := g.Wait(); egErr != nil && !errors.Is(egErr, context.Canceled) {
			cb.logger.DebugCtx(ctx, "Exec: errgroup exit error", egErr)
		}
	}()
	return err
}

// ExecOnce 实现一次性命令编排，逻辑自 transport containerSvc.ExecOnce 下沉，行为不变。
// 命令输出被 limitedBuffer 有界捕获（防大输出打爆内存）后随审计日志落库。
func (cb *containerBiz) ExecOnce(ctx context.Context, stream ExecOnceStream, user *UserInfo, input *ExecOnceInput) error {
	var (
		err    error
		once   = sync.Once{}
		sendMu sync.Mutex
	)
	// 与 Exec 一致：gRPC SendMsg 不保证并发安全，串行化 send loop 与退出错误帧的发送。
	sendMsg := func(resp *container.ExecResponse) error {
		sendMu.Lock()
		defer sendMu.Unlock()
		return stream.Send(resp)
	}

	if running, reason := cb.k8sBiz.IsPodRunning(input.Namespace, input.Pod); !running {
		cb.logger.DebugCtx(ctx, "ExecOnce: pod not running", input.Namespace, input.Pod, reason)
		return errs.NotFound(reason)
	}

	co, err := cb.ResolveContainer(ctx, input.Namespace, input.Pod, input.Container)
	if err != nil {
		return err
	}

	bf := newLimitedBuffer(maxExecOnceLogSize)
	pipe, pipeWriter := io.Pipe()

	closeAll := func() {
		once.Do(func() {
			pipe.Close()
			pipeWriter.Close()
		})
	}
	w := io.MultiWriter(pipeWriter, bf)
	go func() {
		defer closeAll()
		defer cb.logger.HandlePanic("biz.ContainerBiz.ExecOnce: send loop")
		reader := bufio.NewReader(pipe)
		for {
			readByte, err := reader.ReadByte()
			if err != nil {
				return
			}
			if err = sendMsg(&container.ExecResponse{
				Message: []byte{readByte},
			}); err != nil {
				cb.logger.DebugCtx(ctx, "ExecOnce: send response failed", err)
				return
			}
		}
	}()
	startTime := cb.timer.Now()

	err = cb.k8sBiz.Execute(ctx, &Container{
		Namespace: input.Namespace,
		Pod:       input.Pod,
		Container: co,
	}, &ExecuteInput{
		Stdout: w,
		Stderr: w,
		TTY:    false,
		Cmd:    input.Command,
	})
	var exitError *ExecExitError
	if errors.As(err, &exitError) {
		if sendErr := sendMsg(&container.ExecResponse{
			Error: &container.ExecError{
				Code:    int64(exitError.Code),
				Message: exitError.Message,
			},
		}); sendErr != nil {
			cb.logger.DebugCtx(ctx, "ExecOnce: send exit error failed", sendErr)
		}
	}

	closeAll()
	cb.eventBiz.AuditLogWithChange(
		types.EventActionType_Exec,
		user.Name,
		fmt.Sprintf("[ExecOnce]: 用户进入容器执行命令，container: '%s', namespace: '%s', pod： '%s'", co, input.Namespace, input.Pod),
		nil,
		AnyYamlPrettier{
			"namespace": input.Namespace,
			"pod":       input.Pod,
			"container": co,
			"command":   input.Command,
			"result":    bf.String(),
			"error":     toErrStr(err),
			"duration":  cb.timer.Since(startTime).String(),
		},
	)
	cb.logger.DebugCtx(ctx, "ExecOnce: 彻底退出", err)
	return err
}

// resolveLog 解析 pod 并应用可读日志守卫：GetPod 失败原样上抛（错误由最上层 services 统一打印），
// pod 不存在、或非 ShowEvents 且处于 Pending 时返回 NotFound。
// ContainerLog/StreamContainerLog 共用，防止两处 GetPod + 守卫逻辑漂移。
func (cb *containerBiz) resolveLog(ctx context.Context, input *LogInput) (*v1.Pod, error) {
	podInfo, err := cb.k8sBiz.GetPod(input.Namespace, input.Pod)
	if err != nil {
		return nil, err
	}
	if podInfo == nil || (!input.ShowEvents && podInfo.Status.Phase == v1.PodPending) {
		return nil, errs.NotFound(fmt.Sprintf("未找到日志: %s/%s", input.Namespace, input.Pod))
	}
	return podInfo, nil
}

// Log 实现一次性容器日志用例，逻辑自 transport containerSvc.ContainerLog 下沉，行为不变。
// 守卫（NotFound）见 resolveLog；Pending+ShowEvents 聚合本 pod 的 Pod 事件；
// 其余阶段（Running/Succeeded/Failed）统一只取尾部日志，防终止 pod 全量读打爆内存。
func (cb *containerBiz) Log(ctx context.Context, input *LogInput) (*LogResult, error) {
	podInfo, err := cb.resolveLog(ctx, input)
	if err != nil {
		return nil, err
	}

	if podInfo.Status.Phase == v1.PodPending {
		// 进入此分支时 ShowEvents 恒为 true（非 ShowEvents 且 Pending 已被 resolveLog 守卫拦截）。
		var logs []string
		ret, err := cb.k8sBiz.ListEvents(input.Namespace)
		if err != nil {
			cb.logger.DebugCtx(ctx, "ListEvents failed, events section will be empty", err)
		}
		sort.Sort(sortEvents(ret))
		for _, event := range ret {
			if event.Regarding.Kind == "Pod" && event.Regarding.Name == input.Pod {
				logs = append(logs, event.Note)
			}
		}
		return &LogResult{Content: strings.Join(logs, "\n")}, nil
	}

	logs, err := cb.k8sBiz.GetPodLogs(ctx, input.Namespace, input.Pod, &v1.PodLogOptions{
		Container: input.Container,
		TailLines: &tailLines,
	})
	if err != nil {
		return nil, err
	}
	return &LogResult{Content: logs}, nil
}

// LogStream 实现流式容器日志用例，逻辑自 transport containerSvc.StreamContainerLog 下沉，行为不变。
// Running pod 返回实时流；Succeeded/Failed/Pending 与一次性 Log 同源（尾部/事件文本），
// transport 拿到 Content 后逐行切分下发。
func (cb *containerBiz) LogStream(ctx context.Context, input *LogInput) (*LogStreamResult, error) {
	podInfo, err := cb.resolveLog(ctx, input)
	if err != nil {
		return nil, err
	}

	if podInfo.Status.Phase == v1.PodRunning {
		ch, err := cb.k8sBiz.LogStream(ctx, input.Namespace, input.Pod, input.Container)
		if err != nil {
			return nil, err
		}
		return &LogStreamResult{Source: LogSourceLive, Stream: ch}, nil
	}

	res, err := cb.Log(ctx, input)
	if err != nil {
		return nil, err
	}
	return &LogStreamResult{Source: LogSourceContent, Content: res.Content}, nil
}

// tailLines 是日志尾部行数，需要指针传给 PodLogOptions.TailLines，故声明为 var 而非 const。
var tailLines int64 = 1000

// sortEvents 按 k8s ResourceVersion 数值升序排序事件。
type sortEvents []*eventsv1.Event

// Len 返回事件个数。
func (s sortEvents) Len() int {
	return len(s)
}

// Less 判定第 i 个事件应排在第 j 个之前：按 ResourceVersion 数值升序。
func (s sortEvents) Less(i, j int) bool {
	// ResourceVersion 是数字字符串，直接比较字符串会得到错误顺序（如 "9" > "10"），
	// 必须按数值比较；解析失败时回退到字符串比较。
	// 用 ParseUint 而不是 ParseInt：k8s 的 ResourceVersion 语义上是 uint64 单调计数，
	// int64 上限（2^63-1，约 9.2e18）以下没问题，超过后 ParseInt 会解析失败并回退到
	// 字符串比较，导致位数不同的两个大版本号被排错顺序。
	a, errA := strconv.ParseUint(s[i].ResourceVersion, 10, 64)
	b, errB := strconv.ParseUint(s[j].ResourceVersion, 10, 64)
	if errA != nil || errB != nil {
		return s[i].ResourceVersion < s[j].ResourceVersion
	}
	return a < b
}

// Swap 交换两个事件的位置。
func (s sortEvents) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// execSizeQueue 把 gRPC 终端尺寸帧桥接到领域 TerminalSizeQueue 端口，并同步到 recorder。
type execSizeQueue struct {
	ch  chan *TerminalSize
	ctx context.Context
	r   Recorder
}

// newExecSizeQueue 构造把尺寸帧桥接到领域尺寸队列端口的实现。
func newExecSizeQueue(ctx context.Context, ch chan *TerminalSize, r Recorder) TerminalSizeQueue {
	return &execSizeQueue{ch: ch, ctx: ctx, r: r}
}

// Next 返回下一个终端尺寸：channel 关闭或 ctx 取消时返回 nil 结束队列。
func (queue *execSizeQueue) Next() *TerminalSize {
	select {
	case size, ok := <-queue.ch:
		if !ok {
			return nil
		}
		if size.Width > 0 && size.Height > 0 {
			queue.r.Resize(size.Width, size.Height)
		}
		return size
	case <-queue.ctx.Done():
		return nil
	}
}

// maxExecOnceLogSize 限制 ExecOnce 审计日志记录的命令输出大小，防止大输出命令打爆内存。
const maxExecOnceLogSize = 1 << 20 // 1MiB

// limitedBuffer 只保留最近 max 字节的写入内容，用于有界收集命令输出。
type limitedBuffer struct {
	buf []byte
	max int
}

// newLimitedBuffer 构造保留最近 max 字节的有界缓冲区。
func newLimitedBuffer(max int) *limitedBuffer {
	return &limitedBuffer{max: max}
}

// Write 追加内容并裁剪到最近 max 字节（复制到新数组，避免底层数组持续膨胀）。
func (l *limitedBuffer) Write(p []byte) (int, error) {
	l.buf = append(l.buf, p...)
	if len(l.buf) > l.max {
		l.buf = append([]byte(nil), l.buf[len(l.buf)-l.max:]...)
	}
	return len(p), nil
}

// String 返回当前缓冲内容。
func (l *limitedBuffer) String() string {
	return string(l.buf)
}

// toErrStr 把错误转为可安全入审计的非空字符串：nil 错误返回空串。
func toErrStr(e error) string {
	if e == nil {
		return ""
	}
	return e.Error()
}
