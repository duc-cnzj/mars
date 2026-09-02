package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// captureStdout 在窗口内重定向 os.Stdout 并返回捕获的输出。logrus 后端在构造时固化
// output fd（NewLogrusLogger 里 SetOutput(os.Stdout)），因此 logger 必须在 fn 内构造。
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out)
}

// panicNilDeref 触发真实的 nil 指针解引用 panic，复现线上 recovery 日志的
// "runtime error: invalid memory address or nil pointer dereference"。
func panicNilDeref() {
	var p *int
	_ = *p
}

func TestGrpcRunner_RecoveryHandler(t *testing.T) {
	// 回归：panic 日志必须带 goroutine 栈快照定位起源帧。此前 recoveryHandler 只打
	// panic 值（%v），而 nil 解引用等运行时 panic 的 recover 值只是 runtime.Error——
	// 既无 pkg/errors 栈也非 fmt.Formatter，%+v 对它无效，无快照即无从定位 panic 点。
	out := captureStdout(t, func() {
		runner := &grpcRunner{logger: mlog.NewForConfig(nil)}
		func() {
			// 与 grpc_recovery 的 defer→recover→handler 同形：unwind 结束前抓栈，
			// 栈快照才含 panicNilDeref 起源帧（recover 返回后再调用只剩恢复点）。
			defer func() {
				got := runner.recoveryHandler(recover())
				assert.Equal(t, codes.Internal, status.Code(got), "panic 应转 Internal 返回，客户端收到明确失败")
			}()
			panicNilDeref()
		}()
	})
	assert.Contains(t, out, "invalid memory address or nil pointer dereference", "日志应含 panic 值")
	assert.Contains(t, out, "panicNilDeref", "栈快照应含 nil deref 起源函数帧，而非仅恢复点")
}

func TestAuthenticate(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	// 无 token：AuthFromMD 直接报错，不触达 authBiz。
	runner := &grpcRunner{}
	_, err := runner.authenticate(context.TODO())
	assert.Error(t, err)

	md := metadata.New(map[string]string{"authorization": "Bearer xxx"})
	incomingContext := metadata.NewIncomingContext(context.TODO(), md)
	authS := biz.NewMockAuthBiz(m)
	runner = &grpcRunner{authBiz: authS}

	// 无效 token：返回 Unauthenticated。
	authS.EXPECT().VerifyToken(gomock.Any(), "xxx").Return(nil, status.Errorf(codes.Unauthenticated, "Unauthenticated."))
	_, err = runner.authenticate(incomingContext)
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, fromError.Code())

	// 有效 token：注入用户，并经 authBiz.EffectiveRoles 按 users 表接管状态覆盖生效角色——
	// 后台手动降权后即使 JWT 仍带 mars_admin，生效角色也不含管理员（RequireAdmin 据此拒绝）。
	authS.EXPECT().VerifyToken(gomock.Any(), "xxx").Return(&biz.UserInfo{Name: "duc", Email: "duc@x.com", Roles: []string{biz.MarsAdmin}}, nil)
	authS.EXPECT().EffectiveRoles(gomock.Any(), "duc@x.com", []string{biz.MarsAdmin}).Return([]string{}, nil)
	ctx2, err := runner.authenticate(incomingContext)
	assert.Nil(t, err)
	assert.Equal(t, "duc", biz.MustGetUser(ctx2).Name)
	assert.Empty(t, biz.MustGetUser(ctx2).Roles, "降权后生效角色不应含 mars_admin")
}

func TestNewGrpcRunner(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	appMock := app.NewMockApp(m)
	appMock.EXPECT().GrpcRegistry().Return(nil).Times(1)
	appMock.EXPECT().Logger().Return(mlog.NewForConfig(nil)).Times(1)
	appMock.EXPECT().AuthBiz().Return(biz.NewMockAuthBiz(m)).Times(1)

	runner := NewGrpcRunner("test-endpoint", appMock)

	assert.NotNil(t, runner)
	assert.Equal(t, "test-endpoint", runner.(*grpcRunner).endpoint)
}

