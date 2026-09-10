package data

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// Test_k8sRepo_LogStream 覆盖 LogStream 成功路径：日志经假 HTTP 服务流式返回，
// 读循环逐行投递到 channel，EOF 后关闭 channel。
func Test_k8sRepo_LogStream(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("line1\nline2\n"))
	}))
	defer server.Close()

	client, err := kubernetes.NewForConfig(&restclient.Config{Host: server.URL, Timeout: 5 * time.Second})
	require.NoError(t, err)
	mockData := NewMockDataStore(m)
	mockData.EXPECT().K8s().Return(&K8sClient{Client: client}).AnyTimes()
	repo := &k8sRepo{data: mockData, logger: mlog.NewForConfig(nil)}

	ch, err := repo.LogStream(context.TODO(), "ns", "pod", "container")
	require.NoError(t, err)

	var got []string
	for line := range ch {
		got = append(got, string(line))
	}
	assert.Equal(t, []string{"line1\n", "line2\n"}, got)
}

// Test_k8sRepo_LogStream_DropLine 覆盖 LogStream 缓冲满丢行分支（k8s.go 904-905）：
// 服务端一次性写入远超缓冲容量（1000）的行，消费前 sleep 让读循环先灌满缓冲，
// 使后续行落入 select default 丢弃分支；断言读到的行数小于总行数证明确有丢弃。
func Test_k8sRepo_LogStream_DropLine(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	const totalLines = 3000
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		var sb strings.Builder
		for i := 0; i < totalLines; i++ {
			sb.WriteString("line\n")
		}
		_, _ = w.Write([]byte(sb.String()))
	}))
	defer server.Close()

	client, err := kubernetes.NewForConfig(&restclient.Config{Host: server.URL, Timeout: 5 * time.Second})
	require.NoError(t, err)
	mockData := NewMockDataStore(m)
	mockData.EXPECT().K8s().Return(&K8sClient{Client: client}).AnyTimes()
	repo := &k8sRepo{data: mockData, logger: mlog.NewForConfig(nil)}

	ch, err := repo.LogStream(context.TODO(), "ns", "pod", "container")
	require.NoError(t, err)

	// 等待读循环灌满缓冲并处理完全部行后再消费，确保 default 丢弃分支被触发。
	time.Sleep(300 * time.Millisecond)

	var got []string
	for line := range ch {
		got = append(got, string(line))
	}
	assert.Less(t, len(got), totalLines)
}

// Test_k8sRepo_LogStream_StreamError 覆盖 LogStream 打开流失败分支（k8s.go 880-881）：
// 指向已关闭端口，Stream 连接拒绝。
func Test_k8sRepo_LogStream_StreamError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	closedURL := server.URL
	server.Close()

	client, err := kubernetes.NewForConfig(&restclient.Config{Host: closedURL, Timeout: time.Second})
	require.NoError(t, err)
	mockData := NewMockDataStore(m)
	mockData.EXPECT().K8s().Return(&K8sClient{Client: client}).AnyTimes()
	repo := &k8sRepo{data: mockData, logger: mlog.NewForConfig(nil)}

	_, err = repo.LogStream(context.TODO(), "ns", "pod", "container")
	assert.Error(t, err)
}

// Test_executor_Execute 覆盖 Execute 两个错误分支：NewSPDYExecutor 因非法 CA 构造失败
// （k8s.go 1146-1148）、StreamWithContext 对不可达端口握手失败（k8s.go 1150）。
func Test_executor_Execute(t *testing.T) {
	t.Run("NewSPDYExecutor error with invalid CA", func(t *testing.T) {
		client, err := kubernetes.NewForConfig(&restclient.Config{Host: "https://127.0.0.1:1", Timeout: time.Second})
		require.NoError(t, err)
		e := &executor{
			namespace: "ns", pod: "pod", container: "c",
			method: http.MethodPost, cmd: []string{"ls"},
			clientSet: client,
			config: &restclient.Config{
				Host:            "https://127.0.0.1:1",
				TLSClientConfig: restclient.TLSClientConfig{CAData: []byte("invalid-ca")},
			},
		}
		err = e.Execute(context.TODO(), &biz.ExecuteInput{
			Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}, TTY: false,
		})
		assert.Error(t, err)
	})

	t.Run("StreamWithContext error with unreachable host", func(t *testing.T) {
		client, err := kubernetes.NewForConfig(&restclient.Config{Host: "https://127.0.0.1:1", Timeout: time.Second})
		require.NoError(t, err)
		e := &executor{
			namespace: "ns", pod: "pod", container: "c",
			method: http.MethodPost, cmd: []string{"ls"},
			clientSet: client,
			config:    &restclient.Config{Host: "https://127.0.0.1:1", Timeout: time.Second},
		}
		err = e.Execute(context.TODO(), &biz.ExecuteInput{
			Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}, TTY: false,
		})
		assert.Error(t, err)
	})
}

// fakeSPDYExecutor 是 remotecommand.Executor 的测试替身：StreamWithContext 直接
// 返回预置结果，隔离真实 SPDY 握手（集成边界），覆盖 Execute 成功路径 return nil。
type fakeSPDYExecutor struct {
	err error
}

// Stream 兼容 remotecommand.Executor 接口（项目使用 StreamWithContext，此处透传预置结果）。
func (f *fakeSPDYExecutor) Stream(_ remotecommand.StreamOptions) error { return f.err }

// StreamWithContext 返回预置错误，模拟握手成功/失败。
func (f *fakeSPDYExecutor) StreamWithContext(_ context.Context, _ remotecommand.StreamOptions) error {
	return f.err
}

// Test_executor_Execute_Success 覆盖 Execute 成功路径：注入 fakeSPDYExecutor 令
// StreamWithContext 返回 nil，断言方法收尾 return nil（k8s.go 成功分支）。
func Test_executor_Execute_Success(t *testing.T) {
	orig := newSPDYExecutor
	newSPDYExecutor = func(*restclient.Config, string, *url.URL) (remotecommand.Executor, error) {
		return &fakeSPDYExecutor{}, nil
	}
	defer func() { newSPDYExecutor = orig }()

	// RESTClient 构造 exec 请求需要非 nil baseURL，fake clientset 的 RESTClient 不满足；
	// 用真实 clientset 仅构造请求（不发出网络），握手由 seam 注入的 fake executor 承接。
	client, err := kubernetes.NewForConfig(&restclient.Config{Host: "https://127.0.0.1:1", Timeout: time.Second})
	require.NoError(t, err)
	e := &executor{
		namespace: "ns", pod: "pod", container: "c",
		method: http.MethodPost, cmd: []string{"ls"},
		clientSet: client,
		config:    &restclient.Config{Host: "https://127.0.0.1:1", Timeout: time.Second},
	}
	err = e.Execute(context.TODO(), &biz.ExecuteInput{
		Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}, TTY: false,
	})
	assert.NoError(t, err)
}
