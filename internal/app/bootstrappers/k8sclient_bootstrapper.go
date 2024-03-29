package bootstrappers

import (
	"fmt"
	"strings"
	"sync"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/metrics"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/utils"
	"github.com/duc-cnzj/mars/v4/internal/utils/recovery"

	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	runtime2 "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type K8sClientBootstrapper struct{}

func (k *K8sClientBootstrapper) Tags() []string {
	return []string{}
}

func (k *K8sClientBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	var (
		config   *restclient.Config
		err      error
		nsPrefix = app.Config().NsPrefix

		eventFanOutObj = &fanOut[*eventsv1.Event]{
			name:      "event",
			ch:        make(chan contracts.Obj[*eventsv1.Event], 100),
			listeners: make(map[string]chan<- contracts.Obj[*eventsv1.Event]),
		}

		podFanOutObj = &fanOut[*corev1.Pod]{
			name:      "pod",
			ch:        make(chan contracts.Obj[*corev1.Pod], 100),
			listeners: make(map[string]chan<- contracts.Obj[*corev1.Pod]),
		}
	)

	go func() {
		defer recovery.HandlePanic(fmt.Sprintf("[FANOUT]: '%s' Distribute", eventFanOutObj.name))

		eventFanOutObj.Distribute(app.Done())
	}()
	go func() {
		defer recovery.HandlePanic(fmt.Sprintf("[FANOUT]: '%s' Distribute", podFanOutObj.name))

		podFanOutObj.Distribute(app.Done())
	}()

	runtime.ErrorHandlers = []func(err error){
		func(err error) {
			mlog.Warning(err)
		},
	}

	if app.Config().KubeConfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", app.Config().KubeConfig)
		if err != nil {
			return err
		}
	} else {
		config, err = restclient.InClusterConfig()
		if err != nil {
			return err
		}
	}

	// 客户端不限速，有可能会把集群打死。
	config.QPS = -1

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	metrics, err := metricsv.NewForConfig(config)
	if err != nil {
		return err
	}

	inf := informers.NewSharedInformerFactory(clientset, 0)
	svcLister := inf.Core().V1().Services().Lister()
	ingLister := inf.Networking().V1().Ingresses().Lister()
	rsLister := inf.Apps().V1().ReplicaSets().Lister()
	podInf := inf.Core().V1().Pods().Informer()
	podLister := inf.Core().V1().Pods().Lister()
	secretInf := inf.Core().V1().Secrets().Informer()
	secretLister := inf.Core().V1().Secrets().Lister()
	podInf.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: filterPod(nsPrefix),
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				podFanOutObj.ch <- contracts.NewObj[*corev1.Pod](nil, obj.(*corev1.Pod), contracts.Add)
			},
			UpdateFunc: func(oldObj, newObj any) {
				old := oldObj.(*corev1.Pod)
				curr := newObj.(*corev1.Pod)
				if old.ResourceVersion != curr.ResourceVersion {
					select {
					case podFanOutObj.ch <- contracts.NewObj[*corev1.Pod](old, curr, contracts.Update):
					default:
						mlog.Warningf("[INFORMER]: podFanOutObj full")
					}
				}
			},
			DeleteFunc: func(obj any) {
				podFanOutObj.ch <- contracts.NewObj[*corev1.Pod](nil, obj.(*corev1.Pod), contracts.Delete)
			},
		},
	})
	eventInf := inf.Events().V1().Events().Informer()
	eventInf.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: filterEvent(nsPrefix),
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(current any) {
				event := current.(*eventsv1.Event)
				select {
				case eventFanOutObj.ch <- contracts.NewObj[*eventsv1.Event](nil, event, contracts.Add):
				default:
					mlog.Warningf("[INFORMER]: eventFanOutObj full")
				}
			},
		},
	})
	eventLister := inf.Events().V1().Events().Lister()
	inf.Start(app.Done())
	cache.WaitForCacheSync(nil, eventInf.HasSynced, podInf.HasSynced, secretInf.HasSynced)

	app.SetK8sClient(&contracts.K8sClient{
		Client:           clientset,
		MetricsClient:    metrics,
		RestConfig:       config,
		PodInformer:      podInf,
		PodLister:        podLister,
		SecretInformer:   secretInf,
		SecretLister:     secretLister,
		ReplicaSetLister: rsLister,
		ServiceLister:    svcLister,
		IngressLister:    ingLister,
		EventInformer:    eventInf,
		EventLister:      eventLister,
		EventFanOut:      eventFanOutObj,
		PodFanOut:        podFanOutObj,
	})

	return nil
}

func filterEvent(nsPrefix string) func(obj any) bool {
	return func(obj any) bool {
		e, ok := obj.(*eventsv1.Event)
		if !ok {
			return false
		}
		return strings.HasPrefix(e.Namespace, nsPrefix) && e.Regarding.Kind == "Pod" && e.Reason != "Unhealthy"
	}
}

func filterPod(nsPrefix string) func(obj any) bool {
	return func(obj any) bool {
		pod, ok := obj.(*corev1.Pod)
		if !ok {
			return false
		}
		return strings.HasPrefix(pod.Namespace, nsPrefix)
	}
}

type Startable struct {
	c utils.Closeable
}

func (s *Startable) start() bool {
	return s.c.Close()
}

type fanOut[T runtime2.Object] struct {
	name string
	ch   chan contracts.Obj[T]

	started Startable

	listenerMu sync.Mutex
	listeners  map[string]chan<- contracts.Obj[T]
}

func (f *fanOut[T]) AddListener(key string, ch chan<- contracts.Obj[T]) {
	f.listenerMu.Lock()
	defer f.listenerMu.Unlock()
	_, ok := f.listeners[key]
	if ok {
		mlog.Warningf("[FANOUT]: FanOut already exists %s", key)
		return
	}
	mlog.Infof("%s add fanOut listener: %v", f.name, key)
	metrics.K8sInformerFanOutListenerCount.With(prometheus.Labels{"type": f.name}).Inc()
	f.listeners[key] = ch
}

func (f *fanOut[T]) RemoveListener(key string) {
	f.listenerMu.Lock()
	defer f.listenerMu.Unlock()
	mlog.Infof("[FANOUT]: remove listener %s", key)
	_, ok := f.listeners[key]
	if ok {
		delete(f.listeners, key)
		metrics.K8sInformerFanOutListenerCount.With(prometheus.Labels{"type": f.name}).Dec()
	}
}

func (f *fanOut[T]) Distribute(done <-chan struct{}) {
	defer mlog.Debug(fmt.Sprintf("[FANOUT]: '%s' Exit", f.name))
	if !f.started.start() {
		return
	}
	mlog.Infof("[FANOUT]: '%s' start", f.name)
	for {
		select {
		case <-done:
			mlog.Infof("[FANOUT]: '%s' exited!", f.name)
			return
		case obj, ok := <-f.ch:
			if !ok {
				mlog.Warningf("[FANOUT]: '%s' Exit!", f.name)
				return
			}
			func() {
				f.listenerMu.Lock()
				defer f.listenerMu.Unlock()
				for k, s := range f.listeners {
					select {
					case s <- obj:
					default:
						mlog.Warningf("[FANOUT]: '%s' drop %s %v", f.name, k, obj)
					}
				}
			}()
		}
	}
}
