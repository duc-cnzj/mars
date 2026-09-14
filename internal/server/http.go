package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	apiauth "github.com/duc-cnzj/mars/api/v6/proto/auth"
	containerpb "github.com/duc-cnzj/mars/api/v6/proto/container"
	metricspb "github.com/duc-cnzj/mars/api/v6/proto/metrics"
	"github.com/duc-cnzj/mars/v6/frontend"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/server/middlewares"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/stats"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const maxRecvMsgSize = 1 << 20 * 100 // 100 MiB

var defaultMiddlewares = middlewareList{
	middlewares.Recovery,
	middlewares.RouteLogger,
	middlewares.AllowCORS,
}

type apiGateway struct {
	endpoint      string
	port          string
	server        HttpServer
	logger        mlog.Logger
	grpcRegistry  *app.GrpcRegistry
	handler       app.HttpHandler
	newServerFunc func(ctx context.Context, a *apiGateway) (HttpServer, error)
}

// NewApiGateway 构建 HTTP api-gateway 启动器：grpc-gateway 承载 API、gorilla/mux 承载
// 前端与 websocket 路由，端口取自 app.Config().AppPort。返回实现 app.Server。
func NewApiGateway(endpoint string, app app.ServerDeps) app.Server {
	return &apiGateway{
		endpoint:      endpoint,
		port:          app.Config().AppPort,
		logger:        app.Logger().WithModule("server/apiGateway"),
		grpcRegistry:  app.GrpcRegistry(),
		handler:       app.HttpHandler(),
		newServerFunc: initServer,
	}
}

// Run 启动 HTTP 网关：先装配服务（initServer），再启动集群健康检查协程与
// ListenAndServe 协程（后者遇非 ErrServerClosed 错误时打 Error 日志，正常关闭不报错）。
func (a *apiGateway) Run(ctx context.Context) error {
	s, err := a.newServerFunc(ctx, a)
	if err != nil {
		return err
	}

	a.server = s

	go a.handler.TickClusterHealth(ctx.Done())

	go func(s HttpServer) {
		a.logger.Infof("[Server]: start apiGateway runner at :%s.", a.port)
		if err = s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error(err)
		}
	}(s)

	return nil
}

// Shutdown 优雅停止 HTTP 网关：先停业务 handler（含 websocket 面），再停底层 http server。
func (a *apiGateway) Shutdown(ctx context.Context) error {
	a.logger.Info("[Server]: shutdown api-gateway runner.")
	a.handler.Shutdown(ctx)
	return a.server.Shutdown(ctx)
}

// filterMethods 是跳过指标追踪的 gRPC 流方法白名单：这些长连接方法不记录 metadata
// （错误之类），但保有一条访问日志，避免 Trace/指标被高频流冲刷。
// 未导出：仅本包 shouldTagRPC 消费；直接引用 proto 生成的 *_FullMethodName 常量，
// 方法名写错即编译失败（与 biz.publicMethods 同理，不手写 "/pkg.Svc/Method" 路径）。
var filterMethods = map[string]struct{}{
	metricspb.Metrics_StreamTopPod_FullMethodName:           {},
	containerpb.Container_StreamContainerLog_FullMethodName: {},
}

// shouldTagRPC 决定 gRPC 调用是否计入 OpenTelemetry 统计：命中 filterMethods 白名单的
// 长连接流方法不计入（避免高频流冲刷 metrics/trace），其余返回 true。false 分支打一条
// Debugf 便于观测该调用为何被跳过。
func (a *apiGateway) shouldTagRPC(info *stats.RPCTagInfo) bool {
	_, ok := filterMethods[info.FullMethodName]
	a.logger.Debugf("%v\t%v", info.FullMethodName, !ok)
	return !ok
}

// setNosniff 是 grpc-gateway 的 ForwardResponseOption：每个 REST 响应补上
// X-Content-Type-Options: nosniff，禁止浏览器嗅探响应类型（防内容类型混淆）。
func (a *apiGateway) setNosniff(ctx context.Context, writer http.ResponseWriter, message proto.Message) error {
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	return nil
}

