package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authpb "github.com/duc-cnzj/mars/api/v6/proto/auth"
	metricspb "github.com/duc-cnzj/mars/api/v6/proto/metrics"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestNewApiGateway(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := app.NewMockApp(m)

	app.EXPECT().Logger().Return(mlog.NewForConfig(nil)).Times(1)
	app.EXPECT().GrpcRegistry().Return(nil).Times(1)
	app.EXPECT().Config().Return(&config.Config{}).Times(1)
	app.EXPECT().HttpHandler().Return(nil).Times(1)

	gw := NewApiGateway("test-endpoint", app)
	assert.NotNil(t, gw)
}

func Test_apiGateway_Run(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	server := NewMockHttpServer(m)
	handler := app.NewMockHttpHandler(m)
	logger := mlog.NewMockLogger(m)
	gw := &apiGateway{
		handler: handler,
		newServerFunc: func(ctx context.Context, a *apiGateway) (HttpServer, error) {
			return server, nil
		},
		logger: logger,
		port:   "111",
	}

	handler.EXPECT().TickClusterHealth(gomock.Any())
	logger.EXPECT().Infof("[Server]: start apiGateway runner at :%s.", "111").Times(1)
	server.EXPECT().ListenAndServe().Return(assert.AnError).Times(1)
	logger.EXPECT().Error(gomock.Any()).Times(1)
	err := gw.Run(context.TODO())
	time.Sleep(1 * time.Second)
	assert.NoError(t, err)
}

func Test_apiGateway_Shutdown(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	server := NewMockHttpServer(m)
	logger := mlog.NewMockLogger(m)

	handler := app.NewMockHttpHandler(m)
	gw := &apiGateway{
		handler: handler,
		server:  server,
		logger:  logger,
	}
	handler.EXPECT().Shutdown(gomock.Any())
	logger.EXPECT().Info("[Server]: shutdown api-gateway runner.").Times(1)
	server.EXPECT().Shutdown(gomock.Any()).Return(nil).Times(1)
	assert.Nil(t, gw.Shutdown(context.TODO()))
}

func TestMiddlewareList_Wrap(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	middleware := func(logger mlog.Logger, handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "middleware")
			handler.ServeHTTP(w, r)
		})
	}

	middlewareList := middlewareList{middleware}

	wrappedHandler := middlewareList.Wrap(logger, handler)

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, "middleware", rr.Header().Get("X-Test"))
	assert.Equal(t, "test", rr.Body.String())
}

func TestMiddlewareList_Wrap_Empty(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	middlewareList := middlewareList{}

	wrappedHandler := middlewareList.Wrap(logger, handler)

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, "", rr.Header().Get("X-Test"))
	assert.Equal(t, "test", rr.Body.String())
}

func TestHeaderMatcher(t *testing.T) {
	// Test case: tracestate key
	key, ok := headerMatcher("tracestate")
	assert.True(t, ok)
	assert.Equal(t, "tracestate", key)

	// Test case: traceparent key
	key, ok = headerMatcher("traceparent")
	assert.True(t, ok)
	assert.Equal(t, "traceparent", key)

	// Test case: other key
	key, ok = headerMatcher("other")
	assert.False(t, ok)
	assert.Equal(t, "", key)

	// Test case: empty key
	key, ok = headerMatcher("")
	assert.False(t, ok)
	assert.Equal(t, "", key)
}

func Test_initServer(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	server := NewMockHttpServer(m)
	handler := app.NewMockHttpHandler(m)
	handler.EXPECT().RegisterSwaggerUIRoute(gomock.Not(nil)).Times(1)
	handler.EXPECT().RegisterWsRoute(gomock.Not(nil)).Times(1)
	handler.EXPECT().RegisterFileRoute(gomock.Not(nil)).Times(1)
	httpServer, err := initServer(context.TODO(), &apiGateway{
		endpoint:     "x",
		port:         "1000",
		server:       server,
		logger:       mlog.NewForConfig(nil),
		grpcRegistry: &app.GrpcRegistry{},
		handler:      handler,
	})
	assert.Nil(t, err)
	assert.NotNil(t, httpServer)
	assert.Equal(t, httpServer.(*http.Server).Addr, ":1000")
	assert.Equal(t, httpServer.(*http.Server).ReadHeaderTimeout, 5*time.Second)
}

// Test_apiGateway_Run_InitServerError 覆盖 Run 的装配失败分支：initServer 返回错误时
// 直接上抛，不启动任何协程。
func Test_apiGateway_Run_InitServerError(t *testing.T) {
	gw := &apiGateway{
		newServerFunc: func(ctx context.Context, a *apiGateway) (HttpServer, error) {
			return nil, errors.New("boom")
		},
	}
	err := gw.Run(context.TODO())
	assert.Error(t, err)
}

// Test_initServer_EndpointFuncError 覆盖 EndpointFuncs 注册失败分支：任一注册函数返回
// 错误即中止装配并上抛。
func Test_initServer_EndpointFuncError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	_, err := initServer(context.TODO(), &apiGateway{
		endpoint: "x",
		port:     "1000",
		logger:   mlog.NewForConfig(nil),
		grpcRegistry: &app.GrpcRegistry{
			EndpointFuncs: []app.EndpointFunc{
				func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
					return errors.New("boom")
				},
			},
		},
		handler: app.NewMockHttpHandler(m),
	})
	assert.Error(t, err)
}

