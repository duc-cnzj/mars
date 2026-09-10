package middlewares

import (
	"net/http"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
)

type mockHandler struct {
	fn           func(writer http.ResponseWriter, request *http.Request)
	serverCalled int
}

func (m *mockHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	m.serverCalled++
	if m.fn != nil {
		m.fn(writer, request)
	}
}

type mockResponseWriter struct {
	code int
	h    http.Header
}

func (m *mockResponseWriter) Header() http.Header {
	return m.h
}

func (m *mockResponseWriter) Write(bytes []byte) (int, error) {
	return len(bytes), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.code = statusCode
}

func TestAllowCORS(t *testing.T) {
	// 预检请求（OPTIONS + Access-Control-Request-Method）：写预检头并短路，不进下游。
	m := &mockHandler{}
	rw := &mockResponseWriter{h: map[string][]string{}}
	AllowCORS(mlog.NewForConfig(nil), m).ServeHTTP(rw, &http.Request{Header: map[string][]string{"Origin": {"https://mars.com"}, "Access-Control-Request-Method": {"GET"}}, Method: "OPTIONS"})
	assert.Equal(t, "https://mars.com", rw.h["Access-Control-Allow-Origin"][0])
	assert.Equal(t, "Content-Type,Accept,X-Requested-With,Authorization,Accept-Language", rw.h["Access-Control-Allow-Headers"][0])
	assert.Equal(t, "GET,HEAD,POST,PUT,PATCH,DELETE", rw.h["Access-Control-Allow-Methods"][0])
	assert.Equal(t, 0, m.serverCalled, "预检请求必须短路，不得调用下游")

	// 带 Origin 的非预检：写 Allow-Origin 头并透传下游。
	m2 := &mockHandler{}
	rw2 := &mockResponseWriter{h: map[string][]string{}}
	AllowCORS(mlog.NewForConfig(nil), m2).ServeHTTP(rw2, &http.Request{Header: map[string][]string{"Origin": {"https://mars.com"}}, Method: "GET"})
	assert.Equal(t, "https://mars.com", rw2.h["Access-Control-Allow-Origin"][0])
	assert.Equal(t, 1, m2.serverCalled)

	// 无 Origin：不设任何 CORS 头，直接透传下游。
	m3 := &mockHandler{}
	rw3 := &mockResponseWriter{h: map[string][]string{}}
	AllowCORS(mlog.NewForConfig(nil), m3).ServeHTTP(rw3, &http.Request{Method: "GET"})
	assert.Equal(t, 1, m3.serverCalled)
	assert.Empty(t, rw3.h["Access-Control-Allow-Origin"])
}
