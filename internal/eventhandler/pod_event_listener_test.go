package eventhandler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// fakePodPublisher 记录发布调用，供 handle 分支断言。Run 消费链在独立
// goroutine 中并发调用 Publish，字段访问必须加锁，避免与断言 goroutine 竞态。
type fakePodPublisher struct {
	mu     sync.Mutex
	nsIDs  []int64
	err    error
	called int
}

func (f *fakePodPublisher) Publish(nsID int64, pod *corev1.Pod) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.called++
	f.nsIDs = append(f.nsIDs, nsID)
	return f.err
}

// snapshot 返回已发布次数与 nsID 集合的线程安全快照，供并发断言使用。
func (f *fakePodPublisher) snapshot() (int, []int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.called, append([]int64(nil), f.nsIDs...)
}

// newTestPodListener 构造监听者：注入 fake 发布器与真实 mock repo。
func newTestPodListener(m *gomock.Controller) (*PodEventListener, *data.MockK8sRepo, *data.MockNamespaceRepo, *fakePodPublisher) {
	kr := data.NewMockK8sRepo(m)
	nr := data.NewMockNamespaceRepo(m)
	pub := &fakePodPublisher{}
	return &PodEventListener{logger: mlog.NewForConfig(nil), k8sRepo: kr, nsRepo: nr, pub: pub}, kr, nr, pub
}

// TestPodEventListener_Run_StopsOnCtxCancel 覆盖 Run 的常驻语义：
// 消费事件直至 ctx 取消，退出时注销订阅并返回 nil。
func TestPodEventListener_Run_StopsOnCtxCancel(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	l, kr, _, pub := newTestPodListener(m)

	ch := make(chan biz.PodEvent, 1)
	unsubscribed := false
	kr.EXPECT().SubscribePodEvents("pod-watcher").Return(ch, func() { unsubscribed = true })

	// nsRepo 解析 namespace 后发布：验证常驻消费链路。
	nr := data.NewMockNamespaceRepo(m)
	l.nsRepo = nr
	nr.EXPECT().FindByName(gomock.Any(), "ns1").Return(&biz.Namespace{ID: 7}, nil)
	ch <- biz.PodEvent{Type: biz.PodEventAdd, Current: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns1"}}}

	ctx, cancel := context.WithCancel(context.TODO())
	done := make(chan error, 1)
	go func() { done <- l.Run(ctx) }()

	// 事件被消费：等待发布完成（有界等待，经加锁快照读取，避免字段竞态）。
	// 注意 waitFor/tick 必须显式 time.Millisecond：裸整数常量会被当作纳秒，
	// 1µs 后 timer 即触发，Eventually 瞬间失败。
	assert.Eventually(t, func() bool { c, _ := pub.snapshot(); return c == 1 }, time.Second, 10*time.Millisecond)
	_, nsIDs := pub.snapshot()
	assert.Equal(t, int64(7), nsIDs[0])

	cancel()
	assert.NoError(t, <-done)
	assert.True(t, unsubscribed, "ctx 取消后必须注销 informer 订阅")
}

// TestPodEventListener_Handle_UpdatePhaseChanged 覆盖更新事件相位变化分支：
// 新旧相位不同即发布。
func TestPodEventListener_Handle_UpdatePhaseChanged(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	l, _, nr, pub := newTestPodListener(m)

	nr.EXPECT().FindByName(gomock.Any(), "ns1").Return(&biz.Namespace{ID: 7}, nil)
	l.handle(biz.PodEvent{
		Type:    biz.PodEventUpdate,
		Old:     &corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodPending}},
		Current: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns1"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}},
	})
	assert.Equal(t, 1, pub.called)
	assert.Equal(t, int64(7), pub.nsIDs[0])
}