func TestGrpcRunner_Shutdown(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	server := NewMockGrpcServerImp(m)
	runner := &grpcRunner{
		logger: mlog.NewForConfig(nil),
		server: server,
	}

	// Test case 1：ctx 未到期，GracefulStop 完成后 Shutdown 返回 nil。
	server.EXPECT().GracefulStop().Times(1)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	err := runner.Shutdown(ctx)
	assert.Nil(t, err)

	// Test case 2：ctx 已取消且优雅停止迟迟不结束，Shutdown 须走 ctx.Done() 分支返回错误。
	// 若 mock 的 GracefulStop 立即返回，done 与 ctx.Done() 同时就绪，select 随机选择，
	// 约一半概率误返 nil 造成 CI 偶发失败——用 release 阻塞 GracefulStop 使 done 永不就绪，
	// Shutdown 只能经 ctx 取消返回，测试结果确定；stopped 确保 mock 调用完成后测试才结束，
	// 避免 m.Finish() 与 Shutdown 内部 goroutine 竞态（原来靠 time.Sleep 兜底）。
	release := make(chan struct{})
	stopped := make(chan struct{})
	server.EXPECT().GracefulStop().Do(func() {
		<-release
		close(stopped)
	}).Times(1)

	ctx, cancel = context.WithCancel(context.TODO())
	cancel()

	err = runner.Shutdown(ctx)
	assert.NotNil(t, err)
	close(release)
	<-stopped
}

func Test_grpcRunner_initServer(t *testing.T) {
	var ss any
	(&grpcRunner{
		grpcRegistry: &app.GrpcRegistry{
			RegistryFunc: func(s grpc.ServiceRegistrar) {
				ss = s
			},
		},
	}).initServer()

	assert.NotNil(t, ss)
}

