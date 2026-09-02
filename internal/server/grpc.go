package server

import (
	"context"
	"net"
	"runtime/debug"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/server/middlewares"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	grpcRegistry *app.GrpcRegistry
}

// NewGrpcRunner 构建 gRPC 传输层启动器：从 app 取 GrpcRegistry 用于注册服务、
// AuthBiz 用于鉴权与生效角色计算（有效角色经 AuthBiz.EffectiveRoles 按 users 表接管
// 状态解析，不另需 UserBiz），endpoint 为监听地址。返回实现 app.Server 生命周期。
func NewGrpcRunner(
	endpoint string,
	app app.ServerDeps,
) app.Server {
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

// initServer 装配 gRPC 服务：串联恢复/指标/日志/登录/授权/校验拦截器链（Unary+Stream），
// 并把注册表中的服务注册到该 server。链序约定见 initServer 内部注释。
func (g *grpcRunner) initServer() *grpc.Server {
	authFn := func(ctx context.Context) (context.Context, error) {
		return g.authenticate(ctx)
	}
	// 链序约定（从左到右 = 从外到内，最先列出者最外层）：
	// 1. Recovery 必须在最外层——grpc-go 无内置 recover，内层任何拦截器 panic 未被
	//    最外层 Recovery 捕获即击穿进程；Recovery 放最内层只能兜住 handler 自身 panic。
	// 2. Metrics 须置于 Login 之前——Login 失败直接 return，放在后面的拦截器对 401
	//    请求完全不执行，Metrics 在最外层照常计数，认证失败仍进指标。
	// 3. AccessLog 须置于 Login 之后——Login 验签成功把用户注入新 ctx 传给内层拦截器，
	//    AccessLog 只有收到该 ctx 才能打印出调用用户（grpcUser 对无用户 ctx 返回匿名）。
	//    401 请求（Login 失败）的访问审计由 Login 拦截器内的 [auth audit] Warning 日志兜底，
	//    不因 AccessLog 移内而丢认证失败记录。
	// 4. 校验顺序：Login(鉴权) → AccessLog(访问日志) → Auth(授权) → Validator(校验)，
	//    先确认身份、再记录访问、再问权限。
	server := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainStreamInterceptor(
			grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandler(g.recoveryHandler)),
			middlewares.MetricsStreamServerInterceptor(),
			middlewares.LoginStreamServerInterceptor(authFn, g.logger),
			middlewares.AccessLogStreamServerInterceptor(g.logger),
			middlewares.AuthStreamServerInterceptor(),
			middlewares.ValidatorStreamServerInterceptor(),
		),
		grpc.ChainUnaryInterceptor(
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(g.recoveryHandler)),
			middlewares.MetricsUnaryServerInterceptor(),
			middlewares.LoggerUnaryServerInterceptor(g.logger),
			middlewares.LoginUnaryServerInterceptor(authFn, g.logger),
			middlewares.AccessLogUnaryServerInterceptor(g.logger),
			middlewares.AuthUnaryServerInterceptor(),
			middlewares.ValidatorUnaryServerInterceptor(),
		),
	)

	g.grpcRegistry.RegistryFunc(server)

	return server
}

// recoveryHandler 是 gRPC 恢复拦截器的兜底回调：记录 panic 值并返回 Internal 状态错误，
// 避免 panic 击穿进程；返回错误而非 nil 是为了让客户端收到明确的失败——若返回 nil，
// 恢复后的 RPC 会被 grpc_recovery 当成成功空响应交付，panic 被伪装成成功。
//
// 日志必须带 goroutine 栈快照：运行时 panic（nil 指针解引用等）的 recover 值只是
// runtime.Error，既不含 pkg/errors 栈也非 fmt.Formatter，%+v 对它无效——只能现场抓
// debug.Stack() 定位真实 panic 起源帧（与 mlog.HandlePanic 的 panicStack 同思路）。
// grpc_recovery 的 defer 尚未返回，被 unwinding 的帧仍在当前 goroutine 栈上，故此处
// 抓栈能看到 panic 发生点而非仅恢复点。
func (g *grpcRunner) recoveryHandler(p any) error {
	g.logger.Errorf("[Grpc]: recovery error: \n%v\n--- goroutine stack ---\n%s", p, string(debug.Stack()))
	return status.Error(codes.Internal, "internal server error")
}

// authenticate 从 gRPC metadata 提取 Bearer token 并完成鉴权：校验、注入用户与按 users
// 表接管状态计算生效角色统一走 biz.Authenticate，与 HTTP LoginHTTP
// 中间件共用同一鉴权核心——本函数只负责 gRPC 特有的 token 提取（metadata），不再各自实现。
func (g *grpcRunner) authenticate(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}
	// 校验与用户注入统一走 biz.Authenticate，与 HTTP LoginHTTP 中间件
	// 共用同一鉴权核心——本函数只负责 gRPC 特有的 token 提取（metadata）。
	return biz.Authenticate(ctx, g.authBiz, token)
}
