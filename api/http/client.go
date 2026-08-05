// Package http 提供基于 HTTP/JSON 的 grpc-gateway 客户端。
//
// 与 api/grpc/client.go 的 gRPC SDK 共享同一套 proto 生成类型：
// 方法签名、返回类型、错误码全部对齐，调用方切换传输方式时业务代码无需改动。
//
// 每个 service 的 stub 由 gen 从 proto 的 google.api.http 注解生成
// （*.gen.http.go），不要手改。proto 变更后执行：
//
//	go generate ./http/...
//
// 用法：
//
//	cli, err := http.NewHTTPClient(
//	    "http://localhost:6000",
//	    http.WithAuth("admin", "password"),
//	    http.WithTokenAutoRefresh(),
//	)
//	ns, err := cli.Namespace().Create(ctx, &namespace.CreateRequest{...})
//
//go:generate go run ./gen/cmd
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/duc-cnzj/mars/api/v6/http/rest"
	"github.com/duc-cnzj/mars/api/v6/http/transport"
	"github.com/duc-cnzj/mars/api/v6/internal/flight"
	"github.com/duc-cnzj/mars/api/v6/proto/auth"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Client 指向 grpc-gateway 的 HTTP 客户端。
type Client struct {
	baseURL string
	hc      *http.Client

	username, password string
	token              atomic.Value // 存 "Bearer xxx"
	autoRefresh        bool
	tracing            bool

	flights *flight.Group // 合并并发 token 刷新，避免登录风暴
}

// Option 修改 Client 行为。
type Option func(*Client)

// WithAuth 构造时用用户名密码 POST /api/auth/login 换取 token。
func WithAuth(username, password string) Option {
	return func(c *Client) {
		c.username = username
		c.password = password
	}
}

// WithBearerToken 直接注入已签发的 token（自动补 Bearer 前缀）。
func WithBearerToken(token string) Option {
	return func(c *Client) { c.setToken(token) }
}

// WithTokenAutoRefresh 遇到 401 且配置了 WithAuth 时自动重新登录并重试一次。
func WithTokenAutoRefresh() Option {
	return func(c *Client) { c.autoRefresh = true }
}

// WithHTTPClient 替换底层 http.Client（可注入自定义 transport）。
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.hc = hc }
}

// WithTimeout 设置底层 http.Client 的整体超时（含读 body）。
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if c.hc == nil {
			c.hc = &http.Client{Timeout: d}
			return
		}
		c.hc.Timeout = d
	}
}

// WithTracer 接入 OpenTelemetry，为每个 HTTP 请求附加 trace span。
// 底层把 http.Client.Transport 包一层 otelhttp.NewTransport；
// 未显式设置 Transport（或未用 WithHTTPClient）时包装默认 transport。
func WithTracer() Option {
	return func(c *Client) { c.tracing = true }
}

