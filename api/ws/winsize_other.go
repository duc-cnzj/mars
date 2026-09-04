//go:build !unix

package ws

import "os"

// windowSizeSource 在非 unix 平台返回 nil，AutoHandleWindowSize 表现为 no-op
// （无法感知本地终端尺寸变化）。
func windowSizeSource() (current func() (uint32, uint32, bool), changes <-chan os.Signal, stopSrc func()) {
	return nil, nil, func() {}
}
