// Package main 演示 mars HTTP/JSON SDK（github.com/duc-cnzj/mars/api/v6/http）的调用方式。
//
// 与 gRPC SDK（examples/grpc）共享同一套 proto 生成类型：方法签名、返回类型、错误码全部对齐，
// 唯一区别是传输层——HTTP/1.1 JSON（grpc-gateway）而非 HTTP/2 gRPC。切换传输方式时业务代码无需改动。
//
// 连接的是 grpc-gateway 端口（默认 :4000），不是 gRPC 端口（:50000）。
//
// 用法：各动作所需参数均为假造示例值，无需外部传入。
//
//	go run ./examples/http                             # 默认：unary 列出项目空间 + 错误码对齐演示
//	go run ./examples/http -action logs                # server-streaming 拉取 pod 日志（SSE/NDJSON）
//	go run ./examples/http -action exec_once           # server-streaming 执行一次命令，演示错误帧语义
//	go run ./examples/http -action exec_once -timeout 1 # 配合把 command 换成 "sleep 5" 演示超时错误帧 -3
//	go run ./examples/http -action pod_running         # unary 查询 pod 是否 running
//	go run ./examples/http -action version             # unary 查询服务端版本
//	go run ./examples/http -action project             # unary 分页列出可见项目（核心业务域）
//	go run ./examples/http -action top_pod             # server-streaming 实时指标（SSE）
//	go run ./examples/http -action cluster             # unary 集群概览（运维）
//	go run ./examples/http -action webapply            # DryRun 预览部署 yaml（无副作用）
//	go run ./examples/http -action upload              # multipart 上传一个假造的临时文件（HTTP 特有）
//	go run ./examples/http -action download            # 按假造 file-id 二进制下载（HTTP 特有）
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
	"github.com/duc-cnzj/mars/api/v6/proto/cluster"
	"github.com/duc-cnzj/mars/api/v6/proto/container"
	"github.com/duc-cnzj/mars/api/v6/proto/metrics"
	"github.com/duc-cnzj/mars/api/v6/proto/namespace"
	"github.com/duc-cnzj/mars/api/v6/proto/project"
	"github.com/duc-cnzj/mars/api/v6/proto/version"
	"github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func main() {
	var (
		baseURL = flag.String("addr", "http://localhost:4000", "grpc-gateway 地址（默认 :4000）")
		user    = flag.String("user", "admin", "用户名")
		pass    = flag.String("pass", "123456", "密码")
		action  = flag.String("action", "list", "演示动作: list | logs | exec_once | pod_running | version | project | top_pod | cluster | webapply | upload | download")
		timeout = flag.Int64("timeout", 60, "命令最大执行秒数（0=服务端默认 1min，exec_once 动作使用）")
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
	case "exec_once":
		execOnce(ctx, cli, *timeout)
	case "pod_running":
		podRunning(ctx, cli)
	case "version":
		serverVersion(ctx, cli)
	case "project":
		listProjects(ctx, cli)
	case "webapply":
		webApplyPreview(ctx, cli)
	case "top_pod":
		streamTopPod(ctx, cli)
	case "cluster":
		clusterInfo(ctx, cli)
	case "upload":
		uploadFile(ctx, cli)
	case "download":
		downloadFile(ctx, cli)
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

// execOnce 展示 ExecOnce 的一次性命令执行 + 容器执行结果错误帧的消费。
//
// 核心契约（与容器执行结果有关）：
//   - 输出经 ExecResponse.message 逐帧返回，逐帧打印；
//   - 容器执行结果（命令退出码 / 输出截断 / exec 启动失败）经 ExecResponse.error
//     错误帧传达，流以 io.EOF 正常结束，**不提升为 HTTP 5xx**；
//   - io.EOF 是"命令执行完毕"的正常结束信号，不是错误；
//   - 仅当 Recv 返回 mars 自身错误（如连接断开）才是传输层错误。
//
// error 帧 code 语义：
//   - 0-255  容器内命令的非零退出码
//   - -1     命令输出超限被服务端强制截断
//   - -2     容器 exec 启动/执行失败（如命令在容器内不存在）
//   - -3     命令执行超时被服务端强制终止（ExecOnce，timeout_seconds 上限）
//
// 复现不同路径：把 command 换成 "cc" 演示 -2；换成 "sleep 5" 并 -timeout 1 演示 -3；
// 下方默认命令演示退出码 3。
func execOnce(ctx context.Context, cli *http.Client, timeoutSeconds int64) {
	stream, err := cli.Container().ExecOnce(ctx, &container.ExecOnceRequest{
		Namespace:      "duc-abc",
		Pod:            "ng-nginx-594b65865-g975j",
		Container:      "",
		Command:        []string{"bash", "-c", "echo 'hello from exec_once'; echo 'oops to stderr' >&2; exit 3"},
		TimeoutSeconds: timeoutSeconds,
	})
	if err != nil {
		log.Fatalf("ExecOnce 建立流失败: %v", err)
	}
	defer stream.Close()
	for {
		recv, err := stream.Recv()
		// io.EOF 是命令执行完毕的正常结束信号，不是错误。
		if errors.Is(err, io.EOF) {
			fmt.Println("\n-> 流以 io.EOF 正常结束（命令执行完毕）")
			return
		}
		if err != nil {
			log.Fatalf("ExecOnce Recv 传输层错误: %v", err)
		}
		if recv.Error != nil {
			fmt.Printf("-> 错误帧: code=%d %s\n", recv.Error.Code, describeExecErrorCode(recv.Error.Code))
			if recv.Error.Message != "" {
				fmt.Printf("   %s\n", recv.Error.Message)
			}
			continue
		}
		fmt.Print(string(recv.Message))
	}
}

// describeExecErrorCode 解释 ExecError.code 语义，便于 demo 直观展示错误帧类型。
func describeExecErrorCode(code int64) string {
	switch {
	case code == -1:
		return "(命令输出超限被服务端强制截断)"
	case code == -2:
		return "(exec 启动/执行失败，如命令在容器内不存在)"
	case code == -3:
		return "(命令执行超时被服务端强制终止)"
	default:
		return fmt.Sprintf("(容器内命令退出码 %d)", code)
	}
}

// podRunning 展示 unary 查询 pod 运行状态。
func podRunning(ctx context.Context, cli *http.Client) {
	resp, err := cli.Container().IsPodRunning(ctx, &container.IsPodRunningRequest{
		Namespace: "duc-abc",
		Pod:       "ng-nginx-594b65865-g975j",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("pod running=%v reason=%q\n", resp.GetRunning(), resp.GetReason())
}

// serverVersion 展示最简单的 unary 调用：查询服务端版本信息。
func serverVersion(ctx context.Context, cli *http.Client) {
	v, err := cli.Version().Version(ctx, &version.Request{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("mars server: version=%s branch=%s commit=%s go=%s\n",
		v.GetVersion(), v.GetGitBranch(), v.GetGitCommit(), v.GetGoVersion())
}

// listProjects 展示核心业务域 unary 调用：分页列出当前用户可见的项目。
func listProjects(ctx context.Context, cli *http.Client) {
	resp, err := cli.Project().List(ctx, &project.ListRequest{Page: proto.Int32(1), PageSize: proto.Int32(20)})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("projects: page=%d page_size=%d count=%d\n",
		resp.GetPage(), resp.GetPageSize(), resp.GetCount())
	for _, p := range resp.GetItems() {
		fmt.Printf("  project id=%d name=%s namespace=%d branch=%s deploy=%s\n",
			p.GetId(), p.GetName(), p.GetNamespaceId(), p.GetGitBranch(), p.GetDeployStatus())
	}
}

// webApplyPreview 展示核心部署 RPC：WebApplyRequest 的 9 个字段在 literal 中全部写全，
// 每个字段都给假造示例值，读者一眼看全完整请求形状。强制 DryRun=true 只预览渲染出的 yaml，
// 不做实际部署（无副作用）。
func webApplyPreview(ctx context.Context, cli *http.Client) {
	show, _ := cli.Project().Show(ctx, &project.ShowRequest{Id: 1})
	req := &project.WebApplyRequest{
		NamespaceId: 1,
		Name:        "demo-project",
		RepoId:      2,
		GitBranch:   "main",
		GitCommit:   "a1b2c3d4e5f6",
		Config:      "replicaCount: 2\nimage:\n  repository: nginx\n  tag: 1.25\n",
		ExtraValues: []*websocket.ExtraValue{
			{
				Path:  "replicaCount",
				Value: "1",
			},
		},
		Version: proto.Int32(show.Item.Version),
		DryRun:  true,
	}
	resp, err := cli.Project().WebApply(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("webapply(dry-run): %d 个渲染 yaml, dry_run=%v\n", len(resp.GetYamlFiles()), resp.GetDryRun())
	for _, y := range resp.GetYamlFiles() {
		fmt.Println("---")
		fmt.Println(y)
	}
}

// streamTopPod 展示第三种流式能力——指标实时流（SSE）：与日志（streamLogs）/命令（execOnce）不同，
// 这是服务端周期性推送的监控数据，同一 transport.Stream 模式，仅 payload 不同。
func streamTopPod(ctx context.Context, cli *http.Client) {
	stream, err := cli.Metrics().StreamTopPod(ctx, &metrics.TopPodRequest{
		Namespace: "duc-abc",
		Pod:       "ng-nginx-594b65865-g975j",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()
	for {
		m, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("-> top_pod 流结束")
			return
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("top_pod cpu=%s mem=%s time=%s\n",
			m.GetHumanizeCpu(), m.GetHumanizeMemory(), m.GetTime())
	}
}

// clusterInfo 展示运维 unary 调用：集群资源概览。
func clusterInfo(ctx context.Context, cli *http.Client) {
	resp, err := cli.Cluster().ClusterInfo(ctx, &cluster.InfoRequest{})
	if err != nil {
		log.Fatal(err)
	}
	item := resp.GetItem()
	fmt.Printf("cluster status=%s\n", item.GetStatus())
	fmt.Printf("  cpu: total=%s free=%s usage=%s request=%s\n",
		item.GetTotalCpu(), item.GetFreeCpu(), item.GetUsageCpuRate(), item.GetRequestCpuRate())
	fmt.Printf("  mem: total=%s free=%s usage=%s request=%s\n",
		item.GetTotalMemory(), item.GetFreeMemory(), item.GetUsageMemoryRate(), item.GetRequestMemoryRate())
}

// uploadFile 展示 HTTP 特有能力的调用：multipart 上传（gRPC 的 File service 没有上传 RPC）。
// 输入为假造的临时文件（真实内容无关紧要，服务端只关心能收到），返回文件 ID 供 download 使用。
func uploadFile(ctx context.Context, cli *http.Client) {
	f, err := os.CreateTemp("", "mars-upload-*.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if _, err := f.WriteString("fake upload content\n"); err != nil {
		log.Fatal(err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		log.Fatal(err)
	}
	resp, err := cli.File().UploadFile(ctx, f.Name(), f)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("uploaded, file id=%d\n", resp.ID)
}

// downloadFile 展示 HTTP 特有能力的调用：按文件 ID 二进制下载。
// file-id 为假造值（可用 upload 动作返回的真实 id 替换）。返回的 io.ReadCloser 由调用方负责关闭。
func downloadFile(ctx context.Context, cli *http.Client) {
	const fileID = 1
	rc, info, err := cli.File().DownloadFile(ctx, fileID)
	if err != nil {
		log.Fatal(err)
	}
	defer rc.Close()
	fmt.Printf("downloading file=%s size=%d\n", info.Filename, info.Size)
	if _, err := io.Copy(os.Stdout, rc); err != nil {
		log.Fatal(err)
	}
}
