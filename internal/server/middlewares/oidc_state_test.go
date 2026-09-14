package middlewares

import (
	"context"
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOidcCookieSecure(t *testing.T) {
	cases := []struct {
		name    string
		tls     bool
		forward string
		want    bool
	}{
		{name: "无 TLS 且无反代头", want: false},
		{name: "直连 TLS 握手", tls: true, want: true},
		{name: "反代声明 https", forward: "https", want: true},
		{name: "反代声明 HTTPS 大写", forward: "HTTPS", want: true},
		{name: "反代声明 http", forward: "http", want: false},
		{name: "反代头为空串", forward: "", want: false},
		{name: "反代头不可识别", forward: "ftp", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &http.Request{Header: http.Header{}}
			if tc.forward != "" {
				r.Header.Set("X-Forwarded-Proto", tc.forward)
			}
			if tc.tls {
				r.TLS = &tls.ConnectionState{}
			}
			assert.Equal(t, tc.want, oidcCookieSecure(r))
		})
	}
}

func TestWithIsOidcCookieSecure(t *testing.T) {
	// 未写入 ctx：按 false 处理（拿不准时不给 Cookie 加 Secure，否则明文 HTTP 部署会丢 Cookie）。
	assert.False(t, IsOidcCookieSecure(context.TODO()))
	// 写入 true / false 均能原样读回。
	assert.True(t, IsOidcCookieSecure(WithOidcCookieSecure(context.TODO(), true)))
	assert.False(t, IsOidcCookieSecure(WithOidcCookieSecure(context.TODO(), false)))
}

func TestOidcCookieSecureMiddleware(t *testing.T) {
	cases := []struct {
		name string
		req  *http.Request
		want bool
	}{
		{
			name: "明文 HTTP 判定为 false 并透传下游",
			req:  &http.Request{Header: http.Header{}},
			want: false,
		},
		{
			name: "反代声明 https 判定为 true 并透传下游",
			req:  &http.Request{Header: http.Header{"X-Forwarded-Proto": {"https"}}},
			want: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got bool
			called := 0
			h := OidcCookieSecureMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				called++
				got = IsOidcCookieSecure(r.Context())
			}))
			h.ServeHTTP(&mockResponseWriter{h: map[string][]string{}}, tc.req)
			assert.Equal(t, 1, called, "中间件必须透传下游")
			assert.Equal(t, tc.want, got)
			// 判定结果只挂在派生请求的 ctx 上，不改写调用方持有的原始请求。
			assert.Nil(t, tc.req.Context().Value(oidcSecureKey{}))
		})
	}
}
