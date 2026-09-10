package k8sutil

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes/fake"
	restclient "k8s.io/client-go/rest"
	k8stesting "k8s.io/client-go/testing"
	"k8s.io/kubectl/pkg/cmd/exec"
)

func TestRemotePathString(t *testing.T) {
	path := NewRemotePath("/test/path")
	assert.Equal(t, "/test/path", path.String())
}

func TestStripTrailingSlash(t *testing.T) {
	assert.Equal(t, "/test/path", stripTrailingSlash("/test/path/"))
	assert.Equal(t, "/test/path", stripTrailingSlash("/test/path"))
	assert.Equal(t, "", stripTrailingSlash(""))
}

func TestNewCopyOptions(t *testing.T) {
	options := NewCopyOptions(mlog.NewForConfig(nil), &restclient.Config{}, fake.NewSimpleClientset(), 10, &bytes.Buffer{})
	assert.NotNil(t, options)
	assert.NotNil(t, options.logger)
	assert.NotNil(t, options.ClientConfig)
	assert.NotNil(t, options.Clientset)
	assert.NotEmpty(t, options.MaxTries)
	assert.NotNil(t, options.errOut)
}

// fakeUploadFile 是 uploader.File 的最小测试替身：io.Copy 只写不读，缓冲即可。
type fakeUploadFile struct {
	*bytes.Buffer
}

// newFakeUploadFile 构造可写的上传文件替身。
func newFakeUploadFile() *fakeUploadFile {
	return &fakeUploadFile{Buffer: &bytes.Buffer{}}
}

// Name 返回空文件名（测试不需要真实路径）。
func (f *fakeUploadFile) Name() string { return "" }

// Stat 返回 nil（测试不依赖元数据）。
func (f *fakeUploadFile) Stat() (os.FileInfo, error) { return nil, nil }

// Seek 是 no-op。
func (f *fakeUploadFile) Seek(_ int64, _ int) (int64, error) { return 0, nil }

// Close 是 no-op。
func (f *fakeUploadFile) Close() error { return nil }

// spec 是常用测试源描述：pod/命名空间/容器与远端路径齐全。
func spec() CopyFileSpec {
	return CopyFileSpec{
		PodName:       "pod",
		PodNamespace:  "ns",
		ContainerName: "c",
		File:          NewRemotePath("/tmp/x"),
	}
}

// newTestCopyOptions 用 fake clientset 装配拷贝器：pod 不存在时 execute 的
// Run 走 PodClient.Pods().Get 立即返回 NotFound，无需真实集群。
func newTestCopyOptions(maxTries int) *CopyOptions {
	return NewCopyOptions(mlog.NewForConfig(nil), &restclient.Config{}, fake.NewSimpleClientset(), maxTries, &bytes.Buffer{})
}

// TestCopyOptions_CopyFromPod_Success 覆盖 CopyFromPod 成功路径：execute 因 pod
// 不存在快速失败、goroutine 关闭 pipe，主流程 io.Copy 读到 EOF 返回 nil。
func TestCopyOptions_CopyFromPod_Success(t *testing.T) {
	o := newTestCopyOptions(5)
	err := o.CopyFromPod(context.Background(), spec(), newFakeUploadFile())
	assert.NoError(t, err)
}

// TestCopyOptions_CopyFromPod_MaxTriesZero 覆盖 initReadFrom 的 MaxTries==0 分支
// （默认 tar 命令，不带 tail 续传）。
func TestCopyOptions_CopyFromPod_MaxTriesZero(t *testing.T) {
	o := newTestCopyOptions(0)
	err := o.CopyFromPod(context.Background(), spec(), newFakeUploadFile())
	assert.NoError(t, err)
}

