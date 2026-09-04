package biz

// container_test.go 覆盖容器终端用例层 ContainerBiz：Exec/ExecOnce 的流复用编排、
// 尺寸桥接、有界输出缓冲、退出码映射与审计落库。
//
// package biz 的测试不能 import data（data→biz 成环），因此下游全部用嵌入接口 +
// 覆写方法集的 stub 替身；流用假双向/发送流模拟 gRPC server，借助 io.Pipe 的同步
// 语义确定性地触发写错误分支。

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/container"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ---- 假流 ----

// fakeExecStream 是 Exec 用例的假双向流：按 seq 顺序吐帧，耗尽后返回错误结束 recv
// 循环；Send 记录响应，sendErr 非 nil 时触发发送失败（send loop 与退出错误帧共用）。
type fakeExecStream struct {
	seq     []*container.ExecRequest
	sendErr error
	// recvGate 非 nil 时每次 Recv 前阻塞直到 channel 关闭：供 SendError/RecvWriteError
	// 用例拖住 recv loop，防止其先 closeAll 关 pipe 而屏蔽 send loop 失败/recv loop 写失败分支。
	recvGate chan struct{}

	recvCount int
	mu        sync.Mutex
	sent      []*container.ExecResponse
}

func (f *fakeExecStream) Recv() (*container.ExecRequest, error) {
	if f.recvGate != nil {
		<-f.recvGate
	}
	if f.recvCount < len(f.seq) {
		req := f.seq[f.recvCount]
		f.recvCount++
		return req, nil
	}
	return nil, errors.New("recv done")
}

func (f *fakeExecStream) Send(resp *container.ExecResponse) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, resp)
	return f.sendErr
}

func (f *fakeExecStream) Context() context.Context { return context.Background() }

// lastExitCode 返回最后一次收到的退出错误帧的退出码（未收到返回 -1）。
func (f *fakeExecStream) lastExitCode() int64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, r := range f.sent {
		if r.Error != nil {
			return r.Error.Code
		}
	}
	return -1
}

// fakeExecOnceStream 是 ExecOnce 用例的假发送流：Send 记录响应。
// sendErr 使任意帧发送失败（覆盖 send loop 的逐字节转发失败）；
// errFrameErr 仅使错误帧（截断/退出码）发送失败，用于覆盖主流程发错误帧的失败分支
// 而不打断 send loop 对输出流的消费（否则管道提前关闭，截断永远不会触发）。
type fakeExecOnceStream struct {
	sendErr     error
	errFrameErr error
	mu          sync.Mutex
	sent        []*container.ExecResponse
}

func (f *fakeExecOnceStream) Send(resp *container.ExecResponse) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, resp)
	if resp.Error != nil {
		return f.errFrameErr
	}
	return f.sendErr
}

func (f *fakeExecOnceStream) Context() context.Context { return context.Background() }

func (f *fakeExecOnceStream) lastExitCode() int64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, r := range f.sent {
		if r.Error != nil {
			return r.Error.Code
		}
	}
	return -1
}

// findError 返回第一条匹配 code 的错误帧，未找到返回 nil（mutex 保护，防与 send loop 并发读写）。
func (f *fakeExecOnceStream) findError(code int64) *container.ExecError {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, r := range f.sent {
		if r.Error != nil && r.Error.Code == code {
			return r.Error
		}
	}
	return nil
}

// ---- stub 替身 ----

// fakeK8sBizForContainer 覆写 ContainerBiz 用到的三个 k8s 方法，其余由接口兜底。
type fakeK8sBizForContainer struct {
	K8sBiz
	isPodRunning func(namespace, podName string) (bool, string)
	findDefault  func(ctx context.Context, namespace, pod string) (string, error)
	execFn       func(ctx context.Context, c *Container, input *ExecuteInput) error
}

func (f *fakeK8sBizForContainer) IsPodRunning(namespace, podName string) (bool, string) {
	return f.isPodRunning(namespace, podName)
}
func (f *fakeK8sBizForContainer) FindDefaultContainer(ctx context.Context, namespace, pod string) (string, error) {
	return f.findDefault(ctx, namespace, pod)
}
func (f *fakeK8sBizForContainer) Execute(ctx context.Context, c *Container, input *ExecuteInput) error {
	return f.execFn(ctx, c, input)
}