// TestPodEventListener_Handle_UpdateContainerChanged 覆盖更新事件容器就绪变化分支：
// 相位相同但容器 Ready 翻转即发布。
func TestPodEventListener_Handle_UpdateContainerChanged(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	l, _, nr, pub := newTestPodListener(m)

	nr.EXPECT().FindByName(gomock.Any(), "ns1").Return(&biz.Namespace{ID: 7}, nil)
	l.handle(biz.PodEvent{
		Type: biz.PodEventUpdate,
		Old: &corev1.Pod{Status: corev1.PodStatus{
			Phase:             corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{{Name: "app", Ready: false}},
		}},
		Current: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns1"}, Status: corev1.PodStatus{
			Phase:             corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{{Name: "app", Ready: true}},
		}},
	})
	assert.Equal(t, 1, pub.called)
}

// TestPodEventListener_Handle_UpdateUnchangedSkips 覆盖更新事件无变化分支：
// 相位与容器均未变则不发布。
func TestPodEventListener_Handle_UpdateUnchangedSkips(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	l, _, _, pub := newTestPodListener(m)

	// 无 nsRepo EXPECT：若被调用会 panic，反向证明未走发布链。
	l.handle(biz.PodEvent{
		Type:    biz.PodEventUpdate,
		Old:     &corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodRunning, ContainerStatuses: []corev1.ContainerStatus{{Name: "app", Ready: true}}}},
		Current: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns1"}, Status: corev1.PodStatus{Phase: corev1.PodRunning, ContainerStatuses: []corev1.ContainerStatus{{Name: "app", Ready: true}}}},
	})
	assert.Equal(t, 0, pub.called)
}

// TestPodEventListener_Handle_AddPublish 覆盖新增事件分支：直接发布。
func TestPodEventListener_Handle_AddPublish(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	l, _, nr, pub := newTestPodListener(m)

	nr.EXPECT().FindByName(gomock.Any(), "ns1").Return(&biz.Namespace{ID: 7}, nil)
	l.handle(biz.PodEvent{
		Type:    biz.PodEventAdd,
		Current: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns1"}},
	})
	assert.Equal(t, 1, pub.called)
	assert.Equal(t, int64(7), pub.nsIDs[0])
}

// TestPodEventListener_Handle_DeletePublish 覆盖删除事件分支：直接发布。
func TestPodEventListener_Handle_DeletePublish(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	l, _, nr, pub := newTestPodListener(m)

	nr.EXPECT().FindByName(gomock.Any(), "ns1").Return(&biz.Namespace{ID: 7}, nil)
	l.handle(biz.PodEvent{
		Type:    biz.PodEventDelete,
		Current: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns1"}},
	})
	assert.Equal(t, 1, pub.called)
}

// TestPodEventListener_Handle_PublishErrorLogsOnly 覆盖发布失败分支：
// 错误只打日志不 panic。
func TestPodEventListener_Handle_PublishErrorLogsOnly(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	l, _, nr, pub := newTestPodListener(m)
	pub.err = assert.AnError

	nr.EXPECT().FindByName(gomock.Any(), "ns1").Return(&biz.Namespace{ID: 7}, nil)
	assert.NotPanics(t, func() {
		l.handle(biz.PodEvent{
			Type:    biz.PodEventAdd,
			Current: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns1"}},
		})
	})
	assert.Equal(t, 1, pub.called)
}

// TestPodEventListener_Handle_NamespaceNotFoundSkips 覆盖 namespace 解析失败分支：
// 不发布、不中断。
func TestPodEventListener_Handle_NamespaceNotFoundSkips(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	l, _, nr, pub := newTestPodListener(m)

	nr.EXPECT().FindByName(gomock.Any(), "ns1").Return(nil, assert.AnError)
	l.handle(biz.PodEvent{
		Type:    biz.PodEventAdd,
		Current: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns1"}},
	})
	assert.Equal(t, 0, pub.called)
}

// TestPodEventListener_Handle_UnknownTypeNoop 覆盖未知事件类型分支：no-op。
func TestPodEventListener_Handle_UnknownTypeNoop(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	l, _, _, pub := newTestPodListener(m)

	l.handle(biz.PodEvent{Type: biz.PodEventType(99), Current: &corev1.Pod{}})
	assert.Equal(t, 0, pub.called)
}

