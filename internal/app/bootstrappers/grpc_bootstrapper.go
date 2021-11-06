package bootstrappers

import (
	"context"
	"fmt"
	"net"

	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"

	"github.com/duc-cnzj/mars/pkg/cp"

	"github.com/duc-cnzj/mars/internal/utils"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/pkg/auth"
	"github.com/duc-cnzj/mars/pkg/picture"
	"github.com/golang-jwt/jwt"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	if g.server == nil {
		return nil
	}
	mlog.Info("[Server]: shutdown grpcRunner runner.")

	g.server.GracefulStop()

	return nil
}

func (g *grpcRunner) Run(ctx context.Context) error {
	mlog.Infof("[Server]: start grpcRunner runner at %s.", g.endpoint)
	listen, _ := net.Listen("tcp", g.endpoint)
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithOpName(func(method string) string {
				return "[Tracer]: " + method
			})),
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
	picture.RegisterPictureServer(server, new(services.Picture))
	cp.RegisterCpServer(server, new(services.CopyToPod))
	auth.RegisterAuthServer(server, services.NewAuth(app.Config().Prikey(), app.Config().Pubkey(), app.App().Oidc(), app.Config().AdminPassword))

	g.server = server
	go func() {
		if err := server.Serve(listen); err != nil {
			mlog.Error(err)
		}
	}()

	return nil
}

func Authenticate(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}
	parse, err := jwt.ParseWithClaims(token, &services.JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return app.Config().Pubkey(), nil
	})
	if err == nil && parse.Valid {
		claims, ok := parse.Claims.(*services.JwtClaims)
		if ok {
			return services.SetUser(ctx, &claims.UserInfo), nil
		}
	}
	return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated.")
}
