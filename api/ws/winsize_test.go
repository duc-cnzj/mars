package ws

// winsize_test.go：验证 AutoHandleWindowSize 的平台无关内核 autoHandleWindowSize——
// 初始尺寸同步、变化触发再同步、stop 幂等停止监听；注入 fake 的尺寸源与通知通道，
// 不依赖真实 tty。所有共享状态经 channel / atomic 同步，保证 -race 干净。

import (
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestAutoHandleWindowSize(t *testing.T) {
	resizeCh := make(chan [2]uint32, 8)
	resize := func(h, w uint32) { resizeCh <- [2]uint32{h, w} }
	var size atomic.Uint32
	size.Store(80)
	current := func() (uint32, uint32, bool) { return size.Load(), 24, true }
	changes := make(chan os.Signal, 1)
	stopped := make(chan struct{})
	stopSrc := func() { close(stopped) }

	stop := autoHandleWindowSize(current, changes, stopSrc, resize)

	// 初始尺寸立即同步。
	select {
	case r := <-resizeCh:
		if r != [2]uint32{80, 24} {
			t.Fatalf("初始尺寸错误，期望 [80 24]，实际 %v", r)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("初始尺寸未同步")
	}

	// 尺寸变化 → 再同步一次。
	size.Store(120)
	changes <- syscall.SIGWINCH
	select {
	case r := <-resizeCh:
		if r != [2]uint32{120, 24} {
			t.Fatalf("变化后尺寸错误，期望 [120 24]，实际 %v", r)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("尺寸变化未触发再同步")
	}

	// stop 幂等 → stopSrc 恰好调用一次。
	stop()
	stop()
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("stop 后 stopSrc 未被调用")
	}
}

func TestAutoHandleWindowSize_NoTTY(t *testing.T) {
	resizeCh := make(chan [2]uint32, 1)
	resize := func(h, w uint32) { resizeCh <- [2]uint32{h, w} }
	current := func() (uint32, uint32, bool) { return 0, 0, false } // 读不到尺寸
	changes := make(chan os.Signal, 1)
	stopped := make(chan struct{})
	stopSrc := func() { close(stopped) }

	stop := autoHandleWindowSize(current, changes, stopSrc, resize)

	// 初始读不到尺寸 → 不 resize；变化时仍读不到 → 不 resize。
	changes <- syscall.SIGWINCH
	select {
	case <-resizeCh:
		t.Fatal("非 tty 不应 resize")
	case <-time.After(50 * time.Millisecond):
		// 正确：读不到尺寸，不 resize。
	}

	stop()
	stop()
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("stop 后 stopSrc 未被调用")
	}
}

func TestAutoHandleWindowSize_Nil(t *testing.T) {
	// 非 unix 平台 current/changes 为 nil → 返回 no-op stop，不 panic。
	stop := autoHandleWindowSize(nil, nil, func() {}, func(uint32, uint32) {})
	stop()
	stop()
}

// TestTerminal_AutoHandleWindowSize 冒烟：真实走平台 windowSizeSource（unix 下
// 注册 SIGWINCH；非 tty 时 current 返回 false 不 resize），stop 后退出不 panic。
func TestTerminal_AutoHandleWindowSize(t *testing.T) {
	term := &Terminal{}
	stop := term.AutoHandleWindowSize()
	stop()
	stop()
}