// NewHTTPClient 创建一个指向 grpc-gateway 的 HTTP 客户端。
// baseURL 形如 "http://localhost:6000"，末尾可带斜杠。
// 配置了 WithAuth 时会在构造阶段完成登录，登录失败会返回错误。
func NewHTTPClient(baseURL string, opts ...Option) (*Client, error) {
	c := &Client{baseURL: strings.TrimRight(baseURL, "/"), flights: &flight.Group{}}
	for _, o := range opts {
		o(c)
	}
	if c.hc == nil {
		c.hc = &http.Client{}
	}
	if c.tracing {
		base := c.hc.Transport
		if base == nil {
			base = http.DefaultTransport
		}
		c.hc.Transport = otelhttp.NewTransport(base)
	}
	if c.username != "" || c.password != "" {
		if err := c.refreshToken(); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// Close 释放空闲连接。
func (c *Client) Close() error {
	c.hc.CloseIdleConnections()
	return nil
}

// SetBearerToken 运行期替换已签发的 token（自动补 Bearer 前缀）。
// 与 gRPC SDK（api/grpc/client.go）的 Interface.SetBearerToken 对齐。
func (c *Client) SetBearerToken(token string) {
	c.setToken(token)
}

func (c *Client) setToken(token string) {
	// 匹配 "bearer " 含空格，避免 "bearertok" 这类恰好以 bearer 开头但缺空格的
	// token 被误判已带前缀（蓝军审计发现，Wave 25）。
	if token != "" && !strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = "Bearer " + token
	}
	c.token.Store(token)
}

func (c *Client) authToken() string {
	if v := c.token.Load(); v != nil {
		return v.(string)
	}
	return ""
}

// refreshToken 用用户名密码重新登录。内部走 doNoRefresh，避免刷新递归。
// 用 flight 包（internal/flight）合并并发的 401 刷新：同一时刻只有一个 goroutine 真正打登录接口，
// 其余等待复用结果，防止 401 风暴时登录被并发打爆。setToken 放在 fn 里保证只执行一次。
func (c *Client) refreshToken() error {
	if c.username == "" || c.password == "" {
		return errors.New("http: WithAuth 需要 username/password 才能自动登录")
	}
	_, err, _ := c.flights.Do("refresh", func() (interface{}, error) {
		out, err := c.Auth().Login(context.Background(), &auth.LoginRequest{
			Username: c.username,
			Password: c.password,
		})
		if err != nil {
			return nil, err
		}
		c.setToken(out.Token)
		return nil, nil
	})
	return err
}

// do 发送一次请求，401 且开启 autoRefresh 时自动重登并重试一次。
func (c *Client) do(ctx context.Context, method, path string, req, resp proto.Message) error {
	return c.doReq(ctx, method, path, req, resp, true, false)
}

// doNoRefresh 不触发自动重登，用于登录/换 token 等元请求。
func (c *Client) doNoRefresh(ctx context.Context, method, path string, req, resp proto.Message) error {
	return c.doReq(ctx, method, path, req, resp, false, false)
}

// doQuery 将请求字段全部编码为 query 参数，适用于无 body 绑定的方法（含 POST/DELETE）。
func (c *Client) doQuery(ctx context.Context, method, path string, req, resp proto.Message) error {
	return c.doReq(ctx, method, path, req, resp, true, true)
}

func (c *Client) doReq(ctx context.Context, method, path string, req, resp proto.Message, allowRefresh, forceQuery bool) error {
	u := c.baseURL + path
	var body io.Reader
	switch {
	case forceQuery:
		// 无 body 绑定：请求字段扁平化为 query，与 gateway query binding 语义一致。
		// 生成器对 GET/DELETE 等无 body 方法统一走 doQuery，因此此处只有 doQuery 一条 query 路径。
		if q := encodeQuery(req); q != "" {
			u += "?" + q
		}
	default:
		if req != nil {
			data, err := protojson.Marshal(req)
			if err != nil {
				return err
			}
			body = bytes.NewReader(data)
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return err
	}
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	if tok := c.authToken(); tok != "" {
		httpReq.Header.Set("Authorization", tok)
	}

	resp2, err := c.hc.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp2.Body.Close()

	if resp2.StatusCode >= 200 && resp2.StatusCode < 300 {
		if resp == nil {
			return nil
		}
		data, err := io.ReadAll(resp2.Body)
		if err != nil {
			return err
		}
		// DiscardUnknown：服务端未来新增字段不影响老客户端，与 gateway 服务端行为一致。
		return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(data, resp)
	}

	data, _ := io.ReadAll(resp2.Body)
	if ge, ok := parseGatewayError(data); ok {
		if allowRefresh && c.autoRefresh && codes.Code(ge.Code) == codes.Unauthenticated && c.username != "" {
			if err := c.refreshToken(); err != nil {
				return err
			}
			return c.doReq(ctx, method, path, req, resp, false, forceQuery)
		}
		return status.Error(codes.Code(ge.Code), ge.Message)
	}
	return fmt.Errorf("http: unexpected status %d: %s", resp2.StatusCode, strings.TrimSpace(string(data)))
}

// gatewayError 是 grpc-gateway v2 默认错误响应体 {"code":<grpc code>,"message":...}。
type gatewayError struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Details []json.RawMessage `json:"details"`
}

func parseGatewayError(data []byte) (gatewayError, bool) {
	var ge gatewayError
	if err := json.Unmarshal(data, &ge); err != nil {
		return ge, false
	}
	if ge.Code == 0 {
		return ge, false
	}
	return ge, true
}

// errFromStatus 把非 2xx 响应映射为 codes.Error（gateway 错误体）或普通错误。
// 供 Upload/Download 等不走 do() 的原始请求复用。
func (c *Client) errFromStatus(statusCode int, body []byte) error {
	if ge, ok := parseGatewayError(body); ok {
		return status.Error(codes.Code(ge.Code), ge.Message)
	}
	return fmt.Errorf("http: unexpected status %d: %s", statusCode, strings.TrimSpace(string(body)))
}

// encodeQuery 将 proto 请求扁平化为 gateway 期望的 query 参数。
// 字段名用 JSON name（camelCase），与 gateway JSONPb 的 query binding 一致；
// 嵌套消息展开为 a.b=c；repeated 展开为 a=v1&a=v2；零值跳过（proto3 unset 语义）。
func encodeQuery(msg proto.Message) string {
	if msg == nil {
		return ""
	}
	var parts []string
	appendQuery(msg.ProtoReflect(), "", &parts)
	return strings.Join(parts, "&")
}

func appendQuery(m protoreflect.Message, prefix string, parts *[]string) {
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		name := string(fd.JSONName())
		if prefix != "" {
			name = prefix + "." + name
		}
		switch {
		case fd.IsList():
			list := v.List()
			for i := 0; i < list.Len(); i++ {
				*parts = append(*parts, name+"="+scalarString(fd, list.Get(i)))
			}
		case fd.IsMap():
			// query 里 map 极少见，暂不支持
		case fd.Kind() == protoreflect.MessageKind || fd.Kind() == protoreflect.GroupKind:
			appendQuery(v.Message(), name, parts)
		default:
			if isZeroScalar(fd, v) {
				return true
			}
			*parts = append(*parts, name+"="+scalarString(fd, v))
		}
		return true
	})
}

