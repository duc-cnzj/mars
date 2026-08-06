package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// 选项都是纯 setter，不需要网络：用 localhost:1 的 baseURL，且不带 WithAuth 就不会登录。
func TestOptions_Apply(t *testing.T) {
	custom := &http.Client{Timeout: 30 * time.Second}

	c, err := NewClient("http://localhost:1/", // 末尾斜杠应被 trim
		WithBearerToken("abc"),
		WithHTTPClient(custom),
		WithTimeout(15*time.Second),
		WithTokenAutoRefresh(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if c.baseURL != "http://localhost:1" {
		t.Fatalf("baseURL = %q, want %q（末尾斜杠应被裁掉）", c.baseURL, "http://localhost:1")
	}
	if got := c.authToken(); got != "Bearer abc" {
		t.Fatalf("authToken = %q, want %q", got, "Bearer abc")
	}
	if !c.autoRefresh {
		t.Fatal("WithTokenAutoRefresh 未生效")
	}
	if c.hc != custom {
		t.Fatal("WithHTTPClient 未生效")
	}
	if c.hc.Timeout != 15*time.Second {
		t.Fatalf("WithTimeout 未生效: %v", c.hc.Timeout)
	}
}

func TestWithTimeout_CreatesClient(t *testing.T) {
	c, err := NewClient("http://localhost:1", WithTimeout(5*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if c.hc == nil || c.hc.Timeout != 5*time.Second {
		t.Fatalf("WithTimeout 应自建 http.Client: %+v", c.hc)
	}
}

func TestSetBearerToken_Runtime(t *testing.T) {
	c, err := NewClient("http://localhost:1")
	if err != nil {
		t.Fatal(err)
	}
	c.SetBearerToken("runtime-token")
	if got := c.authToken(); got != "Bearer runtime-token" {
		t.Fatalf("SetBearerToken 后 authToken = %q", got)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// WithTracer 应把默认 transport 与自定义 transport 都包成 otelhttp.Transport。
func TestWithTracer_WrapsTransport(t *testing.T) {
	c, err := NewClient("http://localhost:1", WithTracer())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c.hc.Transport.(*otelhttp.Transport); !ok {
		t.Fatalf("默认 transport 应被 otelhttp 包装，实际 %T", c.hc.Transport)
	}

	custom := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent}, nil
	})
	c2, err := NewClient("http://localhost:1", WithTracer(), WithHTTPClient(&http.Client{Transport: custom}))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c2.hc.Transport.(*otelhttp.Transport); !ok {
		t.Fatalf("自定义 transport 应被 otelhttp 包装，实际 %T", c2.hc.Transport)
	}
}

// WithTracer 的请求应注入 traceparent 头（端到端验证 propagation）。
func TestWithTracer_PropagatesTraceContext(t *testing.T) {
	prevTracer, prevProp := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	otel.SetTracerProvider(sdktrace.NewTracerProvider())
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTracer)
		otel.SetTextMapPropagator(prevProp)
	})

	var gotTraceparent string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTraceparent = r.Header.Get("traceparent")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, WithTracer())
	if err != nil {
		t.Fatal(err)
	}
	ctx, span := otel.Tracer("test").Start(context.Background(), "req")
	defer span.End()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/x", nil)
	if _, err := c.hc.Do(req); err != nil {
		t.Fatal(err)
	}
	if gotTraceparent == "" {
		t.Fatal("traceparent 未注入到出站请求")
	}
}

// 全部 15 个 service 访问器都应返回非 nil 客户端（与 gRPC SDK 对齐）。
func TestServiceAccessors_AllWired(t *testing.T) {
	c, err := NewClient("http://localhost:1")
	if err != nil {
		t.Fatal(err)
	}
	for name, svc := range map[string]interface{}{
		"AccessToken": c.AccessToken(),
		"Auth":        c.Auth(),
		"Changelog":   c.Changelog(),
		"Cluster":     c.Cluster(),
		"Container":   c.Container(),
		"Endpoint":    c.Endpoint(),
		"Event":       c.Event(),
		"File":        c.File(),
		"Git":         c.Git(),
		"Metrics":     c.Metrics(),
		"Namespace":   c.Namespace(),
		"Picture":     c.Picture(),
		"Project":     c.Project(),
		"Repo":        c.Repo(),
		"Version":     c.Version(),
	} {
		if svc == nil {
			t.Errorf("访问器 %s 返回 nil", name)
		}
	}
}
