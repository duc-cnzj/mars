package data

import (
	"fmt"
	"strings"
	"sync"

	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/closeable"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	runtime2 "k8s.io/apimachinery/pkg/runtime"
)

// fanOutType 是事件分发类型：informer 回调把增/改/删事件标注后推给监听者。
type fanOutType int

// Add/Update/Delete 是事件变更类型：informer 回调据此标注增/改/删。
const (
	Add fanOutType = iota
	Update
	Delete
)

// filterEvent 构造事件 informer 的过滤函数：只放行目标前缀命名空间下、kind 为 Pod、
// 且 reason 非 "Unhealthy" 的事件（避开无意义的健康检查噪音）。
func filterEvent(nsPrefix string) func(obj any) bool {
	return func(obj any) bool {
		e, ok := obj.(*eventsv1.Event)
		if !ok {
			return false
		}
		return strings.HasPrefix(e.Namespace, nsPrefix) && e.Regarding.Kind == "Pod" && e.Reason != "Unhealthy"
	}
}

// filterPod 构造 Pod informer 的过滤函数：只放行目标前缀命名空间下的 Pod。
func filterPod(nsPrefix string) func(obj any) bool {
	return func(obj any) bool {
		pod, ok := obj.(*corev1.Pod)
		if !ok {
			return false
		}
		return strings.HasPrefix(pod.Namespace, nsPrefix)
	}
}

// startable 是 fanOut 的启动闸门：基于 closeable 保证 Distribute 只真正启动一次。
type startable struct {
	c closeable.Closeable
}

// start 尝试关闭底层 closeable：仅首次调用返回 true（后续调用返回 false 视为已启动）。
func (s *startable) start() bool {
	return s.c.Close()
}

// fanOutInterface 是事件扇出分发端口：监听者按 key 注册/注销，Distribute 消费通道广播。
type fanOutInterface[T runtime2.Object] interface {
	// RemoveListener 注销监听者并关闭其通道。
	RemoveListener(key string)
	// AddListener 注册监听者，接收事件变更对象。
	AddListener(key string, ch chan Obj[T])
	// Distribute 消费事件通道并广播给所有监听者。
	Distribute(done <-chan struct{})
	// RemoveAll 注销全部监听者并关闭所有通道。
	RemoveAll()
}

// Obj 是一次事件变更的封装：携带类型（增/改/删）与新旧对象。
type Obj[T runtime2.Object] interface {
	// Type 返回变更类型（增/改/删）。
	Type() fanOutType
	// Old 返回变更前对象（新增场景为 nil）。
	Old() T
	// Current 返回变更后对象。
	Current() T
}

// fanOut 是 fanOutInterface 的通道扇出实现：单消费通道 + 多监听者广播，丢包时记告警。
type fanOut[T runtime2.Object] struct {
	name string
	ch   chan Obj[T]

	started startable
	logger  mlog.Logger

	listenerMu sync.Mutex
	listeners  map[string]chan<- Obj[T]
}

// newFanOut 构造通道扇出分发器：channel 由外部填充，listeners 按 key 挂接。
// InitK8s 与测试均经它装配 fanOut——回调闭包只写外部持有的 channel、经构造器注入，
// 不直接摸 fanOut 私有字段，保证通道所有权清晰。
func newFanOut[T runtime2.Object](
	logger mlog.Logger,
	name string,
	ch chan Obj[T],
	listeners map[string]chan<- Obj[T],
) fanOutInterface[T] {
	return &fanOut[T]{name: name, ch: ch, logger: logger, listeners: listeners}
}

// AddListener 注册监听者；同 key 重复注册时告警并忽略，避免覆盖已有通道。
func (f *fanOut[T]) AddListener(key string, ch chan Obj[T]) {
	f.listenerMu.Lock()
	defer f.listenerMu.Unlock()
	_, ok := f.listeners[key]
	if ok {
		f.logger.Warningf("[FANOUT]: FanOut already exists %s", key)
		return
	}
	f.logger.Infof("%s add fanOut listener: %v", f.name, key)
	metrics.K8sInformerFanOutListenerCount.With(prometheus.Labels{"type": f.name}).Inc()
	f.listeners[key] = ch
}

// RemoveListener 注销监听者并关闭其通道，同步递减监听数指标。
func (f *fanOut[T]) RemoveListener(key string) {
	f.listenerMu.Lock()
	defer f.listenerMu.Unlock()
	f.logger.Infof("[FANOUT]: remove listener %s", key)
	ch, ok := f.listeners[key]
	if ok {
		delete(f.listeners, key)
		close(ch)
		metrics.K8sInformerFanOutListenerCount.With(prometheus.Labels{"type": f.name}).Dec()
	}
}

// Distribute 消费事件通道并广播给所有监听者：done 关闭或通道关闭时退出；
// 监听者缓冲满则丢弃该事件并记告警（不阻塞分发循环）。
func (f *fanOut[T]) Distribute(done <-chan struct{}) {
	defer f.logger.Debug(fmt.Sprintf("[FANOUT]: '%s' Exit", f.name))
	if !f.started.start() {
		return
	}
	defer f.RemoveAll()
	f.logger.Infof("[FANOUT]: '%s' start", f.name)
	for {
		select {
		case <-done:
			f.logger.Infof("[FANOUT]: '%s' exited!", f.name)
			return
		case obj, ok := <-f.ch:
			if !ok {
				f.logger.Warningf("[FANOUT]: '%s' Exit!", f.name)
				return
			}
			metrics.FanOutChannelLength.With(prometheus.Labels{"name": f.name}).Set(float64(len(f.ch)))
			func() {
				f.listenerMu.Lock()
				defer f.listenerMu.Unlock()
				for k, s := range f.listeners {
					select {
					case s <- obj:
					default:
						f.logger.Warningf("[FANOUT]: '%s' drop %s %v", f.name, k, obj)
					}
				}
			}()
		}
	}
}

// RemoveAll 注销全部监听者并关闭所有通道（Distribute 退出时兜底清理）。
func (f *fanOut[T]) RemoveAll() {
	f.listenerMu.Lock()
	defer f.listenerMu.Unlock()
	for k, s := range f.listeners {
		close(s)
		delete(f.listeners, k)
	}
}

// obj 是 Obj 的实现：携带变更类型与新旧对象。
type obj[T runtime2.Object] struct {
	old, current T
	t            fanOutType
}

// newObj 构造事件变更对象。
func newObj[T runtime2.Object](old T, current T, t fanOutType) Obj[T] {
	return &obj[T]{old: old, current: current, t: t}
}

// sendOrDrop 向事件通道投递事件：通道满时丢弃并记告警，避免 informer 回调阻塞。
// InitK8s 的 Pod/Event 回调共用它，收敛四段重复的 select-default。
func sendOrDrop[T runtime2.Object](ch chan<- Obj[T], obj Obj[T], logger mlog.Logger, tag string) {
	select {
	case ch <- obj:
	default:
		logger.Warningf("[INFORMER]: %s full", tag)
	}
}

// Type 返回变更类型（增/改/删）。
func (o *obj[T]) Type() fanOutType {
	return o.t
}

// Old 返回变更前对象（新增场景为 nil）。
func (o *obj[T]) Old() T {
	return o.old
}

// Current 返回变更后对象。
func (o *obj[T]) Current() T {
	return o.current
}