// echoTestServiceIface 是测试服务的 handler 接口：grpc 的 RegisterService 要求
// ServiceDesc.HandlerType 是接口类型（内部经 reflect.Elem 取接口做 Implements 校验），
// 传具体结构体会直接 panic。
type echoTestServiceIface interface {
	Echo(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
}

// echoTestService 是测试专用的最小 gRPC 服务：用手写 ServiceDesc 注册，避免为单测引入
// proto 代码生成。它让 initServer 装配的真实拦截器链（含 authFn 鉴权闭包）能经真实
// Unary RPC 执行——纯函数级测试无法触达 authFn 闭包内部。
type echoTestService struct{}

func (echoTestService) Echo(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

var echoTestServiceDesc = grpc.ServiceDesc{
	ServiceName: "echo.Test",
	HandlerType: (*echoTestServiceIface)(nil),
	Methods: []grpc.MethodDesc{{
		MethodName: "Echo",
		Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
			in := new(emptypb.Empty)
			if err := dec(in); err != nil {
				return nil, err
			}
			if interceptor == nil {
				return srv.(echoTestService).Echo(ctx, in)
			}
			info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/echo.Test/Echo"}
			handler := func(ctx context.Context, req any) (any, error) {
				return srv.(echoTestService).Echo(ctx, req.(*emptypb.Empty))
			}
			return interceptor(ctx, in, info, handler)
		},
	}},
	Streams:  []grpc.StreamDesc{},
	Metadata: "echo.proto",
}

// panickingAuthorizeService 是授权阶段抛 panic 的测试服务：Authorize 直接 panic，
// 用于验证「Recovery 在最外层能捕获外层拦截器（Auth）的 panic 并返回 Internal」。
// 若 Recovery 被挪回最内层，Auth 的 panic 会穿透所有无 recover 的拦截器——grpc-go
// 自身无内置 recover，进程会被击穿；此测试即为这个链序不变量兜底。
type panickingAuthorizeService struct{}

func (panickingAuthorizeService) Echo(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (panickingAuthorizeService) Authorize(_ context.Context, _ string) (context.Context, error) {
	panic("authorize panic")
}

// newBufconnGRPCRunner 构造一个挂在 bufconn 监听器上的 grpcRunner：服务经 initServer
// 装配，客户端经 bufconn dialer 直连，不需要真实端口。
func newBufconnGRPCRunner(m *gomock.Controller) (*grpcRunner, *bufconn.Listener) {
	authBiz := biz.NewMockAuthBiz(m)
	authBiz.EXPECT().VerifyToken(gomock.Any(), "tok").Return(&biz.UserInfo{Name: "duc"}, nil).AnyTimes()
	authBiz.EXPECT().EffectiveRoles(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil).AnyTimes()
	return &grpcRunner{
		logger:  mlog.NewForConfig(nil),
		authBiz: authBiz,
		grpcRegistry: &app.GrpcRegistry{
			RegistryFunc: func(s grpc.ServiceRegistrar) {
				s.RegisterService(&echoTestServiceDesc, echoTestService{})
			},
		},
	}, bufconn.Listen(1024 * 1024)
}

// Test_grpcRunner_initServer_UnaryRPC 经真实 Unary RPC 驱动 initServer 的拦截器链：
// 携带 Bearer token 的请求命中 authFn → biz.Authenticate 成功并注入用户后进 handler；
// 缺失 token 的请求在 authFn 处即被拒（Unauthenticated）。两个方向覆盖 authFn 闭包的
// 全部执行路径，补上 Test_grpcRunner_initServer 只验注册不验调用的缺口。
func Test_grpcRunner_initServer_UnaryRPC(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	runner, lis := newBufconnGRPCRunner(m)
	defer lis.Close()

	srv := runner.initServer()
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
	)
	assert.Nil(t, err)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	// 有效 token：authFn 通过，handler 正常返回。
	authCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer tok"))
	var resp emptypb.Empty
	assert.Nil(t, conn.Invoke(authCtx, "/echo.Test/Echo", &emptypb.Empty{}, &resp))

	// 缺失 token：authFn 拒绝，返回 Unauthenticated。
	err = conn.Invoke(ctx, "/echo.Test/Echo", &emptypb.Empty{}, &resp)
	assert.NotNil(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

// Test_grpcRunner_initServer_InterceptorPanicRecovered 驱动真实链上的拦截器 panic：
// 服务实现 middlewares.Authorize（AuthUnaryServerInterceptor 的类型断言目标）且 Authorize
// 直接 panic，panic 从 Auth 拦截器冒出，被最外层 Recovery 捕获并转为 Internal 返回客户端。
// 此前只有 recoveryHandler 的单元测试，链上 panic 无任何断言——变异（Recovery 挪回最内）
// 时全部测试仍绿，本测试为「Recovery 必须最外层」的链序不变量补上承重断言。
func Test_grpcRunner_initServer_InterceptorPanicRecovered(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	authBiz := biz.NewMockAuthBiz(m)
	authBiz.EXPECT().VerifyToken(gomock.Any(), "tok").Return(&biz.UserInfo{Name: "duc"}, nil).AnyTimes()
	authBiz.EXPECT().EffectiveRoles(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil).AnyTimes()

	runner := &grpcRunner{
		logger:  mlog.NewForConfig(nil),
		authBiz: authBiz,
		grpcRegistry: &app.GrpcRegistry{
			RegistryFunc: func(s grpc.ServiceRegistrar) {
				s.RegisterService(&echoTestServiceDesc, panickingAuthorizeService{})
			},
		},
	}
	lis := bufconn.Listen(1024 * 1024)
	defer lis.Close()

	srv := runner.initServer()
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
	)
	assert.Nil(t, err)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	authCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer tok"))
	var resp emptypb.Empty
	err = conn.Invoke(authCtx, "/echo.Test/Echo", &emptypb.Empty{}, &resp)
	// 服务端 Auth 拦截器 panic → 最外层 Recovery 捕获 → 客户端收到 Internal（进程未击穿）。
	assert.Equal(t, codes.Internal, status.Code(err))
}

// Test_grpcRunner_initServer_AccessLogPrintsUser 驱动真实 Unary RPC 验证链序不变量：
// AccessLog 位于 Login 之后——携带 Bearer token 的请求先经 Login 注入用户再进 AccessLog，
// 访问日志须打印出调用用户"duc"。此断言为「Login 须在 AccessLog 之前」的承重护栏：
// 若将来有人把 AccessLog 挪回 Login 外层，defer 持原始 ctx 导致 grpcUser 返回匿名，
// Infof 的 user 参数不再等于 "duc"，断言即失败。
func Test_grpcRunner_initServer_AccessLogPrintsUser(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	logger := mlog.NewMockLogger(m)
	// 请求体是 proto 消息（实现 String()），LoggerUnaryServerInterceptor 记一条 Debugf；
	// 用 AnyTimes 容错——本测试的承重断言是下面的 Infof。
	logger.EXPECT().Debugf("[request logger]: method=%s body=%v", "/echo.Test/Echo", gomock.Any()).AnyTimes()
	// 核心断言：AccessLog 打印出 Login 注入的用户名。
	logger.EXPECT().Infof("[Grpc]: user: %v, visit: %v, use: %s.", "duc", "/echo.Test/Echo", gomock.Any()).Times(1)

	authBiz := biz.NewMockAuthBiz(m)
	authBiz.EXPECT().VerifyToken(gomock.Any(), "tok").Return(&biz.UserInfo{Name: "duc"}, nil).AnyTimes()
	authBiz.EXPECT().EffectiveRoles(gomock.Any(), gomock.Any(), gomock.Any()).Return([]string{}, nil).AnyTimes()

	runner := &grpcRunner{
		logger:  logger,
		authBiz: authBiz,
		grpcRegistry: &app.GrpcRegistry{
			RegistryFunc: func(s grpc.ServiceRegistrar) {
				s.RegisterService(&echoTestServiceDesc, echoTestService{})
			},
		},
	}
	lis := bufconn.Listen(1024 * 1024)
	defer lis.Close()

	srv := runner.initServer()
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
	)
	assert.Nil(t, err)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	authCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer tok"))
	var resp emptypb.Empty
	assert.Nil(t, conn.Invoke(authCtx, "/echo.Test/Echo", &emptypb.Empty{}, &resp))
}

// Test_grpcRunner_Run_SuccessAndServe 覆盖 Run 的成功路径：真实端口上 net.Listen、
// initServer 装配、goroutine 内 Serve 拉起，Shutdown 触发的 GracefulStop 让 Serve 返回
// 并被 goroutine 吞掉。此前 Run 全程零覆盖。
func Test_grpcRunner_Run_SuccessAndServe(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	assert.Nil(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	m := gomock.NewController(t)
	defer m.Finish()
	authBiz := biz.NewMockAuthBiz(m)
	runner := &grpcRunner{
		endpoint: fmt.Sprintf("127.0.0.1:%d", port),
		logger:   mlog.NewForConfig(nil),
		authBiz:  authBiz,
		grpcRegistry: &app.GrpcRegistry{
			RegistryFunc: func(s grpc.ServiceRegistrar) {},
		},
	}

	assert.Nil(t, runner.Run(context.TODO()))
	assert.Nil(t, runner.Shutdown(context.TODO()))
}

// Test_grpcRunner_Run_ListenError 覆盖 Run 的失败路径：非法 endpoint 让 net.Listen
// 报错并直接返回，不进入 initServer/Serve。
func Test_grpcRunner_Run_ListenError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	runner := &grpcRunner{
		endpoint: "not-a-valid-address",
		logger:   mlog.NewForConfig(nil),
		authBiz:  biz.NewMockAuthBiz(m),
		grpcRegistry: &app.GrpcRegistry{
			RegistryFunc: func(s grpc.ServiceRegistrar) {},
		},
	}

	assert.Error(t, runner.Run(context.TODO()))
}