func isZeroScalar(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
	switch fd.Kind() {
	case protoreflect.StringKind:
		return v.String() == ""
	case protoreflect.BoolKind:
		return !v.Bool()
	case protoreflect.EnumKind:
		return v.Enum() == 0
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return v.Int() == 0
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return v.Uint() == 0
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return v.Float() == 0
	}
	return false
}

func scalarString(fd protoreflect.FieldDescriptor, v protoreflect.Value) string {
	switch fd.Kind() {
	case protoreflect.StringKind:
		return url.QueryEscape(v.String())
	case protoreflect.BoolKind:
		if v.Bool() {
			return "true"
		}
		return "false"
	case protoreflect.EnumKind:
		return strconv.FormatInt(int64(v.Enum()), 10)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return strconv.FormatInt(v.Int(), 10)
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return strconv.FormatUint(v.Uint(), 10)
	case protoreflect.FloatKind:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32)
	case protoreflect.DoubleKind:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	}
	return fmt.Sprintf("%v", v.Interface())
}

// ops 是 transport.Conn 的适配器：把 Client 的引擎能力暴露给生成 stub（rest/*.gen.http.go），
// 同时保持 Client 公共 API 零膨胀。
type ops struct{ c *Client }

func (o ops) Do(ctx context.Context, method, path string, req, resp proto.Message) error {
	return o.c.do(ctx, method, path, req, resp)
}

func (o ops) DoNoRefresh(ctx context.Context, method, path string, req, resp proto.Message) error {
	return o.c.doNoRefresh(ctx, method, path, req, resp)
}

func (o ops) DoQuery(ctx context.Context, method, path string, req, resp proto.Message) error {
	return o.c.doQuery(ctx, method, path, req, resp)
}

func (o ops) OpenStream(ctx context.Context, method, path string, req proto.Message) (transport.RawStream, error) {
	return o.c.openStream(ctx, method, path, req)
}

// 各 service 的访问器。调用方式与 gRPC SDK（api/grpc/client.go）对齐：cli.Namespace().List(ctx, req)。
// File 除了生成 CRUD 还挂手写自定义路由（multipart 上传/二进制下载/copy_from_pod），
// 通过 FileAPI 嵌入 *rest.FileSvc 实现，调用同样走 cli.File()，见 file_custom.go。
func (c *Client) AccessToken() *rest.AccessTokenSvc { return &rest.AccessTokenSvc{C: ops{c}} }
func (c *Client) Auth() *rest.AuthSvc               { return &rest.AuthSvc{C: ops{c}} }
func (c *Client) Changelog() *rest.ChangelogSvc     { return &rest.ChangelogSvc{C: ops{c}} }
func (c *Client) Cluster() *rest.ClusterSvc         { return &rest.ClusterSvc{C: ops{c}} }
func (c *Client) Container() *rest.ContainerSvc     { return &rest.ContainerSvc{C: ops{c}} }
func (c *Client) Endpoint() *rest.EndpointSvc       { return &rest.EndpointSvc{C: ops{c}} }
func (c *Client) Event() *rest.EventSvc             { return &rest.EventSvc{C: ops{c}} }
func (c *Client) File() *FileAPI                    { return &FileAPI{FileSvc: &rest.FileSvc{C: ops{c}}, c: c} }
func (c *Client) Git() *rest.GitSvc                 { return &rest.GitSvc{C: ops{c}} }
func (c *Client) Metrics() *rest.MetricsSvc         { return &rest.MetricsSvc{C: ops{c}} }
func (c *Client) Namespace() *rest.NamespaceSvc     { return &rest.NamespaceSvc{C: ops{c}} }
func (c *Client) Picture() *rest.PictureSvc         { return &rest.PictureSvc{C: ops{c}} }
func (c *Client) Project() *rest.ProjectSvc         { return &rest.ProjectSvc{C: ops{c}} }
func (c *Client) Repo() *rest.RepoSvc               { return &rest.RepoSvc{C: ops{c}} }
func (c *Client) Version() *rest.VersionSvc         { return &rest.VersionSvc{C: ops{c}} }