// TestCopyOptions_CopyFromPod_CtxCanceled 覆盖 CopyFromPod 错误分支：阻塞 Get
// reactor 让 goroutine 停在 execute（pipe 永不关闭），预取消 ctx 使 io.Copy
// 立即拿到 ctx.Err() 而非 EOF，确定性触发 `if err != io.EOF`。
func TestCopyOptions_CopyFromPod_CtxCanceled(t *testing.T) {
	client := fake.NewSimpleClientset()
	block := make(chan struct{})
	client.PrependReactor("get", "pods", func(_ k8stesting.Action) (bool, runtime.Object, error) {
		<-block
		return true, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "ns"},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		}, nil
	})
	t.Cleanup(func() { close(block) })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	o := NewCopyOptions(mlog.NewForConfig(nil), &restclient.Config{Host: "https://127.0.0.1:1"}, client, 5, &bytes.Buffer{})
	err := o.CopyFromPod(ctx, spec(), newFakeUploadFile())
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// TestTarPipe_Read_CtxDone 覆盖 Read 的 ctx.Done 早退分支。
func TestTarPipe_Read_CtxDone(t *testing.T) {
	pr, pw := io.Pipe()
	defer pw.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tp := &TarPipe{reader: pr, outStream: pw, ctx: ctx, o: newTestCopyOptions(5)}

	n, err := tp.Read(make([]byte, 8))
	assert.Equal(t, 0, n)
	assert.Equal(t, context.Canceled, err)
}

// TestTarPipe_Read_EOF 覆盖 Read 的 io.EOF 分支（写端干净关闭）。
func TestTarPipe_Read_EOF(t *testing.T) {
	pr, pw := io.Pipe()
	tp := &TarPipe{reader: pr, outStream: pw, ctx: context.Background(), o: newTestCopyOptions(5)}
	_ = pw.Close()

	n, err := tp.Read(make([]byte, 8))
	assert.Equal(t, 0, n)
	assert.Equal(t, io.EOF, err)
}

// TestTarPipe_Read_Data 覆盖 Read 的正常数据分支（bytesRead 累加）。
func TestTarPipe_Read_Data(t *testing.T) {
	pr, pw := io.Pipe()
	tp := &TarPipe{reader: pr, outStream: pw, ctx: context.Background(), o: newTestCopyOptions(5)}

	go func() {
		_, _ = pw.Write([]byte("hello"))
		_ = pw.Close()
	}()

	buf := make([]byte, 8)
	n, err := tp.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte("hello"), buf[:5])
	assert.Equal(t, uint64(5), tp.bytesRead)
}

// TestTarPipe_Read_Retry 覆盖 Read 重试分支：CloseWithError 令 pipe 返回非 EOF
// 错误，retries(0) < MaxTries(5) 走 initReadFrom 续传并重置 err。
func TestTarPipe_Read_Retry(t *testing.T) {
	pr, pw := io.Pipe()
	tp := &TarPipe{
		reader: pr, outStream: pw, ctx: context.Background(),
		o:   newTestCopyOptions(5),
		src: spec(),
	}
	_ = pw.CloseWithError(errors.New("boom"))

	n, err := tp.Read(make([]byte, 8))
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, 1, tp.retries)
}

// TestTarPipe_Read_Drop 覆盖 Read 丢弃分支：MaxTries==0 时重试条件恒假，
// 保留非 EOF 错误原样返回。
func TestTarPipe_Read_Drop(t *testing.T) {
	pr, pw := io.Pipe()
	tp := &TarPipe{reader: pr, outStream: pw, ctx: context.Background(), o: newTestCopyOptions(0)}
	_ = pw.CloseWithError(errors.New("boom"))

	n, err := tp.Read(make([]byte, 8))
	assert.Equal(t, 0, n)
	assert.Error(t, err)
}

// TestCopyOptions_execute_ValidateError 覆盖 execute 的 Validate 失败分支：
// 空 ExecOptions 使 Validate 报错并提前返回。
func TestCopyOptions_execute_ValidateError(t *testing.T) {
	o := newTestCopyOptions(5)
	err := o.execute(&exec.ExecOptions{})
	assert.Error(t, err)
}

// TestCopyOptions_execute_RunError 覆盖 execute 的 Run 失败分支：options 齐全
// 通过 Validate，Run 经 fake clientset Get 不存在的 pod 返回 NotFound。
func TestCopyOptions_execute_RunError(t *testing.T) {
	o := newTestCopyOptions(5)

	options := &exec.ExecOptions{
		StreamOptions: exec.StreamOptions{
			IOStreams: genericiooptions.IOStreams{Out: io.Discard, ErrOut: io.Discard},
			Namespace: "ns",
			PodName:   "pod",
		},
		Command: []string{"tar", "cf", "-", "/tmp/x"},
	}
	err := o.execute(options)
	assert.Error(t, err)
}
