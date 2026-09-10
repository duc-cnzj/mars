package grpc

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/auth"
	"github.com/duc-cnzj/mars/api/v6/proto/container"
	"github.com/duc-cnzj/mars/api/v6/proto/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// fakeAuthServer 签发递增 token，非 admin/123456 拒绝；failAfter > 0 时在第 failAfter 次
// 成功签发后开始拒绝（用于模拟刷新时凭据失效）。
type fakeAuthServer struct {
	auth.UnimplementedAuthServer
	issued    int32
	failAfter int32
}

func (f *fakeAuthServer) Login(_ context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	if req.Username != "admin" || req.Password != "123456" ||
		(f.failAfter > 0 && atomic.LoadInt32(&f.issued) >= f.failAfter) {
		return nil, status.Error(codes.Unauthenticated, "bad credentials")
	}
	// token 长度 >6 才能触发 setToken 的 "Bearer " 前缀逻辑。
	return &auth.LoginResponse{Token: fmt.Sprintf("token-%d", atomic.AddInt32(&f.issued, 1))}, nil
}

// fakeVersionServer 可选"第一次必 401"：用于验证 WithTokenAutoRefresh 的刷新重试。
type fakeVersionServer struct {
	version.UnimplementedVersionServer
	failFirst bool
	seen      int32
}

func (f *fakeVersionServer) Version(_ context.Context, _ *version.Request) (*version.Response, error) {
	if f.failFirst && atomic.AddInt32(&f.seen, 1) == 1 {
		return nil, status.Error(codes.Unauthenticated, "stale token")
	}
	return &version.Response{Version: "v6.0.0"}, nil
}

func newBufconnServer(t *testing.T, register func(*grpc.Server)) *bufconn.Listener {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	register(srv)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	return lis
}

func withBufconnDialer(lis *bufconn.Listener) Option {
	return func(c *Client) {
		c.dialOptions = append(c.dialOptions, grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}))
	}
}

func newTestClient(t *testing.T, lis *bufconn.Listener, opts ...Option) *Client {
	t.Helper()
	opts = append(opts, withBufconnDialer(lis))
	cli, err := NewClient("passthrough:///bufnet", opts...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = cli.Close() })
	return cli
}

func TestNewClient_WithAuth_IssuesToken(t *testing.T) {
	lis := newBufconnServer(t, func(s *grpc.Server) { auth.RegisterAuthServer(s, &fakeAuthServer{}) })
	c := newTestClient(t, lis, WithAuth("admin", "123456"))
	if got := c.authToken(); got != "Bearer token-1" {
		t.Fatalf("authToken = %q, want %q", got, "Bearer token-1")
	}
	if !c.hasCredentials() {
		t.Fatal("hasCredentials 应为 true")
	}
}

func TestNewClient_WithAuth_BadCredentials_Error(t *testing.T) {
	lis := newBufconnServer(t, func(s *grpc.Server) { auth.RegisterAuthServer(s, &fakeAuthServer{}) })
	cli, err := NewClient("bufnet", WithAuth("wrong", "creds"), withBufconnDialer(lis))
	if err == nil {
		t.Fatal("错误凭据应返回 error")
	}
	_ = cli
}

// 畸形 target（非法 percent-encoding）构造即失败：透传 error，
// 而非留下 conn==nil 的 client（首次 RPC 会 nil panic）。
func TestNewClient_MalformedTarget_ReturnsError(t *testing.T) {
	if _, err := NewClient("%zz"); err == nil {
		t.Fatal("畸形 target 应返回 error")
	}
}

func TestNewClient_WithBearerToken_NoLogin(t *testing.T) {
	lis := newBufconnServer(t, func(s *grpc.Server) {})
	c := newTestClient(t, lis, WithBearerToken("tokenvalue123"))
	if got := c.authToken(); got != "Bearer tokenvalue123" {
		t.Fatalf("authToken = %q, want %q", got, "Bearer tokenvalue123")
	}
}

func TestNewClient_NoCredentials_NoToken(t *testing.T) {
	lis := newBufconnServer(t, func(s *grpc.Server) {})
	c := newTestClient(t, lis)
	if got := c.authToken(); got != "" {
		t.Fatalf("无凭据 authToken 应为空，实际 %q", got)
	}
	// 各 service 客户端已就位
	if c.Auth() == nil || c.Version() == nil || c.Namespace() == nil {
		t.Fatal("service 客户端不应为 nil")
	}
}

