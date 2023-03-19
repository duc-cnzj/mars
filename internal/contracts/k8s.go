package contracts

//go:generate mockgen -destination ../mock/mock_remote_executor.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts RemoteExecutor
//go:generate mockgen -destination ../mock/mock_pod_copier.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts PodFileCopier
//go:generate mockgen -destination ../mock/mock_archiver.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts Archiver

import (
	"context"
	"io"
	"net/url"

	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/listers/apps/v1"
	v1 "k8s.io/client-go/listers/core/v1"
	networkingv1 "k8s.io/client-go/listers/networking/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

type FanOutType int

const (
	Add FanOutType = iota
	Update
	Delete
)

type FanOutInterface[T runtime.Object] interface {
	RemoveListener(key string)
	AddListener(key string, ch chan<- Obj[T])
	Distribute(done <-chan struct{})
}

type Obj[T runtime.Object] interface {
	Type() FanOutType
	Old() T
	Current() T
}

type obj[T runtime.Object] struct {
	old, current T
	t            FanOutType
}

func NewObj[T runtime.Object](old T, current T, t FanOutType) Obj[T] {
	return &obj[T]{old: old, current: current, t: t}
}

func (o *obj[T]) Type() FanOutType {
	return o.t
}

func (o *obj[T]) Old() T {
	return o.old
}

func (o *obj[T]) Current() T {
	return o.current
}

type K8sClient struct {
	Client        kubernetes.Interface
	MetricsClient versioned.Interface
	RestConfig    *restclient.Config

	PodInformer cache.SharedIndexInformer
	PodLister   v1.PodLister

	SecretInformer cache.SharedIndexInformer
	SecretLister   v1.SecretLister

	ReplicaSetLister appsv1.ReplicaSetLister
	ServiceLister    v1.ServiceLister
	IngressLister    networkingv1.IngressLister

	EventInformer cache.SharedIndexInformer

	EventFanOut FanOutInterface[*eventsv1.Event]
	PodFanOut   FanOutInterface[*corev1.Pod]
}

type ExecUrlBuilder interface {
	URL() *url.URL
}

type ExecRequestBuilder interface {
	BuildExecRequest(namespace, pod string, peo *corev1.PodExecOptions) ExecUrlBuilder
}

type RemoteExecutor interface {
	WithMethod(method string) RemoteExecutor
	WithContainer(namespace, pod, container string) RemoteExecutor
	WithCommand(cmd []string) RemoteExecutor
	Execute(ctx context.Context, clientSet kubernetes.Interface, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error
}

type CopyFileToPodResult struct {
	TargetDir     string
	ErrOut        string
	StdOut        string
	ContainerPath string
	FileName      string
}

type PodFileCopier interface {
	Copy(namespace, pod, container, fpath, targetContainerDir string, clientSet kubernetes.Interface, config *restclient.Config) (*CopyFileToPodResult, error)
}

// Archiver 不是很合理，但暂时这样
type Archiver interface {
	Archive(sources []string, destination string) error
	Open(path string) (io.ReadCloser, error)
	Remove(path string) error
}
