package http

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/container"
	"github.com/duc-cnzj/mars/api/v6/proto/namespace"
)

// WithHeader 单条：自定义 header 到达服务端（unary 路径）。
func TestWithHeader_Applied(t *testing.T) {
	var got string
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Request-ID")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[],"count":0,"page":1,"pageSize":10}`))
	})
	cli, err := NewClient(srv.URL, WithHeader("X-Request-ID", "req-123"))
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	if _, err := cli.Namespace().List(context.Background(), &namespace.ListRequest{}); err != nil {
		t.Fatal(err)
	}
	if got != "req-123" {
		t.Fatalf("X-Request-ID = %q, want %q", got, "req-123")
	}
}

// WithHeaders 批量：多个自定义 header 全部生效；空 key 被忽略（不产生空 header、不 panic）。
func TestWithHeaders_Applied(t *testing.T) {
	var got map[string]string
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		got = map[string]string{
			"X-A": r.Header.Get("X-A"),
			"X-B": r.Header.Get("X-B"),
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[],"count":0,"page":1,"pageSize":10}`))
	})
	cli, err := NewClient(srv.URL, WithHeaders(map[string]string{
		"X-A": "1",
		"X-B": "2",
		"":    "ignored", // 空 key 应被忽略
	}))
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	if _, err := cli.Namespace().List(context.Background(), &namespace.ListRequest{}); err != nil {
		t.Fatal(err)
	}
	if got["X-A"] != "1" || got["X-B"] != "2" {
		t.Fatalf("headers = %v, want X-A=1 X-B=2", got)
	}
}

// 覆盖策略：自定义 Authorization 覆盖 SDK 自动注入的 Bearer token。
func TestWithHeader_OverridesAuthorization(t *testing.T) {
	var gotAuth string
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[],"count":0,"page":1,"pageSize":10}`))
	})
	cli, err := NewClient(srv.URL, WithBearerToken("sdk-token"), WithHeader("Authorization", "Custom cred"))
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	if _, err := cli.Namespace().List(context.Background(), &namespace.ListRequest{}); err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Custom cred" {
		t.Fatalf("Authorization = %q, want 自定义覆盖 %q", gotAuth, "Custom cred")
	}
}

// WithHeader 空 key 忽略：不 panic，仅非空 key 生效。
func TestWithHeader_IgnoresEmptyKey(t *testing.T) {
	var got string
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-OK")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[],"count":0,"page":1,"pageSize":10}`))
	})
	cli, err := NewClient(srv.URL, WithHeader("", "ignored"), WithHeader("X-OK", "yes"))
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	if _, err := cli.Namespace().List(context.Background(), &namespace.ListRequest{}); err != nil {
		t.Fatal(err)
	}
	if got != "yes" {
		t.Fatalf("X-OK = %q, want %q", got, "yes")
	}
}

// 自定义 headers 覆盖全部 5 条请求出口：unary / streaming / upload / download / copy_from_pod。
func TestHeaders_OnAllPaths(t *testing.T) {
	got := make(map[string]string)
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		got[r.URL.Path] = r.Header.Get("X-Custom-Header")
		switch r.URL.Path {
		case "/api/files":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":1}`))
		case "/api/download_file/1":
			_, _ = w.Write([]byte("data"))
		case "/api/copy_from_pod":
			_, _ = w.Write([]byte("data"))
		case "/api/containers/namespaces/devops/pods/p-1/containers/app/stream_logs":
			// 空流：Recv 返回 io.EOF
		default:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"items":[],"count":0,"page":1,"pageSize":10}`))
		}
	})
	cli, err := NewClient(srv.URL, WithHeader("X-Custom-Header", "hdr"))
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	if _, err := cli.Namespace().List(context.Background(), &namespace.ListRequest{}); err != nil {
		t.Fatal(err)
	}

	stream, err := cli.Container().StreamContainerLog(context.Background(), &container.LogRequest{
		Namespace: "devops", Pod: "p-1", Container: "app",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, _ = stream.Recv() // io.EOF，空流正常结束
	_ = stream.Close()

	if _, err := cli.File().UploadFile(context.Background(), "a.txt", strings.NewReader("x")); err != nil {
		t.Fatal(err)
	}
	rc1, _, err := cli.File().DownloadFile(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	_ = rc1.Close()
	rc2, _, err := cli.File().CopyFromPod(context.Background(), &CopyFromPodRequest{Namespace: "n"})
	if err != nil {
		t.Fatal(err)
	}
	_ = rc2.Close()

	for _, path := range []string{
		"/api/namespaces",
		"/api/containers/namespaces/devops/pods/p-1/containers/app/stream_logs",
		"/api/files",
		"/api/download_file/1",
		"/api/copy_from_pod",
	} {
		if got[path] != "hdr" {
			t.Errorf("path %s header = %q, want %q", path, got[path], "hdr")
		}
	}
}