func TestTokenAutoRefresh_RetriesAfterUnauthenticated(t *testing.T) {
	lis := newBufconnServer(t, func(s *grpc.Server) {
		auth.RegisterAuthServer(s, &fakeAuthServer{})
		version.RegisterVersionServer(s, &fakeVersionServer{failFirst: true})
	})
	c := newTestClient(t, lis, WithAuth("admin", "123456"), WithTokenAutoRefresh())
	if got := c.authToken(); got != "Bearer token-1" {
		t.Fatalf("构造期 authToken = %q, want %q", got, "Bearer token-1")
	}
	// 弄脏 token 模拟过期，触发 401 → 自动刷新(tok-2) → 重试成功
	c.SetBearerToken("stale")
	res, err := c.Version().Version(context.Background(), &version.Request{})
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if res.GetVersion() != "v6.0.0" {
		t.Fatalf("Version = %q, want v6.0.0", res.GetVersion())
	}
	if got := c.authToken(); got != "Bearer token-2" {
		t.Fatalf("刷新后 authToken = %q, want %q", got, "Bearer token-2")
	}
}

func TestSetToken_NormalizesPrefix(t *testing.T) {
	c := &Client{}
	for _, tc := range []struct{ in, want string }{
		{"", ""},                          // 空 token 原样保留
		{"abc", "Bearer abc"},             // 短 token 也补前缀（与 http SDK 语义一致）
		{"abcdef", "Bearer abcdef"},       // 恰好 6 字符非 Bearer 也应补前缀（边界回归）
		{"Bearer def", "Bearer def"},      // 已带前缀保持
		{"bearer ghi", "bearer ghi"},      // 大小写不敏感保持
		{"bearertok", "Bearer bearertok"}, // 缺空格（恰以 bearer 开头）也需补前缀
		{"x12345678", "Bearer x12345678"}, // 无前缀加 Bearer
	} {
		c.setToken(tc.in)
		if got := c.authToken(); got != tc.want {
			t.Errorf("setToken(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestClientauth_GetRequestMetadata(t *testing.T) {
	c := &Client{}
	c.setToken("Bearer xyz")
	a := &clientauth{c: c}
	md, err := a.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if md["Authorization"] != "Bearer xyz" {
		t.Fatalf("Authorization = %q", md["Authorization"])
	}
	if a.RequireTransportSecurity() {
		t.Fatal("RequireTransportSecurity 应为 false")
	}
}

func TestOptions_AppendInterceptors(t *testing.T) {
	var unary, stream int
	c := &Client{}
	WithUnaryClientInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		unary++
		return invoker(ctx, method, req, reply, cc, opts...)
	})(c)
	WithStreamClientInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		stream++
		return streamer(ctx, desc, cc, method, opts...)
	})(c)
	if len(c.UnaryClientInterceptors) != 1 || len(c.StreamClientInterceptors) != 1 {
		t.Fatalf("拦截器未追加: unary=%d stream=%d", len(c.UnaryClientInterceptors), len(c.StreamClientInterceptors))
	}
}

// 全部 17 个 service 访问器都应返回非 nil 客户端。
func TestServiceAccessors_AllWired(t *testing.T) {
	lis := newBufconnServer(t, func(s *grpc.Server) {})
	c := newTestClient(t, lis)
	for name, svc := range map[string]interface{}{
		"Auth":        c.Auth(),
		"Repo":        c.Repo(),
		"Changelog":   c.Changelog(),
		"Cluster":     c.Cluster(),
		"Container":   c.Container(),
		"Event":       c.Event(),
		"AccessToken": c.AccessToken(),
		"File":        c.File(),
		"Git":         c.Git(),
		"Metrics":     c.Metrics(),
		"Namespace":   c.Namespace(),
		"Picture":     c.Picture(),
		"Project":     c.Project(),
		"Version":     c.Version(),
		"Endpoint":    c.Endpoint(),
		"Settings":    c.Settings(),
		"User":        c.User(),
	} {
		if svc == nil {
			t.Errorf("访问器 %s 返回 nil", name)
		}
	}
}

func TestClose_NilConn(t *testing.T) {
	c := &Client{}
	if err := c.Close(); err != nil {
		t.Fatalf("nil conn Close 应返回 nil，实际 %v", err)
	}
}

