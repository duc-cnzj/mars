package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/duc-cnzj/mars/api/v5"
	"github.com/duc-cnzj/mars/api/v5/container"
	"golang.org/x/crypto/ssh/terminal"
)

// 定义一个结构体来存储窗口尺寸
type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// 获取终端窗口的当前尺寸
func getWinsize() (*winsize, error) {
	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		os.Stdout.Fd(),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)),
	)
	if int(retCode) == -1 {
		return nil, errno
	}
	return ws, nil
}

func main() {
	client, _ := api.NewClient("localhost:50000", api.WithAuth("admin", "123456"))
	defer client.Close()
	exec, err := client.Container().Exec(context.TODO())
	if err != nil {
		log.Println(err)
		return
	}
	ns := "devops-duc"
	pod := "nginx-54bff68475-k69gh"

	// 捕获 SIGINT 和 SIGWINCH 信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)

	// 创建上下文和取消函数
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 获取当前终端的状态
	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("failed to set terminal to raw mode: %v", err)
	}
	// 确保在程序退出时恢复终端状态
	defer terminal.Restore(int(os.Stdin.Fd()), oldState)

	// 启动一个 goroutine 来处理信号
	go func() {
		for {
			sig := <-sigChan
			switch sig {
			case syscall.SIGWINCH:
				// 获取新的窗口尺寸
				ws, err := getWinsize()
				if err != nil {
					fmt.Println("Error getting window size:", err)
					continue
				}
				exec.Send(&container.ExecRequest{
					Namespace: ns,
					Pod:       pod,
					SizeQueue: &container.TerminalSize{
						Width:  uint32(ws.Col),
						Height: uint32(ws.Row),
					},
				})
				// 输出新的窗口尺寸
				fmt.Printf("Window size changed: %d rows, %d columns\n", ws.Row, ws.Col)
			}
		}
	}()

	defer func() {
		log.Println("CloseSend: ", exec.CloseSend())
	}()

	// 启动一个 goroutine 来接收服务端消息
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				recv, err := exec.Recv()
				if err != nil {
					cancel()
					os.Exit(1)
					return
				}
				if recv.Error != nil {
					fmt.Printf("code=%v msg=%v", recv.Error.Code, recv.Error.Message)
					os.Exit(int(recv.Error.Code))
					return
				}
				fmt.Print(string(recv.Message))
			}
		}
	}()

	// 发送初始命令
	err = exec.Send(&container.ExecRequest{
		Namespace: ns,
		Pod:       pod,
		Command:   []string{"sh"},
	})
	if err != nil {
		log.Println(err)
		return
	}

	// 读取用户输入并发送给服务端
	reader := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			b, err := reader.ReadByte()
			if err != nil {
				if err == io.EOF {
					fmt.Println("Received EOF (Ctrl+D), exiting...")
					cancel()
					return
				}
				if err == bufio.ErrBufferFull {
					continue
				}
				log.Println("Reader error:", err)
				cancel()
				return
			}
			err = exec.Send(&container.ExecRequest{
				Namespace: ns,
				Pod:       pod,
				Message:   []byte{b},
			})
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}
