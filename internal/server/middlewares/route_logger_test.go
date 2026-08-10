package middlewares

import (
	"context"
	"net/http"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/namespace"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"go.uber.org/mock/gomock"
)

func TestRouteLogger(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	l := mlog.NewMockLogger(controller)
	m := &mockHandler{fn: func(writer http.ResponseWriter, request *http.Request) {}}
	rw := &mockResponseWriter{h: map[string][]string{}}
	l.EXPECT().Debugf("[Http]: method: %v, url: %v, use %v", gomock.Any()).Times(1)
	RouteLogger(l, m).ServeHTTP(rw, &http.Request{})
}

func TestLoggerUnaryServerInterceptor(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	logger := mlog.NewMockLogger(m)
	handler := func(ctx context.Context, req any) (any, error) {
		return nil, nil
	}

	// Test case: request with String method
	req := &namespace.CreateRequest{}
	logger.EXPECT().Debugf("[request logger]: method=%s body=%v", "/test.method", req.String()).Times(1)
	interceptor := LoggerUnaryServerInterceptor(logger)
	_, err := interceptor(context.TODO(), req, &grpc.UnaryServerInfo{FullMethod: "/test.method"}, handler)
	assert.Nil(t, err)

	// Test case: request without String() 方法——不打印日志，直接透传 handler。
	called := 0
	_, err = interceptor(context.TODO(), "plain-string", &grpc.UnaryServerInfo{FullMethod: "/plain.method"}, func(ctx context.Context, req any) (any, error) {
		called++
		return nil, nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, called)
}
