---
title: 🧩 SDK 接入
lang: zh-CN
---

# 🧩 SDK 接入

mars 的 SDK 在仓库内 [api/](https://github.com/duc-cnzj/mars/tree/master/api) 模块，同时提供 **gRPC** 与 **HTTP/JSON** 两套客户端，两者共享同一套 proto 生成类型。

## 安装

```bash
go get github.com/duc-cnzj/mars/api/v6
```

import 路径：

| 用途 | 路径 |
|---|---|
| gRPC 客户端 | `github.com/duc-cnzj/mars/api/v6/grpc` |
| HTTP/JSON 客户端 | `github.com/duc-cnzj/mars/api/v6/http` |
| proto 类型 | `github.com/duc-cnzj/mars/api/v6/proto/*` |

## gRPC 客户端（连 :50000）

```go
package main

import (
    client "github.com/duc-cnzj/mars/api/v6/grpc"
)

func main() {
    c, err := client.NewClient("localhost:50000",
        client.WithAuth("admin", "123456"),
    )
    if err != nil {
        panic(err)
    }
    defer c.Close()
}
```

## HTTP/JSON 客户端（连 gateway :4000）

```go
package main

import (
    "time"

    client "github.com/duc-cnzj/mars/api/v6/http"
)

func main() {
    c, err := client.NewClient("http://localhost:4000",
        client.WithAuth("admin", "123456"),
        client.WithTokenAutoRefresh(), // 遇 401 自动重登并重试一次
        client.WithTimeout(30*time.Second),
    )
    if err != nil {
        panic(err)
    }
    defer c.Close()
}
```

::: tip
HTTP 客户端连接的是 **grpc-gateway 端口（默认 :4000）**，不是 gRPC 端口（:50000）。切换传输方式时业务代码无需改动——方法签名、返回类型、错误码全部对齐。
:::

## 调用服务方法

```go
// 与 gRPC SDK 的 cli.Namespace().List(ctx, req) 签名完全一致
ctx := context.Background()
list, err := c.Namespace().List(ctx, &namespace.ListRequest{})
if err != nil {
    log.Fatal(err)
}
```

更多服务：`Project()`、`Container()`、`Event()`、`Changelog()`、`Metrics()`、`Git()` 等，见 [API 参考](./api-reference.md)。

## 示例

仓库内 [examples/](https://github.com/duc-cnzj/mars/tree/master/examples) 提供了完整可运行示例：

| 示例 | 说明 |
|---|---|
| `examples/apply` | 创建命名空间并通过 helm 部署应用 |
| `examples/exec` / `examples/execonce` | 打开容器终端 / 执行一次命令 |
| `examples/copyfile` | 向容器拷贝文件 |
| `examples/http` | HTTP/JSON SDK 演示：列表、流式日志、上传、下载 |
