package data

import (
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newPodFan 构造一个空的 pod fanOut 测试对象（经 newFanOut 装配，与生产 InitK8s 一致）。
func newPodFan(input chan Obj[*corev1.Pod]) *fanOut[*corev1.Pod] {
	return newFanOut[*corev1.Pod](mlog.NewForConfig(nil), "pod", input, map[string]chan<- Obj[*corev1.Pod]{}).(*fanOut[*corev1.Pod])
}

// TestFanOut_AddListener_DuplicateKey 覆盖重复 key 注册的告警分支（已存在则忽略新 listener）。
func TestFanOut_AddListener_DuplicateKey(t *testing.T) {
	fan := newPodFan(make(chan Obj[*corev1.Pod]))
	ch1 := make(chan Obj[*corev1.Pod], 1)
	ch2 := make(chan Obj[*corev1.Pod], 1)
	fan.AddListener("k", ch1)
	fan.AddListener("k", ch2) // 已存在 → Warningf 分支
	assert.Len(t, fan.listeners, 1)
	// map 中保存的仍是 ch1：向其中发送、从 ch1 收到即证同一通道
	select {
	case fan.listeners["k"] <- nil:
	default:
		t.Fatal("listener channel not writable")
	}
	<-ch1
}

// TestFanOut_RemoveAll 覆盖批量关闭：所有 listener 通道被 close 并从 map 移除。
func TestFanOut_RemoveAll(t *testing.T) {
	fan := newPodFan(make(chan Obj[*corev1.Pod]))
	a := make(chan Obj[*corev1.Pod])
	b := make(chan Obj[*corev1.Pod])
	fan.AddListener("a", a)
	fan.AddListener("b", b)

	fan.RemoveAll()
	assert.Empty(t, fan.listeners)
	_, aOK := <-a
	_, bOK := <-b
	assert.False(t, aOK)
	assert.False(t, bOK)
}

// TestFanOut_Distribute_StartedGuard 覆盖已启动守卫：closeable 原子 CAS 在
// 已关闭状态下返回 false → Distribute 早退。预置 started 状态避免并发竞态。
func TestFanOut_Distribute_StartedGuard(t *testing.T) {
	fan := newPodFan(make(chan Obj[*corev1.Pod], 1))
	fan.started.c.Close() // 预置已启动 → start() 返回 false
	fan.Distribute(make(chan struct{}))
	// 不 panic、不阻塞即通过
}

// TestFanOut_Distribute_ChannelClosed 覆盖输入通道关闭 → 分发 goroutine 退出。
func TestFanOut_Distribute_ChannelClosed(t *testing.T) {
	input := make(chan Obj[*corev1.Pod])
	fan := newPodFan(input)
	done := make(chan struct{})
	exited := make(chan struct{})
	go func() {
		fan.Distribute(done)
		close(exited)
	}()
	close(input) // f.ch 关闭 → !ok 分支退出
	<-exited
}

// TestFanOut_Distribute_DropFullListener 覆盖分发时的丢弃分支：listener 无缓冲
// 且无接收方 → select 命中 default（drop）。靠 close(input) 保证 obj 先被处理、
// 再因通道关闭退出，且 <-exited 同步等待，避免竞态。
func TestFanOut_Distribute_DropFullListener(t *testing.T) {
	input := make(chan Obj[*corev1.Pod], 1)
	fan := newPodFan(input)
	listener := make(chan Obj[*corev1.Pod]) // 无缓冲、无接收 → 必 drop
	fan.AddListener("l", listener)

	exited := make(chan struct{})
	go func() {
		fan.Distribute(make(chan struct{}))
		close(exited)
	}()

	input <- newObj[*corev1.Pod](nil, &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p"}}, Add)
	close(input) // 处理完 obj（命中 drop）后经 !ok 分支退出
	<-exited
}
