// Package grpc 提供基于 gRPC 的 mars 客户端 SDK。
//
// 与 api/http 的 HTTP/JSON 客户端共享同一套 proto 生成类型，方法签名、返回类型、
// 错误码全部对齐，调用方切换传输方式时业务代码无需改动。
//
// 支持认证（WithAuth/WithBearerToken）、token 自动刷新（WithTokenAutoRefresh）、
// 链路追踪（WithTracer）与拦截器注入，并用 singleflight 合并并发的 token 刷新。
package grpc

import (
	"context"
	"crypto/tls"
	"io"
	"strings"
	"sync/atomic"

	"github.com/cenkalti/backoff/v4"
	"github.com/duc-cnzj/mars/api/v6/internal/flight"
	"github.com/duc-cnzj/mars/api/v6/proto/auth"
	"github.com/duc-cnzj/mars/api/v6/proto/changelog"
	"github.com/duc-cnzj/mars/api/v6/proto/cluster"
	"github.com/duc-cnzj/mars/api/v6/proto/container"
	"github.com/duc-cnzj/mars/api/v6/proto/endpoint"
	"github.com/duc-cnzj/mars/api/v6/proto/event"
	"github.com/duc-cnzj/mars/api/v6/proto/file"
	"github.com/duc-cnzj/mars/api/v6/proto/git"
	"github.com/duc-cnzj/mars/api/v6/proto/metrics"
	"github.com/duc-cnzj/mars/api/v6/proto/namespace"
	"github.com/duc-cnzj/mars/api/v6/proto/picture"
	"github.com/duc-cnzj/mars/api/v6/proto/project"
	"github.com/duc-cnzj/mars/api/v6/proto/repo"
	"github.com/duc-cnzj/mars/api/v6/proto/token"
	"github.com/duc-cnzj/mars/api/v6/proto/version"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Interface 是 mars gRPC 客户端的门面：一个连接暴露全部 15 个 service 客户端，
// 并支持 Close 与运行期替换 token。
type Interface interface {
	io.Closer
	SetBearerToken(string)

	Auth() auth.AuthClient
	Picture() picture.PictureClient
	Version() version.VersionClient
	Cluster() cluster.ClusterClient
	Changelog() changelog.ChangelogClient
	Event() event.EventClient
	Container() container.ContainerClient
	File() file.FileClient
	Git() git.GitClient
	Namespace() namespace.NamespaceClient
	Project() project.ProjectClient
	Endpoint() endpoint.EndpointClient
	Metrics() metrics.MetricsClient
	AccessToken() token.AccessTokenClient
	Repo() repo.RepoClient
}

// Client 是 Interface 的默认实现。内部状态字段不导出，行为由 Option 配置；
// UnaryClientInterceptors/StreamClientInterceptors 两个导出切片仅便于测试注入，
// 运行期仍只经 Option 追加。
type Client struct {
	singleflight flight.Group

	UnaryClientInterceptors  []grpc.UnaryClientInterceptor
	StreamClientInterceptors []grpc.StreamClientInterceptor
	username, password       string
	authTokenValue           atomic.Value

	tls *tls.Config

	conn        *grpc.ClientConn
	dialOptions []grpc.DialOption

	auth        auth.AuthClient
	changelog   changelog.ChangelogClient
	cluster     cluster.ClusterClient
	container   container.ContainerClient
	endpoint    endpoint.EndpointClient
	event       event.EventClient
	file        file.FileClient
	git         git.GitClient
	metrics     metrics.MetricsClient
	namespace   namespace.NamespaceClient
	picture     picture.PictureClient
	project     project.ProjectClient
	repo        repo.RepoClient
	accessToken token.AccessTokenClient
	version     version.VersionClient
}

var _ Interface = (*Client)(nil)

// noAutoRefreshMethods 是禁止自动刷新重试的方法全集：这些方法自身的 401 表示凭据错误，
// 刷新 token 无济于事，且会递归触发 getToken → singleflight.Do 形成自我死锁。
var noAutoRefreshMethods = map[string]struct{}{
	auth.Auth_Login_FullMethodName:    {},
	auth.Auth_Exchange_FullMethodName: {},
}

// NewClient 建立到 addr（host:port）的 gRPC 连接并返回门面客户端。
// 配置了 WithAuth 时会在构造阶段完成登录换取 token，失败返回错误。
func NewClient(addr string, opts ...Option) (Interface, error) {
	c := &Client{}

	for _, opt := range opts {
		opt(c)
	}

	// buildDialOptions 恒注入传输凭据（tls==nil 时 insecure，否则 WithTransportCredentials 已加 TLS）。
	// grpc.NewClient 对畸形 target（如含非法 percent-encoding 的 "%zz"）会构造失败返回 error，
	// 必须透传而非丢弃——丢弃会留下 conn==nil 的 client，首次 RPC nil pointer panic
	// （蓝军交叉审计实证，Wave 25；此前探针采样漏掉该 target 类别）。
	dial, err := grpc.NewClient(addr, c.buildDialOptions()...)
	if err != nil {
		return nil, err
	}
	c.conn = dial

	c.auth = auth.NewAuthClient(dial)
	c.changelog = changelog.NewChangelogClient(dial)
	c.cluster = cluster.NewClusterClient(dial)
	c.container = container.NewContainerClient(dial)
	c.event = event.NewEventClient(dial)
	c.git = git.NewGitClient(dial)
	c.metrics = metrics.NewMetricsClient(dial)
	c.namespace = namespace.NewNamespaceClient(dial)
	c.picture = picture.NewPictureClient(dial)
	c.project = project.NewProjectClient(dial)
	c.version = version.NewVersionClient(dial)
	c.file = file.NewFileClient(dial)
	c.endpoint = endpoint.NewEndpointClient(dial)
	c.accessToken = token.NewAccessTokenClient(dial)
	c.repo = repo.NewRepoClient(dial)

	if c.password != "" || c.username != "" {
		if err := c.getToken(); err != nil {
			// conn 已创建，登录失败时释放，避免构造失败路径泄漏 gRPC 连接。
			_ = c.Close()
			return nil, err
		}
	}

	return c, nil
}

// SetBearerToken 运行期替换已签发的 token（自动补 Bearer 前缀）。
func (c *Client) SetBearerToken(token string) {
	c.setToken(token)
}

func (c *Client) hasCredentials() bool {
	return c.username != "" && c.password != ""
}

// Close 关闭底层 gRPC 连接。
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

// 各 service 访问器：返回对应 service 的 gRPC 客户端，方法与 HTTP SDK（api/http/client.go）
// 对齐（cli.Namespace().List(ctx, req)）。连接生命周期由 Client 统一管理。
func (c *Client) Auth() auth.AuthClient {
	return c.auth
}

func (c *Client) Repo() repo.RepoClient {
	return c.repo
}

func (c *Client) Changelog() changelog.ChangelogClient {
	return c.changelog
}

func (c *Client) Cluster() cluster.ClusterClient {
	return c.cluster
}

func (c *Client) Container() container.ContainerClient {
	return c.container
}

func (c *Client) Event() event.EventClient {
	return c.event
}

func (c *Client) AccessToken() token.AccessTokenClient {
	return c.accessToken
}

func (c *Client) File() file.FileClient {
	return c.file
}

func (c *Client) Git() git.GitClient {
	return c.git
}

func (c *Client) Metrics() metrics.MetricsClient {
	return c.metrics
}

func (c *Client) Namespace() namespace.NamespaceClient {
	return c.namespace
}

func (c *Client) Picture() picture.PictureClient {
	return c.picture
}

func (c *Client) Project() project.ProjectClient {
	return c.project
}

func (c *Client) Version() version.VersionClient {
	return c.version
}

func (c *Client) Endpoint() endpoint.EndpointClient {
	return c.endpoint
}

func (c *Client) authToken() string {
	v := c.authTokenValue.Load()
	if v != nil {
		return v.(string)
	}
	return ""
}

func (c *Client) buildDialOptions() []grpc.DialOption {
	if c.tls == nil {
		c.dialOptions = append(c.dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	if !c.hasCredentials() && c.authTokenValue.Load() == nil {
		c.dialOptions = append(c.dialOptions, grpc.WithPerRPCCredentials(&clientauth{c: c}))
	}

	c.dialOptions = append(c.dialOptions,
		grpc.WithChainStreamInterceptor(c.StreamClientInterceptors...),
		grpc.WithChainUnaryInterceptor(c.UnaryClientInterceptors...),
	)

	return c.dialOptions
}

func (c *Client) getToken() error {
	login, err, _ := c.singleflight.Do("Retry", func() (interface{}, error) {
		return c.auth.Login(context.TODO(), &auth.LoginRequest{
			Username: c.username,
			Password: c.password,
		})
	})
	if err != nil {
		return err
	}

	c.setToken(login.(*auth.LoginResponse).Token)
	return nil
}

func (c *Client) setToken(token string) {
	// 与 http SDK（api/http/client.go setToken）语义保持一致：
	// 非空且未带 "Bearer " 前缀（大小写不敏感）时自动补前缀。
	// 匹配含空格，避免 "bearertok" 这类恰好以 bearer 开头但缺空格的 token 被误判已带前缀。
	if token != "" && !strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = "Bearer " + token
	}
	c.authTokenValue.Store(token)
}

// skipAutoRefresh 判断某次调用是否应跳过自动刷新重试：
// 未配置凭据时无 token 可刷新；Login/Exchange 自身 401 表示凭据错误，重试救不回来，
// 且会递归进入 getToken → flight.Do 形成自我死锁，必须原样返回错误。
func (c *Client) skipAutoRefresh(method string) bool {
	if !c.hasCredentials() {
		return true
	}
	_, skip := noAutoRefreshMethods[method]
	return skip
}

// Option 以函数式配置修改 Client。
type Option func(*Client)

type clientauth struct {
	c *Client
}

func (a *clientauth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"Authorization": a.c.authToken(),
	}, nil
}

func (a *clientauth) RequireTransportSecurity() bool {
	return false
}

// WithAuth 使用用户名密码在构造时登录换取 token，并把 token 作为每个 RPC 的 Authorization 元数据。
func WithAuth(username, password string) Option {
	return func(c *Client) {
		c.username = username
		c.password = password
		c.dialOptions = append(c.dialOptions, grpc.WithPerRPCCredentials(&clientauth{c: c}))
	}
}

// WithBearerToken 直接注入已签发的 token（自动补 Bearer 前缀）。
func WithBearerToken(token string) Option {
	return func(c *Client) {
		c.setToken(token)
		c.dialOptions = append(c.dialOptions, grpc.WithPerRPCCredentials(&clientauth{c: c}))
	}
}

// WithTokenAutoRefresh 遇到 codes.Unauthenticated（且配置了 WithAuth）时自动重新登录并
// 重试，覆盖 unary 与 server-streaming 两类调用。Login/Exchange 自身不触发刷新，
// 以免凭据错误时递归 getToken 与 singleflight 形成自我死锁。
func WithTokenAutoRefresh() Option {
	return func(c *Client) {
		c.UnaryClientInterceptors = append(c.UnaryClientInterceptors,
			func(
				ctx context.Context,
				method string,
				req, reply interface{},
				cc *grpc.ClientConn,
				invoker grpc.UnaryInvoker,
				opts ...grpc.CallOption) error {
				err := invoker(ctx, method, req, reply, cc, opts...)
				if c.skipAutoRefresh(method) || status.Code(err) != codes.Unauthenticated {
					return err
				}
				operation := func() error {
					if gerr := c.getToken(); gerr != nil {
						return gerr
					}
					return invoker(ctx, method, req, reply, cc, opts...)
				}
				var bf backoff.BackOff = backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
				bf = backoff.WithMaxRetries(bf, 5)
				return backoff.Retry(operation, bf)
			})
		c.StreamClientInterceptors = append(c.StreamClientInterceptors,
			func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
				cs, err := streamer(ctx, desc, cc, method, opts...)
				if err != nil {
					return nil, err
				}
				// 仅纯 server-streaming 支持重建流；client/bidi 流重建没有语义意义。
				if c.skipAutoRefresh(method) || !desc.ServerStreams || desc.ClientStreams {
					return cs, nil
				}
				return &autoRefreshStream{
					c: c, ctx: ctx, desc: desc, cc: cc,
					method: method, streamer: streamer, opts: opts,
					ClientStream: cs,
				}, nil
			})
	}
}

