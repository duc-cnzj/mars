package client

import (
	"context"
	"crypto/tls"
	"io"
	"strings"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"

	"github.com/duc-cnzj/mars-client/v4/auth"
	"github.com/duc-cnzj/mars-client/v4/changelog"
	"github.com/duc-cnzj/mars-client/v4/cluster"
	"github.com/duc-cnzj/mars-client/v4/container"
	"github.com/duc-cnzj/mars-client/v4/endpoint"
	"github.com/duc-cnzj/mars-client/v4/event"
	"github.com/duc-cnzj/mars-client/v4/file"
	"github.com/duc-cnzj/mars-client/v4/git"
	"github.com/duc-cnzj/mars-client/v4/gitconfig"
	"github.com/duc-cnzj/mars-client/v4/metrics"
	"github.com/duc-cnzj/mars-client/v4/namespace"
	"github.com/duc-cnzj/mars-client/v4/picture"
	"github.com/duc-cnzj/mars-client/v4/project"
	"github.com/duc-cnzj/mars-client/v4/version"

	"github.com/cenkalti/backoff/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

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
	GitConfig() gitconfig.GitConfigClient

	Namespace() namespace.NamespaceClient
	Project() project.ProjectClient
	Endpoint() endpoint.EndpointClient
	Metrics() metrics.MetricsClient
}

type Client struct {
	singleflight Group

	UnaryClientInterceptors  []grpc.UnaryClientInterceptor
	StreamClientInterceptors []grpc.StreamClientInterceptor
	username, password       string
	authTokenValue           atomic.Value

	tls *tls.Config

	conn        *grpc.ClientConn
	dialOptions []grpc.DialOption
	tracer      trace.Tracer

	auth      auth.AuthClient
	changelog changelog.ChangelogClient
	cluster   cluster.ClusterClient
	container container.ContainerClient
	event     event.EventClient
	git       git.GitClient
	gitConfig gitconfig.GitConfigClient
	metrics   metrics.MetricsClient
	namespace namespace.NamespaceClient
	picture   picture.PictureClient
	project   project.ProjectClient
	version   version.VersionClient
	file      file.FileClient
	endpoint  endpoint.EndpointClient
}

func NewClient(addr string, opts ...Option) (Interface, error) {
	c := &Client{}

	for _, opt := range opts {
		opt(c)
	}

	dial, err := grpc.Dial(addr, c.buildDialOptions()...)

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
	c.gitConfig = gitconfig.NewGitConfigClient(dial)
	c.metrics = metrics.NewMetricsClient(dial)
	c.namespace = namespace.NewNamespaceClient(dial)
	c.picture = picture.NewPictureClient(dial)
	c.project = project.NewProjectClient(dial)
	c.version = version.NewVersionClient(dial)
	c.file = file.NewFileClient(dial)
	c.endpoint = endpoint.NewEndpointClient(dial)

	if c.password != "" || c.username != "" {
		if err := c.getToken(); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Client) SetBearerToken(token string) {
	c.setToken(token)
}

func (c *Client) hasCredentials() bool {
	return c.username != "" && c.password != ""
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

func (c *Client) Auth() auth.AuthClient {
	return c.auth
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

func (c *Client) File() file.FileClient {
	return c.file
}

func (c *Client) Git() git.GitClient {
	return c.git
}

func (c *Client) GitConfig() gitconfig.GitConfigClient {
	return c.gitConfig
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

	c.dialOptions = append(c.dialOptions, grpc.WithChainUnaryInterceptor(c.UnaryClientInterceptors...))
	c.dialOptions = append(c.dialOptions, grpc.WithChainStreamInterceptor(c.StreamClientInterceptors...))

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
	if len(token) > 6 && !strings.EqualFold("Bearer", token[0:6]) {
		token = "Bearer " + token
	}
	c.authTokenValue.Store(token)
}

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

func WithAuth(username, password string) Option {
	return func(c *Client) {
		c.username = username
		c.password = password
		c.dialOptions = append(c.dialOptions, grpc.WithPerRPCCredentials(&clientauth{c: c}))
	}
}

func WithBearerToken(token string) Option {
	return func(c *Client) {
		c.setToken(token)
		c.dialOptions = append(c.dialOptions, grpc.WithPerRPCCredentials(&clientauth{c: c}))
	}
}

// WithTokenAutoRefresh
// TODO c.StreamClientInterceptors 有点难搞，好在目前没用到，之后用到了需要搞一下
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
				operation := func() error {
					if gerr := c.getToken(); gerr != nil {
						return gerr
					}
					return invoker(ctx, method, req, reply, cc, opts...)
				}
				var bf backoff.BackOff = backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
				bf = backoff.WithMaxRetries(bf, 5)

				if c.hasCredentials() && status.Code(err) == codes.Unauthenticated && method != "/Auth/Login" {
					return backoff.Retry(operation, bf)
				}
				return err
			})
	}
}

// WithTransportCredentials 暂时不支持
//func WithTransportCredentials(tls *tls.Config) Option {
//	return func(c *Client) {
//		c.tls = tls
//		c.dialOptions = append(c.dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(tls)))
//	}
//}

func WithUnaryClientInterceptor(op grpc.UnaryClientInterceptor) Option {
	return func(c *Client) {
		c.UnaryClientInterceptors = append(c.UnaryClientInterceptors, op)
	}
}

func WithStreamClientInterceptor(op grpc.StreamClientInterceptor) Option {
	return func(c *Client) {
		c.StreamClientInterceptors = append(c.StreamClientInterceptors, op)
	}
}

func WithTracer(tracer trace.Tracer) Option {
	return func(c *Client) {
		c.tracer = tracer
		c.UnaryClientInterceptors = append(c.UnaryClientInterceptors, TraceUnaryClientInterceptor(c.tracer))
	}
}

type GatewayCarrier metadata.MD

func (hc GatewayCarrier) Get(key string) string {
	vals := metadata.MD(hc).Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func (hc GatewayCarrier) Set(key string, value string) {
	metadata.MD(hc).Set(key, value)
}

func (hc GatewayCarrier) Keys() []string {
	keys := make([]string, 0, len(hc))
	for k := range hc {
		keys = append(keys, k)
	}
	return keys
}

func TraceUnaryClientInterceptor(tracer trace.Tracer) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start, span := tracer.Start(ctx, "TraceUnaryClientInterceptor: "+method)
		defer span.End()
		span.SetAttributes(attribute.String("method", method))
		ctxt := propagation.TraceContext{}
		md := metadata.MD{}
		if outMD, found := metadata.FromOutgoingContext(ctx); found {
			md = outMD
		}
		ctxt.Inject(start, GatewayCarrier(md))
		ctx = metadata.NewOutgoingContext(ctx, metadata.MD(md))

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			span.SetStatus(otelcodes.Error, err.Error())
		}
		return err
	}
}
