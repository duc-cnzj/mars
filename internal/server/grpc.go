package server

import (
	"context"
	"net"

	"github.com/duc-cnzj/mars/v6/internal/application"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/server/middlewares"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

// GrpcServerImp 是 *grpc.Server 的最小抽象：仅暴露启动（Serve）与优雅停止（GracefulStop），
// 供 grpcRunner 持有时可注入测试替身（见 grpc_test.go 的 fakeGrpcServer）。
type GrpcServerImp interface {
	GracefulStop()
	Serve(lis net.Listener) error
}

type grpcRunner struct {
	server       GrpcServerImp
	endpoint     string
	logger       mlog.Logger
	authBiz      biz.AuthBiz
	grpcRegistry *application.GrpcRegistry
}

// NewGrpcRunner 构建 gRPC 传输层启动器：从 app 取 GrpcRegistry 用于注册服务、
// AuthBiz 用于鉴权，endpoint 为监听地址。返回实现 application.Server 生命周期。
func NewGrpcRunner(
	endpoint string,
	app application.ServerDeps,
) application.Server {
	return &grpcRunner{
		grpcRegistry: app.GrpcRegistry(),
		endpoint:     endpoint,
		logger:       app.Logger().WithModule("server/grpcRunner"),
		authBiz:      app.AuthBiz(),
	}
}

// Shutdown 优雅停止 gRPC 服务：goroutine 内 GracefulStop，等待完成或 ctx 超时。
func (g *grpcRunner) Shutdown(ctx context.Context) error {
	defer g.logger.Info("[Server]: shutdown grpcRunner runner.")
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

// Run 启动 gRPC 服务：监听 endpoint，构建并注册拦截器链与服务，goroutine 内 Serve。
func (g *grpcRunner) Run(ctx context.Context) error {
	g.logger.Infof("[Server]: start grpcRunner runner at %s.", g.endpoint)
	listen, err := net.Listen("tcp", g.endpoint)
	if err != nil {
		return err
	}
	g.server = g.initServer()
	go func() {
		if err := g.server.Serve(listen); err != nil {
			g.logger.Error(err)
		}
	}()

	return nil
}

// initServer 装配 gRPC 服务：串联登录/授权/校验/指标/日志/恢复拦截器链（Unary+Stream），
// 并把注册表中的服务注册到该 server。
func (g *grpcRunner) initServer() *grpc.Server {
	authFn := func(ctx context.Context) (context.Context, error) {
		return authenticate(ctx, g.authBiz)
	}
	server := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainStreamInterceptor(
			middlewares.LoginStreamServerInterceptor(authFn),
			middlewares.AuthStreamServerInterceptor(),
			middlewares.ValidatorStreamServerInterceptor(),
			grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandler(g.recoveryHandler)),
			middlewares.MetricsStreamServerInterceptor(g.logger),
		),
		grpc.ChainUnaryInterceptor(
			middlewares.LoggerUnaryServerInterceptor(g.logger),
			middlewares.LoginUnaryServerInterceptor(authFn),
			middlewares.MetricsServerInterceptor(g.logger),
			middlewares.AuthUnaryServerInterceptor(),
			middlewares.ValidatorUnaryServerInterceptor(),
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(g.recoveryHandler)),
		),
	)

	g.grpcRegistry.RegistryFunc(server)

	return server
}

// recoveryHandler 是 gRPC 恢复拦截器的兜底回调：记录 panic 值并返回 nil，避免 panic 击穿进程。
func (g *grpcRunner) recoveryHandler(p any) error {
	g.logger.Errorf("[Grpc]: recovery error: \n%v", p)
	return nil
}

// authenticate 从 gRPC metadata 提取 Bearer token 并调用 biz.Authenticate 完成鉴权：
// 与 HTTP 侧登录中间件共用同一鉴权核心，本函数只做 gRPC 特有的 token 提取。
func authenticate(ctx context.Context, auth biz.AuthBiz) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}
	// 校验与用户注入统一走 biz.Authenticate，与 HTTP LoginHTTP 中间件共用同一
	// 鉴权核心——本函数只负责 gRPC 特有的 token 提取（metadata），不再各自实现。
	return biz.Authenticate(ctx, auth, token)
}
