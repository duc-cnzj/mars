// Package main 演示 ExecOnce 一次性命令执行的调用方式（非 tty，一次请求一条命令）。
//
// 与 Exec（交互式）的区别：ExecOnce 不进入交互终端，服务端跑完命令即结束流，
// 适合一次性脚本/巡检，无需挂终端。
//
// 核心契约（这也是本 demo 的重点）：
//   - 输出经 ExecResponse.message 逐帧返回，逐帧打印即可；
//   - 容器执行结果（命令退出码 / 输出截断 / exec 启动失败）经 ExecResponse.error
//     错误帧传达，流以 io.EOF 正常结束，**不提升为传输层 gRPC 错误**；
//   - io.EOF 是"命令执行完毕"的正常结束信号，不是错误；
//   - 仅当 Recv 返回 mars 自身错误（如连接断开）才是传输层错误。
//
// error 帧 code 语义：
//   - 0-255  容器内命令的非零退出码（命令已启动并结束）
//   - -1     命令输出超限被服务端强制截断
//   - -2     容器 exec 启动/执行失败（如命令在容器内不存在）
//   - -3     命令执行超时被服务端强制终止（timeout_seconds 上限）
//
// 用法（flag 驱动，可复现不同错误路径）：
//
//	# 默认：正常输出 + 退出码
//	go run ./examples/grpc/execonce
//
//	# 命令不存在 → code=-2 错误帧
//	go run ./examples/grpc/execonce -ns duc-abc -pod ng-nginx-594b65865-g975j cc
//
//	# 正常退出码 3 → code=3 错误帧
//	go run ./examples/grpc/execonce -ns duc-abc -pod ng-nginx-594b65865-g975j bash -c 'exit 3'
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/duc-cnzj/mars/api/v6/grpc"
	"github.com/duc-cnzj/mars/api/v6/proto/container"
	"github.com/duc-cnzj/mars/v6/examples/execerr"
)

func main() {
	var (
		addr    = flag.String("addr", "localhost:50000", "gRPC 地址（非 gateway :4000）")
		user    = flag.String("user", "admin", "用户名")
		pass    = flag.String("pass", "123456", "密码")
		ns      = flag.String("ns", "duc-abc", "命名空间")
		pod     = flag.String("pod", "ng-nginx-594b65865-g975j", "pod 名")
		ctr     = flag.String("container", "", "容器名（默认使用 pod 首个容器）")
		timeout = flag.Int64("timeout", 60, "命令最大执行秒数（0=服务端默认 1min）")
	)
	flag.Parse()
	cmd := flag.Args()
	if len(cmd) == 0 {
		// 默认命令：输出到 stdout/stderr，并以退出码 3 结束，演示 code=3 错误帧。
		cmd = []string{"bash", "-c", "echo 'hello from execonce'; echo 'oops to stderr' >&2; exit 3"}
	}

	cli, err := grpc.NewClient(*addr, grpc.WithAuth(*user, *pass))
	if err != nil {
		log.Fatalf("NewClient: %v", err)
	}
	defer cli.Close()

	runExecOnce(cli, &container.ExecOnceRequest{
		Namespace:      *ns,
		Pod:            *pod,
		Container:      *ctr,
		Command:        cmd,
		TimeoutSeconds: *timeout,
	})
}

// runExecOnce 执行一条 ExecOnce 命令并逐帧消费结果。
//
// 关键点：容器执行结果走 error 帧 + io.EOF 正常结束，只有 Recv 返回 mars 自身错误才是传输层错误。
func runExecOnce(cli *grpc.Client, req *container.ExecOnceRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := cli.Container().ExecOnce(ctx, req)
	if err != nil {
		log.Fatalf("ExecOnce 建立流失败: %v", err)
	}
	defer stream.CloseSend()

	for {
		recv, err := stream.Recv()
		// io.EOF 是命令执行完毕的正常结束信号，不是错误。
		if err == io.EOF {
			fmt.Println("\n-> 流以 io.EOF 正常结束（命令执行完毕）")
			return
		}
		if err != nil {
			log.Fatalf("ExecOnce Recv 传输层错误: %v", err)
		}
		if recv.Error != nil {
			fmt.Printf("-> 错误帧: code=%d %s\n", recv.Error.Code, execerr.Describe(recv.Error.Code))
			if recv.Error.Message != "" {
				fmt.Printf("   %s\n", recv.Error.Message)
			}
			continue
		}
		fmt.Print(string(recv.Message))
	}
}
