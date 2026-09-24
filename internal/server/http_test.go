package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	authpb "github.com/duc-cnzj/mars/api/v6/proto/auth"
	metricspb "github.com/duc-cnzj/mars/api/v6/proto/metrics"
	projectpb "github.com/duc-cnzj/mars/api/v6/proto/project"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/server/middlewares"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
	"google.golang.org/protobuf/proto"
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

// Test_apiGateway_setOidcStateCookie_Settings 覆盖 Settings 响应下发 state Cookie：
// 取所有 provider 共用的 state 写入 HttpOnly + SameSite=Lax + 限定路径的 Cookie。
func Test_apiGateway_setOidcStateCookie_Settings(t *testing.T) {
	gw := &apiGateway{}

	t.Run("有 provider 时下发 state", func(t *testing.T) {
		rr := httptest.NewRecorder()
		err := gw.setOidcStateCookie(context.TODO(), rr, &authpb.SettingsResponse{
			Items: []*authpb.SettingsResponse_OidcSetting{
				{Name: "a", State: "shared-state"},
				{Name: "b", State: "shared-state"},
			},
		})
		assert.Nil(t, err)
		cookies := rr.Result().Cookies()
		assert.Len(t, cookies, 1)
		assert.Equal(t, middlewares.OidcStateCookieName, cookies[0].Name)
		assert.Equal(t, "shared-state", cookies[0].Value)
		assert.Equal(t, middlewares.OidcStateCookiePath, cookies[0].Path)
		assert.Equal(t, middlewares.OidcStateCookieMaxAge, cookies[0].MaxAge)
		assert.True(t, cookies[0].HttpOnly)
		assert.Equal(t, http.SameSiteLaxMode, cookies[0].SameSite)
		// ctx 未标记 HTTPS：不加 Secure，否则明文 HTTP 部署的浏览器会直接丢弃 Cookie。
		assert.False(t, cookies[0].Secure)
	})

	t.Run("HTTPS 请求加 Secure", func(t *testing.T) {
		rr := httptest.NewRecorder()
		ctx := middlewares.WithOidcCookieSecure(context.TODO(), true)
		err := gw.setOidcStateCookie(ctx, rr, &authpb.SettingsResponse{
			Items: []*authpb.SettingsResponse_OidcSetting{{Name: "a", State: "s"}},
		})
		assert.Nil(t, err)
		cookies := rr.Result().Cookies()
		assert.Len(t, cookies, 1)
		assert.True(t, cookies[0].Secure)
	})

	t.Run("无 provider 时不下发", func(t *testing.T) {
		rr := httptest.NewRecorder()
		assert.Nil(t, gw.setOidcStateCookie(context.TODO(), rr, &authpb.SettingsResponse{}))
		assert.Empty(t, rr.Result().Cookies())
	})
}

// Test_apiGateway_setOidcStateCookie_Exchange 覆盖 Exchange 响应清除 state Cookie：
// 换发成功即一次性消费，旧 state 不可重放（Max-Age=0 即删除）。
func Test_apiGateway_setOidcStateCookie_Exchange(t *testing.T) {
	gw := &apiGateway{}
	rr := httptest.NewRecorder()
	assert.Nil(t, gw.setOidcStateCookie(context.TODO(), rr, &authpb.ExchangeResponse{Token: "t"}))

	cookies := rr.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, middlewares.OidcStateCookieName, cookies[0].Name)
	assert.Equal(t, "", cookies[0].Value)
	assert.Equal(t, middlewares.OidcStateCookiePath, cookies[0].Path)
	assert.Equal(t, -1, cookies[0].MaxAge)
	assert.True(t, cookies[0].HttpOnly)
}

// Test_apiGateway_setOidcStateCookie_OtherResponse 覆盖非 auth 响应：不写任何 Cookie、
// 返回 nil（该回调挂在所有 REST 响应上，必须对无关消息类型无副作用）。
func Test_apiGateway_setOidcStateCookie_OtherResponse(t *testing.T) {
	gw := &apiGateway{}
	rr := httptest.NewRecorder()
	assert.Nil(t, gw.setOidcStateCookie(context.TODO(), rr, &emptypb.Empty{}))
	assert.Empty(t, rr.Result().Cookies())
}

// recordedProjectCall 记录一次经网关转发到 gRPC client 的调用。
type recordedProjectCall struct {
	method string
	req    proto.Message
}

