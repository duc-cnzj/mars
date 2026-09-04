// Package main 演示 mars WebSocket SDK（github.com/duc-cnzj/mars/api/v6/ws）的
// 终端能力，调用方式刻意压到最简：一条 OpenTerminal 就完成
// 「连接→鉴权→打开容器终端」全部交互，sessionID 由 SDK 自动生成。
//
// 之后只需要：
//   - term.Pump(in, stdout, toast)        一次性接好 stdin→远端、远端→stdout/toast 三通道；
//   - term.AutoHandleWindowSize()        自动跟随本地窗口尺寸变化同步远端 pty；
//   - term.Write(p)                      把本地击键/粘贴当 stdin 转发给容器；
//   - term.Stdout()                      接收容器输出；
//   - term.Done()                        会话结束信号。
//
// 本 demo 默认以 raw 模式运行（-raw=false 可切回 canonical）：
//   - raw 模式下关掉本地行缓冲与回显，每个按键字节即时透传给远端 shell，
//     由远端 readline 解释——tab 补全、方向键、clear 等交互得以生效；
//   - Ctrl+C 在 raw 模式下只是一个字节 0x03，透传给远端 shell（由远端发 SIGINT），
//     所以退出请用远端 `exit` 命令，而非本地 Ctrl+C；
//   - 退出时自动恢复本地终端设置。
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/api/v6/ws"
)

func main() {
	var (
		wsURL = "ws://127.0.0.1:4000/ws"
		ns    = "ductest-test"
		pod   = "nginx-7bccd5989b-6mq4k"
		ctr   = "ng"
		user  = "admin"
		pass  = "123456"
	)

	// WithAuth 每次启动实时登录换取新 JWT——硬编码静态 token 必然过期失效，会被服务器
	// 静默拒绝（鉴权失败不回错误帧），表现为"能输入但没有返回"。
	cli, err := ws.NewClient(wsURL, ws.WithAuth(user, pass))
	if err != nil {
		fmt.Fprintln(os.Stderr, "[ws] NewClient:", err)
		os.Exit(1)
	}
	defer cli.Close()

	// 单调用：内部自动等待连接/鉴权就绪、生成 sessionID、发起 ExecShell，
	// 并内置鉴权竞态兜底（服务器若回"认证中，请稍等~"会自动重发直到 shell 打开）。
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	fmt.Fprintf(os.Stderr, "[ws] 连接 %s，打开容器终端 %s/%s/%s ...\n", wsURL, ns, pod, ctr)
	term, err := cli.OpenTerminal(ctx, &websocket.Container{Namespace: ns, Pod: pod, Container: ctr})
	if err != nil {
		fmt.Fprintln(os.Stderr, "[ws] OpenTerminal:", err)
		os.Exit(1)
	}
	defer term.Close()
	fmt.Fprintf(os.Stderr, "[ws] 终端已打开（sessionID=%s）：直接输入命令回车执行；Ctrl+C 退出\n", term.ID())

	// 数据面：term.Pump 接好 stdin→远端、远端→stdout、远端→toast 三通道，
	// 并按 -raw 开关切本地终端模式（默认 raw，get tab 补全/方向键/clear）。
	// stop 既停转发又恢复本地终端设置，故必须 defer。
	stop := term.Pump(
		os.Stdin,
		func(data []byte) { _, _ = os.Stdout.Write(data) },
		func(data []byte) { fmt.Fprintf(os.Stderr, "[toast] %s\n", data) },
		//ws.WithRawMode(*raw),
	)
	defer stop()

	// 控制面：自动跟随本地终端窗口尺寸变化（初始尺寸 + SIGWINCH）同步远端 pty。
	term.AutoHandleWindowSize()

	// 控制面：SIGINT/SIGTERM（Ctrl+C / kill）→ 优雅退出。
	quit := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	go func() {
		sig := <-sigCh
		fmt.Fprintf(os.Stderr, "\n[ws] 收到信号 %v，退出\n", sig)
		close(quit)
	}()

	// 等待远端会话结束或收到退出信号；defer 已负责 Close。
	select {
	case <-term.Done():
		fmt.Fprintln(os.Stderr, "[ws] 远端会话已结束")
	case <-quit:
	}
}
