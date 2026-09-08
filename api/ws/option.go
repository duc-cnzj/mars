package ws

// option.go 定义 ws 客户端的构造期可选项（Option），
// 风格对齐 api/http 与 api/grpc：opts 在 NewClient 内逐个应用到 *Client。

import (
	"context"
	"net/http"

	"github.com/cenkalti/backoff/v4"
	"github.com/gorilla/websocket"
)

// Option 是 NewClient 的构造期选项，返回的闭包接收 *Client 完成副作用配置。
type Option func(*Client)

// authCreds 是 WithAuth 登录凭据的内存载体。
type authCreds struct {
	username string
	password string
}

// WithBearerToken 直接注入已签发的 JWT，连接时作为 HandleAuthorize 帧的 token 发送。
// 适合调用方自行管理 token 生命周期；token 过期后需重建或配合 WithTokenProvider。
func WithBearerToken(token string) Option {
	return func(c *Client) {
		c.tokenProvider = func(context.Context) (string, error) {
			return token, nil
		}
	}
}

// WithAuth 用用户名密码在连接时登录换取 token（复用 api/http 的 Auth().Login，
// base URL 由 ws URL 的 scheme 推导：ws→http / wss→https，去掉路径）。
// 每次（重）连接都会重新登录取新 token，token 过期后自动续期。
func WithAuth(username, password string) Option {
	return func(c *Client) {
		c.httpAuth = &authCreds{username: username, password: password}
	}
}

// WithTokenProvider 自定义 token 提供函数：每次（重）连接时调用一次取新 token，
// 是最灵活的鉴权原语（可做缓存、OIDC exchange、多租户取值等）。
func WithTokenProvider(fn func(ctx context.Context) (string, error)) Option {
	return func(c *Client) {
		c.tokenProvider = fn
	}
}

// WithHTTPClient 注入底层 *http.Client，仅配合 WithAuth 登录使用
// （例如自定义 transport / TLS / 代理）。
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithDialer 注入自定义 websocket 拨号器（例如自定义 TLS、代理、握手超时）。
func WithDialer(d *websocket.Dialer) Option {
	return func(c *Client) {
		if d != nil {
			c.dialer = d
		}
	}
}

// WithReconnectBackoff 自定义断线重连的退避策略。
// 默认 exponential（起始约 500ms，翻倍，上限 60s），连不上会持续退避重试。
func WithReconnectBackoff(b backoff.BackOff) Option {
	return func(c *Client) {
		if b != nil {
			c.backoff = b
		}
	}
}
