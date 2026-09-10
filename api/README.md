# mars API SDK

<div align="center">

English | [简体中文](README_zh-CN.md)

</div>

The client SDK module for mars (`github.com/duc-cnzj/mars/api/v6`). It ships three clients — **gRPC**, **HTTP/JSON** (grpc-gateway) and **WebSocket** — that share the same set of proto-generated types: method signatures, return types and error codes are fully aligned, so switching transports requires no changes to your business code. WebSocket (`api/ws`) carries the one capability gRPC and HTTP cannot express: **bidirectional, real-time container terminal** access (single public entry point `OpenTerminal`).

## Three transports, one set of types

| Dimension | gRPC SDK (`api/grpc`) | HTTP SDK (`api/http`) | WebSocket SDK (`api/ws`) |
|---|---|---|---|
| Transport | HTTP/2 gRPC | HTTP/1.1 JSON (grpc-gateway) | WebSocket (binary protobuf frames) |
| Client | `grpc.NewClient(addr, opts...)` | `http.NewClient(baseURL, opts...)` | `ws.NewClient(url, opts...)` |
| Accessor | `cli.Namespace().List(ctx, req)` | `cli.Namespace().List(ctx, req)` | `cli.OpenTerminal(ctx, container)` terminal entry point |
| Streaming | native gRPC stream | server-streaming → SSE/NDJSON |
| Server needed | mars gRPC port (e.g. `:50000`) | mars gateway port (e.g. `:4000`) |

Both packages expose the same 17 service accessors: `Auth/Repo/Changelog/Cluster/Container/Event/AccessToken/File/Git/Metrics/Namespace/Picture/Project/Version/Endpoint/Settings/User`.

### Capability differences: gRPC-only vs HTTP-only

gRPC has **84** methods, HTTP also has **84** (81 generated from proto `google.api.http` annotations + 3 hand-written); **81 are shared** (every generated HTTP stub has a matching gRPC counterpart with an identical signature). There are only two kinds of difference, and both are called out explicitly in the generator or the hand-written source, so you can verify them yourself.

**gRPC-only (3) — no HTTP route exists:**

| Method | Streaming kind | Why HTTP lacks it |
|---|---|---|
| `Container.Exec` | bidi | Unrepresentable in HTTP/JSON; needs WebSocket (the mars ws channel carries the terminal) |
| `Container.StreamCopyToPod` | client | Unrepresentable in HTTP/JSON; needs WebSocket |
| `Project.Apply` | server | proto has **no `google.api.http` annotation**, so the gateway does not expose it; the HTTP-side alternative is `Project.WebApply` |

> The first two are impossible purely because of their streaming direction (client/bidi streaming) under HTTP/1.1 JSON; `Project.Apply` is server-streaming but its `.proto` carries no http annotation — the gateway has no route for it at all, so the HTTP SDK naturally has no method. `Container.ExecOnce` / `Container.StreamContainerLog` / `Metrics.StreamTopPod` are all server-streaming methods that **do** carry the annotation, so both SDKs offer them (native gRPC stream on one side, HTTP SSE on the other).

**HTTP-only (3) — defined in no proto at all, with no gRPC counterpart:**

| Method | HTTP route | Notes |
|---|---|---|
| `FileAPI.UploadFile` | `POST /api/files` | multipart upload, returns a file ID |
| `FileAPI.DownloadFile` | `GET /api/download_file/{id}` | binary download, returns a stream plus metadata |
| `FileAPI.CopyFromPod` | `POST /api/copy_from_pod` | copy a file from a pod to the caller |

> Mind the direction: `Container.CopyToPod` (copy *into* a pod) exists in both SDKs, while `FileAPI.CopyFromPod` (copy *out of* a pod) is HTTP-only. The gRPC `File` service only has `List/Delete/DiskInfo/MaxUploadSize/ShowRecords` — it has **no** upload/download RPC.

## Installation

```bash
go get -u github.com/duc-cnzj/mars/api/v6/grpc
go get -u github.com/duc-cnzj/mars/api/v6/http
```

## gRPC usage

