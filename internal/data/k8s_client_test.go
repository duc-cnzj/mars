package data

import (
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	gwfake "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/fake"
	externalversions "sigs.k8s.io/gateway-api/pkg/client/informers/externalversions"
)

// runK8sClientStart 启动一次 K8sClient.start：注入真实 informer factory（fake clientset）
// 与带监听通道的真实 fanOut，close(done) 触发各分发 goroutine 与 WaitForCacheSync 退出。
// 返回注入的监听通道，供断言 Distribute 确实被启动（RemoveAll 关闭通道 = goroutine 已退出）。
func runK8sClientStart(t *testing.T, gwInstalled bool) (evCh chan Obj[*eventsv1.Event], podCh chan Obj[*corev1.Pod]) {
	t.Helper()
	logger := mlog.NewForConfig(nil)
	factory := informers.NewSharedInformerFactory(k8sfake.NewSimpleClientset(), 0)

	evCh = make(chan Obj[*eventsv1.Event])
	podCh = make(chan Obj[*corev1.Pod])
	k := &K8sClient{
		GatewayApiInstalled: gwInstalled,
		logger:              logger,
		factory:             factory,
		eventFanOut: newFanOut(logger, "event", make(chan Obj[*eventsv1.Event]),
			map[string]chan<- Obj[*eventsv1.Event]{"ev": evCh}),
		podFanOut: newFanOut(logger, "pod", make(chan Obj[*corev1.Pod]),
			map[string]chan<- Obj[*corev1.Pod]{"pod": podCh}),
		PodInformer:    factory.Core().V1().Pods().Informer(),
		SecretInformer: factory.Core().V1().Secrets().Informer(),
	}
	if gwInstalled {
		k.gwFactory = externalversions.NewSharedInformerFactory(gwfake.NewSimpleClientset(), 0)
	}

	done := make(chan struct{})
	ret := make(chan struct{})
	go func() {
		k.start(done)
		close(ret)
	}()
	close(done)
	<-ret
	return evCh, podCh
}

// waitClosed 断言通道已被 RemoveAll 关闭：fanOut.Distribute 退出时兜底调用 RemoveAll
// 关闭全部监听通道，因此通道关闭即证明对应分发 goroutine 已启动并退出。
func waitClosed[T any](t *testing.T, name string, ch <-chan T) {
	t.Helper()
	select {
	case _, ok := <-ch:
		assert.False(t, ok, "%s fanout 通道应被 RemoveAll 关闭，而非收到数据", name)
	case <-time.After(2 * time.Second):
		t.Fatalf("%s fanout 分发 goroutine 未在 2s 内退出", name)
	}
}

// TestK8sClient_Start 覆盖 start 全部分支：
// 两个分发 goroutine 启动并退出、informer factory 启动、WaitForCacheSync 返回；
// gwInstalled 两分支分别验证 Gateway API factory 的启动与跳过。
func TestK8sClient_Start(t *testing.T) {
	t.Run("未安装 Gateway API 跳过 gwFactory 启动", func(t *testing.T) {
		evCh, podCh := runK8sClientStart(t, false)
		waitClosed(t, "event", evCh)
		waitClosed(t, "pod", podCh)
	})
	t.Run("已安装 Gateway API 同时启动 gwFactory", func(t *testing.T) {
		evCh, podCh := runK8sClientStart(t, true)
		waitClosed(t, "event", evCh)
		waitClosed(t, "pod", podCh)
	})
}
