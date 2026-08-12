// Package main 演示 mars HTTP/JSON SDK（github.com/duc-cnzj/mars/api/v6/http）的调用方式。
//
// 与 gRPC SDK（api/grpc）共享同一套 proto 生成类型：方法签名、返回类型、错误码全部对齐，
// 唯一区别是传输层——HTTP/1.1 JSON（grpc-gateway）而非 HTTP/2 gRPC。切换传输方式时业务代码无需改动。
//
// 连接的是 grpc-gateway 端口（默认 :4000），不是 gRPC 端口（:50000）。
//
// 用法：
//
//	go run ./examples/http                       # 默认：unary 列出项目空间 + 错误码对齐演示
//	go run ./examples/http -action logs          # server-streaming 拉取 pod 日志（SSE/NDJSON）
//	go run ./examples/http -action upload -file ./x.tar.gz   # multipart 上传（HTTP 特有）
//	go run ./examples/http -action download -file-id 1       # 二进制下载（HTTP 特有）
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/duc-cnzj/mars/api/v6/http"
	"github.com/duc-cnzj/mars/api/v6/proto/container"
	"github.com/duc-cnzj/mars/api/v6/proto/namespace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	var (
		baseURL = flag.String("addr", "http://localhost:4000", "grpc-gateway 地址（默认 :4000）")
		user    = flag.String("user", "admin", "用户名")
		pass    = flag.String("pass", "123456", "密码")
		action  = flag.String("action", "list", "演示动作: list | logs | upload | download")
		file    = flag.String("file", "", "upload 动作要上传的本地文件路径")
		fileID  = flag.Int("file-id", 0, "download 动作要下载的文件 id")
	)
	flag.Parse()

	// 构造客户端：WithAuth 在构造阶段完成登录换取 token；
	// WithTokenAutoRefresh 遇 401 自动重登并重试一次；WithTimeout 限制整体超时。
	// 想接入链路追踪再加 http.WithTracer()。
	cli, err := http.NewClient(*baseURL,
		http.WithAuth(*user, *pass),
		http.WithTokenAutoRefresh(),
		http.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatalf("NewClient: %v", err)
	}
	defer cli.Close()

	ctx := context.Background()
	switch *action {
	case "logs":
		streamLogs(ctx, cli)
	case "upload":
		uploadFile(ctx, cli, *file)
	case "download":
		downloadFile(ctx, cli, *fileID)
	default:
		listNamespaces(ctx, cli)
	}
}

// listNamespaces 展示 unary 调用：与 gRPC SDK 的 cli.Namespace().List(ctx, req) 签名完全一致。
func listNamespaces(ctx context.Context, cli *http.Client) {
	ns, err := cli.Namespace().List(ctx, &namespace.ListRequest{})
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range ns.GetItems() {
		fmt.Printf("namespace: id=%d name=%s projects=%d private=%v\n",
			item.GetId(), item.GetName(), len(item.GetProjects()), item.GetPrivate())
	}

	// 错误处理：gateway 把 gRPC 错误码还原成 codes.Error，调用方用 status.Code 判断，与 gRPC SDK 对齐。
	_, err = cli.Namespace().Show(ctx, &namespace.ShowRequest{Id: -1})
	if err != nil && status.Code(err) == codes.NotFound {
		fmt.Println("-> Show(id=-1) 返回 codes.NotFound，错误码已对齐")
	} else {
		fmt.Printf("-> Show(id=-1) err: %v\n", err)
	}
}

// streamLogs 展示 server-streaming：HTTP/1.1 下 gateway 输出 NDJSON 或 SSE，SDK 自动兼容两种格式。
// io.EOF 表示流正常结束；流中途错误以 google.rpc.Status envelope 返回，还原成 codes.Error。
func streamLogs(ctx context.Context, cli *http.Client) {
	stream, err := cli.Container().StreamContainerLog(ctx, &container.LogRequest{
		Namespace: "ductest-cool",
		Pod:       "nginx-54bff68475-k69gh",
		Container: "ng",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()
	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(msg.GetLog())
	}
}

// uploadFile 展示 HTTP 特有能力的调用：multipart 上传（gRPC 的 File service 没有上传 RPC）。
// 返回文件 ID，后续可用 download 动作下载。
func uploadFile(ctx context.Context, cli *http.Client, path string) {
	if path == "" {
		log.Fatal("-file 必填")
	}
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	resp, err := cli.File().UploadFile(ctx, f.Name(), f)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("uploaded, file id=%d\n", resp.ID)
}

// downloadFile 展示 HTTP 特有能力的调用：按文件 ID 二进制下载。
// 返回的 io.ReadCloser 由调用方负责关闭，元信息（文件名/大小）从响应头解析。
func downloadFile(ctx context.Context, cli *http.Client, id int) {
	rc, info, err := cli.File().DownloadFile(ctx, id)
	if err != nil {
		log.Fatal(err)
	}
	defer rc.Close()
	fmt.Printf("downloading file=%s size=%d\n", info.Filename, info.Size)
	if _, err := io.Copy(os.Stdout, rc); err != nil {
		log.Fatal(err)
	}
}
