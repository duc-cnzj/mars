package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewApiGateway(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := application.NewMockApp(m)

	app.EXPECT().Logger().Return(mlog.NewForConfig(nil)).Times(1)
	app.EXPECT().GrpcRegistry().Return(nil).Times(1)
	app.EXPECT().Config().Return(&config.Config{}).Times(1)
	app.EXPECT().HttpHandler().Return(nil).Times(1)

	gw := NewApiGateway("test-endpoint", app)
	assert.NotNil(t, gw)
}

func Test_apiGateway_Run(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	server := NewMockHttpServer(m)
	handler := application.NewMockHttpHandler(m)
	logger := mlog.NewMockLogger(m)
	gw := &apiGateway{
		handler: handler,
		newServerFunc: func(ctx context.Context, a *apiGateway) (HttpServer, error) {
			return server, nil
		},
		logger: logger,
		port:   "111",
	}

	handler.EXPECT().TickClusterHealth(gomock.Any())
	logger.EXPECT().Infof("[Server]: start apiGateway runner at :%s.", "111").Times(1)
	server.EXPECT().ListenAndServe().Return(assert.AnError).Times(1)
	logger.EXPECT().Error(gomock.Any()).Times(1)
	err := gw.Run(context.TODO())
	time.Sleep(1 * time.Second)
	assert.NoError(t, err)
}

func Test_apiGateway_Shutdown(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	server := NewMockHttpServer(m)
	logger := mlog.NewMockLogger(m)

	handler := application.NewMockHttpHandler(m)
	gw := &apiGateway{
		handler: handler,
		server:  server,
		logger:  logger,
	}
	handler.EXPECT().Shutdown(gomock.Any())
	logger.EXPECT().Info("[Server]: shutdown api-gateway runner.").Times(1)
	server.EXPECT().Shutdown(gomock.Any()).Return(nil).Times(1)
	assert.Nil(t, gw.Shutdown(context.TODO()))
}

func TestMiddlewareList_Wrap(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	middleware := func(logger mlog.Logger, handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "middleware")
			handler.ServeHTTP(w, r)
		})
	}

	middlewareList := middlewareList{middleware}

	wrappedHandler := middlewareList.Wrap(logger, handler)

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, "middleware", rr.Header().Get("X-Test"))
	assert.Equal(t, "test", rr.Body.String())
}

func TestMiddlewareList_Wrap_Empty(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	middlewareList := middlewareList{}

	wrappedHandler := middlewareList.Wrap(logger, handler)

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, "", rr.Header().Get("X-Test"))
	assert.Equal(t, "test", rr.Body.String())
}

func TestHeaderMatcher(t *testing.T) {
	// Test case: tracestate key
	key, ok := headerMatcher("tracestate")
	assert.True(t, ok)
	assert.Equal(t, "tracestate", key)

	// Test case: traceparent key
	key, ok = headerMatcher("traceparent")
	assert.True(t, ok)
	assert.Equal(t, "traceparent", key)

	// Test case: other key
	key, ok = headerMatcher("other")
	assert.False(t, ok)
	assert.Equal(t, "", key)

	// Test case: empty key
	key, ok = headerMatcher("")
	assert.False(t, ok)
	assert.Equal(t, "", key)
}

func Test_initServer(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	server := NewMockHttpServer(m)
	handler := application.NewMockHttpHandler(m)
	handler.EXPECT().RegisterSwaggerUIRoute(gomock.Not(nil)).Times(1)
	handler.EXPECT().RegisterWsRoute(gomock.Not(nil)).Times(1)
	handler.EXPECT().RegisterFileRoute(gomock.Not(nil)).Times(1)
	httpServer, err := initServer(context.TODO(), &apiGateway{
		endpoint:     "x",
		port:         "1000",
		server:       server,
		logger:       mlog.NewForConfig(nil),
		grpcRegistry: &application.GrpcRegistry{},
		handler:      handler,
	})
	assert.Nil(t, err)
	assert.NotNil(t, httpServer)
	assert.Equal(t, httpServer.(*http.Server).Addr, ":1000")
	assert.Equal(t, httpServer.(*http.Server).ReadHeaderTimeout, 5*time.Second)
}
