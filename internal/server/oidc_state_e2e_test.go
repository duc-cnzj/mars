package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apiauth "github.com/duc-cnzj/mars/api/v6/proto/auth"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/server/middlewares"
	"github.com/duc-cnzj/mars/v6/internal/services"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

// oidcHarness 汇总端到端测试所需的三件套：装好的网关卡，以及服务层三个 mock。
type oidcHarness struct {
	h         http.Handler
	authBiz   *biz.MockAuthBiz
	userBiz   *biz.MockUserBiz
	eventRepo *data.MockEventRepo
}

// newOidcHarness 用真实 initServer 装配网关 + 真实 services.NewAuthSvc（只 mock 其下层的 biz）。
//
// overWire 决定 grpc 服务端注册方式，两条路径都必须验证，因为 metadata 的落点不同：
//   - false：RegisterAuthHandlerServer 进程内直调，走 AnnotateIncomingContext（不上网络）；
//   - true ：RegisterAuthHandlerFromEndpoint 真实拨号，走 AnnotateContext→NewOutgoingContext。
//
// 只有后者与生产装配一致，能暴露「进程内测试全绿、真拨号时 Cookie 没上网络」这类盲区。
func newOidcHarness(t *testing.T, overWire bool) *oidcHarness {
	t.Helper()
	m := gomock.NewController(t)
	authBiz := biz.NewMockAuthBiz(m)
	userBiz := biz.NewMockUserBiz(m)
	eventRepo := data.NewMockEventRepo(m)

	authSvc := services.NewAuthSvc(services.AuthSvcDeps{
		EventBiz: biz.NewEventBiz(eventRepo),
		Logger:   mlog.NewForConfig(nil),
		AuthBiz:  authBiz,
		UserBiz:  userBiz,
	})

	endpoint := "x"
	register := func(ctx context.Context, mux *runtime.ServeMux, _ string, _ []grpc.DialOption) error {
		return apiauth.RegisterAuthHandlerServer(ctx, mux, authSvc)
	}
	if overWire {
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		// 带上 Validator 拦截器：wire 用例要顺带证明新增的 state 字段必填规则真的生效
		// （进程内直调变体会绕过拦截器，故该校验只能在 wire 用例里验证）。
		gs := grpc.NewServer(grpc.UnaryInterceptor(middlewares.ValidatorUnaryServerInterceptor()))
		apiauth.RegisterAuthServer(gs, authSvc)
		go func() { _ = gs.Serve(lis) }()
		t.Cleanup(gs.Stop)
		endpoint = lis.Addr().String()
		register = func(ctx context.Context, mux *runtime.ServeMux, addr string, opts []grpc.DialOption) error {
			return apiauth.RegisterAuthHandlerFromEndpoint(ctx, mux, addr, opts)
		}
	}

	handler := app.NewMockHttpHandler(m)
	handler.EXPECT().RegisterSwaggerUIRoute(gomock.Not(nil)).Times(1)
	handler.EXPECT().RegisterWsRoute(gomock.Not(nil)).Times(1)
	handler.EXPECT().RegisterFileRoute(gomock.Not(nil)).Times(1)

	httpServer, err := initServer(context.TODO(), &apiGateway{
		endpoint: endpoint,
		port:     "1000",
		logger:   mlog.NewForConfig(nil),
		grpcRegistry: &app.GrpcRegistry{
			EndpointFuncs: []app.EndpointFunc{register},
		},
		handler: handler,
	})
	require.NoError(t, err)
	return &oidcHarness{
		h:         httpServer.(*http.Server).Handler,
		authBiz:   authBiz,
		userBiz:   userBiz,
		eventRepo: eventRepo,
	}
}

// oidcTestAuthConfig 是端到端测试用的 OIDC provider 配置：AuthCodeURL 需要一个可用的授权端点。
func oidcTestAuthConfig() biz.OidcConfig {
	return biz.OidcConfig{
		"test": {
			Config: oauth2.Config{
				ClientID:    "cid",
				RedirectURL: "https://app.example/api/auth/exchange",
				Endpoint:    oauth2.Endpoint{AuthURL: "https://idp.example/auth"},
				Scopes:      []string{"openid"},
			},
		},
	}
}

// postExchange 发一起 exchange 请求；cookie 为空表示不带 Cookie 头。
func postExchange(h http.Handler, cookie, body string, tlsOn bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", "/api/auth/exchange", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	if tlsOn {
		req.TLS = &tls.ConnectionState{}
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

// assertSettingsIssuesCookie 断言 settings 下发的 state Cookie 属性齐全，返回 state 值。
// 同时覆盖 Secure 推导：明文 HTTP 不加、TLS 握手加。
func assertSettingsIssuesCookie(t *testing.T, h http.Handler) string {
	t.Helper()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/api/auth/settings", nil))
	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())

	cookies := rr.Result().Cookies()
	require.Len(t, cookies, 1, "settings 必须下发 state Cookie，Set-Cookie=%v", rr.Header().Values("Set-Cookie"))
	assert.NotEmpty(t, cookies[0].Value)
	assert.Equal(t, middlewares.OidcStateCookieName, cookies[0].Name)
	assert.Equal(t, middlewares.OidcStateCookiePath, cookies[0].Path)
	assert.Equal(t, middlewares.OidcStateCookieMaxAge, cookies[0].MaxAge)
	assert.True(t, cookies[0].HttpOnly, "state 是凭据，不得被 JS 读取")
	assert.Equal(t, http.SameSiteLaxMode, cookies[0].SameSite)
	assert.False(t, cookies[0].Secure, "明文 HTTP 部署不得加 Secure，否则浏览器丢弃 Cookie 致 SSO 整体失败")

	// HTTPS 直连（TLS 握手）经 OidcCookieSecureMiddleware → ctx → ForwardResponseOption 加 Secure。
	tlsRR := httptest.NewRecorder()
	tlsReq := httptest.NewRequest("GET", "/api/auth/settings", nil)
	tlsReq.TLS = &tls.ConnectionState{}
	h.ServeHTTP(tlsRR, tlsReq)
	require.Equal(t, http.StatusOK, tlsRR.Code, tlsRR.Body.String())
	tlsCookies := tlsRR.Result().Cookies()
	require.Len(t, tlsCookies, 1)
	assert.True(t, tlsCookies[0].Secure, "HTTPS 请求下发的 state Cookie 必须带 Secure")

	return cookies[0].Value
}