// setOidcStateCookie 是 grpc-gateway 的 ForwardResponseOption：把 OIDC 登录 state 经
// Set-Cookie 下发（Settings）与清除（Exchange）。
//
// 之所以要由 HTTP 层来写：state 必须绑定到「发起登录的那个浏览器」，而浏览器身份唯一
// 不可伪造的载体就是 Cookie——攻击者能自己申请到一份合法的 state+code，服务端存储 state
// 或签名 state 都拦不住，只有「攻击者写不了受害者浏览器上的本域 Cookie」这一条能拦住
// （完整契约与理由见 middlewares.OidcStateCookieName 的注释）。
func (a *apiGateway) setOidcStateCookie(ctx context.Context, writer http.ResponseWriter, message proto.Message) error {
	// Secure 判定由 OidcCookieSecureMiddleware 经 ctx 传入：本回调拿不到 *http.Request。
	secure := middlewares.IsOidcCookieSecure(ctx)

	switch resp := message.(type) {
	case *apiauth.SettingsResponse:
		// 所有 provider 共用同一个 state（state 绑定的是「这次登录尝试」而非某个 provider，
		// 且 Cookie 只有一个槽位），故取任一项即可；未配置 provider 时无需下发。
		if len(resp.Items) == 0 {
			return nil
		}
		http.SetCookie(writer, &http.Cookie{
			Name:     middlewares.OidcStateCookieName,
			Value:    resp.Items[0].State,
			Path:     middlewares.OidcStateCookiePath,
			MaxAge:   middlewares.OidcStateCookieMaxAge,
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		})
	case *apiauth.ExchangeResponse:
		// 一次性消费：换发成功即清 Cookie，旧 state 不可重放。
		http.SetCookie(writer, &http.Cookie{
			Name:     middlewares.OidcStateCookieName,
			Value:    "",
			Path:     middlewares.OidcStateCookiePath,
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		})
	}

	return nil
}

// initServer 装配 HTTP 网关：构建 grpc-gateway ServeMux（headers/forward/JSON 编解码）、
// grpc 拨号选项（OpenTelemetry 过滤、最大接收消息）、注册 API 路由/文件/ws/swagger/前端路由，
// 最终用中间件链 + otelhttp 包裹返回可启动的 http.Server。
// 路由注册顺序是硬约束：/api（gmux）、/ws、swagger 必须先于前端 SPA 兜底（/{any:.*}），
// 否则会被兜底吞成 index.html；前端 /resources 静态文件与 /{any:.*} 兜底最后注册。
func initServer(ctx context.Context, a *apiGateway) (HttpServer, error) {
	router := mux.NewRouter()

	gmux := runtime.NewServeMux(
		runtime.WithUnescapingMode(runtime.UnescapingModeAllExceptSlash),
		runtime.WithOutgoingHeaderMatcher(headerMatcher),
		runtime.WithIncomingHeaderMatcher(headerMatcher),
		runtime.WithForwardResponseOption(a.setNosniff),
		runtime.WithForwardResponseOption(a.setOidcStateCookie),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseEnumNumbers:  false,
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	opts := []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(
			otelgrpc.WithFilter(a.shouldTagRPC),
		)),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxRecvMsgSize)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	for _, f := range a.grpcRegistry.EndpointFuncs {
		if err := f(ctx, gmux, a.endpoint, opts); err != nil {
			return nil, err
		}
	}

	router.HandleFunc("/ping", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.Write([]byte("pong"))
	})

	a.handler.RegisterFileRoute(gmux)
	a.handler.RegisterWsRoute(router)
	// /api 前缀统一交给 grpc-gateway：必须先于前端 SPA 兜底注册，否则 /api/xxx
	// 会被 LoadFrontendRoutes 的 /{any:.*} 兜底吞成 index.html。全仓 HTTP API
	// 均约定挂在 /api 下（proto http 注解 + 文件路由，见 fileHandler.RegisterFileRoute）。
	// 外层多包一层：把「本次请求是否经 HTTPS」经 ctx 带进 gmux，供 state Cookie 决定 Secure。
	router.PathPrefix("/api/").Handler(middlewares.OidcCookieSecureMiddleware(gmux))
	// swagger 文档路由先于前端兜底注册，避免 /docs/ 与 /doc/swagger.json 被 SPA 兜底拦截。
	a.handler.RegisterSwaggerUIRoute(router)
	frontend.LoadFrontendRoutes(router)

	s := &http.Server{
		Addr: ":" + a.port,
		Handler: defaultMiddlewares.Wrap(
			a.logger,
			otelhttp.NewHandler(
				router,
				"grpc-gateway",
				otelhttp.WithFilter(func(request *http.Request) bool {
					return strings.HasPrefix(request.URL.Path, "/api")
				}),
				otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
					return fmt.Sprintf("grpc-gateway [%s] %s", r.Method, r.URL.Path)
				}),
			),
		),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return s, nil
}

type middlewareList []func(logger mlog.Logger, handler http.Handler) http.Handler

// Wrap 把中间件列表按逆序套在外层（最后一个先执行）：空列表直接返回原 handler。
func (m middlewareList) Wrap(logger mlog.Logger, r http.Handler) (h http.Handler) {
	if len(m) == 0 {
		return r
	}
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](logger, r)
		r = h
	}
	return
}

// headerMatcher 决定 gRPC-gateway 出入站 header 映射：trace 相关 header（traceparent/
// tracestate）白名单放行，其余回落到 runtime.DefaultHeaderMatcher 默认行为。
func headerMatcher(key string) (string, bool) {
	key = strings.ToLower(key)
	switch key {
	case "tracestate":
		fallthrough
	case "traceparent":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}
