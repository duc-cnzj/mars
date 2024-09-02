package proxy

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHttpProxyClient(t *testing.T) {
	client := NewHttpProxyClient("http://localhost:8080")
	assert.NotNil(t, client)
	assert.Equal(t, 2*time.Minute, client.Timeout)
}

func Test_proxyFunc(t *testing.T) {
	proxyUrl := "http://localhost:8080"
	f := proxyFunc(proxyUrl)
	req := &http.Request{}
	u, err := f(req)
	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, proxyUrl, u.String())
}

func Test_proxyFunc2(t *testing.T) {
	proxyUrl := ""
	f := proxyFunc(proxyUrl)
	req := &http.Request{}
	u, err := f(req)
	assert.Nil(t, u)
	assert.Nil(t, err)
}
