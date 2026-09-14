package services

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"sort"

	apiauth "github.com/duc-cnzj/mars/api/v6/proto/auth"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/server/middlewares"
	"github.com/duc-cnzj/mars/v6/internal/util/rand"
	"github.com/spf13/cast"
	"google.golang.org/grpc/metadata"
)

// oidcCookieMetadataKey 是 grpc-gateway 把 HTTP Cookie 头映射进 gRPC metadata 后的键名：
// DefaultHeaderMatcher 会给 isPermanentHTTPHeader 白名单内的头（Cookie 在列）加上
// "grpcgateway-" 前缀；键名小写由 metadata 包自身保证（MD.Get 先做 ToLower）。
const oidcCookieMetadataKey = "grpcgateway-cookie"

var _ apiauth.AuthServer = (*authSvc)(nil)

// authSvc 是 apiauth.AuthServer 的 gRPC 实现：提供登录、用户信息、登录设置
// 与 OIDC 授权码换取，审计事件经 eventBiz 落库，登录成功后经 userBiz 同步用户投影，
// 由 NewAuthSvc 构造。
type authSvc struct {
	apiauth.UnimplementedAuthServer

	logger   mlog.Logger
	authBiz  biz.AuthBiz
	eventBiz biz.EventBiz
	userBiz  biz.UserBiz
}

// AuthSvcDeps 收口 NewAuthSvc 的构造依赖，由 wire 按字段注入。
type AuthSvcDeps struct {
	EventBiz biz.EventBiz
	Logger   mlog.Logger
	AuthBiz  biz.AuthBiz
	UserBiz  biz.UserBiz
}

// NewAuthSvc 收口认证服务的构造依赖，由 wire 按字段注入。
func NewAuthSvc(deps AuthSvcDeps) apiauth.AuthServer {
	return &authSvc{
		logger:   deps.Logger.WithModule("services/auth"),
		eventBiz: deps.EventBiz,
		authBiz:  deps.AuthBiz,
		userBiz:  deps.UserBiz,
	}
}

// Login 处理用户名密码登录：校验凭证后签发登录凭证，并落登录审计日志。
func (a *authSvc) Login(ctx context.Context, request *apiauth.LoginRequest) (*apiauth.LoginResponse, error) {
	loginResp, err := a.authBiz.Login(ctx, &biz.LoginInput{
		Username: request.Username,
		Password: request.Password,
	})
	if err != nil {
		return nil, logError(ctx, a.logger, err)
	}

	a.eventBiz.AuditLog(
		types.EventActionType_Login,
		loginResp.UserInfo.Name,
		loginResp.UserInfo.Email,
		fmt.Sprintf("用户 '%s' email: '%s' 登录了系统", loginResp.UserInfo.Name, loginResp.UserInfo.Email),
	)

	// 登录成功即同步用户投影（与 OIDC Exchange 行为对齐）：admin 账号同样落 users 表，
	// 不存在则创建、存在则推进最近登录；登录身份携带的角色（内置超管 mars_admin）随
	// 投影写入/同步。投影写库失败仅记日志不阻断登录——凭证已校验、登录事件已落库，
	// users 只是管理投影，该用户下次登录会由 SyncLoginUser 自动补回。
	if err := a.userBiz.SyncLoginUser(ctx, loginResp.UserInfo.Email, loginResp.UserInfo.Name, loginResp.UserInfo.Roles); err != nil {
		a.logger.ErrorCtx(ctx, err)
	}

	return &apiauth.LoginResponse{
		Token:     loginResp.Token,
		ExpiresIn: loginResp.ExpiredIn,
	}, nil
}

// Info 返回当前登录用户信息：用户由鉴权拦截器（middlewares.Login*ServerInterceptor）
// 统一验签后经 biz.SetUser 注入 ctx，本方法只做「取 ctx 用户 → 映射响应」，不再自行验签
// （消除与拦截器的双重验签）。取用户用 biz.MustGetUser（与 AccessBiz/services 全仓惯例
// 一致）：双链路（gRPC 拦截器 + HTTP gateway 经 RegisterAuthHandlerFromEndpoint 回环 dial
// 到同一 gRPC server）必注入用户，ctx 无用户即编程错误（panic，由 grpc_recovery 兜底）。
//
// IsGray 是唯一需要额外读库的字段：灰度标记不在 JWT 里（后台可随时改，签进 token 会变成
// 登录时刻的死快照）。前端的底栏灰度标记与「启动对齐灰度路由 cookie」都以本字段为权威口径，
// 所以每次 /api/auth/info 都要取当前库值——这正是「后台一改开关，用户刷新页面即生效」的落点。
// 读库失败降级为「非灰度」（fail-open 到稳定版）并记日志：灰度只是发布通道，不该因一次
// 查询失败把整个登录态恢复打断（用户会看到「打不开页面」，而问题其实只是发布通道未知）。
func (a *authSvc) Info(ctx context.Context, req *apiauth.InfoRequest) (*apiauth.InfoResponse, error) {
	user := biz.MustGetUser(ctx)
	isGray, err := a.userBiz.IsGray(ctx, user.Email)
	if err != nil {
		a.logger.ErrorCtx(ctx, err)
	}
	return &apiauth.InfoResponse{
		Id:           cast.ToInt32(user.ID),
		Avatar:       user.Picture,
		Name:         user.Name,
		Email:        user.Email,
		LogoutUrl:    user.LogoutUrl,
		Roles:        user.Roles,
		IsSuperAdmin: user.IsSuperAdmin(),
		IsGray:       isGray,
	}, nil
}

