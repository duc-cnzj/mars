package proxy

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"
)

// NewHttpProxyClient 支持 socks5://xx http[s]://xx
func NewHttpProxyClient(proxyUrl string) *http.Client {
	return &http.Client{
		Timeout: 2 * time.Minute,
		Transport: &http.Transport{
			Proxy: proxyFunc(proxyUrl),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // #nosec G402
			},
			MaxConnsPerHost: 1000,
		},
	}
}

func proxyFunc(proxyUrl string) func(r *http.Request) (*url.URL, error) {
	return func(r *http.Request) (*url.URL, error) {
		parse, _ := url.Parse(proxyUrl)
		if parse != nil && parse.Host != "" {
			return parse, nil
		}
		return nil, nil
	}
}
