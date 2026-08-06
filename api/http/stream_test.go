package http

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/namespace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// openStream 错误分支 -------------------------------------------------

// 非 GET/DELETE 方法 + 非 nil 请求：直接报错（当前只支持 query 绑定）。
func TestOpenStream_UnsupportedMethod(t *testing.T) {
	cli, err := NewClient("http://example.com")
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	if _, err := cli.openStream(context.Background(), http.MethodPost, "/api/x", &namespace.ListRequest{}); err == nil {
		t.Fatal("want error for non GET/DELETE with req")
	}
}

// 非 GET/DELETE + nil 请求：不报错，但 NewRequest 仍会构造（走常规路径）。
func TestOpenStream_PostNilReq(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{}"))
	})
	cli, err := NewClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	rc, err := cli.openStream(context.Background(), http.MethodPost, "/api/x", nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = rc.Close()
}

// baseURL 非法：NewRequestWithContext 失败。
func TestOpenStream_BadURL(t *testing.T) {
	if _, err := badBaseClient(t).openStream(context.Background(), http.MethodGet, "/api/x", &namespace.ListRequest{}); err == nil {
		t.Fatal("want error")
	}
}

// 网络错误：hc.Do 失败。
func TestOpenStream_NetworkError(t *testing.T) {
	if _, err := netErrClient(t).openStream(context.Background(), http.MethodGet, "/api/x", &namespace.ListRequest{}); err == nil {
		t.Fatal("want error")
	}
}

// 带 token：Authorization 头正确注入。
func TestOpenStream_WithToken(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("Authorization = %q", got)
		}
		_, _ = w.Write([]byte("{}"))
	})
	cli, err := NewClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	cli.SetBearerToken("tok")
	rc, err := cli.openStream(context.Background(), http.MethodGet, "/api/x", &namespace.ListRequest{})
	if err != nil {
		t.Fatal(err)
	}
	_ = rc.Close()
}

// Recv 边界 -------------------------------------------------

// SSE 流在无尾随空行时直接 EOF：残留的 data 事件应在 EOF 处 flush。
func TestEventStream_Recv_SSE_FlushOnEOF(t *testing.T) {
	rc := io.NopCloser(strings.NewReader(`data: {"log":"tail"}`)) // 无尾随换行
	es := &eventStream{rc: rc, br: bufio.NewReader(rc)}
	var out namespace.ListResponse
	if err := es.Recv(&out); err != nil {
		t.Fatalf("Recv err = %v, want 成功 flush 最后一条 data", err)
	}
}

// 底层 reader 返回非 EOF 错误：原样透传。
func TestEventStream_Recv_ReadError(t *testing.T) {
	rc := io.NopCloser(errReader{err: errors.New("read boom")})
	es := &eventStream{rc: rc, br: bufio.NewReader(rc)}
	var out namespace.ListResponse
	if err := es.Recv(&out); err == nil || !strings.Contains(err.Error(), "read boom") {
		t.Fatalf("err = %v, want read boom", err)
	}
}

// 空流直接 EOF。
func TestEventStream_Recv_EmptyEOF(t *testing.T) {
	rc := io.NopCloser(strings.NewReader(""))
	es := &eventStream{rc: rc, br: bufio.NewReader(rc)}
	var out namespace.ListResponse
	if err := es.Recv(&out); err != io.EOF {
		t.Fatalf("err = %v, want io.EOF", err)
	}
}

// streamErrorFromEnvelope 错误分支 -------------------------------------------------

// envelope 体不是合法 JSON：返回解析失败错误。
func TestStreamErrorFromEnvelope_BadJSON(t *testing.T) {
	if err := streamErrorFromEnvelope([]byte(`not-json`)); err == nil {
		t.Fatal("want error")
	}
}

// openStream 流建立 401 自动刷新 -------------------------------------------------

// 流建立请求 401 → WithTokenAutoRefresh 刷新 token 后重试一次（与 unary 对齐）。
func TestOpenStream_401_AutoRefreshRetries(t *testing.T) {
	var calls int
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			_, _ = w.Write([]byte(`{"token":"tok2"}`))
		case "/api/x":
			calls++
			if calls == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"code":16,"message":"token expired"}`))
				return
			}
			if got := r.Header.Get("Authorization"); got != "Bearer tok2" {
				t.Errorf("重试 Authorization = %q, want %q", got, "Bearer tok2")
			}
			_, _ = w.Write([]byte("{}"))
		}
	})
	cli, err := NewClient(srv.URL, WithAuth("admin", "123456"), WithTokenAutoRefresh())
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	rc, err := cli.openStream(context.Background(), http.MethodGet, "/api/x", &namespace.ListRequest{})
	if err != nil {
		t.Fatalf("openStream 应自动刷新后成功: %v", err)
	}
	defer rc.Close()
	if calls != 2 {
		t.Fatalf("服务端调用次数 = %d, want 2（401 + 刷新后重试）", calls)
	}
}

// 流建立 401 但 refreshToken 失败：返回登录错误，不吞错。
// 注意：WithAuth 构造期即登录一次，login 端点须先成功再失败。
func TestOpenStream_401_RefreshLoginFails(t *testing.T) {
	var loginCalls int
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			loginCalls++
			if loginCalls == 1 {
				_, _ = w.Write([]byte(`{"token":"tok1"}`)) // 构造期成功
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code":13,"message":"login boom"}`)) // 刷新失败
		case "/api/x":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"code":16,"message":"token expired"}`))
		}
	})
	cli, err := NewClient(srv.URL, WithAuth("admin", "123456"), WithTokenAutoRefresh())
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	if _, err := cli.openStream(context.Background(), http.MethodGet, "/api/x", &namespace.ListRequest{}); err == nil {
		t.Fatal("refreshToken 失败时 openStream 应返回 error")
	}
	if loginCalls != 2 {
		t.Fatalf("login 调用次数 = %d, want 2（构造 + 刷新）", loginCalls)
	}
}

// 流建立 401 但未开启 autoRefresh：直接还原 Unauthenticated 错误。
func TestOpenStream_401_NoAutoRefresh_ReturnsError(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			_, _ = w.Write([]byte(`{"token":"tok1"}`)) // 构造期成功
		case "/api/x":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"code":16,"message":"token expired"}`))
		}
	})
	cli, err := NewClient(srv.URL, WithAuth("admin", "123456"))
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	_, err = cli.openStream(context.Background(), http.MethodGet, "/api/x", &namespace.ListRequest{})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("code = %v, want Unauthenticated", status.Code(err))
	}
}

// 流建立 4xx（非 401）：errFromStatus 还原错误。
func TestOpenStream_404_ReturnsError(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":5,"message":"no such stream"}`))
	})
	cli, err := NewClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	_, err = cli.openStream(context.Background(), http.MethodGet, "/api/x", &namespace.ListRequest{})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("code = %v, want NotFound", status.Code(err))
	}
}
