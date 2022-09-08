package contracts

import (
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
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

	EventInformer cache.SharedIndexInformer

	EventFanOut FanOutInterface[*eventsv1.Event]
	PodFanOut   FanOutInterface[*corev1.Pod]
}