```go
package main

import (
	"context"

	"github.com/duc-cnzj/mars/api/v6/grpc"
	"github.com/duc-cnzj/mars/api/v6/proto/namespace"
)

func main() {
	// If WithAuth is configured, the client logs in immediately to obtain a token
	// and returns an error on failure.
	c, err := grpc.NewClient("127.0.0.1:50000",
		grpc.WithAuth("admin", "123456"),
		grpc.WithTokenAutoRefresh(), // re-login and retry on 401 (5 exponential backoffs by default)
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

### gRPC options

| Option | Effect |
|---|---|
| `WithAuth(username, password)` | Log in at construction time and attach the token to the Authorization metadata of every RPC |
| `WithBearerToken(token)` | Inject an already-issued token directly (the `Bearer` prefix is added automatically) |
| `WithTokenAutoRefresh()` | On `codes.Unauthenticated` (and with WithAuth configured), re-login and retry, covering unary and server-streaming; a 401 from Login/Exchange itself is returned as-is (retrying bad credentials is pointless and would also self-deadlock the singleflight) |
| `WithUnaryClientInterceptor(op)` | Append a unary interceptor |
| `WithStreamClientInterceptor(op)` | Append a streaming interceptor |
| `WithTracer()` | Wire up OpenTelemetry (otelgrpc client stats handler) |
| `WithTransportCredentials(tlsCfg)` | Establish a TLS connection using a custom `tls.Config` (mTLS included); plaintext insecure is never the default |

Replacing the token at runtime: `c.SetBearerToken("...")`.

## HTTP usage

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

> Fully runnable examples live in [`examples/http`](../examples/http) at the repository root: unary calls,
> error-code alignment, server-streaming (SSE), and the HTTP-only capabilities (multipart upload / binary
> download). The examples connect to the gateway on `:4000`; `mars serve`'s `app_port` defaults to `:6000`,
> while `:4000` is typical for a local port-forward or docker port mapping.

### HTTP options

| Option | Effect |
|---|---|
| `WithAuth(username, password)` | Exchange credentials for a token via `POST /api/auth/login` at construction time |
| `WithBearerToken(token)` | Inject an already-issued token directly (the `Bearer` prefix is added automatically) |
| `WithTokenAutoRefresh()` | On 401 (and with WithAuth configured), re-login and retry once; a 401 from Login/Exchange itself is returned as-is |
| `WithHTTPClient(hc)` | Replace the underlying `*http.Client` (useful for injecting a custom transport) |
| `WithHeader(key, value)` | Attach a custom header to every request (injected at construction time, immutable afterwards); a same-named header overrides the SDK's automatic `Content-Type`/`Accept`/`Authorization` (Set semantics, applied last); an empty key is ignored. Handy for correlation IDs such as `X-Request-ID` or for passing headers through to your business logic |
| `WithHeaders(headers)` | Attach a batch of custom headers, same semantics as `WithHeader` |
| `WithTimeout(d)` | Set the overall timeout of the underlying http.Client |
| `WithTracer()` | Wire up OpenTelemetry; requests carry traces (the Transport is wrapped with otelhttp under the hood) |

## Option comparison

Both SDKs share `WithAuth` / `WithBearerToken` / `WithTokenAutoRefresh` / `WithTracer` with identical semantics. Transport-specific differences:

| Transport-specific | Notes |
|---|---|
| gRPC `WithTransportCredentials(tlsCfg)` | No equivalent needed on the HTTP side (swapping the transport via `WithHTTPClient` covers TLS/proxies) |
| gRPC `WithUnaryClientInterceptor` / `WithStreamClientInterceptor` | Interceptor injection is a native gRPC mechanism with no HTTP counterpart |
| HTTP `WithHTTPClient(hc)` / `WithTimeout(d)` | Direct control over `*http.Client`; no gRPC counterpart (connection settings go through dial options) |
| HTTP `WithHeader` / `WithHeaders` | Client-level custom headers that override the SDK's automatic headers (Set semantics); no gRPC counterpart (custom metadata goes through interceptors) |

## WebSocket SDK (`api/ws`)

`api/ws` targets a single core scenario: **spawning an interactive shell inside a given container and reading from / writing to it** (the bidi, real-time capability gRPC and HTTP cannot express). It therefore exposes exactly **one entry point**, `OpenTerminal` — connecting, authenticating, generating the sessionID, opening the shell and handling the auth race are all handled internally, so the caller only receives a readable/writable `Terminal`. The endpoint is `ws(s)://<host>/ws`, authenticated with the same JWT used by HTTP/gRPC.

```go
cli, err := ws.NewClient("ws://127.0.0.1:4000/ws", ws.WithAuth("admin", "123456"))
if err != nil { panic(err) }
defer cli.Close()

// A single call performs the whole "connect → authenticate → open terminal" interaction;
// the sessionID is generated automatically by the SDK.
ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
defer cancel()
term, err := cli.OpenTerminal(ctx, &websocket.Container{Namespace: "ns", Pod: "p", Container: "c"})
if err != nil { panic(err) }
defer term.Close()

// Data plane: one call wires up all three channels — stdin→remote, remote→stdout, and
// remote→toast (raw mode by default, see the raw-mode notes below). The returned stop
// halts forwarding and restores the local terminal, so it must be deferred.
stop := term.Pump(os.Stdin,
	func(d []byte) { _, _ = os.Stdout.Write(d) },
	func(d []byte) { /* out-of-band messages (toast) */ },
)
defer stop()

// Control plane: automatically follow local terminal window-size changes
// (initial size plus SIGWINCH) and sync them to the remote pty.
term.AutoHandleWindowSize()

// Manual operation is still available: term.Write(p) sends stdin, term.Stdout() receives
// output, term.Resize(h, w), and term.ID() returns the auto-generated sessionID.

// Session ends (process exit / kicked / explicit Close)
<-term.Done()
```

