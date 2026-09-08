//go:build unix

package ws

import (
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

// winsize 承载终端窗口尺寸（ioctl 用）。
type winsize struct {
	Row, Col, Xpixel, Ypixel uint16
}

// windowSizeSource 提供平台相关的本地终端尺寸能力（unix 用 ioctl + SIGWINCH）：
//   - current 返回当前窗口尺寸，读取失败（如非 tty）ok=false；
//   - changes 在窗口尺寸变化时触发（SIGWINCH 信号）；
//   - stopSrc 停止尺寸变化监听（signal.Stop）。
func windowSizeSource() (current func() (uint32, uint32, bool), changes <-chan os.Signal, stopSrc func()) {
	changesCh := make(chan os.Signal, 1)
	signal.Notify(changesCh, syscall.SIGWINCH)
	current = func() (uint32, uint32, bool) {
		sz := &winsize{}
		_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(),
			uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(sz))) // #nosec G103 -- 标准 ioctl
		if errno != 0 {
			return 0, 0, false
		}
		// 注：ioctl 成功分支为 tty 集成边界（单测跑在非 tty 环境，os.Stdout 非终端，
		// errno 必非 0），故下方真实读取不覆盖，由 windowSizeSource 的 fake 注入分支测试兜底。
		return uint32(sz.Row), uint32(sz.Col), true
	}
	return current, changesCh, func() { signal.Stop(changesCh) }
}