// 回归：WithAuth + WithTokenAutoRefresh 且运行时凭据错误时，NewClient 必须返回 error
// 而不是挂死。旧实现 guard 写错方法路径（"/Auth/Login"），Login 401 时递归触发
// getToken → flight.Do 自我死锁。
func TestTokenAutoRefresh_BadRuntimeCreds_NoDeadlock(t *testing.T) {
	lis := newBufconnServer(t, func(s *grpc.Server) { auth.RegisterAuthServer(s, &fakeAuthServer{}) })
	done := make(chan error, 1)
	go func() {
		_, err := NewClient("passthrough:///bufnet",
			WithAuth("admin", "wrong-pass"),
			WithTokenAutoRefresh(),
			withBufconnDialer(lis),
		)
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("错误凭据应返回 error")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("NewClient 挂起（自动刷新与 singleflight 自我死锁）")
	}
}

// selfSignedServerTLS 生成内存中的自签名证书，用作 bufconn TLS 服务端凭据。
func selfSignedServerTLS(t *testing.T) *tls.Config {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "bufnet"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost"},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{{
		Certificate: [][]byte{der},
		PrivateKey:  key,
	}}}
}

// TLS 传输凭据：自签名证书 + InsecureSkipVerify 走完整握手完成一次 unary RPC。
func TestWithTransportCredentials_TLS_RoundTrip(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer(grpc.Creds(credentials.NewTLS(selfSignedServerTLS(t))))
	version.RegisterVersionServer(srv, &fakeVersionServer{})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	c := newTestClient(t, lis, WithTransportCredentials(&tls.Config{InsecureSkipVerify: true}))
	res, err := c.Version().Version(context.Background(), &version.Request{})
	if err != nil {
		t.Fatalf("Version over TLS: %v", err)
	}
	if res.GetVersion() != "v6.0.0" {
		t.Fatalf("Version = %q, want v6.0.0", res.GetVersion())
	}
}

// fakeContainerStreamServer 校验 Authorization 元数据：携带 stale token 时流返回 401，
// 否则下发两行日志。
type fakeContainerStreamServer struct {
	container.UnimplementedContainerServer
}

func (f *fakeContainerStreamServer) StreamContainerLog(_ *container.LogRequest, stream grpc.ServerStreamingServer[container.LogResponse]) error {
	// SetBearerToken("stale") 经 setToken 补前缀后到达的 Authorization 是 "Bearer stale"。
	if md, _ := metadata.FromIncomingContext(stream.Context()); len(md.Get("authorization")) == 0 || md.Get("authorization")[0] == "Bearer stale" {
		return status.Error(codes.Unauthenticated, "stale token")
	}
	if err := stream.Send(&container.LogResponse{Log: "line1"}); err != nil {
		return err
	}
	return stream.Send(&container.LogResponse{Log: "line2"})
}

// server-streaming 自动刷新：流首个消息遇 401 → 刷新 token → 重建流 → 继续读取。
func TestStreamAutoRefresh_ReestablishesAfterUnauthenticated(t *testing.T) {
	lis := newBufconnServer(t, func(s *grpc.Server) {
		auth.RegisterAuthServer(s, &fakeAuthServer{})
		container.RegisterContainerServer(s, &fakeContainerStreamServer{})
	})
	c := newTestClient(t, lis, WithAuth("admin", "123456"), WithTokenAutoRefresh())
	if got := c.authToken(); got != "Bearer token-1" {
		t.Fatalf("构造期 authToken = %q, want %q", got, "Bearer token-1")
	}
	c.SetBearerToken("stale") // 弄脏 token 模拟过期

	stream, err := c.Container().StreamContainerLog(context.Background(), &container.LogRequest{
		Namespace: "ns", Pod: "pod", Container: "app",
	})
	if err != nil {
		t.Fatalf("StreamContainerLog: %v", err)
	}
	var logs []string
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("Recv: %v", err)
		}
		logs = append(logs, resp.GetLog())
	}
	if len(logs) != 2 || logs[0] != "line1" || logs[1] != "line2" {
		t.Fatalf("logs = %v, want [line1 line2]", logs)
	}
	if got := c.authToken(); got != "Bearer token-2" {
		t.Fatalf("刷新后 authToken = %q, want %q", got, "Bearer token-2")
	}
}

