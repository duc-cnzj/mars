package executor

import (
	"io"

	"github.com/duc-cnzj/mars/internal/contracts"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type defaultRemoteExecutor struct {
	method                    string
	namespace, pod, container string
	cmd                       []string
}

func NewDefaultRemoteExecutor() *defaultRemoteExecutor {
	return &defaultRemoteExecutor{}
}

func (e *defaultRemoteExecutor) WithMethod(method string) contracts.RemoteExecutor {
	e.method = method
	return e
}

func (e *defaultRemoteExecutor) WithContainer(namespace, pod, container string) contracts.RemoteExecutor {
	e.namespace = namespace
	e.pod = pod
	e.container = container
	return e
}

func (e *defaultRemoteExecutor) WithCommand(cmd []string) contracts.RemoteExecutor {
	e.cmd = cmd
	return e
}

func (e *defaultRemoteExecutor) Execute(clientSet kubernetes.Interface, cfg *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error {
	peo := &v1.PodExecOptions{
		Stdin:     stdin != nil,
		Stdout:    stdout != nil,
		Stderr:    stderr != nil,
		TTY:       tty,
		Container: e.container,
		Command:   e.cmd,
	}

	req := clientSet.CoreV1().
		RESTClient().
		Post().
		Namespace(e.namespace).
		Resource("pods").
		SubResource("exec").
		Name(e.pod)

	exec, err := remotecommand.NewSPDYExecutor(cfg, e.method, req.VersionedParams(peo, scheme.ParameterCodec).URL())
	if err != nil {
		return err
	}

	return exec.Stream(remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            stdout,
		Stderr:            stderr,
		Tty:               tty,
		TerminalSizeQueue: terminalSizeQueue,
	})
}
