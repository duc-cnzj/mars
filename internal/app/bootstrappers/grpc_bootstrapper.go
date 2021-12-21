package bootstrappers

import (
	"context"
	"fmt"
	"net"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/grpc/services"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/pkg/auth"
	"github.com/duc-cnzj/mars/pkg/changelog"
	"github.com/duc-cnzj/mars/pkg/cluster"
	"github.com/duc-cnzj/mars/pkg/cp"
	"github.com/duc-cnzj/mars/pkg/event"
	"github.com/duc-cnzj/mars/pkg/gitlab"
	"github.com/duc-cnzj/mars/pkg/mars"
	rpcmetrics "github.com/duc-cnzj/mars/pkg/metrics"
	"github.com/duc-cnzj/mars/pkg/namespace"
	"github.com/duc-cnzj/mars/pkg/picture"
	"github.com/duc-cnzj/mars/pkg/project"
	"github.com/duc-cnzj/mars/pkg/version"
)

var grpcEndpoint string

func init() {
	port, err := utils.GetFreePort()
	if err != nil {
		panic("There are no free ports for grpc server")
	}
	grpcEndpoint = fmt.Sprintf("localhost:%d", port)
}

type GrpcBootstrapper struct{}

func (g *GrpcBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	app.AddServer(&grpcRunner{endpoint: grpcEndpoint})

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
	listen, _ := net.Listen("tcp", g.endpoint)
	server := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			grpc_opentracing.StreamServerInterceptor(traceWithOpName()),
			grpc_auth.StreamServerInterceptor(Authenticate),
			grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
				mlog.Error("[Grpc]: recovery error: ", p)
				return nil
			})),
			func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
				user, err := services.GetUser(ctx)
				if err == nil {
					mlog.Infof("[Grpc]: user: %v, visit: %v.", user.Name, info.FullMethod)
				}

				return handler(srv, ss)
			},
			grpc_prometheus.StreamServerInterceptor,
		),
		grpc.ChainUnaryInterceptor(
			grpc_opentracing.UnaryServerInterceptor(traceWithOpName()),
			grpc_auth.UnaryServerInterceptor(Authenticate),
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				user, err := services.GetUser(ctx)
				if err == nil {
					mlog.Infof("[Grpc]: user: %v, visit: %v.", user.Name, info.FullMethod)
				}

				return handler(ctx, req)
			},
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
				mlog.Error("[Grpc]: recovery error: ", p)
				return nil
			})),
			grpc_prometheus.UnaryServerInterceptor,
		),
	)

	grpc_prometheus.Register(server)

	cluster.RegisterClusterServer(server, new(services.Cluster))
	gitlab.RegisterGitlabServer(server, new(services.Gitlab))
	mars.RegisterMarsServer(server, new(services.Mars))
	namespace.RegisterNamespaceServer(server, new(services.Namespace))
	project.RegisterProjectServer(server, new(services.Project))
	picture.RegisterPictureServer(server, new(services.Picture))
	cp.RegisterCpServer(server, new(services.CopyToPod))
	auth.RegisterAuthServer(server, services.NewAuth(app.Auth(), app.Oidc(), app.Config().AdminPassword))
	rpcmetrics.RegisterMetricsServer(server, new(services.Metrics))
	version.RegisterVersionServer(server, new(services.VersionService))
	changelog.RegisterChangelogServer(server, new(services.Changelog))
	event.RegisterEventServer(server, new(services.EventSvc))

	g.server = server
	go func() {
		if err := server.Serve(listen); err != nil {
			mlog.Error(err)
		}
	}()

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
		return services.SetUser(ctx, &verifyToken.UserInfo), nil

	}

	return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated.")
}