// fakeFileBizForContainer 只覆写 NewRecorder，把测试注入的假 recorder 返回给用例。
type fakeFileBizForContainer struct {
	FileBiz
	recorder func(user *UserInfo, container *Container) Recorder
}

func (f *fakeFileBizForContainer) NewRecorder(user *UserInfo, container *Container) Recorder {
	return f.recorder(user, container)
}

// fakeEventBizForContainer 只覆写 Exec/ExecOnce 用到的两个审计方法。
type fakeEventBizForContainer struct {
	EventBiz
	fileAudit func(action types.EventActionType, username, operatorEmail, msg string, fileID int, duration time.Duration)
	audit     func(action types.EventActionType, username, operatorEmail, msg string, oldS, newS YamlPrettier)
}

func (f *fakeEventBizForContainer) FileAuditLogWithDuration(action types.EventActionType, username, operatorEmail, msg string, fileID int, duration time.Duration) {
	f.fileAudit(action, username, operatorEmail, msg, fileID, duration)
}
func (f *fakeEventBizForContainer) AuditLogWithChange(action types.EventActionType, username, operatorEmail, msg string, oldS, newS YamlPrettier) {
	f.audit(action, username, operatorEmail, msg, oldS, newS)
}

// fakeRecorderForContainer 实现 Recorder 全部方法，供审计路径读 File/User/Duration。
type fakeRecorderForContainer struct {
	user      *UserInfo
	file      *File
	container *Container
	dur       time.Duration
	mu        sync.Mutex
	w, h      uint16
}

func (r *fakeRecorderForContainer) Resize(width, height uint16) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.w, r.h = width, height
}
func (r *fakeRecorderForContainer) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(p), nil
}
func (r *fakeRecorderForContainer) Close() error            { return nil }
func (r *fakeRecorderForContainer) SetShell(string)         {}
func (r *fakeRecorderForContainer) GetShell() string        { return "" }
func (r *fakeRecorderForContainer) Duration() time.Duration { return r.dur }
func (r *fakeRecorderForContainer) size() (uint16, uint16) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.w, r.h
}

// File/User/Container 被 Exec 的 closeAll 审计路径解引用，未注入的字段用零值兜底避免 nil 解引用。
func (r *fakeRecorderForContainer) File() *File {
	if r.file == nil {
		return &File{}
	}
	return r.file
}
func (r *fakeRecorderForContainer) User() *UserInfo {
	if r.user == nil {
		return &UserInfo{}
	}
	return r.user
}
func (r *fakeRecorderForContainer) Container() *Container {
	if r.container == nil {
		return &Container{}
	}
	return r.container
}

// newTestContainerBiz 组装 ContainerBiz 测试实例，依赖由测试注入的 fake 替身。
func newTestContainerBiz(k K8sBiz, f FileBiz, e EventBiz) *containerBiz {
	return NewContainerBiz(mlog.NewForConfig(nil), k, f, e, timer.NewReal()).(*containerBiz)
}

// ---- ResolveContainer ----

func TestResolveContainer_GivenContainer(t *testing.T) {
	// 显式指定容器时不触达 k8s（findDefault 置 panic 证明未调用）。
	k := &fakeK8sBizForContainer{findDefault: func(ctx context.Context, ns, pod string) (string, error) {
		panic("FindDefaultContainer should not be called")
	}}
	cb := newTestContainerBiz(k, &fakeFileBizForContainer{}, &fakeEventBizForContainer{})
	got, err := cb.ResolveContainer(context.TODO(), "a", "b", "c")
	assert.NoError(t, err)
	assert.Equal(t, "c", got)
}

func TestResolveContainer_DefaultContainer(t *testing.T) {
	k := &fakeK8sBizForContainer{findDefault: func(ctx context.Context, ns, pod string) (string, error) {
		return "default-c", nil
	}}
	cb := newTestContainerBiz(k, &fakeFileBizForContainer{}, &fakeEventBizForContainer{})
	got, err := cb.ResolveContainer(context.TODO(), "a", "b", "")
	assert.NoError(t, err)
	assert.Equal(t, "default-c", got)
}