// autoRefreshStream 包装 server-streaming 客户端流：RecvMsg 遇到 Unauthenticated 时自动
// 刷新 token 并重建流（重发请求）后继续读取。只重试一次，避免 token 持续失效时无限循环。
type autoRefreshStream struct {
	grpc.ClientStream

	c         *Client
	ctx       context.Context
	desc      *grpc.StreamDesc
	cc        *grpc.ClientConn
	method    string
	streamer  grpc.Streamer
	opts      []grpc.CallOption
	reqMsg    interface{}
	refreshed bool
}

// SendMsg 记录首个请求消息，重建流时需要用同一请求重新打开 RPC。
func (s *autoRefreshStream) SendMsg(m interface{}) error {
	if s.reqMsg == nil {
		s.reqMsg = m
	}
	return s.ClientStream.SendMsg(m)
}

func (s *autoRefreshStream) RecvMsg(m interface{}) error {
	err := s.ClientStream.RecvMsg(m)
	if status.Code(err) != codes.Unauthenticated || s.refreshed {
		return err
	}
	if gerr := s.c.getToken(); gerr != nil {
		return gerr
	}
	s.refreshed = true
	// 重建流：streamer 只负责 NewStream，请求消息需显式重发，否则服务端会一直等待请求头。
	cs, serr := s.streamer(s.ctx, s.desc, s.cc, s.method, s.opts...)
	if serr != nil {
		return serr
	}
	if serr := cs.SendMsg(s.reqMsg); serr != nil {
		return serr
	}
	s.ClientStream = cs
	return cs.RecvMsg(m)
}

// WithTransportCredentials 使用自定义 tls.Config 建立 TLS 连接（含双向 mTLS 场景）。
// 不调用时默认使用明文 insecure 传输。
func WithTransportCredentials(tlsCfg *tls.Config) Option {
	return func(c *Client) {
		c.tls = tlsCfg
		c.dialOptions = append(c.dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
	}
}

// WithUnaryClientInterceptor 追加一个 unary 客户端拦截器。
func WithUnaryClientInterceptor(op grpc.UnaryClientInterceptor) Option {
	return func(c *Client) {
		c.UnaryClientInterceptors = append(c.UnaryClientInterceptors, op)
	}
}

// WithStreamClientInterceptor 追加一个 streaming 客户端拦截器。
func WithStreamClientInterceptor(op grpc.StreamClientInterceptor) Option {
	return func(c *Client) {
		c.StreamClientInterceptors = append(c.StreamClientInterceptors, op)
	}
}

// WithTracer 接入 OpenTelemetry，为每个 RPC 附加 trace 统计 handler。
func WithTracer() Option {
	return func(c *Client) {
		c.dialOptions = append(c.dialOptions, grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
	}
}