// Settings 返回可用的 OIDC 登录方式：拼出带一次性 state 的授权码 URL，按名字排序后返回，
// 供前端渲染登录页。
//
// 所有 provider 共用同一个 state：state 绑定的是「这一次浏览器发起的登录尝试」而非某个
// provider，且 state Cookie 只有一个槽位，per-provider 各发一个会让回调时无从比对。
// state 由 HTTP 层随本响应写入 Cookie（见 server.(*apiGateway).setOidcStateCookie），
// Exchange 侧再拿回调回传的 state 与该 Cookie 比对。
func (a *authSvc) Settings(ctx context.Context, request *apiauth.SettingsRequest) (*apiauth.SettingsResponse, error) {
	settings, err := a.authBiz.Settings(ctx)
	if err != nil {
		return nil, logError(ctx, a.logger, err)
	}
	state := rand.String(32)

	var items = make([]*apiauth.SettingsResponse_OidcSetting, 0, len(settings))
	for name, setting := range settings {
		items = append(items, &apiauth.SettingsResponse_OidcSetting{
			Enabled:            true,
			Name:               name,
			Url:                setting.Config.AuthCodeURL(state),
			EndSessionEndpoint: setting.EndSessionEndpoint,
			State:              state,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return &apiauth.SettingsResponse{Items: items}, nil
}

// Exchange 用 OIDC 授权码换发登录凭证：换发编排（遍历 provider/验签/claims 解码）
// 已下沉 biz.AuthBiz.Exchange，这里只做 transport 份内事——防 CSRF 校验、签名、审计与响应映射。
func (a *authSvc) Exchange(ctx context.Context, request *apiauth.ExchangeRequest) (*apiauth.ExchangeResponse, error) {
	// 先过 CSRF 门卫，再做任何换发动作：未经校验的 code 不该触达 IdP（既是安全边界，
	// 也避免攻击者拿本接口当探测 IdP 是否可达的探针）。
	if err := verifyOidcState(ctx, request.State); err != nil {
		return nil, logError(ctx, a.logger, err)
	}

	userinfo, err := a.authBiz.Exchange(ctx, request.Code)
	if err != nil {
		return nil, logError(ctx, a.logger, err)
	}

	data, err := a.authBiz.Sign(ctx, userinfo)
	if err != nil {
		// 与其他 service 一致：直接返回原始错误，不额外包装 codes.Internal（避免丢失原始错误码）。
		return nil, logError(ctx, a.logger, err)
	}
	a.eventBiz.AuditLogWithRequest(
		types.EventActionType_Login,
		userinfo.Name,
		userinfo.Email,
		fmt.Sprintf("用户 '%s' email: '%s' 登录了系统", userinfo.Name, userinfo.Email),
		request,
	)
	// 登录成功即同步用户投影：不存在则创建、存在则推进最近登录；SSO id_token 携带的
	// 角色（mars_admin）随投影写入/同步，使用户管理页能反映 SSO 身份。投影写库失败仅记
	// 日志不阻断登录——OIDC 凭证已校验、登录事件已落库，users 只是管理投影，该用户下次
	// 登录会由 SyncLoginUser 自动补回。
	if err := a.userBiz.SyncLoginUser(ctx, userinfo.Email, userinfo.Name, userinfo.Roles); err != nil {
		a.logger.ErrorCtx(ctx, err)
	}

	return &apiauth.ExchangeResponse{
		Token:     data.Token,
		ExpiresIn: data.ExpiredIn,
	}, nil
}

// verifyOidcState 校验 OIDC 回调回传的 state 与当初下发到浏览器的 state Cookie 是否一致，
// 是防「登录 CSRF」的唯一有效手段。
//
// 为什么不是「存起来」或「签名」：攻击者可以自己正常走一遍 /api/auth/settings，拿到一份
// 完全合法的 state（无论它被服务端存了还是签了名），再自己完成 IdP 登录拿到 code，然后把
// 这对 (state, code) 塞给受害者浏览器——那两种方案都会照单放行。state 的职责是绑定「发起
// 登录的那个浏览器」，而浏览器身份唯一不可伪造的载体是 Cookie（攻击者写不了本域 Cookie），
// 所以必须拿回传值和 Cookie 比对（契约见 middlewares.OidcStateCookieName）。
//
// 用 InvalidArgument 而非 Unauthenticated：与 biz.AuthBiz.Exchange 全部失败时的错误码保持一致，
// 同时避开前端对 401 的「清 token + 跳登出」副作用（那条路径留给真正的会话过期）。
// 对外不区分「state 不匹配」与「code 换发失败」，避免给攻击者留下探测 oracle。
func verifyOidcState(ctx context.Context, state string) error {
	cookieValue, ok := oidcStateFromMetadata(ctx)
	if !ok {
		return errs.InvalidArgument("OIDC 登录校验失败")
	}
	// 常数时间比对：state 是可直接换取登录凭证的凭据，随机串逐字节比较会泄露前缀信息。
	if subtle.ConstantTimeCompare([]byte(cookieValue), []byte(state)) != 1 {
		return errs.InvalidArgument("OIDC 登录校验失败")
	}
	return nil
}

// oidcStateFromMetadata 从 gRPC metadata 取回 HTTP 请求的 Cookie 头并解析出 state 值。
// 原生 gRPC 调用不带该 metadata（Cookie 是纯 HTTP 概念），取不到即返回 false，
// 由调用方按校验失败处理。
func oidcStateFromMetadata(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
	raw := md.Get(oidcCookieMetadataKey)
	if len(raw) == 0 {
		return "", false
	}
	cookies, err := http.ParseCookie(raw[0])
	if err != nil {
		return "", false
	}
	for _, c := range cookies {
		if c.Name == middlewares.OidcStateCookieName {
			return c.Value, true
		}
	}
	return "", false
}
