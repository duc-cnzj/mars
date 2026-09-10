package proxy

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPProxyClient(t *testing.T) {
	client := NewHTTPProxyClient("http://localhost:8080")
	assert.NotNil(t, client)
	assert.Equal(t, 2*time.Minute, client.Timeout)
}

func TestProxyFunc(t *testing.T) {
	proxyURL := "http://localhost:8080"
	f := proxyFunc(proxyURL)
	req := &http.Request{}
	u, err := f(req)
	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, proxyURL, u.String())
}

func TestProxyFuncEmpty(t *testing.T) {
	f := proxyFunc("")
	req := &http.Request{}
	u, err := f(req)
	assert.Nil(t, u)
	assert.Nil(t, err)
}