// recordingProjectClient 是 projectpb.ProjectClient 的记录型替身，只实现本用例涉及的一元方法。
// 嵌入接口使未实现的方法保持 nil 指针——若请求被路由到预期之外的方法，调用会 panic，
// 把"走错路由"直接暴露成失败而不是静默通过。
type recordingProjectClient struct {
	projectpb.ProjectClient
	mu   sync.Mutex
	seen []recordedProjectCall
}

func (c *recordingProjectClient) add(method string, req proto.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seen = append(c.seen, recordedProjectCall{method: method, req: req})
}

// take 取走最近一次调用并清空记录（用例间互不串味）。
func (c *recordingProjectClient) take() recordedProjectCall {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.seen) == 0 {
		return recordedProjectCall{}
	}
	last := c.seen[len(c.seen)-1]
	c.seen = nil
	return last
}

func (c *recordingProjectClient) Show(_ context.Context, req *projectpb.ShowRequest, _ ...grpc.CallOption) (*projectpb.ShowResponse, error) {
	c.add("Show", req)
	return &projectpb.ShowResponse{}, nil
}

func (c *recordingProjectClient) ShowByName(_ context.Context, req *projectpb.ShowByNameRequest, _ ...grpc.CallOption) (*projectpb.ShowResponse, error) {
	c.add("ShowByName", req)
	return &projectpb.ShowResponse{}, nil
}

func (c *recordingProjectClient) Version(_ context.Context, req *projectpb.VersionRequest, _ ...grpc.CallOption) (*projectpb.VersionResponse, error) {
	c.add("Version", req)
	return &projectpb.VersionResponse{}, nil
}

func (c *recordingProjectClient) WebApplyByName(_ context.Context, req *projectpb.WebApplyByNameRequest, _ ...grpc.CallOption) (*projectpb.WebApplyResponse, error) {
	c.add("WebApplyByName", req)
	return &projectpb.WebApplyResponse{}, nil
}

func (c *recordingProjectClient) WebApply(_ context.Context, req *projectpb.WebApplyRequest, _ ...grpc.CallOption) (*projectpb.WebApplyResponse, error) {
	c.add("WebApply", req)
	return &projectpb.WebApplyResponse{}, nil
}

// newProjectRouteTestServer 用与生产完全一致的装配（initServer：UnescapingModeAllExceptSlash
// + JSONPb + /api 前缀 + 同一中间件链）注册 project 路由，返回可直接 ServeHTTP 的 handler。
// 断言对象是「HTTP 请求 → 哪个 RPC + 路径/查询参数怎么解码」——只看"注册没报错"证明不了路由形状。
func newProjectRouteTestServer(t *testing.T) (http.Handler, *recordingProjectClient) {
	t.Helper()
	m := gomock.NewController(t)
	t.Cleanup(m.Finish)
	handler := app.NewMockHttpHandler(m)
	handler.EXPECT().RegisterSwaggerUIRoute(gomock.Not(nil)).Times(1)
	handler.EXPECT().RegisterWsRoute(gomock.Not(nil)).Times(1)
	handler.EXPECT().RegisterFileRoute(gomock.Not(nil)).Times(1)

	cli := &recordingProjectClient{}
	reg := &app.GrpcRegistry{EndpointFuncs: []app.EndpointFunc{
		func(ctx context.Context, mux *runtime.ServeMux, _ string, _ []grpc.DialOption) error {
			return projectpb.RegisterProjectHandlerClient(ctx, mux, cli)
		},
	}}
	s, err := initServer(context.TODO(), &apiGateway{
		endpoint:     "x",
		port:         "1000",
		logger:       mlog.NewForConfig(nil),
		grpcRegistry: reg,
		handler:      handler,
	})
	assert.NoError(t, err)
	return s.(*http.Server).Handler, cli
}

