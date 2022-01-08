package bootstrappers

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/duc-cnzj/mars/internal/validator"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/duc-cnzj/mars/client/auth"
	"github.com/duc-cnzj/mars/client/changelog"
	"github.com/duc-cnzj/mars/client/cluster"
	"github.com/duc-cnzj/mars/client/container_copy"
	"github.com/duc-cnzj/mars/client/event"
	"github.com/duc-cnzj/mars/client/gitserver"
	"github.com/duc-cnzj/mars/client/mars"
	rpcmetrics "github.com/duc-cnzj/mars/client/metrics"
	"github.com/duc-cnzj/mars/client/namespace"
	"github.com/duc-cnzj/mars/client/picture"
	"github.com/duc-cnzj/mars/client/project"
	"github.com/duc-cnzj/mars/client/version"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	marsauth "github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/grpc/services"
	"github.com/duc-cnzj/mars/internal/mlog"
)

type GrpcBootstrapper struct{}

func (g *GrpcBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	app.AddServer(&grpcRunner{endpoint: fmt.Sprintf("localhost:%s", app.Config().GrpcPort)})

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
	//p := x509.NewCertPool()
	//file, err := os.ReadFile("../certs/ca.pem")
	//if err != nil {
	//	return err
	//}
	//p.AppendCertsFromPEM(file)
	//cert, err := tls.LoadX509KeyPair("../certs/server.pem", "../certs/server-key.pem")
	//if err != nil {
	//	return err
	//}
	//c := credentials.NewTLS(&tls.Config{
	//	Certificates: []tls.Certificate{cert},
	//	ClientAuth:   tls.RequireAndVerifyClientCert,
	//	ClientCAs:    p,
	//})
	server := grpc.NewServer(
		//grpc.Creds(c),
		grpc.ChainStreamInterceptor(
			grpc_opentracing.StreamServerInterceptor(traceWithOpName()),
			grpc_auth.StreamServerInterceptor(Authenticate),
			marsauth.StreamServerInterceptor(),
			grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
				mlog.Error("[Grpc]: recovery error: ", p)
				return nil
			})),
			func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
				defer func(t time.Time) {
					user, err := marsauth.GetUser(ctx)
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
			validator.UnaryServerInterceptor(),
			marsauth.UnaryServerInterceptor(),
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				defer func(t time.Time) {
					user, err := marsauth.GetUser(ctx)
					if err == nil {
						mlog.Infof("[Grpc]: user: %v, visit: %v, use: %s.", user.Name, info.FullMethod, time.Since(t))
					}
				}(time.Now())

				return handler(ctx, req)
			},
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
				bf := make([]byte, 1024*5)
				runtime.Stack(bf, false)
				mlog.Error("[Grpc]: recovery error: ", p, string(bf))
				return nil
			})),
			grpc_prometheus.UnaryServerInterceptor,
		),
	)

	grpc_prometheus.Register(server)

	cluster.RegisterClusterServer(server, new(services.Cluster))
	gitserver.RegisterGitServerServer(server, new(services.GitServer))
	mars.RegisterMarsServer(server, new(services.Mars))
	namespace.RegisterNamespaceServer(server, new(services.Namespace))
	project.RegisterProjectServer(server, new(services.Project))
	picture.RegisterPictureServer(server, new(services.Picture))
	container_copy.RegisterContainerCopyServer(server, new(services.ContainerCopy))
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
		return marsauth.SetUser(ctx, &verifyToken.UserInfo), nil
	}

	return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated.")
}
