# mars API SDK

mars 的客户端 SDK 模块（`github.com/duc-cnzj/mars/api/v6`）。提供 **gRPC** 与 **HTTP/JSON**（grpc-gateway）两套对等客户端，共享同一批 proto 生成类型：方法签名、返回类型、错误码全部对齐，调用方切换传输方式时业务代码无需改动。

## 两种传输，一套类型

| 维度 | gRPC SDK (`api/grpc`) | HTTP SDK (`api/http`) |
|---|---|---|
| 传输 | HTTP/2 gRPC | HTTP/1.1 JSON（grpc-gateway） |
| 客户端 | `grpc.NewClient(addr, opts...)` | `http.NewClient(baseURL, opts...)` |
| 访问器 | `cli.Namespace().List(ctx, req)` | `cli.Namespace().List(ctx, req)` |
| 流式 | 原生 gRPC stream | server-streaming → SSE/NDJSON |
| 需要服务端 | mars gRPC 端口（如 `:50000`） | mars gateway 端口（如 `:4000`） |

两个包都暴露同一批 15 个 service 访问器：`Auth/Repo/Changelog/Cluster/Container/Event/AccessToken/File/Git/Metrics/Namespace/Picture/Project/Version/Endpoint`。

### 能力差异：gRPC 特有 vs HTTP 特有

gRPC 共 **65** 个方法，HTTP **61** 个；其中 **61 个共享**（每个 HTTP stub 方法在 gRPC 都有对应，签名一致）。差异只有两类，生成器/手写代码在源码里都有明确注释，可复核。

**gRPC 特有（4 个）—— HTTP 生成器诚实跳过：**

| 方法 | streaming 类型 | HTTP 缺失原因 |
|---|---|---|
| `Container.Exec` | bidi | HTTP/JSON 无解，需要 WebSocket（mars ws 通道承载终端） |
| `Container.StreamCopyToPod` | client | HTTP/JSON 无解，需要 WebSocket |
| `Container.ExecOnce` | server | proto **无 `google.api.http` 注解**，grpc-gateway 不暴露该路由 |
| `Project.Apply` | server | proto **无 `google.api.http` 注解**，gateway 不暴露；HTTP 侧替代是 `Project.WebApply` |

> 前两个是流式方向本身（client/bidi streaming）在 HTTP/1.1 JSON 下无解；后两个是 server-streaming 但 `.proto` 没配 http 注解——gateway 根本没有对应 HTTP 路由，HTTP SDK 自然没有方法。`Container.StreamContainerLog` / `Metrics.StreamTopPod` 是**配了注解**的 server-streaming，两套 SDK 都有（gRPC 原生流 / HTTP SSE）。

**HTTP 特有（3 个）—— 不在任何 proto 里，gRPC 无对应：**

| 方法 | HTTP 路由 | 说明 |
|---|---|---|
| `FileAPI.UploadFile` | `POST /api/files` | multipart 上传，返回文件 ID |
| `FileAPI.DownloadFile` | `GET /api/download_file/{id}` | 二进制下载，返回流 + 元信息 |
| `FileAPI.CopyFromPod` | `POST /api/copy_from_pod` | 从 pod 拷贝文件到本地 |

> 注意方向性：`Container.CopyToPod`（拷入 pod）两套 SDK 都有；`FileAPI.CopyFromPod`（从 pod 拷出）只有 HTTP 有。gRPC 的 `File` service 只有 `List/Delete/DiskInfo/MaxUploadSize/ShowRecords`，**没有**上传/下载 RPC。

## 安装

```bash
go get -u github.com/duc-cnzj/mars/api/v6/grpc
go get -u github.com/duc-cnzj/mars/api/v6/http
```

## gRPC 用法

```go
package main

import (
	"context"

	"github.com/duc-cnzj/mars/api/v6/grpc"
	"github.com/duc-cnzj/mars/api/v6/proto/namespace"
)

func main() {
	// 构造时若配置了 WithAuth，会立即登录换取 token，失败返回 error。
	c, err := grpc.NewClient("127.0.0.1:50000",
		grpc.WithAuth("admin", "123456"),
		grpc.WithTokenAutoRefresh(), // 401 时自动重登并重试（默认最多 5 次指数退避）
	)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	ns, err := c.Namespace().List(context.Background(), &namespace.ListRequest{})
	if err != nil {
		panic(err)
	}
	_ = ns
}
```

### gRPC Option

| Option | 作用 |
|---|---|
| `WithAuth(username, password)` | 构造时登录换取 token，挂到每个 RPC 的 Authorization 元数据 |
| `WithBearerToken(token)` | 直接注入已签发 token（自动补 `Bearer` 前缀） |
| `WithTokenAutoRefresh()` | 遇 `codes.Unauthenticated`（且配置了 WithAuth）自动重登重试，覆盖 unary 与 server-streaming；Login/Exchange 自身 401 原样返回（凭据错误重试无意义，也避免 singleflight 自死锁） |
| `WithUnaryClientInterceptor(op)` | 追加 unary 拦截器 |
| `WithStreamClientInterceptor(op)` | 追加 streaming 拦截器 |
| `WithTracer()` | 接入 OpenTelemetry（otelgrpc client stats handler） |
| `WithTransportCredentials(tlsCfg)` | 使用自定义 `tls.Config` 建立 TLS 连接（含 mTLS）；不调用默认明文 insecure |

