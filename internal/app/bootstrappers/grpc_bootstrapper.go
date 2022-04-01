package bootstrappers

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"time"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/duc-cnzj/mars-client/v4/auth"
	"github.com/duc-cnzj/mars-client/v4/changelog"
	"github.com/duc-cnzj/mars-client/v4/cluster"
	"github.com/duc-cnzj/mars-client/v4/container"
	"github.com/duc-cnzj/mars-client/v4/endpoint"
	"github.com/duc-cnzj/mars-client/v4/event"
	"github.com/duc-cnzj/mars-client/v4/file"
	"github.com/duc-cnzj/mars-client/v4/gitproject"
	rpcmetrics "github.com/duc-cnzj/mars-client/v4/metrics"
	"github.com/duc-cnzj/mars-client/v4/namespace"
	"github.com/duc-cnzj/mars-client/v4/picture"
	"github.com/duc-cnzj/mars-client/v4/project"
	"github.com/duc-cnzj/mars-client/v4/version"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	marsauthorizor "github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/grpc/services"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/validator"
)

type GrpcBootstrapper struct{}

func (g *GrpcBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	app.AddServer(&grpcRunner{endpoint: fmt.Sprintf("0.0.0.0:%s", app.Config().GrpcPort)})

	return nil
}

type grpcRunner struct {
	server   *grpc.Server
	endpoint string
}

func (g *grpcRunner) Shutdown(ctx context.Context) error {
	defer mlog.Info("[Server]: shutdown grpcRunner runner.")
	if g.server == nil {
		return nil
	}

	done := make(chan struct{}, 1)
	go func() {
		g.server.GracefulStop()
		done <- struct{}{}
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (g *grpcRunner) Run(ctx context.Context) error {
	mlog.Infof("[Server]: start grpcRunner runner at %s.", g.endpoint)
	listen, err := net.Listen("tcp", g.endpoint)
	if err != nil {
		return err
	}
	server := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			grpc_opentracing.StreamServerInterceptor(traceWithOpName()),
			grpc_auth.StreamServerInterceptor(Authenticate),
			marsauthorizor.StreamServerInterceptor(),
			validator.StreamServerInterceptor(),
			grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandler(func(p any) (err error) {
				bf := make([]byte, 1024*5)
				runtime.Stack(bf, false)
				mlog.Error("[Grpc]: recovery error: ", p, string(bf))
				return nil
			})),
			func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
				defer func(t time.Time) {
					user, err := marsauthorizor.GetUser(ctx)
					if err == nil {
						mlog.Infof("[Grpc]: user: %v, visit: %v, use: %s.", user.Name, info.FullMethod, time.Since(t))
					}
				}(time.Now())

				return handler(srv, ss)
			},
			grpc_prometheus.StreamServerInterceptor,
		),
		grpc.ChainUnaryInterceptor(
			grpc_opentracing.UnaryServerInterceptor(traceWithOpName()),
			grpc_auth.UnaryServerInterceptor(Authenticate),
			marsauthorizor.UnaryServerInterceptor(),
			validator.UnaryServerInterceptor(),
			func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
				defer func(t time.Time) {
					user, err := marsauthorizor.GetUser(ctx)
					if err == nil {
						mlog.Infof("[Grpc]: user: %v, visit: %v, use: %s.", user.Name, info.FullMethod, time.Since(t))
					}
				}(time.Now())

				return handler(ctx, req)
			},
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(func(p any) (err error) {
				bf := make([]byte, 1024*5)
				runtime.Stack(bf, false)
				mlog.Error("[Grpc]: recovery error: ", p, string(bf))
				return nil
			})),
			grpc_prometheus.UnaryServerInterceptor,
		),
	)

	grpc_prometheus.Register(server)

	cluster.RegisterClusterServer(server, new(services.ClusterSvc))
	gitproject.RegisterGitProjectServer(server, new(services.GitProjectSvc))
	gitproject.RegisterGitProjectConfigServer(server, new(services.GitProjectConfigSvc))
	namespace.RegisterNamespaceServer(server, new(services.NamespaceSvc))
	project.RegisterProjectServer(server, new(services.ProjectSvc))
	picture.RegisterPictureServer(server, new(services.PictureSvc))
	container.RegisterContainerSvcServer(server, new(services.ContainerSvc))
	auth.RegisterAuthServer(server, services.NewAuthSvc(app.Auth(), app.Oidc(), app.Config().AdminPassword))
	rpcmetrics.RegisterMetricsServer(server, new(services.MetricsSvc))
	version.RegisterVersionServer(server, new(services.VersionSvc))
	changelog.RegisterChangelogServer(server, new(services.ChangelogSvc))
	event.RegisterEventServer(server, new(services.EventSvc))
	file.RegisterFileSvcServer(server, new(services.FileSvc))
	endpoint.RegisterEndpointServer(server, new(services.EndpointSvc))

	g.server = server
	go func(server *grpc.Server) {
		if err := server.Serve(listen); err != nil {
			mlog.Error(err)
		}
	}(server)

	return nil
}

func traceWithOpName() grpc_opentracing.Option {
	return grpc_opentracing.WithOpName(func(method string) string {
		return "[Tracer]: " + method
	})
}

func Authenticate(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}
	if verifyToken, b := app.Auth().VerifyToken(token); b {
		return marsauthorizor.SetUser(ctx, &verifyToken.UserInfo), nil
	}

	return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated.")
}