// TestPodEventListener_NewPodEventListener 验证构造器注入模块化 logger。
func TestPodEventListener_NewPodEventListener(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	l := NewPodEventListener(mlog.NewForConfig(nil), data.NewMockK8sRepo(m), data.NewMockNamespaceRepo(m), &fakePodPublisher{})
	assert.NotNil(t, l)
	assert.NotNil(t, l.logger)
}

// TestContainerStatusChanged 覆盖容器状态对比各分支：
// 数量不一致、Ready 翻转、Ready 一致、容器集合变化均为 true，完全一致为 false。
func TestContainerStatusChanged(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	status := func(ready ...bool) []corev1.ContainerStatus {
		var out []corev1.ContainerStatus
		for i, r := range ready {
			out = append(out, corev1.ContainerStatus{Name: string(rune('a' + i)), Ready: r})
		}
		return out
	}

	t.Run("容器数量不一致返回 true", func(t *testing.T) {
		old := &corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: status(true)}}
		cur := &corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: status(true, true)}}
		assert.True(t, containerStatusChanged(logger, old, cur))
	})

	t.Run("Ready 翻转返回 true", func(t *testing.T) {
		old := &corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: status(false, true)}}
		cur := &corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: status(true, true)}}
		assert.True(t, containerStatusChanged(logger, old, cur))
	})

	t.Run("容器集合变化返回 true", func(t *testing.T) {
		old := &corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: status(true, true)}}
		cur := &corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: status(true, true, true)}}
		assert.True(t, containerStatusChanged(logger, old, cur))
	})

	t.Run("完全一致返回 false", func(t *testing.T) {
		old := &corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: status(true, true)}}
		cur := &corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: status(true, true)}}
		assert.False(t, containerStatusChanged(logger, old, cur))
	})

	t.Run("新容器不在旧集合返回 true", func(t *testing.T) {
		old := &corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: status(true)}}
		cur := &corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{Name: "b", Ready: true}, {Name: "a", Ready: true}}}}
		assert.True(t, containerStatusChanged(logger, old, cur))
	})
}

// TestPodEventListener_Run_StopsOnChannelClose 覆盖 Run 消费 channel 关闭分支：
// informer 订阅 channel 被 close 时记录警告并返回 nil，不再阻塞消费。
func TestPodEventListener_Run_StopsOnChannelClose(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	l, kr, _, pub := newTestPodListener(m)

	ch := make(chan biz.PodEvent)
	unsubscribed := false
	kr.EXPECT().SubscribePodEvents("pod-watcher").Return(ch, func() { unsubscribed = true })
	close(ch)

	assert.NoError(t, l.Run(context.TODO()))
	assert.True(t, unsubscribed, "channel 关闭退出后必须注销 informer 订阅")
	assert.Equal(t, 0, pub.called)
}

// TestPodEventListener_Handle_UpdatePublishErrorLogsOnly 覆盖更新事件发布失败分支：
// 相位变化后 UPDATE 路径发布失败只打日志不 panic（与 ADD/DELETE 路径互为镜像）。
func TestPodEventListener_Handle_UpdatePublishErrorLogsOnly(t *testing.T) {
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	l, _, nr, pub := newTestPodListener(m)
	pub.err = assert.AnError

	nr.EXPECT().FindByName(gomock.Any(), "ns1").Return(&biz.Namespace{ID: 7}, nil)
	assert.NotPanics(t, func() {
		l.handle(biz.PodEvent{
			Type:    biz.PodEventUpdate,
			Old:     &corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodPending}},
			Current: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p1", Namespace: "ns1"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}},
		})
	})
	assert.Equal(t, 1, pub.called)
}

// TestPodEventPublisher_var 编译期断言窄接口可被 fake 实现。
var _ PodEventPublisher = (*fakePodPublisher)(nil)
