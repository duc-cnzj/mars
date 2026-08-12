package k8sutil

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/exec"
)

// FileCopy 定义从 Pod 复制文件到 uploader.File 的端口。
type FileCopy interface {
	// CopyFromPod 从远端 Pod 按 src 描述拉取文件内容写入 file。
	CopyFromPod(ctx context.Context, src CopyFileSpec, file uploader.File) error
}

// CopyOptions 持有执行一次 Pod 文件复制所需的客户端配置：rest 配置、clientset 与重试次数。
type CopyOptions struct {
	MaxTries int

	ClientConfig *restclient.Config
	Clientset    kubernetes.Interface

	errOut io.Writer
	logger mlog.Logger
}

// NewCopyOptions 构造文件复制器，绑定客户端配置、最大重试次数与错误输出。
func NewCopyOptions(
	logger mlog.Logger,
	clientConfig *restclient.Config,
	clientset kubernetes.Interface,
	maxTries int,
	errOut io.Writer,
) *CopyOptions {
	return &CopyOptions{
		logger:       logger,
		MaxTries:     maxTries,
		ClientConfig: clientConfig,
		Clientset:    clientset,
		errOut:       errOut,
	}
}

// CopyFromPod 从远端 Pod 按 src 打包拉取文件内容写入 destFile。
func (o *CopyOptions) CopyFromPod(ctx context.Context, src CopyFileSpec, destFile uploader.File) error {
	reader := newTarPipe(ctx, src, o)
	if _, err := io.Copy(destFile, reader); err != nil {
		if err != io.EOF {
			return err
		}
	}
	return nil
}

// TarPipe 是 io.Reader 实现：经 kubectl exec 远端 tar 打包文件流，支持断点续传重试。
type TarPipe struct {
	src       CopyFileSpec
	o         *CopyOptions
	reader    *io.PipeReader
	outStream *io.PipeWriter
	bytesRead uint64
	retries   int
	ctx       context.Context
}

// newTarPipe 构造 TarPipe，并启动后台 goroutine 通过 kubectl exec tar 打包远端文件。
func newTarPipe(ctx context.Context, src CopyFileSpec, o *CopyOptions) *TarPipe {
	t := new(TarPipe)
	t.src = src
	t.o = o
	t.ctx = ctx
	t.initReadFrom(0)
	return t
}

// initReadFrom 重建 io.Pipe 并启动一次 tar 流读取；断点续传时从 n 字节处恢复。
func (t *TarPipe) initReadFrom(n uint64) {
	t.reader, t.outStream = io.Pipe()
	options := &exec.ExecOptions{
		StreamOptions: exec.StreamOptions{
			IOStreams: genericiooptions.IOStreams{
				In:     nil,
				Out:    t.outStream,
				ErrOut: t.o.errOut,
			},

			Namespace:     t.src.PodNamespace,
			PodName:       t.src.PodName,
			ContainerName: t.src.ContainerName,
		},

		Command:  []string{"tar", "cf", "-", t.src.File.String()},
		Executor: &exec.DefaultRemoteExecutor{},
	}
	if t.o.MaxTries != 0 {
		options.Command = []string{"sh", "-c", fmt.Sprintf("tar cf - %s | tail -c+%d", t.src.File, n)}
	}

	go func() {
		defer t.outStream.Close()
		if err := t.o.execute(options); err != nil {
			t.o.logger.Error(err)
		}
	}()
}

// Read 从 tar 流读取数据；中断时按 MaxTries 重试续传，context 取消则立即返回。
func (t *TarPipe) Read(p []byte) (n int, err error) {
	select {
	case <-t.ctx.Done():
		return 0, t.ctx.Err()
	default:
		n, err = t.reader.Read(p)
		if err != nil {
			if err == io.EOF {
				// 处理读取到文件末尾的情况
				return n, io.EOF
			}
			if t.o.MaxTries < 0 || t.retries < t.o.MaxTries {
				t.retries++
				t.o.logger.Warningf("Resuming copy at %d bytes, retry %d/%d\n", t.bytesRead, t.retries, t.o.MaxTries)
				t.initReadFrom(t.bytesRead + 1)
				err = nil
			} else {
				t.o.logger.Warningf("Dropping out copy after %d retries err: %v\n", t.retries, err)
			}
		} else {
			t.bytesRead += uint64(n)
		}
		return n, err
	}
}

// execute 补齐 ExecOptions 的客户端配置并执行远端命令；Namespace/PodName/ContainerName
// 已由 initReadFrom 从 src 设置，此处只补客户端引用。
func (o *CopyOptions) execute(options *exec.ExecOptions) error {
	options.Config = o.ClientConfig
	options.PodClient = o.Clientset.CoreV1()

	if err := options.Validate(); err != nil {
		return err
	}

	return options.Run()
}

// CopyFileSpec 描述一次从 Pod 复制文件的源：Pod 名/命名空间/容器与远端路径。
type CopyFileSpec struct {
	PodName      string
	PodNamespace string
	// ContainerName 可选的容器名：留空时由远端默认容器承接。
	ContainerName string
	File          PathSpec
}

// PathSpec 是远端路径的抽象：String 返回路径字符串。
type PathSpec interface {
	// String 返回路径字符串。
	String() string
}

// RemotePath 表示远端 unix 风格路径：方法始终用 `/` 分隔，不处理窗口风格。
type RemotePath struct {
	file string
}

// NewRemotePath 构造远端路径：把窗口风格分隔符统一为 unix 风格并去除尾部斜杠。
func NewRemotePath(fileName string) RemotePath {
	// 远端容器按 linux 处理，故需把窗口风格分隔符转为 unix 风格以保证一致。
	file := strings.ReplaceAll(stripTrailingSlash(fileName), `\`, "/")
	return RemotePath{file: file}
}

// String 返回 unix 风格的路径字符串。
func (p RemotePath) String() string {
	return p.file
}

// stripTrailingSlash 去除路径尾部斜杠（unix/windows 两种风格），根路径 "/" 除外。
func stripTrailingSlash(file string) string {
	if len(file) == 0 {
		return file
	}
	if file != "/" && strings.HasSuffix(string(file[len(file)-1]), "/") {
		return file[:len(file)-1]
	}
	return file
}
