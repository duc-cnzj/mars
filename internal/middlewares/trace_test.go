package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/propagation"

	marsauthorizor "github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/testutil"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	trace2 "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestTracingIgnoreFn(t *testing.T) {
	assert.False(t, TracingIgnoreFn("/api/xxx"))
	assert.True(t, TracingIgnoreFn("/xxx"))
	assert.True(t, TracingIgnoreFn("/ws"))
	assert.True(t, TracingIgnoreFn("/api/containers/namespaces/{namespace}/pods/{pod}/containers/{container}/stream_logs"))
	assert.True(t, TracingIgnoreFn("/api/metrics/namespace/{namespace}/pods/{pod}/stream"))
}

type testHandler struct {
	requestHandler func(*http.Request)
}

func (t *testHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	t.requestHandler(request)
}

func TestTracingWrapper(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	tp := trace.NewTracerProvider()
	tracer := tp.Tracer("test")
	app.EXPECT().GetTracer().Return(tracer)
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/version", nil)
	TracingWrapper(&testHandler{
		requestHandler: func(request *http.Request) {
			span := trace2.SpanFromContext(request.Context())
			rwSpan := span.(trace.ReadWriteSpan)
			assert.Equal(t, "[GET]: /api/version", rwSpan.Name())
			var m = make(map[string]struct{})
			for _, value := range rwSpan.Attributes() {
				m[string(value.Key)] = struct{}{}
			}
			assert.Len(t, m, 4)
			_, ok := m["url"]
			assert.True(t, ok)
			_, ok = m["method"]
			assert.True(t, ok)
			_, ok = m["user-agent"]
			assert.True(t, ok)
			_, ok = m["component"]
			assert.True(t, ok)
		},
	}).ServeHTTP(w, request)
}

func TestTracingWrapper2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	tp := trace.NewTracerProvider()
	tracer := tp.Tracer("test")
	app.EXPECT().GetTracer().Return(tracer)
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/version", nil)
	request.Header.Set("traceparent", "00-873ef9392cbadea7563141ed35098771-ea43520520f4cb4c-01")
	TracingWrapper(&testHandler{
		requestHandler: func(request *http.Request) {
			span := trace2.SpanFromContext(request.Context())
			rwSpan := span.(trace.ReadWriteSpan)
			assert.Equal(t, "873ef9392cbadea7563141ed35098771", rwSpan.Parent().TraceID().String())
			assert.Equal(t, "ea43520520f4cb4c", rwSpan.Parent().SpanID().String())
			assert.True(t, rwSpan.Parent().IsRemote())
		},
	}).ServeHTTP(w, request)
}

func TestGatewayCarrier(t *testing.T) {
	gc := &GatewayCarrier{}
	assert.Empty(t, gc.Get("key"))
	gc.Set("key", "value")
	assert.Equal(t, "value", gc.Get("key"))
	assert.Equal(t, "value", gc.Get("Key"))
	assert.Equal(t, "value", gc.Get("KEY"))
	assert.Equal(t, []string{"key"}, gc.Keys())
}

type tracerWrap struct {
	ctx  context.Context
	span trace2.Span
	t    trace2.Tracer
}

func (t *tracerWrap) Start(ctx context.Context, spanName string, opts ...trace2.SpanStartOption) (context.Context, trace2.Span) {
	start, span := t.t.Start(ctx, spanName, opts...)
	t.ctx = start
	t.span = span
	return start, span
}

func TestTraceUnaryClientInterceptor(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	tp := trace.NewTracerProvider()
	tracer := tp.Tracer("test")
	tw := &tracerWrap{t: tracer}
	app.EXPECT().GetTracer().Return(tw)
	TraceUnaryClientInterceptor(context.TODO(), "test", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		md, _ := metadata.FromOutgoingContext(ctx)
		get := md.Get("traceparent")
		assert.NotEmpty(t, get)
		return errors.New("err")
	})
	s := tw.span.(trace.ReadWriteSpan)
	assert.False(t, s.Parent().HasTraceID())
	assert.Equal(t, codes.Error, s.Status().Code)
	assert.Equal(t, "err", s.Status().Description)
	assert.Equal(t, "TraceUnaryClientInterceptor: test", s.Name())
	mm := make(map[string]string)
	for _, value := range s.Attributes() {
		mm[string(value.Key)] = value.Value.AsString()
	}
	assert.Equal(t, mm["method"], "test")
}

