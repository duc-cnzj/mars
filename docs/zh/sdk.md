---
title: 🔌 SDK 对接
lang: zh-CN
---

# 🔌 SDK 对接

网页上能做的事，用代码都能做。Mars 自带 Go SDK，封装了和前端同一套 API——不用离开终端，就能自动化部署、往 Pod 里拷文件、执行命令、拉日志。

## 两种传输方式，同一套 API

SDK 提供两种传输方式，它们共用**同一套生成的类型、方法签名和错误码**。切换传输方式不需要改业务代码。

| 传输方式 | 包 | 端口 | 适合场景 |
|---|---|---|---|
| **gRPC** | `.../api/v6/grpc` | `50000` | 服务间调用、流式（部署进度、日志、exec）|
| **HTTP/JSON** | `.../api/v6/http` | `4000` | 脚本、不方便走 HTTP/2 的环境 |

## 安装

```bash
go get github.com/duc-cnzj/mars/api/v6/grpc
go get github.com/duc-cnzj/mars/api/v6/http
```

## 连接与认证

两种客户端都在构造时传入用户名和密码。HTTP 客户端还能在 `401` 时自动重登。

```go
// gRPC
client, _ := grpc.NewClient("localhost:50000", grpc.WithAuth("admin", "123456"))
defer client.Close()

// HTTP/JSON
cli, err := http.NewClient("http://localhost:4000",
    http.WithAuth("admin", "123456"),
    http.WithTokenAutoRefresh(), // 遇 401 自动重新登录并重试一次
    http.WithTimeout(30*time.Second),
)
defer cli.Close()
```

> 日常使用建议用非 admin 账号。访问 token 按「持有即有权」处理——见 [权限管理](./access-control.md)。

## 列出命名空间

```go
ns, err := cli.Namespace().List(ctx, &namespace.ListRequest{})
if err != nil {
    log.Fatal(err)
}
for _, item := range ns.GetItems() {
    fmt.Printf("namespace: id=%d name=%s projects=%d\n",
        item.GetId(), item.GetName(), len(item.GetProjects()))
}
```

## 部署应用

`Apply` 是流式调用——会一直推送部署进度，直到 `Metadata.End` 为止。用 `ExtraValues` 可以临时覆盖 chart 里的值。

```go
input := &project.ApplyRequest{
    NamespaceId: 20,   // 鼠标悬停命名空间名称时会显示
    RepoId:      107,  // 后台仓库管理页面可见
    Atomic:      false,
    ExtraValues: []*websocket.ExtraValue{
        {Path: "replicaCount", Value: "1"},
    },
}
apply, err := client.Project().Apply(ctx, input)
if err != nil {
    log.Fatal(err)
}
for {
    recv, err := apply.Recv()
    if err != nil { log.Fatal(err) }
    if recv.Metadata.End {
        fmt.Println(recv.Metadata.Message, recv.Project)
        return
    }
    fmt.Println(recv.Metadata.Message)
}
```

## 拉取容器日志

HTTP 下的服务端流会输出 NDJSON / SSE——SDK 自动兼容两种格式。`io.EOF` 表示流正常结束。

```go
stream, err := cli.Container().StreamContainerLog(ctx, &container.LogRequest{
    Namespace: "devops-demo",
    Pod:       "nginx-54bff68475-k69gh",
    Container: "ng",
})
if err != nil { log.Fatal(err) }
defer stream.Close()
for {
    msg, err := stream.Recv()
    if errors.Is(err, io.EOF) { break }
    if err != nil { log.Fatal(err) }
    fmt.Println(msg.GetLog())
}
```

## 往 Pod 里拷贝文件

把文件分块流式发到 `StreamCopyToPod`。Pod 有多个容器时传 `Container`，否则用默认容器。

```go
cp, _ := c.Container().StreamCopyToPod(ctx)
src, _ := os.Open("./config.yaml")
defer src.Close()
buf := make([]byte, 1024*1024)
for {
    n, err := src.Read(buf)
    if err != nil { break } // io.EOF
    cp.Send(&container.StreamCopyToPodRequest{
        FileName:  src.Name(),
        Data:      buf[:n],
        Namespace: "devops-demo",
        Pod:       "nginx-54bff68475-k69gh",
    })
}
resp, _ := cp.CloseAndRecv()
fmt.Printf("uploaded: %v\n", resp)
```

## HTTP 独有能力：上传与下载

文件传输只有 HTTP 版——gRPC 的 File 服务没有上传接口。

```go
// 上传，返回文件 id，之后用它下载
resp, err := cli.File().UploadFile(ctx, name, file)
fmt.Println("file id:", resp.ID)

// 下载，返回 io.ReadCloser；文件名/大小从响应头解析
rc, info, err := cli.File().DownloadFile(ctx, id)
defer rc.Close()
fmt.Printf("file=%s size=%d\n", info.Filename, info.Size)
```

## 错误处理

错误携带 gRPC 状态码，用 `status.Code` / `status.Convert` 判断——HTTP 和 gRPC 客户端返回同样的码。

```go
_, err = cli.Namespace().Show(ctx, &namespace.ShowRequest{Id: -1})
if status.Code(err) == codes.NotFound {
    fmt.Println("Show(id=-1) -> codes.NotFound")
}
```

## 运行示例

每个能力在仓库的 `examples/` 下都有可运行的示例：

| 示例 | 演示什么 | 运行方式 |
|---|---|---|
| `examples/http` | 列出命名空间、拉日志、上传 / 下载文件 | `go run ./examples/http` |
| `examples/apply` | 部署应用并调整副本数（流式）| `go run ./examples/apply` |
| `examples/copyfile` | 往 Pod 里拷贝文件 | `go run ./examples/copyfile` |
| `examples/exec` | 在容器里开交互式终端 | `go run ./examples/exec` |
| `examples/execonce` | 在容器里执行单条命令 | `go run ./examples/execonce` |

> 示例默认连 `localhost`。把它们指向你的 Mars 地址，例如 `go run ./examples/http -addr http://你的主机:4000`。