// assertExchangeRejects 断言两类攻击形态都被拒，且**不得触达换发逻辑**（调用方不设
// authBiz.Exchange 期望，门卫一旦漏过，gomock 会以 Unexpected call 直接判败）。
func assertExchangeRejects(t *testing.T, h http.Handler, state string) {
	t.Helper()
	body := fmt.Sprintf(`{"code":"c","state":%q}`, state)

	assert.Equal(t, http.StatusBadRequest, postExchange(h, "", body, false).Code,
		"无 state Cookie 必须拒绝（跨站表单 POST 不带任何 Cookie 的形态）")
	assert.Equal(t, http.StatusBadRequest,
		postExchange(h, middlewares.OidcStateCookieName+"=attacker-own-state", body, false).Code,
		"Cookie 存在但值不匹配必须拒绝（攻击者拿自己申请的合法 state 塞给受害者的形态）")
}

// assertExchangeAccepts 断言 Cookie 与回传 state 一致时放行，并清除一次性 Cookie。
// 调用方须先设好 authBiz.Exchange/Sign、userBiz.SyncLoginUser、eventRepo 的期望。
func assertExchangeAccepts(t *testing.T, h http.Handler, state string) {
	t.Helper()
	rr := postExchange(h, middlewares.OidcStateCookieName+"="+state,
		fmt.Sprintf(`{"code":"c","state":%q}`, state), false)
	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	assert.Contains(t, rr.Body.String(), "signed")

	cleared := rr.Result().Cookies()
	require.Len(t, cleared, 1)
	assert.Equal(t, -1, cleared[0].MaxAge, "换发成功即一次性消费，必须清除 state Cookie")
}

// expectExchangeSuccess 设好换发成功的全套 mock 期望（服务层从 Exchange 到落库的四步）。
func expectExchangeSuccess(h *oidcHarness) {
	userinfo := &biz.UserInfo{Name: "duc", Email: "duc@example.com", Roles: []string{biz.MarsAdmin}}
	h.authBiz.EXPECT().Exchange(gomock.Any(), "c").Return(userinfo, nil)
	h.authBiz.EXPECT().Sign(gomock.Any(), userinfo).Return(&biz.LoginResponse{Token: "signed", ExpiredIn: 3600}, nil)
	h.userBiz.EXPECT().SyncLoginUser(gomock.Any(), "duc@example.com", "duc", []string{biz.MarsAdmin}).Return(nil)
	h.eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Login, "duc", "duc@example.com",
		"用户 'duc' email: 'duc@example.com' 登录了系统", gomock.Any(),
	)
}

// Test_initServer_OidcStateCookie_InProcess 验证进程内直调装配下的完整防线：
// HTTP Cookie 头 → metadata → 门卫比对 → 拒绝/放行。
func Test_initServer_OidcStateCookie_InProcess(t *testing.T) {
	h := newOidcHarness(t, false)
	h.authBiz.EXPECT().Settings(gomock.Any()).Return(oidcTestAuthConfig(), nil).Times(2)

	state := assertSettingsIssuesCookie(t, h.h)
	assertExchangeRejects(t, h.h, state)
	expectExchangeSuccess(h)
	assertExchangeAccepts(t, h.h, state)
}

// Test_initServer_OidcStateCookie_OverGRPCWire 与上一条完全同形，唯一区别是 gateway 真拨号到
// 一个真实 gRPC server（与生产装配一致）。这是本改动最关键的一条证据：state Cookie 必须经
// grpc-gateway 的 OutgoingContext 真的上网络、被服务端 metadata 收到——若只验证进程内直调，
// 「Cookie 没上网络」这类故障会全绿通过。
func Test_initServer_OidcStateCookie_OverGRPCWire(t *testing.T) {
	h := newOidcHarness(t, true)
	h.authBiz.EXPECT().Settings(gomock.Any()).Return(oidcTestAuthConfig(), nil).Times(2)

	state := assertSettingsIssuesCookie(t, h.h)
	assertExchangeRejects(t, h.h, state)

	// 缺 state 字段：新增的必填规则由 Validator 拦截器挡下（400），不得进业务逻辑。
	missingField := postExchange(h.h, middlewares.OidcStateCookieName+"="+state, `{"code":"c"}`, false)
	assert.Equal(t, http.StatusBadRequest, missingField.Code, "state 缺失必须被必填规则拒绝")

	expectExchangeSuccess(h)
	assertExchangeAccepts(t, h.h, state)
}
