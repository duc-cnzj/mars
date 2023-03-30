package bootstrappers

import (
	"context"
	"fmt"
	"net"
	"runtime"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	marsauthorizor "github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/grpc/services"
	"github.com/duc-cnzj/mars/v4/internal/middlewares"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

type GrpcBootstrapper struct{}

func (g *GrpcBootstrapper) Tags() []string {
	return []string{"api", "grpc"}
}

func (g *GrpcBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	app.AddServer(&grpcRunner{endpoint: fmt.Sprintf("0.0.0.0:%s", app.Config().GrpcPort)})

	return nil
}

type grpcServerImp interface {
	GracefulStop()
	Serve(lis net.Listener) error
}

type grpcRunner struct {
	server   grpcServerImp
	endpoint string
}

func (g *grpcRunner) Shutdown(ctx context.Context) error {
	defer mlog.Info("[Server]: shutdown grpcRunner runner.")
	if g.server == nil {
		return nil
	}

	done := make(chan struct{})
	go func() {
		g.server.GracefulStop()
		close(done)
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
	g.server = g.initServer()
	go func(server grpcServerImp) {
		if err := server.Serve(listen); err != nil {
			mlog.Error(err)
		}
	}(g.server)

	return nil
}

func (g *grpcRunner) initServer() *grpc.Server {
	server := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			grpc_auth.StreamServerInterceptor(authenticate),
			marsauthorizor.StreamServerInterceptor(),
			middlewares.StreamServerInterceptor(),
			grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandler(recoveryHandler)),
			middlewares.MetricsStreamServerInterceptor,
		),
		grpc.ChainUnaryInterceptor(
			grpc_auth.UnaryServerInterceptor(authenticate),
			middlewares.MetricsServerInterceptor,
			middlewares.TraceUnaryServerInterceptor,
			marsauthorizor.UnaryServerInterceptor(),
			middlewares.UnaryServerInterceptor(),
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(recoveryHandler)),
		),
	)

	for _, registryFunc := range services.RegisteredServers() {
		registryFunc(server, app.App())
	}
	return server
}

func recoveryHandler(p any) error {
	bf := make([]byte, 1024*5)
	n := runtime.Stack(bf, false)
	bf = bf[:n]
	mlog.Error("[Grpc]: recovery error: ", p, string(bf))
	return nil
}

func authenticate(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}
	if verifyToken, b := app.Auth().VerifyToken(token); b {
		return marsauthorizor.SetUser(ctx, &verifyToken.UserInfo), nil
	}

	return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated.")
}