运行期替换 token：`c.SetBearerToken("...")`。

## HTTP 用法

```go
package main

import (
	"context"

	"github.com/duc-cnzj/mars/api/v6/http"
	"github.com/duc-cnzj/mars/api/v6/proto/namespace"
)

func main() {
	c, err := http.NewClient("http://127.0.0.1:4000",
		http.WithAuth("admin", "123456"),
		http.WithTokenAutoRefresh(),
		http.WithTimeout(30*time.Second),
	)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	ns, err := c.Namespace().List(context.Background(), &namespace.ListRequest{})
	if err != nil {
		panic(err)
	}
	_ = ns
}
```

> 完整可运行示例见仓库根 [`examples/http`](../examples/http)：覆盖 unary 调用、错误码对齐、
> server-streaming（SSE）、以及 HTTP 特有能力（multipart 上传 / 二进制下载）。示例默认连接 gateway `:4000`；
> 服务器端 `mars serve` 的 `app_port` 默认 `:6000`，`:4000` 常见于本地 port-forward / docker 端口映射。

### HTTP Option

| Option | 作用 |
|---|---|
| `WithAuth(username, password)` | 构造时 `POST /api/auth/login` 换取 token |
| `WithBearerToken(token)` | 直接注入已签发 token（自动补 `Bearer` 前缀） |
| `WithTokenAutoRefresh()` | 遇 401（且配置了 WithAuth）自动重登并重试一次；Login/Exchange 自身 401 原样返回 |
| `WithHTTPClient(hc)` | 替换底层 `*http.Client`（可注入自定义 transport） |
| `WithTimeout(d)` | 设置底层 http.Client 整体超时 |
| `WithTracer()` | 接入 OpenTelemetry，请求注入 trace（底层用 otelhttp 包装 Transport） |

## Option 对照

两套 SDK 共享 `WithAuth` / `WithBearerToken` / `WithTokenAutoRefresh` / `WithTracer`，语义一致。传输层特有差异：

| 传输特有 | 说明 |
|---|---|
| gRPC `WithTransportCredentials(tlsCfg)` | HTTP 侧无需该 Option（`WithHTTPClient` 换 transport 即覆盖 TLS/代理） |
| gRPC `WithUnaryClientInterceptor` / `WithStreamClientInterceptor` | 拦截器注入是 gRPC 原生机制，HTTP 无对应 |
| HTTP `WithHTTPClient(hc)` / `WithTimeout(d)` | 直接操控 `*http.Client`，gRPC 无对应（连接配置走 dial options） |

## Server-streaming

HTTP/JSON 下，带 `google.api.http` 注解的 server-streaming 方法由 gateway 输出 NDJSON（`{"result": <msg>}`）或标准 SSE（`data: {...}`），SDK 自动兼容两种格式：

```go
// 例：container.StreamContainerLog / StreamTopPod
stream, err := c.Container().StreamContainerLog(ctx, &container.StreamContainerLogRequest{...})
if err != nil {
	panic(err)
}
defer stream.Close()
for {
	msg, err := stream.Recv() // io.EOF = 流正常结束
	if err != nil {
		if errors.Is(err, io.EOF) {
			break
		}
		panic(err)
	}
	_ = msg
}
```

流中途错误以 `google.rpc.Status` envelope 返回，还原成 `codes.Error`，与 unary 错误码通用。client/bidi streaming（`Exec`、`StreamCopyToPod`、`Apply`）在 HTTP/JSON 下无解（需要 WebSocket），生成器诚实跳过，仅 gRPC SDK 可用。

## 生成工作流

proto 变更后重跑生成器，产物落在 `api/http/rest/`：

```bash
cd api
go generate ./http/...      # 触发 go:generate → go run ./gen/cmd
go run ./http/gen/cmd       # 等价手跑
```

生成器保证 **rest/ 目录 100% 等于当前 proto 的产物**：

- unary + `google.api.http` 注解 → 生成 HTTP stub；
- server-streaming + 注解 → 生成 SSE stub；
- client/bidi streaming、无注解方法、自定义路由 → 跳过/手写；
- 生成后清理 rest/ 里不在集合内的孤儿 `*.gen.http.go`（proto 删掉 service 不留垃圾文件），绝不触碰手写文件。

`api/http/gen_test.go` 的 `TestGeneratedStubsUpToDate` 做双向漂移校验：已提交 stub ⊆ 生成器输出 且 无孤儿。改 proto 后必须先重新生成再提交，否则测试挂。

## 质量保证

```bash
go build ./...             # 全量编译
go vet ./...               # 静态检查
go test ./...              # 单测（手写生产代码 100%，rest/ 生成 stub 由漂移测试兜底）
```

- grpc 包：bufconn 内存 gRPC 测试，覆盖登录/token 前缀/自动刷新/singleflight 去重/拦截器注入/15 个访问器；
- http/transport 包：fake stream/conn 覆盖泛型流工厂与错误传播；
- internal/flight 包：singleflight 去重、DoChan、Forget 语义全覆盖。

## 相关

- proto 定义与 gateway 服务端：仓库根 `api/proto/`、`internal/...`；
- 更多使用示例：[examples](https://github.com/duc-cnzj/mars/tree/master/examples)。
