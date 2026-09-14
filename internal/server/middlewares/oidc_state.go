package middlewares

import (
	"context"
	"net/http"
	"strings"
)

// OIDC 登录 state 的 Cookie 契约（单一事实来源，HTTP 层与 gRPC 服务层共用）。
//
// state 是防「登录 CSRF」的一次性随机串：/api/auth/settings 下发时写入本 Cookie 并拼进
// IdP 授权 URL，/api/auth/exchange 回调时比对两者是否一致。
//
// 为什么必须落在 Cookie 上：state 的职责是绑定「发起登录的那个浏览器」。只用服务端存储
// 或签名 state 都不够——攻击者可以自己正常走一遍 settings 申请到一份合法的 (state, code)，
// 再把它塞进受害者浏览器，那两种方案都会照单放行；只有「服务端下发、攻击者无法伪造」的
// Cookie 才能挡住这种「攻击者自发起」的登录 CSRF（同源策略保证攻击者写不了本域 Cookie）。
const (
	// OidcStateCookieName 是 state Cookie 名。
	OidcStateCookieName = "mars_oidc_state"

	// OidcStateCookiePath 限定 Cookie 只在 auth 接口上收发，缩小暴露面。
	// 下发（settings）与校验（exchange）都落在该路径下，故清理 Cookie 必须用同一值，
	// 否则浏览器按路径匹配不到、清不掉。
	OidcStateCookiePath = "/api/auth"

	// OidcStateCookieMaxAge 是 state Cookie 有效期（秒）：state 只在「点 SSO 跳 IdP →
	// 回调换 token」这段窗口内有意义，5 分钟足够覆盖人工在 IdP 上输入账号密码的耗时。
	OidcStateCookieMaxAge = 300
)

// oidcSecureKey 是「当前请求是否经 HTTPS」的 ctx 键。未导出：读写都经下面两个函数，
// 避免键类型散落（对齐 auth 包用值类型作 ctx 键的既有做法）。
type oidcSecureKey struct{}

// oidcCookieSecure 判定请求是否经 HTTPS：直连看 TLS 握手，反代终止 TLS 时看
// X-Forwarded-Proto。判错的代价不对称——把明文 HTTP 误判成 HTTPS 会让 Cookie 被浏览器
// 直接丢弃、SSO 登录整体失败，故拿不准时一律按 false（不加 Secure 属性）处理。
func oidcCookieSecure(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

// WithOidcCookieSecure 把 HTTPS 判定结果写入 ctx。
func WithOidcCookieSecure(ctx context.Context, secure bool) context.Context {
	return context.WithValue(ctx, oidcSecureKey{}, secure)
}

// IsOidcCookieSecure 读回 WithOidcCookieSecure 写入的判定结果，未写入时按 false 处理。
// grpc-gateway 的 ForwardResponseOption 只拿得到 ctx、拿不到 *http.Request，
// 是本中间件要把判定结果搬进 ctx 的唯一理由。
func IsOidcCookieSecure(ctx context.Context) bool {
	secure, _ := ctx.Value(oidcSecureKey{}).(bool)
	return secure
}

// OidcCookieSecureMiddleware 把「本请求是否经 HTTPS」写入 ctx 后透传下游，
// 供 ForwardResponseOption 决定 state Cookie 要不要带 Secure 属性。
func OidcCookieSecureMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(WithOidcCookieSecure(r.Context(), oidcCookieSecure(r))))
	})
}
