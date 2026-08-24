---
title: 🧩 SDK
lang: en-US
---

# 🧩 SDK

The mars SDK lives in the in-repo [api/](https://github.com/duc-cnzj/mars/tree/master/api) module, providing both **gRPC** and **HTTP/JSON** clients that share the same generated proto types.

## Install

```bash
go get github.com/duc-cnzj/mars/api/v6
```

Import paths:

| Purpose | Path |
|---|---|
| gRPC client | `github.com/duc-cnzj/mars/api/v6/grpc` |
| HTTP/JSON client | `github.com/duc-cnzj/mars/api/v6/http` |
| proto types | `github.com/duc-cnzj/mars/api/v6/proto/*` |

## gRPC Client (connect :50000)

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

## HTTP/JSON Client (connect the gateway :4000)

```go
package main

import (
    "time"

    client "github.com/duc-cnzj/mars/api/v6/http"
)

func main() {
    c, err := client.NewClient("http://localhost:4000",
        client.WithAuth("admin", "123456"),
        client.WithTokenAutoRefresh(), // auto re-login and retry once on 401
        client.WithTimeout(30*time.Second),
    )
    if err != nil {
        panic(err)
    }
    defer c.Close()
}
```

::: tip
The HTTP client connects to the **grpc-gateway port (default :4000)**, not the gRPC port (:50000). Switching transports requires no business-code changes — method signatures, return types and error codes are fully aligned.
:::

## Calling Service Methods

```go
// identical signature to the gRPC SDK: cli.Namespace().List(ctx, req)
ctx := context.Background()
list, err := c.Namespace().List(ctx, &namespace.ListRequest{})
if err != nil {
    log.Fatal(err)
}
```

More services: `Project()`, `Container()`, `Event()`, `Changelog()`, `Metrics()`, `Git()`, etc. — see [API Reference](./api-reference.md).

## Examples

Complete runnable examples live in [examples/](https://github.com/duc-cnzj/mars/tree/master/examples):

| Example | Description |
|---|---|
| `examples/apply` | create a namespace and deploy an app via Helm |
| `examples/exec` / `examples/execonce` | open a container terminal / run a one-shot command |
| `examples/copyfile` | copy a file into a container |
| `examples/http` | HTTP/JSON SDK demo: list, stream logs, upload, download |
