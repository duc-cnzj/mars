---
title: 🔌 SDK & API
lang: en-US
---

# 🔌 SDK & API

Everything you can do in the web UI you can also do from code. Mars ships a Go SDK that wraps the same API the UI uses — automate deployments, copy files into pods, run commands, and stream logs without leaving your terminal.

## Two transports, one API

The SDK is available over two transports that share the **same generated types, method signatures and error codes**. Switching transports changes no business code.

| Transport | Package | Port | Best for |
|---|---|---|---|
| **gRPC** | `.../api/v6/grpc` | `50000` | Server-to-server, streaming (deploy progress, logs, exec) |
| **HTTP/JSON** | `.../api/v6/http` | `4000` | Scripts and environments that can't speak HTTP/2 |

## Install

```bash
go get github.com/duc-cnzj/mars/api/v6/grpc
go get github.com/duc-cnzj/mars/api/v6/http
```

## Connect and authenticate

Both clients take a username and password at construction time. The HTTP client can also auto-refresh the token on `401`.

```go
// gRPC
client, _ := grpc.NewClient("localhost:50000", grpc.WithAuth("admin", "123456"))
defer client.Close()

// HTTP/JSON
cli, err := http.NewClient("http://localhost:4000",
    http.WithAuth("admin", "123456"),
    http.WithTokenAutoRefresh(), // re-login on 401 and retry once
    http.WithTimeout(30*time.Second),
)
defer cli.Close()
```

> Prefer a non-admin account for everyday use. Access tokens behave as "whoever holds the token has the power" — see [Permissions](./access-control.md).

## List your namespaces

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

## Deploy an app

`Apply` is a streaming call — it reports deployment progress until `Metadata.End` is set. Pass `ExtraValues` to override chart values on the fly.

```go
input := &project.ApplyRequest{
    NamespaceId: 20,   // shown when hovering a namespace name in the UI
    RepoId:      107,  // visible on the repository management page
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

## Stream container logs

Server-streaming over HTTP emits NDJSON / SSE — the SDK handles both transparently. `io.EOF` means the stream ended normally.

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

## Copy a file into a pod

Stream file chunks into `StreamCopyToPod`. If the pod runs several containers, pass `Container`; otherwise the default container is used.

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

## HTTP extras: upload & download

File transfer is HTTP-only — the gRPC File service has no upload RPC.

```go
// upload, returns a file id for later download
resp, err := cli.File().UploadFile(ctx, name, file)
fmt.Println("file id:", resp.ID)

// download, returns an io.ReadCloser; filename/size come from response headers
rc, info, err := cli.File().DownloadFile(ctx, id)
defer rc.Close()
fmt.Printf("file=%s size=%d\n", info.Filename, info.Size)
```

## Error handling

Errors carry gRPC status codes. Check them with `status.Code` / `status.Convert` — HTTP and gRPC clients return the same codes.

```go
_, err = cli.Namespace().Show(ctx, &namespace.ShowRequest{Id: -1})
if status.Code(err) == codes.NotFound {
    fmt.Println("Show(id=-1) -> codes.NotFound")
}
```

## Run the examples

Every capability has a runnable example in the repository under `examples/`:

| Example | What it shows | Run it |
|---|---|---|
| `examples/http` | list namespaces, stream logs, upload / download files | `go run ./examples/http` |
| `examples/apply` | deploy an app and scale replicas (streaming) | `go run ./examples/apply` |
| `examples/copyfile` | copy a file into a pod | `go run ./examples/copyfile` |
| `examples/exec` | interactive shell in a container | `go run ./examples/exec` |
| `examples/execonce` | run a one-off command in a container | `go run ./examples/execonce` |

> The examples default to `localhost`. Point them at your Mars host, e.g. `go run ./examples/http -addr http://your-host:4000`.
