package server

import (
	"context"
	"net"
	"runtime"

	"github.com/duc-cnzj/mars/v4/internal/application"
	marsauthorizor "github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/server/middlewares"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServerImp interface {
	GracefulStop()
	Serve(lis net.Listener) error
}

type grpcRunner struct {
	server   grpcServerImp
	endpoint string
	app      application.App
}

func NewGrpcRunner(
	endpoint string,
	app application.App,
) application.Server {
	return &grpcRunner{endpoint: endpoint, app: app}
}

func (g *grpcRunner) Shutdown(ctx context.Context) error {
	defer g.app.Logger().Info("[Server]: shutdown grpcRunner runner.")
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
	g.app.Logger().Infof("[Server]: start grpcRunner runner at %s.", g.endpoint)
	listen, err := net.Listen("tcp", g.endpoint)
	if err != nil {
		return err
	}
	g.server = g.initServer()
	go func(server grpcServerImp) {
		if err := server.Serve(listen); err != nil {
			g.app.Logger().Error(err)
		}
	}(g.server)

	return nil
}

func (g *grpcRunner) initServer() *grpc.Server {
	authFn := func(ctx context.Context) (context.Context, error) {
		return authenticate(ctx, g.app.Auth())
	}
	server := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			grpc_auth.StreamServerInterceptor(authFn),
			marsauthorizor.StreamServerInterceptor(),
			middlewares.StreamServerInterceptor(),
			grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandler(g.recoveryHandler)),
			middlewares.MetricsStreamServerInterceptor,
		),
		grpc.ChainUnaryInterceptor(
			middlewares.LoggerUnaryServerInterceptor(g.app.Logger()),
			grpc_auth.UnaryServerInterceptor(authFn),
			middlewares.MetricsServerInterceptor,
			middlewares.TraceUnaryServerInterceptor,
			marsauthorizor.UnaryServerInterceptor(),
			middlewares.UnaryServerInterceptor(),
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(g.recoveryHandler)),
		),
	)

	g.app.GrpcRegistry().RegistryFunc(server)

	return server
}

func (g *grpcRunner) recoveryHandler(p any) error {
	bf := make([]byte, 1024*5)
	n := runtime.Stack(bf, false)
	bf = bf[:n]
	g.app.Logger().Errorf("[Grpc]: recovery error: \n%v", bf)
	return nil
}

func authenticate(ctx context.Context, auth marsauthorizor.Auth) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}
	if verifyToken, b := auth.VerifyToken(token); b {
		return marsauthorizor.SetUser(ctx, verifyToken.UserInfo), nil
	}

	return nil, status.Errorf(codes.Unauthenticated, "Unauthenticated.")
}
