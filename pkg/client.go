package client

import (
	"context"
	"crypto/tls"
	"strings"
	"sync/atomic"

	"github.com/cenkalti/backoff/v4"
	"github.com/duc-cnzj/mars/client/auth"
	"github.com/duc-cnzj/mars/client/changelog"
	"github.com/duc-cnzj/mars/client/cluster"
	"github.com/duc-cnzj/mars/client/container_copy"
	"github.com/duc-cnzj/mars/client/event"
	"github.com/duc-cnzj/mars/client/gitserver"
	"github.com/duc-cnzj/mars/client/mars"
	"github.com/duc-cnzj/mars/client/metrics"
	"github.com/duc-cnzj/mars/client/namespace"
	"github.com/duc-cnzj/mars/client/picture"
	"github.com/duc-cnzj/mars/client/project"
	"github.com/duc-cnzj/mars/client/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type Interface interface {
	Auth() auth.AuthClient
	Changelog() changelog.ChangelogClient
	Cluster() cluster.ClusterClient
	ContainerCopy() container_copy.ContainerCopyClient
	Event() event.EventClient
	GitServer() gitserver.GitServerClient
	Mars() mars.MarsClient
	Metrics() metrics.MetricsClient
	Namespace() namespace.NamespaceClient
	Picture() picture.PictureClient
	Project() project.ProjectClient
	Version() version.VersionClient
}

type Client struct {
	singleflight Group

	UnaryClientInterceptors  []grpc.UnaryClientInterceptor
	StreamClientInterceptors []grpc.StreamClientInterceptor
	username, password       string
	authTokenValue           atomic.Value

	tls *tls.Config

	dialOptions []grpc.DialOption

	auth          auth.AuthClient
	changelog     changelog.ChangelogClient
	cluster       cluster.ClusterClient
	containerCopy container_copy.ContainerCopyClient
	event         event.EventClient
	gitServer     gitserver.GitServerClient
	mars          mars.MarsClient
	metrics       metrics.MetricsClient
	namespace     namespace.NamespaceClient
	picture       picture.PictureClient
	project       project.ProjectClient
	version       version.VersionClient
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

	c.auth = auth.NewAuthClient(dial)
	c.changelog = changelog.NewChangelogClient(dial)
	c.cluster = cluster.NewClusterClient(dial)
	c.containerCopy = container_copy.NewContainerCopyClient(dial)
	c.event = event.NewEventClient(dial)
	c.gitServer = gitserver.NewGitServerClient(dial)
	c.mars = mars.NewMarsClient(dial)
	c.metrics = metrics.NewMetricsClient(dial)
	c.namespace = namespace.NewNamespaceClient(dial)
	c.picture = picture.NewPictureClient(dial)
	c.project = project.NewProjectClient(dial)
	c.version = version.NewVersionClient(dial)

	if c.password != "" || c.username != "" {
		if err := c.getToken(); err != nil {
			return nil, err
		}
	}

	return c, nil
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

func (c *Client) ContainerCopy() container_copy.ContainerCopyClient {
	return c.containerCopy
}

func (c *Client) Event() event.EventClient {
	return c.event
}

func (c *Client) GitServer() gitserver.GitServerClient {
	return c.gitServer
}

func (c *Client) Mars() mars.MarsClient {
	return c.mars
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

func (c *Client) authToken() string {
	v := c.authTokenValue.Load()
	if v != nil {
		return v.(string)
	}
	return ""
}

func (c *Client) buildDialOptions() []grpc.DialOption {
	if c.tls == nil {
		c.dialOptions = append(c.dialOptions, grpc.WithInsecure())
	}

	c.dialOptions = append(c.dialOptions, grpc.WithChainUnaryInterceptor(c.UnaryClientInterceptors...))
	c.dialOptions = append(c.dialOptions, grpc.WithChainStreamInterceptor(c.StreamClientInterceptors...))

	return c.dialOptions
}

func (c *Client) getToken() error {
	login, err, _ := c.singleflight.Do("Retry", func() (interface{}, error) {
		return c.auth.Login(context.TODO(), &auth.AuthLoginRequest{
			Username: c.username,
			Password: c.password,
		})
	})
	if err != nil {
		return err
	}

	c.setToken(login.(*auth.AuthLoginResponse).Token)
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

type clienttokenauth struct {
	token string
}

func (a *clienttokenauth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"Authorization": a.token,
	}, nil
}

func (a *clienttokenauth) RequireTransportSecurity() bool {
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
		c.dialOptions = append(c.dialOptions, grpc.WithPerRPCCredentials(&clienttokenauth{token: token}))
	}
}

// WithRefresh
// TODO c.StreamClientInterceptors 有点难搞，好在目前没用到，之后用到了需要搞一下
func WithRefresh() Option {
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

				if status.Code(err) == codes.Unauthenticated && method != "/Auth/Login" {
					return backoff.Retry(operation, bf)
				}
				return err
			})
	}
}

func WithTransportCredentials(tls *tls.Config) Option {
	return func(c *Client) {
		c.tls = tls
		c.dialOptions = append(c.dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(tls)))
	}
}

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