// Test_initServer_RoutesAndClosures 通过真实请求驱动 initServer 装配的整条链路：用
// EndpointFunc 注册一个 grpc-gateway 路由，经 httptest 请求触发 ping 闭包、
// ForwardResponseOption（nosniff）闭包、otelhttp 过滤与 span 名格式化闭包。
func Test_initServer_RoutesAndClosures(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	handler := app.NewMockHttpHandler(m)
	handler.EXPECT().RegisterSwaggerUIRoute(gomock.Not(nil)).Times(1)
	handler.EXPECT().RegisterWsRoute(gomock.Not(nil)).Times(1)
	handler.EXPECT().RegisterFileRoute(gomock.Not(nil)).Times(1)

	grpcRegistry := &app.GrpcRegistry{
		EndpointFuncs: []app.EndpointFunc{
			func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
				// 生产环境 grpc-gateway 路由全部挂在 /api 下（proto http 注解约定），
				// 测试用 /api 前缀与 initServer 的 PathPrefix("/api/") 绑定对齐。
				return mux.HandlePath("GET", "/api/test/{name}",
					func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
						w.Write([]byte("gateway:" + pathParams["name"]))
					})
			},
		},
	}

	httpServer, err := initServer(context.TODO(), &apiGateway{
		endpoint:     "x",
		port:         "1000",
		logger:       mlog.NewForConfig(nil),
		grpcRegistry: grpcRegistry,
		handler:      handler,
	})
	assert.Nil(t, err)
	h := httpServer.(*http.Server).Handler

	// /ping：直接注册的处理函数。
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/ping", nil))
	assert.Equal(t, "pong", rr.Body.String())

	// 经 EndpointFunc 注册的 grpc-gateway 路由：验证 EndpointFuncs 循环装配链路。
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/api/test/foo", nil))
	assert.Equal(t, "gateway:foo", rr.Body.String())

	// /api 前缀与非 /api 前缀：分别覆盖 otelhttp.WithFilter 的 true/false 分支。
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/api/anything", nil))
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/nope", nil))
}

// Test_initServer_SpaFallback 覆盖 SPA 兜底路由（历史 P0 回归点）：/admin/* 这类多段前端
// 路由刷新/深链接时直接 GET 服务器，必须回 index.html 让前端 Router 接管；而 /api 前缀
// 仍由 grpc-gateway 处理，不能被 SPA 兜底吞掉。
func Test_initServer_SpaFallback(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	handler := app.NewMockHttpHandler(m)
	handler.EXPECT().RegisterSwaggerUIRoute(gomock.Not(nil)).Times(1)
	handler.EXPECT().RegisterWsRoute(gomock.Not(nil)).Times(1)
	handler.EXPECT().RegisterFileRoute(gomock.Not(nil)).Times(1)

	grpcRegistry := &app.GrpcRegistry{
		EndpointFuncs: []app.EndpointFunc{
			func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
				return mux.HandlePath("GET", "/api/version",
					func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
						w.Write([]byte(`{"api":"version"}`))
					})
			},
		},
	}

	httpServer, err := initServer(context.TODO(), &apiGateway{
		endpoint:     "x",
		port:         "1000",
		logger:       mlog.NewForConfig(nil),
		grpcRegistry: grpcRegistry,
		handler:      handler,
	})
	assert.Nil(t, err)
	h := httpServer.(*http.Server).Handler

	// 多段前端路由：刷新 /admin/cluster 必须回 index.html（本次 bug 的复现点）。
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/admin/cluster", nil))
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/html; charset=utf-8", rr.Header().Get("Content-Type"))
	assert.Contains(t, rr.Body.String(), "<!doctype html>")

	// 单段前端路由：/admin 同样回 index.html（既有行为不回退）。
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/admin", nil))
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "<!doctype html>")

	// /api 前缀不受 SPA 兜底影响，仍由 grpc-gateway 处理。
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/api/version", nil))
	assert.Equal(t, `{"api":"version"}`, rr.Body.String())
}

// Test_apiGateway_shouldTagRPC 覆盖 gRPC 统计过滤判定：白名单内方法不计入（返回 false）、
// 白名单外方法计入（返回 true），两次均打 Debugf。
func Test_apiGateway_shouldTagRPC(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	logger := mlog.NewMockLogger(m)
	gw := &apiGateway{logger: logger}

	logger.EXPECT().Debugf("%v\t%v", metricspb.Metrics_StreamTopPod_FullMethodName, false).Times(1)
	assert.False(t, gw.shouldTagRPC(&stats.RPCTagInfo{FullMethodName: metricspb.Metrics_StreamTopPod_FullMethodName}))

	logger.EXPECT().Debugf("%v\t%v", authpb.Auth_Login_FullMethodName, true).Times(1)
	assert.True(t, gw.shouldTagRPC(&stats.RPCTagInfo{FullMethodName: authpb.Auth_Login_FullMethodName}))
}

// Test_apiGateway_setNosniff 覆盖 ForwardResponseOption：REST 响应补 X-Content-Type-Options:
// nosniff 头，返回值恒为 nil。
func Test_apiGateway_setNosniff(t *testing.T) {
	gw := &apiGateway{}
	rr := httptest.NewRecorder()
	assert.Nil(t, gw.setNosniff(context.TODO(), rr, &emptypb.Empty{}))
	assert.Equal(t, "nosniff", rr.Header().Get("X-Content-Type-Options"))
}