func TestResolveContainer_DefaultContainerError(t *testing.T) {
	k := &fakeK8sBizForContainer{findDefault: func(ctx context.Context, ns, pod string) (string, error) {
		return "", errors.New("no default")
	}}
	cb := newTestContainerBiz(k, &fakeFileBizForContainer{}, &fakeEventBizForContainer{})
	_, err := cb.ResolveContainer(context.TODO(), "a", "b", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no default")
}

// ---- EnsurePodRunning ----

func TestContainerBiz_EnsurePodRunning_Running(t *testing.T) {
	k := &fakeK8sBizForContainer{isPodRunning: func(ns, pod string) (bool, string) {
		return true, ""
	}}
	cb := newTestContainerBiz(k, &fakeFileBizForContainer{}, &fakeEventBizForContainer{})
	assert.NoError(t, cb.EnsurePodRunning(context.Background(), "a", "b"))
}

func TestContainerBiz_EnsurePodRunning_NotRunning(t *testing.T) {
	k := &fakeK8sBizForContainer{isPodRunning: func(ns, pod string) (bool, string) {
		return false, "pod down"
	}}
	cb := newTestContainerBiz(k, &fakeFileBizForContainer{}, &fakeEventBizForContainer{})
	err := cb.EnsurePodRunning(context.Background(), "a", "b")
	assert.Equal(t, codes.NotFound, status.Code(err))
	assert.Contains(t, status.Convert(err).Message(), "pod down")
}

// ---- execSizeQueue ----

func TestExecSizeQueue_Next_ContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	cancel()
	queue := newExecSizeQueue(ctx, make(chan *TerminalSize, 1), &fakeRecorderForContainer{})
	assert.Nil(t, queue.Next())
}

func TestExecSizeQueue_Next_ChannelClosed(t *testing.T) {
	ch := make(chan *TerminalSize, 1)
	close(ch)
	queue := newExecSizeQueue(context.TODO(), ch, &fakeRecorderForContainer{})
	assert.Nil(t, queue.Next())
}

func TestExecSizeQueue_Next_SizeReceived(t *testing.T) {
	reco := &fakeRecorderForContainer{}
	ch := make(chan *TerminalSize, 1)
	expected := &TerminalSize{Width: 10, Height: 20}
	ch <- expected
	queue := newExecSizeQueue(context.TODO(), ch, reco)
	assert.Equal(t, expected, queue.Next())
	w, h := reco.size()
	assert.Equal(t, uint16(10), w)
	assert.Equal(t, uint16(20), h)
}

func TestExecSizeQueue_Next_ZeroSizeNoResize(t *testing.T) {
	// 宽或高为 0 的帧（未设置窗口）只透传尺寸，不触发 Resize。
	reco := &fakeRecorderForContainer{}
	ch := make(chan *TerminalSize, 1)
	expected := &TerminalSize{Width: 0, Height: 20}
	ch <- expected
	queue := newExecSizeQueue(context.TODO(), ch, reco)
	assert.Equal(t, expected, queue.Next())
	w, h := reco.size()
	assert.Equal(t, uint16(0), w)
	assert.Equal(t, uint16(0), h)
}

// ---- limitedBuffer / toErrStr ----

func TestLimitedBuffer_KeepsSmallOutput(t *testing.T) {
	lb := newLimitedBuffer(100)
	_, err := lb.Write([]byte("abc"))
	assert.Nil(t, err)
	assert.Equal(t, "abc", lb.String())
}

func TestLimitedBuffer_TruncatesToTail(t *testing.T) {
	lb := newLimitedBuffer(10)
	_, err := lb.Write([]byte("hello world foo bar"))
	assert.Nil(t, err)
	// 只保留最后 10 字节，避免大输出把内存打爆。
	assert.Equal(t, "ld foo bar", lb.String())
}

func TestLimitedBuffer_MultipleWritesTruncate(t *testing.T) {
	lb := newLimitedBuffer(10)
	_, _ = lb.Write([]byte("012345"))
	_, _ = lb.Write([]byte("6789abcdef"))
	assert.Equal(t, "6789abcdef", lb.String())
}

func TestLimitedBuffer_SingleHugeWrite(t *testing.T) {
	lb := newLimitedBuffer(8)
	big := make([]byte, 1<<20)
	for i := range big {
		big[i] = 'x'
	}
	_, err := lb.Write(big)
	assert.Nil(t, err)
	assert.Len(t, lb.String(), 8)
	assert.Equal(t, "xxxxxxxx", lb.String())
}

func TestToErrStr(t *testing.T) {
	assert.Equal(t, "", toErrStr(nil))
	assert.Equal(t, "error", toErrStr(errors.New("error")))
}

// ---- Exec ----