// server-streaming 自动刷新但重登失败：流应返回登录错误，而非挂起或吞错。
func TestStreamAutoRefresh_RefreshLoginFails_ReturnsError(t *testing.T) {
	lis := newBufconnServer(t, func(s *grpc.Server) {
		auth.RegisterAuthServer(s, &fakeAuthServer{failAfter: 1}) // 构造期签发后拒绝续期
		container.RegisterContainerServer(s, &fakeContainerStreamServer{})
	})
	c := newTestClient(t, lis, WithAuth("admin", "123456"), WithTokenAutoRefresh())
	if got := c.authToken(); got != "Bearer token-1" {
		t.Fatalf("构造期 authToken = %q, want %q", got, "Bearer token-1")
	}
	c.SetBearerToken("stale")

	stream, err := c.Container().StreamContainerLog(context.Background(), &container.LogRequest{
		Namespace: "ns", Pod: "pod", Container: "app",
	})
	if err != nil {
		t.Fatalf("StreamContainerLog: %v", err)
	}
	if _, err := stream.Recv(); err == nil {
		t.Fatal("重登失败时 Recv 应返回 error")
	}
}

func TestSkipAutoRefresh(t *testing.T) {
	c := &Client{}
	if !c.skipAutoRefresh("/version.Version/Version") {
		t.Fatal("无凭据时任何方法都应跳过刷新")
	}
	c.username, c.password = "admin", "123456"
	if !c.skipAutoRefresh(auth.Auth_Login_FullMethodName) {
		t.Fatal("Login 自身应跳过刷新")
	}
	if !c.skipAutoRefresh(auth.Auth_Exchange_FullMethodName) {
		t.Fatal("Exchange 自身应跳过刷新")
	}
	if c.skipAutoRefresh("/version.Version/Version") {
		t.Fatal("有凭据的业务方法不应跳过")
	}
}

func TestWithTracer_AppendsStatsHandler(t *testing.T) {
	c := &Client{}
	before := len(c.dialOptions)
	WithTracer()(c)
	if len(c.dialOptions) != before+1 {
		t.Fatalf("WithTracer 应追加一个 dial option，before=%d after=%d", before, len(c.dialOptions))
	}
}

// fakeClientStream 仅用于测试流拦截器的包装判断，不实际收发。
type fakeClientStream struct {
	grpc.ClientStream
}

func (f *fakeClientStream) RecvMsg(interface{}) error { return nil }
func (f *fakeClientStream) SendMsg(interface{}) error { return nil }

// 流拦截器只应包装纯 server-streaming；bidi/client/unary 直接透传原始流。
func TestWithTokenAutoRefresh_StreamInterceptor_WrapsOnlyServerStreams(t *testing.T) {
	c := &Client{username: "admin", password: "123456"}
	WithTokenAutoRefresh()(c)
	inter := c.StreamClientInterceptors[0]

	for _, tc := range []struct {
		name string
		desc *grpc.StreamDesc
		want bool // 是否应被 autoRefreshStream 包装
	}{
		{"server-streaming", &grpc.StreamDesc{ServerStreams: true}, true},
		{"bidi", &grpc.StreamDesc{ServerStreams: true, ClientStreams: true}, false},
		{"client-streaming", &grpc.StreamDesc{ClientStreams: true}, false},
		{"unary-like", &grpc.StreamDesc{}, false},
	} {
		cs, err := inter(context.Background(), tc.desc, nil, "/x.X/Y",
			func(_ context.Context, _ *grpc.StreamDesc, _ *grpc.ClientConn, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
				return &fakeClientStream{}, nil
			}, nil)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if _, wrapped := cs.(*autoRefreshStream); wrapped != tc.want {
			t.Errorf("%s: wrapped=%v, want %v", tc.name, wrapped, tc.want)
		}
	}
}

// fakeAuthClient 直接注入 c.auth，绕开真实 bufconn 的 Login 服务，
// 用于在纯单元级构造出"getToken 成功/失败"两种精确场景。
type fakeAuthClient struct {
	token string
	err   error
}

func (f *fakeAuthClient) Login(_ context.Context, _ *auth.LoginRequest, _ ...grpc.CallOption) (*auth.LoginResponse, error) {
	return &auth.LoginResponse{Token: f.token}, f.err
}
func (f *fakeAuthClient) Info(context.Context, *auth.InfoRequest, ...grpc.CallOption) (*auth.InfoResponse, error) {
	return nil, nil
}
func (f *fakeAuthClient) Settings(context.Context, *auth.SettingsRequest, ...grpc.CallOption) (*auth.SettingsResponse, error) {
	return nil, nil
}
func (f *fakeAuthClient) Exchange(context.Context, *auth.ExchangeRequest, ...grpc.CallOption) (*auth.ExchangeResponse, error) {
	return nil, nil
}

