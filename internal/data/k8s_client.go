package data

import (
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/listers/apps/v1"
	v1 "k8s.io/client-go/listers/core/v1"
	eventsv1lister "k8s.io/client-go/listers/events/v1"
	networkingv1 "k8s.io/client-go/listers/networking/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	"sigs.k8s.io/gateway-api/pkg/client/informers/externalversions"
	gatewaylisterv1 "sigs.k8s.io/gateway-api/pkg/client/listers/apis/v1"
)

// K8sClient 是 k8s 访问客户端容器：持有 clientset/metricsClient/restConfig 与
// informer/lister 及事件扇出，InitK8s 装配后供 k8sRepo/helmRepo 读取。
type K8sClient struct {
	GatewayApiInstalled bool
	HTTPRouteLister     gatewaylisterv1.HTTPRouteLister
	gwFactory           externalversions.SharedInformerFactory

	logger        mlog.Logger
	factory       informers.SharedInformerFactory
	Client        kubernetes.Interface
	MetricsClient versioned.Interface
	RestConfig    *restclient.Config

	PodInformer cache.SharedIndexInformer
	PodLister   v1.PodLister

	SecretInformer cache.SharedIndexInformer
	SecretLister   v1.SecretLister

	ReplicaSetLister  appsv1.ReplicaSetLister
	DeploymentLister  appsv1.DeploymentLister
	StatefulSetLister appsv1.StatefulSetLister
	DaemonSetLister   appsv1.DaemonSetLister
	ServiceLister     v1.ServiceLister
	IngressLister     networkingv1.IngressLister

	eventFanOut fanOutInterface[*eventsv1.Event]
	podFanOut   fanOutInterface[*corev1.Pod]
	EventLister eventsv1lister.EventLister
}

// start 启动事件扇出分发与 informer 工厂，并等待 Pod/Secret informer 首次同步。
// 由 InitK8s 在装配完成后调用，done 关闭即触发各分发循环退出。
func (k *K8sClient) start(done <-chan struct{}) {
	go func() {
		defer k.logger.HandlePanic("[FANOUT]: event Distribute")

		k.eventFanOut.Distribute(done)
	}()
	go func() {
		defer k.logger.HandlePanic("[FANOUT]: pod Distribute")

		k.podFanOut.Distribute(done)
	}()
	k.factory.Start(done)
	if k.GatewayApiInstalled {
		k.gwFactory.Start(done)
	}
	cache.WaitForCacheSync(done, k.PodInformer.HasSynced, k.SecretInformer.HasSynced)
}
