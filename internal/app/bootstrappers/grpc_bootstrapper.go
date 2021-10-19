package bootstrappers

import (
	"context"
	"net"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/grpc/services"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/pkg/cluster"
	"github.com/duc-cnzj/mars/pkg/gitlab"
	"github.com/duc-cnzj/mars/pkg/mars"
	"github.com/duc-cnzj/mars/pkg/namespace"
	"github.com/duc-cnzj/mars/pkg/project"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

var grpcEndpoint = "localhost:9999"

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
	mlog.Info("[Server]: shutdown grpcRunner runner.")

	g.server.GracefulStop()

	return nil
}

func (g *grpcRunner) Run(ctx context.Context) error {
	mlog.Debug("[Server]: start grpcRunner runner.")
	listen, _ := net.Listen("tcp", g.endpoint)
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
				mlog.Error("[Grpc]: recovery error: ", p)
				return nil
			})),
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				mlog.Debugf("[Grpc]: Method %v Called.", info.FullMethod)
				return handler(ctx, req)
			},
			grpc_prometheus.UnaryServerInterceptor,
		),
	)

	grpc_prometheus.Register(server)

	cluster.RegisterClusterServer(server, new(services.Cluster))
	gitlab.RegisterGitlabServer(server, new(services.Gitlab))
	mars.RegisterMarsServer(server, new(services.Mars))
	namespace.RegisterNamespaceServer(server, new(services.Namespace))
	project.RegisterProjectServer(server, new(services.Project))

	g.server = server
	go func() {
		if err := server.Serve(listen); err != nil {
			mlog.Error(err)
		}
	}()

	return nil
}