// fakeRecvStream 允许单独配置 RecvMsg/SendMsg 的返回，用于验证重建流的失败路径。
type fakeRecvStream struct {
	grpc.ClientStream
	recvErr error
	sendErr error
}

func (f *fakeRecvStream) RecvMsg(interface{}) error { return f.recvErr }
func (f *fakeRecvStream) SendMsg(interface{}) error { return f.sendErr }

// unary 刷新重试中 getToken 失败：应返回登录错误而非吞掉；用超时 ctx 控制 backoff 提前终止。
func TestTokenAutoRefresh_UnaryGetTokenFails_ReturnsError(t *testing.T) {
	lis := newBufconnServer(t, func(s *grpc.Server) {
		auth.RegisterAuthServer(s, &fakeAuthServer{failAfter: 1}) // 构造期签发后拒绝续期
		version.RegisterVersionServer(s, &fakeVersionServer{failFirst: true})
	})
	c := newTestClient(t, lis, WithAuth("admin", "123456"), WithTokenAutoRefresh())
	if got := c.authToken(); got != "Bearer token-1" {
		t.Fatalf("构造期 authToken = %q, want %q", got, "Bearer token-1")
	}
	c.SetBearerToken("stale")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := c.Version().Version(ctx, &version.Request{})
	if err == nil {
		t.Fatal("getToken 失败时 unary 调用应返回 error")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		// 合法退出路径：backoff 感知 ctx 超时后提前终止，避免长时间退避。
		return
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("err code = %v, want Unauthenticated（登录凭据失效）", status.Code(err))
	}
}

// stream 拦截器里 NewStream 本身失败：应原样透传 streamer 的错误。
func TestWithTokenAutoRefresh_StreamNewStreamFails(t *testing.T) {
	c := &Client{username: "admin", password: "123456"}
	WithTokenAutoRefresh()(c)
	inter := c.StreamClientInterceptors[0]

	_, err := inter(context.Background(), &grpc.StreamDesc{ServerStreams: true}, nil, "/x.Y/Z",
		func(context.Context, *grpc.StreamDesc, *grpc.ClientConn, string, ...grpc.CallOption) (grpc.ClientStream, error) {
			return nil, errors.New("newstream boom")
		}, nil)
	if err == nil || err.Error() != "newstream boom" {
		t.Fatalf("err = %v, want newstream boom", err)
	}
}

// autoRefreshStream 重建流的 NewStream 失败：返回 streamer 错误，且已标记 refreshed。
func TestAutoRefreshStream_ReestablishNewStreamFails(t *testing.T) {
	c := &Client{username: "admin", password: "123456"}
	c.auth = &fakeAuthClient{token: "token-2"}

	s := &autoRefreshStream{
		c:      c,
		ctx:    context.Background(),
		desc:   &grpc.StreamDesc{ServerStreams: true},
		method: "/container.Container/StreamContainerLog",
		streamer: func(context.Context, *grpc.StreamDesc, *grpc.ClientConn, string, ...grpc.CallOption) (grpc.ClientStream, error) {
			return nil, errors.New("reestablish boom")
		},
		ClientStream: &fakeRecvStream{recvErr: status.Error(codes.Unauthenticated, "stale")},
	}
	err := s.RecvMsg(&container.LogResponse{})
	if err == nil || err.Error() != "reestablish boom" {
		t.Fatalf("err = %v, want reestablish boom", err)
	}
	if !s.refreshed {
		t.Fatal("重建流前应已标记 refreshed")
	}
}

// autoRefreshStream 重建后重发请求消息失败：应返回 SendMsg 的错误。
func TestAutoRefreshStream_ReestablishSendMsgFails(t *testing.T) {
	c := &Client{username: "admin", password: "123456"}
	c.auth = &fakeAuthClient{token: "token-2"}

	s := &autoRefreshStream{
		c:      c,
		ctx:    context.Background(),
		desc:   &grpc.StreamDesc{ServerStreams: true},
		method: "/container.Container/StreamContainerLog",
		streamer: func(context.Context, *grpc.StreamDesc, *grpc.ClientConn, string, ...grpc.CallOption) (grpc.ClientStream, error) {
			return &fakeRecvStream{sendErr: errors.New("send boom")}, nil
		},
		ClientStream: &fakeRecvStream{recvErr: status.Error(codes.Unauthenticated, "stale")},
	}
	err := s.RecvMsg(&container.LogResponse{})
	if err == nil || err.Error() != "send boom" {
		t.Fatalf("err = %v, want send boom", err)
	}
}