func TestTraceUnaryClientInterceptor1(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	tp := trace.NewTracerProvider()
	tracer := tp.Tracer("test")
	tw := &tracerWrap{t: tracer}
	app.EXPECT().GetTracer().Return(tw)

	base := context.TODO()
	start, span := tracer.Start(base, "base")
	defer span.End()

	err := TraceUnaryClientInterceptor(metadata.NewOutgoingContext(start, metadata.MD{"a": []string{"b"}}), "test", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		md, _ := metadata.FromOutgoingContext(ctx)
		get := md.Get("traceparent")
		assert.Equal(t, md.Get("a")[0], "b")
		assert.NotEmpty(t, get)
		return nil
	})
	assert.Nil(t, err)
	s := tw.span.(trace.ReadWriteSpan)
	assert.True(t, s.Parent().HasTraceID())
	assert.Equal(t, codes.Unset, s.Status().Code)
	assert.Equal(t, "TraceUnaryClientInterceptor: test", s.Name())
	mm := make(map[string]string)
	for _, value := range s.Attributes() {
		mm[string(value.Key)] = value.Value.AsString()
	}
	assert.Equal(t, mm["method"], "test")
}

func TestTraceUnaryServerInterceptor(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	tp := trace.NewTracerProvider()
	tracer := tp.Tracer("test")
	tw := &tracerWrap{t: tracer}
	app.EXPECT().GetTracer().Return(tw)
	TraceUnaryServerInterceptor(marsauthorizor.SetUser(context.TODO(), &contracts.UserInfo{
		OpenIDClaims: contracts.OpenIDClaims{
			Name:  "duc",
			Email: "1025434218@qq.com",
		},
	}), nil, &grpc.UnaryServerInfo{
		FullMethod: "test",
	}, func(ctx context.Context, req any) (any, error) {
		return nil, errors.New("xxx")
	})
	s := tw.span.(trace.ReadWriteSpan)
	assert.Equal(t, codes.Error, s.Status().Code)
	assert.Equal(t, "xxx", s.Status().Description)
	assert.Equal(t, "test", s.Name())
	mm := make(map[string]string)
	for _, value := range s.Attributes() {
		mm[string(value.Key)] = value.Value.AsString()
	}
	assert.Equal(t, mm["user"], "duc")
	assert.Equal(t, mm["email"], "1025434218@qq.com")
	assert.False(t, s.Parent().HasTraceID())
}

func TestTraceUnaryServerInterceptor1(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	tp := trace.NewTracerProvider()
	tracer := tp.Tracer("test")
	tw := &tracerWrap{t: tracer}
	app.EXPECT().GetTracer().Return(tw)

	base := context.TODO()
	start, span := tracer.Start(base, "base")
	defer span.End()

	ctxt := propagation.TraceContext{}
	md := metadata.MD{}
	ctxt.Inject(start, GatewayCarrier(md))

	res, err := TraceUnaryServerInterceptor(metadata.NewIncomingContext(context.TODO(), md), nil, &grpc.UnaryServerInfo{
		FullMethod: "test",
	}, func(ctx context.Context, req any) (any, error) {
		return "xxx", nil
	})
	assert.Nil(t, err)
	assert.Equal(t, "xxx", res)
	s := tw.span.(trace.ReadWriteSpan)
	assert.True(t, s.Parent().HasTraceID())
	mm := make(map[string]string)
	for _, value := range s.Attributes() {
		mm[string(value.Key)] = value.Value.AsString()
	}
	assert.Equal(t, mm["user"], "")
	assert.Equal(t, mm["email"], "")
	assert.Equal(t, codes.Unset, s.Status().Code)
}