func TestContainerBiz_Exec_PodNotRunning(t *testing.T) {
	k := &fakeK8sBizForContainer{isPodRunning: func(ns, pod string) (bool, string) {
		return false, "pod down"
	}}
	cb := newTestContainerBiz(k, &fakeFileBizForContainer{}, &fakeEventBizForContainer{})
	err := cb.Exec(context.Background(), &fakeExecStream{}, &UserInfo{}, &ExecInput{Namespace: "a", Pod: "b"})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestContainerBiz_Exec_FindDefaultContainerError(t *testing.T) {
	k := &fakeK8sBizForContainer{
		isPodRunning: func(ns, pod string) (bool, string) { return true, "" },
		findDefault: func(ctx context.Context, ns, pod string) (string, error) {
			return "", errors.New("no default")
		},
	}
	cb := newTestContainerBiz(k, &fakeFileBizForContainer{}, &fakeEventBizForContainer{})
	err := cb.Exec(context.Background(), &fakeExecStream{}, &UserInfo{}, &ExecInput{Namespace: "a", Pod: "b"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no default")
}

// TestContainerBiz_Exec_SuccessExitError 覆盖快乐路径：首屏输入、尺寸队列（先推成功再
// 队列满）、输出回传、退出码帧与审计落库。execFn 用 goroutine 持续消费 stdin，避免
// io.Pipe 同步语义把 recv goroutine 的 writer.Write 卡死。
func TestContainerBiz_Exec_SuccessExitError(t *testing.T) {
	k := &fakeK8sBizForContainer{
		isPodRunning: func(ns, pod string) (bool, string) { return true, "" },
		execFn: func(ctx context.Context, c *Container, input *ExecuteInput) error {
			go func() {
				buf := make([]byte, 256)
				for {
					if _, err := input.Stdin.Read(buf); err != nil {
						return
					}
				}
			}()
			_, _ = input.Stdout.Write([]byte("hello"))
			_, _ = input.Stderr.Write([]byte("err"))
			return &ExecExitError{Code: 2}
		},
	}
	reco := &fakeRecorderForContainer{user: &UserInfo{Name: "mars"}, file: &File{ID: 1}, dur: time.Second}
	file := &fakeFileBizForContainer{recorder: func(u *UserInfo, c *Container) Recorder { return reco }}
	var audited bool
	event := &fakeEventBizForContainer{fileAudit: func(action types.EventActionType, username, operatorEmail, msg string, fileID int, duration time.Duration) {
		audited = true
		assert.Equal(t, types.EventActionType_Exec, action)
		assert.Equal(t, "mars", username)
		assert.Equal(t, 1, fileID)
	}}
	cb := newTestContainerBiz(k, file, event)
	stream := &fakeExecStream{seq: []*container.ExecRequest{
		{Namespace: "a", Pod: "b", Container: "c", Command: []string{"ls"},
			Message: []byte("world"), SizeQueue: &container.TerminalSize{Width: 10, Height: 20}},
		{Message: []byte("tail"), SizeQueue: &container.TerminalSize{Width: 15, Height: 30}},
	}}
	err := cb.Exec(context.Background(), stream, &UserInfo{Name: "mars"}, &ExecInput{
		Namespace:    "a",
		Pod:          "b",
		Container:    "c",
		Command:      []string{"ls"},
		FirstMessage: []byte("hi"),
		InitialSize:  &container.TerminalSize{Width: 100, Height: 200},
	})
	assert.Error(t, err)
	// 初始终端窗口已预先应用到 recorder。
	w, h := reco.size()
	assert.Equal(t, uint16(100), w)
	assert.Equal(t, uint16(200), h)
	assert.True(t, audited)
	// 退出错误帧已发送，退出码 2。
	assert.Equal(t, int64(2), stream.lastExitCode())
}

// TestContainerBiz_Exec_FirstMessageWriteError 覆盖 recv goroutine 首帧写错误分支：
// execFn 不消费 stdin，FirstMessage 的 writer.Write 阻塞到 closeAll 关闭 reader 后
// 返回 ErrClosedPipe——只记日志、不 panic、不中断 Exec。
func TestContainerBiz_Exec_FirstMessageWriteError(t *testing.T) {
	release := make(chan struct{})
	k := &fakeK8sBizForContainer{
		isPodRunning: func(ns, pod string) (bool, string) { return true, "" },
		execFn: func(ctx context.Context, c *Container, input *ExecuteInput) error {
			<-release
			return &ExecExitError{Code: 4}
		},
	}
	reco := &fakeRecorderForContainer{user: &UserInfo{Name: "mars"}, file: &File{ID: 1}, dur: time.Second}
	file := &fakeFileBizForContainer{recorder: func(u *UserInfo, c *Container) Recorder { return reco }}
	var audited bool
	event := &fakeEventBizForContainer{fileAudit: func(action types.EventActionType, username, operatorEmail, msg string, fileID int, duration time.Duration) {
		audited = true
	}}
	cb := newTestContainerBiz(k, file, event)
	stream := &fakeExecStream{}
	done := make(chan error, 1)
	go func() {
		done <- cb.Exec(context.Background(), stream, &UserInfo{Name: "mars"}, &ExecInput{
			Namespace:    "a",
			Pod:          "b",
			Container:    "c",
			FirstMessage: []byte("hi"),
		})
	}()
	close(release)
	assert.Error(t, <-done)
	assert.True(t, audited)
}

// TestContainerBiz_Exec_SendError 覆盖 send loop 的发送失败分支：stream.Send 恒报错，
// send loop 记 ErrorCtx 后退出，Exec 返回下层错误。
func TestContainerBiz_Exec_SendError(t *testing.T) {
	release := make(chan struct{})
	k := &fakeK8sBizForContainer{
		isPodRunning: func(ns, pod string) (bool, string) { return true, "" },
		execFn: func(ctx context.Context, c *Container, input *ExecuteInput) error {
			go func() {
				buf := make([]byte, 256)
				for {
					if _, err := input.Stdin.Read(buf); err != nil {
						return
					}
				}
			}()
			// 写字节后阻塞：等 send loop 消费并触发发送失败后，测试才 close(release)
			// 放行 execFn 返回。若不阻塞，closeAll 可能先关 pipe，send loop 走 EOF 分支，
			// 发送失败分支成为竞态缺口。
			_, _ = input.Stdout.Write([]byte("hello"))
			<-release
			return &ExecExitError{Code: 3}
		},
	}
	reco := &fakeRecorderForContainer{user: &UserInfo{Name: "mars"}, file: &File{ID: 1}, dur: time.Second}
	file := &fakeFileBizForContainer{recorder: func(u *UserInfo, c *Container) Recorder { return reco }}
	event := &fakeEventBizForContainer{fileAudit: func(action types.EventActionType, username, operatorEmail, msg string, fileID int, duration time.Duration) {
	}}
	cb := newTestContainerBiz(k, file, event)
	// recvGate 拖住 recv loop：否则 seq 耗尽后 recv loop 先 closeAll 关 pipe，
	// execFn 写的字节落入已关管道，send loop 走 EOF 分支，发送失败分支永不可达。
	// seq 携带一帧残片：close(recvGate) 放行后 recv loop 写已关管道，顺带覆盖写失败分支。
	stream := &fakeExecStream{
		seq:      []*container.ExecRequest{{Message: []byte("x")}},
		sendErr:  errors.New("send boom"),
		recvGate: make(chan struct{}),
	}
	done := make(chan error, 1)
	go func() {
		done <- cb.Exec(context.Background(), stream, &UserInfo{Name: "mars"}, &ExecInput{
			Namespace: "a", Pod: "b", Container: "c",
		})
	}()
	// 轮询 send loop 已发送过至少一帧：read loop 消费字节后 sendMsg 失败 → closeAll
	// 关 pipe（确定性已关），此时 recv loop 仍被 recvGate 拖住。
	assert.Eventually(t, func() bool {
		stream.mu.Lock()
		defer stream.mu.Unlock()
		return len(stream.sent) > 0
	}, time.Second, 10*time.Millisecond)
	close(release)
	close(stream.recvGate)
	assert.Error(t, <-done)
}

// ---- ExecOnce ----

func TestContainerBiz_ExecOnce_PodNotRunning(t *testing.T) {
	k := &fakeK8sBizForContainer{isPodRunning: func(ns, pod string) (bool, string) {
		return false, "pod down"
	}}
	cb := newTestContainerBiz(k, &fakeFileBizForContainer{}, &fakeEventBizForContainer{})
	err := cb.ExecOnce(context.Background(), &fakeExecOnceStream{}, &UserInfo{}, &ExecOnceInput{Namespace: "a", Pod: "b"})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestContainerBiz_ExecOnce_FindDefaultContainerError(t *testing.T) {
	k := &fakeK8sBizForContainer{
		isPodRunning: func(ns, pod string) (bool, string) { return true, "" },
		findDefault: func(ctx context.Context, ns, pod string) (string, error) {
			return "", errors.New("no default")
		},
	}
	cb := newTestContainerBiz(k, &fakeFileBizForContainer{}, &fakeEventBizForContainer{})
	err := cb.ExecOnce(context.Background(), &fakeExecOnceStream{}, &UserInfo{}, &ExecOnceInput{Namespace: "a", Pod: "b"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no default")
}

// TestContainerBiz_ExecOnce_Success 覆盖默认容器回落 + 命令输出有界捕获 + 审计落库
// （错误为空串、result 为 stdout 内容）。
func TestContainerBiz_ExecOnce_Success(t *testing.T) {
	k := &fakeK8sBizForContainer{
		isPodRunning: func(ns, pod string) (bool, string) { return true, "" },
		findDefault: func(ctx context.Context, ns, pod string) (string, error) {
			return "c", nil
		},
		execFn: func(ctx context.Context, c *Container, input *ExecuteInput) error {
			_, _ = input.Stdout.Write([]byte("hello"))
			return nil
		},
	}
	var captured AnyYamlPrettier
	event := &fakeEventBizForContainer{audit: func(action types.EventActionType, username, operatorEmail, msg string, oldS, newS YamlPrettier) {
		assert.Equal(t, types.EventActionType_Exec, action)
		assert.Equal(t, "admin", username)
		assert.Nil(t, oldS)
		captured, _ = newS.(AnyYamlPrettier)
	}}
	cb := newTestContainerBiz(k, &fakeFileBizForContainer{}, event)
	err := cb.ExecOnce(context.Background(), &fakeExecOnceStream{}, &UserInfo{Name: "admin"}, &ExecOnceInput{
		Namespace: "a", Pod: "b", Command: []string{"ls"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "a", captured["namespace"])
	assert.Equal(t, "b", captured["pod"])
	assert.Equal(t, "c", captured["container"])
	assert.Equal(t, []string{"ls"}, captured["command"])
	assert.Equal(t, "hello", captured["result"])
	assert.Equal(t, "", captured["error"])
	assert.NotEmpty(t, captured["duration"])
}

// TestContainerBiz_ExecOnce_ExitError 覆盖退出码帧发送与审计里错误信息透传。
func TestContainerBiz_ExecOnce_ExitError(t *testing.T) {
	k := &fakeK8sBizForContainer{
		isPodRunning: func(ns, pod string) (bool, string) { return true, "" },
		findDefault: func(ctx context.Context, ns, pod string) (string, error) {
			return "c", nil
		},
		execFn: func(ctx context.Context, c *Container, input *ExecuteInput) error {
			_, _ = input.Stdout.Write([]byte("out"))
			return &ExecExitError{Code: 1, Message: "boom"}
		},
	}
	var captured AnyYamlPrettier
	event := &fakeEventBizForContainer{audit: func(action types.EventActionType, username, operatorEmail, msg string, oldS, newS YamlPrettier) {
		captured, _ = newS.(AnyYamlPrettier)
	}}
	cb := newTestContainerBiz(k, &fakeFileBizForContainer{}, event)
	stream := &fakeExecOnceStream{}
	err := cb.ExecOnce(context.Background(), stream, &UserInfo{Name: "admin"}, &ExecOnceInput{
		Namespace: "a", Pod: "b", Command: []string{"ls"},
	})
	assert.Error(t, err)
	assert.Equal(t, int64(1), stream.lastExitCode())
	assert.Equal(t, "out", captured["result"])
	assert.Contains(t, captured["error"].(string), "boom")
}

// TestContainerBiz_ExecOnce_SendError 覆盖 ExecOnce send loop 的发送失败分支。
func TestContainerBiz_ExecOnce_SendError(t *testing.T) {
	k := &fakeK8sBizForContainer{
		isPodRunning: func(ns, pod string) (bool, string) { return true, "" },
		findDefault: func(ctx context.Context, ns, pod string) (string, error) {
			return "c", nil
		},
		execFn: func(ctx context.Context, c *Container, input *ExecuteInput) error {
			_, _ = input.Stdout.Write([]byte("out"))
			// 返回退出错误让 ExecOnce 的返回值非 nil（send loop 的失败只记日志不改变返回值）。
			return &ExecExitError{Code: 1}
		},
	}
	event := &fakeEventBizForContainer{audit: func(action types.EventActionType, username, operatorEmail, msg string, oldS, newS YamlPrettier) {}}
	cb := newTestContainerBiz(k, &fakeFileBizForContainer{}, event)
	stream := &fakeExecOnceStream{sendErr: errors.New("send boom")}
	err := cb.ExecOnce(context.Background(), stream, &UserInfo{Name: "admin"}, &ExecOnceInput{
		Namespace: "a", Pod: "b",
	})
	assert.Error(t, err)
}

// TestContainerBiz_ExecOnce_ExitErrorSendFailure 覆盖退出码错误帧发送失败分支：
// errFrameErr 只让错误帧失败，不打断 send loop，覆盖主流程发退出码帧失败只记日志。
func TestContainerBiz_ExecOnce_ExitErrorSendFailure(t *testing.T) {
	k := &fakeK8sBizForContainer{
		isPodRunning: func(ns, pod string) (bool, string) { return true, "" },
		findDefault: func(ctx context.Context, ns, pod string) (string, error) {
			return "c", nil
		},
		execFn: func(ctx context.Context, c *Container, input *ExecuteInput) error {
			_, _ = input.Stdout.Write([]byte("out"))
			return &ExecExitError{Code: 1, Message: "boom"}
		},
	}
	event := &fakeEventBizForContainer{audit: func(action types.EventActionType, username, operatorEmail, msg string, oldS, newS YamlPrettier) {}}
	cb := newTestContainerBiz(k, &fakeFileBizForContainer{}, event)
	stream := &fakeExecOnceStream{errFrameErr: errors.New("frame boom")}
	err := cb.ExecOnce(context.Background(), stream, &UserInfo{Name: "admin"}, &ExecOnceInput{
		Namespace: "a", Pod: "b", Command: []string{"ls"},
	})
	assert.Error(t, err)
	assert.Equal(t, int64(1), stream.lastExitCode())
}

// ---- execOnceDeadline / cappedWriter 单元 ----

func TestExecOnceDeadline(t *testing.T) {
	assert.Equal(t, time.Duration(30)*time.Second, execOnceDeadline(0), "0 用默认 30s")
	assert.Equal(t, time.Duration(30)*time.Second, execOnceDeadline(-5), "负值用默认 30s")
	assert.Equal(t, time.Duration(5)*time.Second, execOnceDeadline(5), "正数透传")
}

// errWriter 恒返回错误，用于 cappedWriter 底层写失败分支。
type errWriter struct{}

func (errWriter) Write(p []byte) (int, error) { return 0, errors.New("write boom") }

func TestCappedWriter(t *testing.T) {
	var buf bytes.Buffer
	cancelled := false
	c := newCappedWriter(&buf, 10, func() { cancelled = true })

	// 上限内：完整转发，不截断不取消。
	n, err := c.Write([]byte("12345"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.False(t, c.isTruncated())
	assert.Equal(t, "12345", buf.String())

	// 单次越界：转发剩余 5 字节后封顶截断，并触发 cancel。
	n, err = c.Write([]byte("abcdefghij"))
	assert.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.True(t, c.isTruncated())
	assert.True(t, cancelled)
	assert.Equal(t, "12345abcde", buf.String())

	// 已截断：后续 Write 直接丢弃（不转发、不重复 cancel）。
	cancelled = false
	n, err = c.Write([]byte("zzz"))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, "12345abcde", buf.String())
	assert.False(t, cancelled)
}

func TestCappedWriter_NilCancelAndWriteError(t *testing.T) {
	// nil cancel 不 panic。
	c := newCappedWriter(io.Discard, 10, nil)
	_, err := c.Write([]byte("abc"))
	assert.NoError(t, err)
	assert.False(t, c.isTruncated())

	// 底层写错误透传（剩余计算不触发截断）。
	c2 := newCappedWriter(errWriter{}, 10, nil)
	n, err := c2.Write([]byte("hello"))
	assert.Error(t, err)
	assert.Equal(t, 5, n)

	// 越界时的部分转发若底层写失败：同样透传错误（不置截断）。
	c3 := newCappedWriter(errWriter{}, 10, nil)
	n, err = c3.Write(make([]byte, 20))
	assert.Error(t, err)
	assert.Equal(t, 20, n)
}

// ---- ExecOnce 超时 wiring 与输出截断 ----

// TestContainerBiz_ExecOnce_TimeoutWiring 验证请求超时被接进 Execute 的 ctx deadline。
func TestContainerBiz_ExecOnce_TimeoutWiring(t *testing.T) {
	k := &fakeK8sBizForContainer{
		isPodRunning: func(ns, pod string) (bool, string) { return true, "" },
		findDefault: func(ctx context.Context, ns, pod string) (string, error) {
			return "c", nil
		},
		execFn: func(ctx context.Context, c *Container, input *ExecuteInput) error {
			deadline, ok := ctx.Deadline()
			assert.True(t, ok, "Execute 应收到带 deadline 的 ctx")
			assert.WithinDuration(t, time.Now().Add(5*time.Second), deadline, 2*time.Second)
			return nil
		},
	}
	event := &fakeEventBizForContainer{audit: func(action types.EventActionType, username, operatorEmail, msg string, oldS, newS YamlPrettier) {}}
	cb := newTestContainerBiz(k, &fakeFileBizForContainer{}, event)
	err := cb.ExecOnce(context.Background(), &fakeExecOnceStream{}, &UserInfo{Name: "admin"}, &ExecOnceInput{
		Namespace: "a", Pod: "b", Command: []string{"ls"}, TimeoutSeconds: 5,
	})
	assert.NoError(t, err)
}

// TestContainerBiz_ExecOnce_Truncation 覆盖输出超限：发一条明确的截断错误帧而非静默断流。
func TestContainerBiz_ExecOnce_Truncation(t *testing.T) {
	big := make([]byte, maxExecOnceStreamSize+16)
	k := &fakeK8sBizForContainer{
		isPodRunning: func(ns, pod string) (bool, string) { return true, "" },
		findDefault: func(ctx context.Context, ns, pod string) (string, error) {
			return "c", nil
		},
		execFn: func(ctx context.Context, c *Container, input *ExecuteInput) error {
			_, _ = input.Stdout.Write(big)
			return nil
		},
	}
	event := &fakeEventBizForContainer{audit: func(action types.EventActionType, username, operatorEmail, msg string, oldS, newS YamlPrettier) {}}
	cb := newTestContainerBiz(k, &fakeFileBizForContainer{}, event)
	stream := &fakeExecOnceStream{}
	err := cb.ExecOnce(context.Background(), stream, &UserInfo{Name: "admin"}, &ExecOnceInput{
		Namespace: "a", Pod: "b", Command: []string{"cat"},
	})
	assert.NoError(t, err)
	truncFrame := stream.findError(execOnceTruncatedCode)
	assert.NotNil(t, truncFrame, "应发送一条输出超限截断错误帧")
	assert.Contains(t, truncFrame.Message, "上限")
}

// TestContainerBiz_ExecOnce_TruncationSendError 覆盖截断错误帧发送失败分支：send 报错时
// 只记 Debug 日志不 panic，用例仍正常返回。
func TestContainerBiz_ExecOnce_TruncationSendError(t *testing.T) {
	big := make([]byte, maxExecOnceStreamSize+16)
	k := &fakeK8sBizForContainer{
		isPodRunning: func(ns, pod string) (bool, string) { return true, "" },
		findDefault: func(ctx context.Context, ns, pod string) (string, error) {
			return "c", nil
		},
		execFn: func(ctx context.Context, c *Container, input *ExecuteInput) error {
			_, _ = input.Stdout.Write(big)
			return nil
		},
	}
	event := &fakeEventBizForContainer{audit: func(action types.EventActionType, username, operatorEmail, msg string, oldS, newS YamlPrettier) {}}
	cb := newTestContainerBiz(k, &fakeFileBizForContainer{}, event)
	// errFrameErr 仅让截断错误帧发送失败，不打断 send loop 对 1MiB 输出的消费，
	// 故截断仍会触发，从而覆盖"截断帧发送失败只记日志"分支。
	stream := &fakeExecOnceStream{errFrameErr: errors.New("frame boom")}
	err := cb.ExecOnce(context.Background(), stream, &UserInfo{Name: "admin"}, &ExecOnceInput{
		Namespace: "a", Pod: "b", Command: []string{"cat"},
	})
	assert.NoError(t, err)
}
