package proxy

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"
)

// NewHTTPProxyClient 构造一个走指定代理、信任自签证书的 HTTP 客户端。
// 代理地址支持 socks5:// 与 http(s):// 两种协议；proxyURL 为空时不走代理。
func NewHTTPProxyClient(proxyURL string) *http.Client {
	return &http.Client{
		Timeout: 2 * time.Minute,
		Transport: &http.Transport{
			Proxy: proxyFunc(proxyURL),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // #nosec G402 信任自签证书
			},
			MaxConnsPerHost: 1000,
		},
	}
}

// proxyFunc 将代理地址解析为 http.Transport 的 Proxy 函数；
// 代理地址无效时返回 nil（表示直连，不走代理）。
func proxyFunc(proxyURL string) func(r *http.Request) (*url.URL, error) {
	return func(r *http.Request) (*url.URL, error) {
		parse, _ := url.Parse(proxyURL)
		if parse != nil && parse.Host != "" {
			return parse, nil
		}
		return nil, nil
	}
}
