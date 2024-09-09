package k8s

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/uploader"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/exec"
)

type FileCopy interface {
	CopyFromPod(ctx context.Context, src CopyFileSpec, file uploader.File) error
}

// CopyOptions have the data required to perform the copy operation
type CopyOptions struct {
	Namespace string
	Container string
	MaxTries  int

	ClientConfig *restclient.Config
	Clientset    kubernetes.Interface

	errOut io.Writer
	logger mlog.Logger
}

// NewCopyOptions creates the options for copy
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

func (o *CopyOptions) CopyFromPod(ctx context.Context, src CopyFileSpec, destFile uploader.File) error {
	reader := newTarPipe(ctx, src, o)
	o.Namespace = src.PodNamespace
	o.Container = src.ContainerName
	if _, err := io.Copy(destFile, reader); err != nil {
		if err != io.EOF {
			return err
		}
	}
	return nil
}

type TarPipe struct {
	src       CopyFileSpec
	o         *CopyOptions
	reader    *io.PipeReader
	outStream *io.PipeWriter
	bytesRead uint64
	retries   int
	ctx       context.Context
}

func newTarPipe(ctx context.Context, src CopyFileSpec, o *CopyOptions) *TarPipe {
	t := new(TarPipe)
	t.src = src
	t.o = o
	t.ctx = ctx
	t.initReadFrom(0)
	return t
}

func (t *TarPipe) initReadFrom(n uint64) {
	t.reader, t.outStream = io.Pipe()
	options := &exec.ExecOptions{
		StreamOptions: exec.StreamOptions{
			IOStreams: genericiooptions.IOStreams{
				In:     nil,
				Out:    t.outStream,
				ErrOut: t.o.errOut,
			},

			Namespace: t.src.PodNamespace,
			PodName:   t.src.PodName,
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

func (o *CopyOptions) execute(options *exec.ExecOptions) error {
	if len(options.Namespace) == 0 {
		options.Namespace = o.Namespace
	}

	if len(o.Container) > 0 {
		options.ContainerName = o.Container
	}

	options.Config = o.ClientConfig
	options.PodClient = o.Clientset.CoreV1()

	if err := options.Validate(); err != nil {
		return err
	}

	return options.Run()
}

type CopyFileSpec struct {
	PodName      string
	PodNamespace string
	// ContainerName optional
	ContainerName string
	File          PathSpec
}

type PathSpec interface {
	String() string
}

// RemotePath represents always UNIX path, its methods will use path
// package which is always using `/`
type RemotePath struct {
	file string
}

func NewRemotePath(fileName string) RemotePath {
	// we assume remote file is a linux container but we need to convert
	// windows path separators to unix style for consistent processing
	file := strings.ReplaceAll(stripTrailingSlash(fileName), `\`, "/")
	return RemotePath{file: file}
}

func (p RemotePath) String() string {
	return p.file
}

// strips trailing slash (if any) both unix and windows style
func stripTrailingSlash(file string) string {
	if len(file) == 0 {
		return file
	}
	if file != "/" && strings.HasSuffix(string(file[len(file)-1]), "/") {
		return file[:len(file)-1]
	}
	return file
}