// Test_initServer_ProjectRoutes_NewByNameEndpoints 覆盖新增的按名字寻址路由在真实网关上的
// 解析结果，并回归守护既有路由不被新路由吃掉——新路由 `by_name/{ns}/{name}` 是 project 服务下
// 唯一的五段模式，与既有 `{id}` / `{id}/version` 共存，路由优先级错一位就会静默错投。
func Test_initServer_ProjectRoutes_NewByNameEndpoints(t *testing.T) {
	h, cli := newProjectRouteTestServer(t)

	t.Run("GET by_name/{ns}/{name} → ShowByName 且两段变量各自解码", func(t *testing.T) {
		h.ServeHTTP(httptest.NewRecorder(),
			httptest.NewRequest(http.MethodGet, "/api/projects/by_name/mars-demo/web", nil))

		got := cli.take()
		if !assert.Equal(t, "ShowByName", got.method) {
			return
		}
		req := got.req.(*projectpb.ShowByNameRequest)
		assert.Equal(t, "mars-demo", req.GetNamespace())
		assert.Equal(t, "web", req.GetName())
	})

	t.Run("GET /{id} 仍路由到 Show，未被 by_name 抢占", func(t *testing.T) {
		h.ServeHTTP(httptest.NewRecorder(),
			httptest.NewRequest(http.MethodGet, "/api/projects/123", nil))

		got := cli.take()
		if !assert.Equal(t, "Show", got.method) {
			return
		}
		assert.Equal(t, int32(123), got.req.(*projectpb.ShowRequest).GetId())
	})

	t.Run("GET /{id}/version 三段字面量路由不受影响", func(t *testing.T) {
		h.ServeHTTP(httptest.NewRecorder(),
			httptest.NewRequest(http.MethodGet, "/api/projects/123/version", nil))

		got := cli.take()
		if !assert.Equal(t, "Version", got.method) {
			return
		}
		assert.Equal(t, int32(123), got.req.(*projectpb.VersionRequest).GetId())
	})

	t.Run("POST apply_by_name 走 WebApplyByName，body 字段被解码", func(t *testing.T) {
		body := `{"namespace":"mars-demo","name":"web","gitBranch":"dev"}`
		req := httptest.NewRequest(http.MethodPost, "/api/projects/apply_by_name", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		h.ServeHTTP(httptest.NewRecorder(), req)

		got := cli.take()
		if !assert.Equal(t, "WebApplyByName", got.method) {
			return
		}
		byName := got.req.(*projectpb.WebApplyByNameRequest)
		assert.Equal(t, "mars-demo", byName.GetNamespace())
		assert.Equal(t, "web", byName.GetName())
		assert.Equal(t, "dev", byName.GetGitBranch())
	})

	t.Run("POST apply 仍走 WebApply，未被 apply_by_name 抢占", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/projects/apply", strings.NewReader(`{"namespaceId":1,"repoId":2}`))
		req.Header.Set("Content-Type", "application/json")
		h.ServeHTTP(httptest.NewRecorder(), req)

		got := cli.take()
		if !assert.Equal(t, "WebApply", got.method) {
			return
		}
		assert.Equal(t, int32(1), got.req.(*projectpb.WebApplyRequest).GetNamespaceId())
		assert.Equal(t, int32(2), got.req.(*projectpb.WebApplyRequest).GetRepoId())
	})

	// 路径解码与段数边界：生产网关用 UnescapingModeAllExceptSlash，实测行为一并钉住——
	// %20 正常解码为空格，%2F（编码斜杠）留在参数内不切段（否则会变成一次路由越权）。
	t.Run("路径变量的转义解码：%20 解码、%2F 不切段", func(t *testing.T) {
		h.ServeHTTP(httptest.NewRecorder(),
			httptest.NewRequest(http.MethodGet, "/api/projects/by_name/mars%20demo/my%20proj", nil))
		got := cli.take()
		if !assert.Equal(t, "ShowByName", got.method) {
			return
		}
		req := got.req.(*projectpb.ShowByNameRequest)
		assert.Equal(t, "mars demo", req.GetNamespace())
		assert.Equal(t, "my proj", req.GetName())

		h.ServeHTTP(httptest.NewRecorder(),
			httptest.NewRequest(http.MethodGet, "/api/projects/by_name/a%2Fb/c", nil))
		got = cli.take()
		if !assert.Equal(t, "ShowByName", got.method, "编码斜杠不得切出新路径段") {
			return
		}
		assert.Equal(t, "a/b", got.req.(*projectpb.ShowByNameRequest).GetNamespace())
	})

	// 段数不匹配必须落 404，不能被 `/api/projects/{id}` 兜成"id 解析失败"或错投其他方法。
	t.Run("段数不匹配落 404，不越界匹配", func(t *testing.T) {
		for _, p := range []string{
			"/api/projects/by_name/mars-demo",   // 少了 name
			"/api/projects/by_name/ns/n/extra",  // 多了段
			"/api/projects/apply_by_name/extra", // apply_by_name 不接受子路径
		} {
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, p, nil))
			assert.Equal(t, http.StatusNotFound, rr.Code, p)
			assert.Empty(t, cli.take().method, p)
		}
	})
}
