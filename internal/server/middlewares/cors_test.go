package middlewares

import (
	"net/http"
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/mlog"
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
	m := &mockHandler{}
	rw := &mockResponseWriter{h: map[string][]string{}}
	AllowCORS(mlog.NewForConfig(nil), m).ServeHTTP(rw, &http.Request{Header: map[string][]string{"Origin": {"https://mars.com"}, "Access-Control-Request-Method": {"GET"}}, Method: "OPTIONS"})
	assert.Equal(t, "https://mars.com", rw.h["Access-Control-Allow-Origin"][0])
	assert.Equal(t, "Content-Type,Accept,X-Requested-With,Authorization,Accept-Language", rw.h["Access-Control-Allow-Headers"][0])
	assert.Equal(t, "GET,HEAD,POST,PUT,PATCH,DELETE", rw.h["Access-Control-Allow-Methods"][0])

	m2 := &mockHandler{}
	AllowCORS(mlog.NewForConfig(nil), m2).ServeHTTP(rw, &http.Request{Header: map[string][]string{"Origin": {"https://mars.com"}}, Method: "GET"})
	assert.Equal(t, 1, m2.serverCalled)
}
