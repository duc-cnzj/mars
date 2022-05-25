package middlewares

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/opentracing/opentracing-go"

	"github.com/stretchr/testify/assert"
)

func TestTracingIgnoreFn(t *testing.T) {
	assert.False(t, TracingIgnoreFn(context.TODO(), "/api/xxx"))
	assert.True(t, TracingIgnoreFn(context.TODO(), "/xxx"))
	assert.True(t, TracingIgnoreFn(context.TODO(), "/ws"))
	assert.True(t, TracingIgnoreFn(context.TODO(), "/api/containers/namespaces/{namespace}/pods/{pod}/containers/{container}/stream_logs"))
	assert.True(t, TracingIgnoreFn(context.TODO(), "/api/metrics/namespace/{namespace}/pods/{pod}/stream"))
}

type mockHandlerTrace struct {
	req *http.Request
}

func (m *mockHandlerTrace) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	m.req = request
}
func TestTracingWrapper(t *testing.T) {
	m := &mockHandlerTrace{}
	rw := &mockResponseWriter{}
	r := &http.Request{URL: &url.URL{Path: "/api/xxx"}}
	ctx := context.TODO()
	r = r.WithContext(ctx)
	TracingWrapper(m).ServeHTTP(rw, r)
	fromContext := opentracing.SpanFromContext(m.req.Context())
	assert.NotNil(t, fromContext)
}

func TestTracingWrapper1(t *testing.T) {
	m := &mockHandlerTrace{}
	rw := &mockResponseWriter{}
	r := &http.Request{URL: &url.URL{Path: "/xxx"}}
	ctx := context.TODO()
	r = r.WithContext(ctx)
	TracingWrapper(m).ServeHTTP(rw, r)
	fromContext := opentracing.SpanFromContext(m.req.Context())
	assert.Nil(t, fromContext)
}
