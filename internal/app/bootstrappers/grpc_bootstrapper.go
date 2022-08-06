package bootstrappers

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"time"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	marsauthorizor "github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/grpc/services"
	"github.com/duc-cnzj/mars/internal/metrics"
	"github.com/duc-cnzj/mars/internal/middlewares"
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
					user, e := marsauthorizor.GetUser(ctx)
					if e == nil {
						mlog.Infof("[Grpc]: user: %v, visit: %v, use: %s.", user.Name, info.FullMethod, time.Since(t))
					}
				}(time.Now())

				e := handler(srv, ss)
				if e != nil {
					metrics.GrpcRequestTotalFail.With(prometheus.Labels{"method": info.FullMethod}).Inc()
					metrics.GrpcErrorCount.With(prometheus.Labels{"method": info.FullMethod}).Inc()
				} else {
					metrics.GrpcRequestTotalSuccess.With(prometheus.Labels{"method": info.FullMethod}).Inc()
				}

				return e
			},
		),
		grpc.ChainUnaryInterceptor(
			grpc_auth.UnaryServerInterceptor(Authenticate),
			func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
				defer func(t time.Time) {
					user := &contracts.UserInfo{}
					if u, err := marsauthorizor.GetUser(ctx); err == nil {
						user = u
					}
					mlog.Infof("[Grpc]: user: %v, visit: %v, use: %s.", user.Name, info.FullMethod, time.Since(t))
					metrics.GrpcLatency.With(prometheus.Labels{"method": info.FullMethod}).Observe(time.Since(t).Seconds())
				}(time.Now())

				i, err := handler(ctx, req)
				if err != nil {
					metrics.GrpcRequestTotalFail.With(prometheus.Labels{"method": info.FullMethod}).Inc()
					metrics.GrpcErrorCount.With(prometheus.Labels{"method": info.FullMethod}).Inc()
				} else {
					metrics.GrpcRequestTotalSuccess.With(prometheus.Labels{"method": info.FullMethod}).Inc()
				}

				return i, err
			},
			middlewares.TraceUnaryServerInterceptor,
			marsauthorizor.UnaryServerInterceptor(),
			validator.UnaryServerInterceptor(),
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(func(p any) (err error) {
				bf := make([]byte, 1024*5)
				runtime.Stack(bf, false)
				mlog.Error("[Grpc]: recovery error: ", p, string(bf))
				return nil
			})),
		),
	)

	for _, registryFunc := range services.RegisteredServers() {
		registryFunc(server, app.App())
	}

	g.server = server
	go func(server *grpc.Server) {
		if err := server.Serve(listen); err != nil {
			mlog.Error(err)
		}
	}(server)

	return nil
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