**Raw mode (on by default)**: `Pump` switches the local terminal to raw mode by default (disabling local line buffering and echo, so every keystroke byte is passed straight through to the remote shell) and lets the remote readline interpret it, which gives you **tab completion, arrow keys, clear** and the rest of the interactive experience. The trade-off is that **in raw mode Ctrl+C is just a `0x03` byte** forwarded to the remote shell (which then raises SIGINT) — exit with the remote `exit` command rather than local Ctrl+C. Use `ws.WithRawMode(false)` to fall back to canonical mode (local echo, Ctrl+C raises a local SIGINT). `stop` restores the local terminal settings automatically.

- `Client` keeps a background goroutine alive, reconnecting and re-authenticating with a backoff strategy on disconnect; `Close()` is idempotent.
- The `Terminal` high-level abstraction: `Pump(in, stdout, toast, opts...)` (data-plane orchestration, returns stop, raw mode by default), `AutoHandleWindowSize()` (follows the local window size), `Write` (stdin) / `Resize` / `Stdout` / `Toast` / `Done` / `Close` / `ID` (the auto-generated sessionID).
- Fully runnable examples live in [`examples/ws`](../examples/ws) at the repository root.

### WebSocket options

| Option | Notes |
|---|---|
| `WithBearerToken(token)` | Inject an already-issued JWT; HandleAuthorize is sent on connect |
| `WithAuth(user, pass)` | Log in through `api/http` on connect to obtain a token, auto-renewed on every reconnect |
| `WithTokenProvider(fn)` | Custom token source — the most flexible option (caching, OIDC exchange, etc.) |
| `WithHTTPClient(hc)` | Injects the underlying `*http.Client` used solely for the `WithAuth` login |
| `WithDialer(d)` | Inject a custom ws dialer (TLS/proxy/handshake timeout) |
| `WithReconnectBackoff(b)` | Custom backoff strategy for reconnects |

## Server-streaming

Under HTTP/JSON, server-streaming methods carrying a `google.api.http` annotation are emitted by the gateway as NDJSON (`{"result": <msg>}`) or standard SSE (`data: {...}`); the SDK accepts both formats transparently:

```go
// e.g. container.StreamContainerLog / StreamTopPod
stream, err := c.Container().StreamContainerLog(ctx, &container.StreamContainerLogRequest{...})
if err != nil {
	panic(err)
}
defer stream.Close()
for {
	msg, err := stream.Recv() // io.EOF = clean end of stream
	if err != nil {
		if errors.Is(err, io.EOF) {
			break
		}
		panic(err)
	}
	_ = msg
}
```

Mid-stream errors come back as a `google.rpc.Status` envelope and are restored to a `codes.Error`, so error codes work exactly as they do for unary calls. Only three methods are unreachable over HTTP/JSON: the client/bidi streaming `Exec` and `StreamCopyToPod` (which need WebSocket), plus `Project.Apply`, which carries no http annotation — the generator honestly skips them, leaving them available in the gRPC SDK only.

## Generation workflow

After changing a proto, re-run the generator; the output lands in `api/http/rest/`:

```bash
cd api
go generate ./http/...      # triggers go:generate → go run ./gen/cmd
go run ./http/gen/cmd       # equivalent to running it by hand
```

The generator guarantees that **the rest/ directory is 100% equivalent to the output of the current proto**:

- unary + `google.api.http` annotation → generate an HTTP stub;
- server-streaming + annotation → generate an SSE stub;
- client/bidi streaming, un-annotated methods, custom routes → skip, or hand-write;
- after generating, sweep orphan `*.gen.http.go` files in rest/ that are not part of the set (deleting a service from the proto leaves no junk behind), and never touch hand-written files.

`TestGeneratedStubsUpToDate` in `api/http/gen_test.go` performs a two-way drift check: committed stubs ⊆ generator output, and no orphans. After changing a proto you must regenerate before committing, or the test fails.

## Quality assurance

```bash
go build ./...             # full compile
go vet ./...               # static analysis
go test ./...              # unit tests (hand-written production code at 100%; rest/ generated stubs are covered by the drift test)
```

- grpc package: bufconn in-memory gRPC tests covering login / token prefix / auto-refresh / singleflight dedup / interceptor injection / the 17 service accessors;
- http/transport package: fake stream/conn covering the generic stream factory and error propagation;
- internal/flight package: full coverage of singleflight dedup (concurrent calls on the same key execute once, the rest share the result, and nothing is cached after completion).

## See also

- proto definitions and gateway server: repository root `api/proto/`, `internal/...`;
- More usage examples: [examples](https://github.com/duc-cnzj/mars/tree/master/examples).
</content>
</invoke>
