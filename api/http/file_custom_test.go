package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// errTransport 是固定返回错误的 RoundTripper，用于确定性地触发 hc.Do 的网络错误分支。
type errTransport struct{ err error }

func (t errTransport) RoundTrip(*http.Request) (*http.Response, error) { return nil, t.err }

// errReader 让 io.Copy / io.ReadAll 读文件时失败，用于触发上传流复制错误分支。
type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) { return 0, r.err }

// badBaseClient 返回 baseURL 含空格的客户端：URL 构造必然失败（NewRequestWithContext error）。
func badBaseClient(t *testing.T) *Client {
	t.Helper()
	cli, err := NewHTTPClient("http://exa mple.com")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cli.Close() })
	return cli
}

// netErrClient 返回 hc.Do 必然失败的客户端（自定义 RoundTripper 返回错误）。
func netErrClient(t *testing.T) *Client {
	t.Helper()
	cli, err := NewHTTPClient("http://example.com",
		WithHTTPClient(&http.Client{Transport: errTransport{err: errors.New("boom")}}))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cli.Close() })
	return cli
}

// UploadFile 错误分支 -------------------------------------------------

// URL 构造失败：baseURL 非法 → NewRequestWithContext error。
func TestFileSvc_UploadFile_BadURL(t *testing.T) {
	_, err := badBaseClient(t).File().UploadFile(context.Background(), "a.txt", strings.NewReader("x"))
	if err == nil {
		t.Fatal("want error")
	}
}

// 网络错误：hc.Do 失败。
func TestFileSvc_UploadFile_NetworkError(t *testing.T) {
	_, err := netErrClient(t).File().UploadFile(context.Background(), "a.txt", strings.NewReader("x"))
	if err == nil {
		t.Fatal("want error")
	}
}

// 源 reader 报错：multipart 复制文件内容失败 → io.Copy error。
func TestFileSvc_UploadFile_CopyError(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("不应到达服务端")
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	_, err = cli.File().UploadFile(context.Background(), "a.txt", errReader{err: errors.New("read boom")})
	if err == nil {
		t.Fatal("want error")
	}
}

// 带 token：Authorization 头正确注入（UploadFile token 分支）。
func TestFileSvc_UploadFile_WithToken(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("Authorization = %q", got)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1}`))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	cli.SetBearerToken("tok")
	if _, err := cli.File().UploadFile(context.Background(), "a.txt", strings.NewReader("x")); err != nil {
		t.Fatal(err)
	}
}

// 服务端返回非 201：gateway 错误体 → codes.Error；非 gateway 兜底 → unexpected status。
func TestFileSvc_UploadFile_NonCreated(t *testing.T) {
	t.Run("gateway错误体还原为codes.Error", func(t *testing.T) {
		srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"code":5,"message":"not found"}`))
		})
		cli, err := NewHTTPClient(srv.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer cli.Close()
		_, err = cli.File().UploadFile(context.Background(), "a.txt", strings.NewReader("x"))
		if status.Code(err) != codes.NotFound {
			t.Errorf("code = %v, want NotFound", status.Code(err))
		}
	})

	t.Run("code为0的body走unexpected兜底", func(t *testing.T) {
		// {"code":0} 被 parseGatewayError 判为非法错误体 → 走到 errFromStatus 的兜底分支。
		srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code":0,"message":"boom"}`))
		})
		cli, err := NewHTTPClient(srv.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer cli.Close()
		_, err = cli.File().UploadFile(context.Background(), "a.txt", strings.NewReader("x"))
		if err == nil || !strings.Contains(err.Error(), "unexpected status") {
			t.Errorf("err = %v, want unexpected status 兜底", err)
		}
	})
}

// 服务端返回 200（而非 201）：也应视为成功并解析 body（状态码判定回归）。
func TestFileSvc_UploadFile_200(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":42}`))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	out, err := cli.File().UploadFile(context.Background(), "a.txt", strings.NewReader("x"))
	if err != nil {
		t.Fatalf("200 响应不应报错: %v", err)
	}
	if out == nil || out.ID != 42 {
		t.Fatalf("out = %+v, want id=42", out)
	}
}

