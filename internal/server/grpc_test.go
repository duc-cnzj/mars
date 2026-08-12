package server

import (
	"context"
	"errors"
	"fmt"
	"net"
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

func TestGrpcRunner_RecoveryHandler(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	logger := mlog.NewMockLogger(m)
	auth := biz.NewMockAuthBiz(m)

	runner := &grpcRunner{
		logger:  logger,
		authBiz: auth,
	}

	// Test case: recoveryHandler 记录 panic 值并返回 Internal 错误，
	// 客户端收到明确失败而非成功空响应。
	err := errors.New("test error")
	logger.EXPECT().Errorf("[Grpc]: recovery error: \n%v", err).Times(1)

	got := runner.recoveryHandler(err)
	assert.Error(t, got)
	assert.Equal(t, codes.Internal, status.Code(got))
}

func TestAuthenticate(t *testing.T) {
	_, err := authenticate(context.TODO(), nil)
	assert.Error(t, err)
	md := metadata.New(map[string]string{"authorization": "Bearer xxx"})

	m := gomock.NewController(t)
	defer m.Finish()
	authS := biz.NewMockAuthBiz(m)
	authS.EXPECT().VerifyToken(gomock.Any(), "xxx").Return(nil, status.Errorf(codes.Unauthenticated, "Unauthenticated."))
	incomingContext := metadata.NewIncomingContext(context.TODO(), md)
	_, err = authenticate(incomingContext, authS)
	fromError, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, fromError.Code())

	authS.EXPECT().VerifyToken(gomock.Any(), "xxx").Return(&biz.UserInfo{Name: "duc"}, nil)
	ctx2, err := authenticate(incomingContext, authS)
	assert.Nil(t, err)
	assert.Equal(t, "duc", biz.MustGetUser(ctx2).Name)
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

	server.EXPECT().GracefulStop().Times(2)
	// Test case: Shutdown completes before context deadline
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	err := runner.Shutdown(ctx)
	assert.Nil(t, err)

	// Test case: Shutdown does not complete before context deadline
	ctx, cancel = context.WithCancel(context.TODO())
	cancel()

	err = runner.Shutdown(ctx)
	assert.NotNil(t, err)
	time.Sleep(time.Second)
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