// 服务端 201 但 body 不是合法 JSON：json.Unmarshal error。
func TestFileSvc_UploadFile_BadJSON(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`not-json`))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	if _, err := cli.File().UploadFile(context.Background(), "a.txt", strings.NewReader("x")); err == nil {
		t.Fatal("want error")
	}
}

// DownloadFile 错误分支 -------------------------------------------------

// URL 构造失败。
func TestFileSvc_DownloadFile_BadURL(t *testing.T) {
	if _, _, err := badBaseClient(t).File().DownloadFile(context.Background(), 1); err == nil {
		t.Fatal("want error")
	}
}

// 网络错误。
func TestFileSvc_DownloadFile_NetworkError(t *testing.T) {
	if _, _, err := netErrClient(t).File().DownloadFile(context.Background(), 1); err == nil {
		t.Fatal("want error")
	}
}

// 服务端 4xx：errFromStatus 还原错误。
func TestFileSvc_DownloadFile_ServerError(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":5,"message":"no such file"}`))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	if _, _, err := cli.File().DownloadFile(context.Background(), 99); status.Code(err) != codes.NotFound {
		t.Errorf("code = %v, want NotFound", status.Code(err))
	}
}

// 带 token。
func TestFileSvc_DownloadFile_WithToken(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("Authorization = %q", got)
		}
		_, _ = w.Write([]byte("data"))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	cli.SetBearerToken("tok")
	rc, _, err := cli.File().DownloadFile(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
}

// CopyFromPod 错误分支 -------------------------------------------------

// URL 构造失败。
func TestFileSvc_CopyFromPod_BadURL(t *testing.T) {
	req := &CopyFromPodRequest{Namespace: "n", Pod: "p", Container: "c", FilePath: "/x"}
	if _, _, err := badBaseClient(t).File().CopyFromPod(context.Background(), req); err == nil {
		t.Fatal("want error")
	}
}

// 网络错误。
func TestFileSvc_CopyFromPod_NetworkError(t *testing.T) {
	req := &CopyFromPodRequest{Namespace: "n", Pod: "p", Container: "c", FilePath: "/x"}
	if _, _, err := netErrClient(t).File().CopyFromPod(context.Background(), req); err == nil {
		t.Fatal("want error")
	}
}

// 服务端 4xx。
func TestFileSvc_CopyFromPod_ServerError(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":3,"message":"bad req"}`))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	req := &CopyFromPodRequest{Namespace: "n", Pod: "p", Container: "c", FilePath: "/x"}
	if _, _, err := cli.File().CopyFromPod(context.Background(), req); status.Code(err) != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", status.Code(err))
	}
}

// 带 token。
func TestFileSvc_CopyFromPod_WithToken(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("Authorization = %q", got)
		}
		_, _ = w.Write([]byte("data"))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	cli.SetBearerToken("tok")
	req := &CopyFromPodRequest{Namespace: "n", Pod: "p", Container: "c", FilePath: "/x"}
	rc, _, err := cli.File().CopyFromPod(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
}

// 服务端返回 Content-Disposition：下载元信息解析 filename（CopyFromPod 侧 CD 分支）。
func TestFileSvc_CopyFromPod_ContentDisposition(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="b.txt"`)
		_, _ = w.Write([]byte("data"))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	req := &CopyFromPodRequest{Namespace: "n", Pod: "p", Container: "c", FilePath: "/x"}
	rc, info, err := cli.File().CopyFromPod(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	if info.Filename != "b.txt" {
		t.Errorf("filename = %q, want b.txt", info.Filename)
	}
}

// 下载无 Content-Disposition 头：Filename 保持空，不应 panic。
func TestFileSvc_DownloadFile_NoContentDisposition(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("data"))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	rc, info, err := cli.File().DownloadFile(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	if info.Filename != "" {
		t.Errorf("filename = %q, want empty", info.Filename)
	}
	if _, err := io.Copy(io.Discard, rc); err != nil {
		t.Fatal(err)
	}
}
