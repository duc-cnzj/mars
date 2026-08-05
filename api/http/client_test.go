package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/internal/flight"
	"github.com/duc-cnzj/mars/api/v6/proto/auth"
	"github.com/duc-cnzj/mars/api/v6/proto/container"
	"github.com/duc-cnzj/mars/api/v6/proto/file"
	"github.com/duc-cnzj/mars/api/v6/proto/git"
	"github.com/duc-cnzj/mars/api/v6/proto/namespace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ptr[T any](v T) *T { return &v }

func newTestServer(t *testing.T, h http.HandlerFunc) *httptest.Server {
	t.Helper()
	s := httptest.NewServer(h)
	t.Cleanup(s.Close)
	return s
}

// SetBearerToken：运行期替换 token，自动补 Bearer 前缀。
func TestClient_SetBearerToken(t *testing.T) {
	cli, err := NewHTTPClient("http://example.com")
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	cli.SetBearerToken("abc")
	if got := cli.authToken(); got != "Bearer abc" {
		t.Errorf("token = %q, want %q", got, "Bearer abc")
	}
	cli.SetBearerToken("Bearer xyz") // 已带前缀则原样透传
	if got := cli.authToken(); got != "Bearer xyz" {
		t.Errorf("token = %q, want %q", got, "Bearer xyz")
	}
}

// setToken 前缀归一化：与 gRPC SDK 语义一致，非空且未带 "Bearer " 前缀一律补。
func TestSetToken_NormalizesPrefix(t *testing.T) {
	c := &Client{}
	for _, tc := range []struct{ in, want string }{
		{"", ""},                          // 空 token 原样
		{"abc", "Bearer abc"},             // 短 token 补前缀
		{"Bearer def", "Bearer def"},      // 已带前缀保持
		{"bearer ghi", "bearer ghi"},      // 大小写不敏感保持
		{"bearertok", "Bearer bearertok"}, // 缺空格（恰以 bearer 开头）也需补前缀
		{"x12345678", "Bearer x12345678"}, // 无前缀加 Bearer
	} {
		c.setToken(tc.in)
		if got := c.authToken(); got != tc.want {
			t.Errorf("setToken(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// refreshToken 并发去重：10 个 goroutine 同时刷新，只发 1 次登录请求。
func TestClient_refresh_singleflight(t *testing.T) {
	var (
		mu      sync.Mutex
		logins  int
		started = make(chan struct{})
		release = make(chan struct{})
		barrier atomic.Bool // 构造登录已过，之后进入"刷新同步模式"
	)
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/login" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		mu.Lock()
		logins++
		mu.Unlock()
		if !barrier.CompareAndSwap(false, true) {
			// 构造期那次 login 直接放行；之后的第一次刷新 login 进入同步阻塞。
			close(started)
			<-release
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"t"}`))
	})
	cli, err := NewHTTPClient(srv.URL, WithAuth("a", "b")) // 构造登录 1 次
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	cli.setToken("stale") // 模拟过期 token
	mu.Lock()
	logins = 0 // 只统计刷新阶段
	mu.Unlock()

	firstDone := make(chan error, 1)
	go func() { firstDone <- cli.refreshToken() }()
	<-started // 刷新阶段的第一次 login 已在途

	var wg sync.WaitGroup
	for i := 0; i < 9; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := cli.refreshToken(); err != nil {
				t.Error(err)
			}
		}()
	}
	close(release) // 放行，singleflight 只让一个 login 通过
	wg.Wait()
	if err := <-firstDone; err != nil {
		t.Error(err)
	}

	mu.Lock()
	defer mu.Unlock()
	if logins != 1 {
		t.Errorf("logins = %d, want 1（singleflight 合并并发刷新）", logins)
	}
	if got := cli.authToken(); got != "Bearer t" {
		t.Errorf("token = %q, want %q", got, "Bearer t")
	}
}

// 构造：WithAuth 在 NewHTTPClient 阶段自动登录并注入 token。
func TestNewHTTPClient_login(t *testing.T) {
	var loginCalled bool
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/login" && r.Method == http.MethodPost {
			loginCalled = true
			var req auth.LoginRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Error(err)
			}
			if req.Username != "admin" || req.Password != "pwd" {
				t.Errorf("bad login body: %s", req.String())
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"token":"tok123","expiresIn":3600}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	cli, err := NewHTTPClient(srv.URL, WithAuth("admin", "pwd"))
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	if !loginCalled {
		t.Error("expected login call during construction")
	}
	if got := cli.authToken(); got != "Bearer tok123" {
		t.Errorf("token = %q, want %q", got, "Bearer tok123")
	}
}

// do()：GET 请求字段扁平化为 query，字段名用 camelCase。
func TestClient_GET_query(t *testing.T) {
	var gotURI string
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotURI = r.URL.RequestURI()
		if r.URL.Path != "/api/namespaces" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[],"count":0,"page":1,"pageSize":10}`))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	_, err = cli.Namespace().List(context.Background(), &namespace.ListRequest{
		Page:     ptr[int32](1),
		PageSize: ptr[int32](10),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotURI, "page=1") || !strings.Contains(gotURI, "pageSize=10") {
		t.Errorf("query = %q, want page=1&pageSize=10", gotURI)
	}
}

// do()：POST 请求编码为 protojson body，零值字段省略。
func TestClient_POST_body(t *testing.T) {
	var body string
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"item":{"id":1},"exists":false}`))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	_, err = cli.Namespace().Create(context.Background(), &namespace.CreateRequest{
		Namespace:      "devops",
		IgnoreIfExists: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, `"namespace":"devops"`) || !strings.Contains(body, `"ignoreIfExists":true`) {
		t.Errorf("body = %s", body)
	}
}

// do()：gateway 错误体映射为 codes.Error，业务判断与 gRPC SDK 通用。
func TestClient_error_mapping(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":7,"message":"没有权限"}`))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	_, err = cli.Namespace().List(context.Background(), &namespace.ListRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	if status.Code(err) != codes.PermissionDenied {
		t.Errorf("code = %v, want PermissionDenied", status.Code(err))
	}
	if got := status.Convert(err).Message(); got != "没有权限" {
		t.Errorf("message = %q", got)
	}
}

// do()：401 且开启 autoRefresh 时自动重登并重试一次。
func TestClient_auto_refresh(t *testing.T) {
	var logins, infos int
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/login":
			logins++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"token":"t2"}`))
		case "/api/auth/info":
			infos++
			if strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ") != "t2" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"code":16,"message":"unauthenticated"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":1,"name":"admin"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	cli, err := NewHTTPClient(srv.URL, WithAuth("a", "b"), WithTokenAutoRefresh())
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	cli.setToken("t1") // 模拟旧 token 过期

	info, err := cli.Auth().Info(context.Background(), &auth.InfoRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "admin" {
		t.Errorf("name = %q", info.Name)
	}
	if logins != 2 {
		t.Errorf("logins = %d, want 2（构造登录 + 刷新）", logins)
	}
	if infos != 2 {
		t.Errorf("infos = %d, want 2（首次 + 重试）", infos)
	}
}

// 特殊路由：multipart 上传 POST /api/files，返回文件 ID。
func TestFileSvc_UploadFile(t *testing.T) {
	var contentType, reqBody string
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/files" || r.Method != http.MethodPost {
			t.Errorf("path/method = %s %s", r.Method, r.URL.Path)
		}
		contentType = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		reqBody = string(b)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":42}`))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	out, err := cli.File().UploadFile(context.Background(), "deploy.yaml", strings.NewReader("content"))
	if err != nil {
		t.Fatal(err)
	}
	if out.ID != 42 {
		t.Errorf("id = %d", out.ID)
	}
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		t.Errorf("content-type = %q", contentType)
	}
	if !strings.Contains(reqBody, `filename="deploy.yaml"`) || !strings.Contains(reqBody, "content") {
		t.Errorf("body = %q", reqBody)
	}
}

// 特殊路由：二进制下载 GET /api/download_file/{id}，解析 Content-Disposition。
func TestFileSvc_DownloadFile(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/download_file/7" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Disposition", `attachment; filename="a.txt"`)
		_, _ = w.Write([]byte("hello"))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	rc, info, err := cli.File().DownloadFile(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	if info.Filename != "a.txt" {
		t.Errorf("filename = %q", info.Filename)
	}
	data, _ := io.ReadAll(rc)
	if string(data) != "hello" {
		t.Errorf("data = %q", string(data))
	}
}

// 特殊路由：POST /api/copy_from_pod 从 pod 拷文件，返回二进制流。
func TestFileSvc_CopyFromPod(t *testing.T) {
	var gotBody string
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/copy_from_pod" || r.Method != http.MethodPost {
			t.Errorf("path/method = %s %s", r.Method, r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = w.Write([]byte("pod-file-content"))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	rc, _, err := cli.File().CopyFromPod(context.Background(), &CopyFromPodRequest{
		Namespace: "devops",
		Pod:       "pod-1",
		Container: "app",
		FilePath:  "/etc/config.yaml",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	if !strings.Contains(gotBody, `"namespace":"devops"`) || !strings.Contains(gotBody, `"filepath":"/etc/config.yaml"`) {
		t.Errorf("body = %q", gotBody)
	}
	data, _ := io.ReadAll(rc)
	if string(data) != "pod-file-content" {
		t.Errorf("data = %q", string(data))
	}
}

// 路径模板替换：{id} 正确落到 path 上，IsExists 走 query。
func TestClient_path_templates(t *testing.T) {
	paths := make(map[string]bool)
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		paths[r.Method+" "+r.URL.Path] = true
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/namespaces/5":
			_, _ = w.Write([]byte(`{}`))
		case r.URL.Path == "/api/namespaces/5/update_desc":
			_, _ = w.Write([]byte(`{}`))
		case r.URL.Path == "/api/record_files/9":
			_, _ = w.Write([]byte(`{}`))
		case r.URL.Path == "/api/namespaces/exists":
			_, _ = w.Write([]byte(`{}`))
		default:
			_, _ = w.Write([]byte(`{}`))
		}
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	_, _ = cli.Namespace().Show(context.Background(), &namespace.ShowRequest{Id: 5})
	_, _ = cli.Namespace().UpdateDesc(context.Background(), &namespace.UpdateDescRequest{Id: 5, Desc: "x"})
	_, _ = cli.File().ShowRecords(context.Background(), &file.ShowRecordsRequest{Id: 9})
	_, _ = cli.Namespace().IsExists(context.Background(), &namespace.IsExistsRequest{Name: "devops"})

	for _, want := range []string{"GET /api/namespaces/5", "POST /api/namespaces/5/update_desc", "GET /api/record_files/9", "POST /api/namespaces/exists"} {
		if !paths[want] {
			t.Errorf("missing request %q, got %v", want, paths)
		}
	}
}

// %2F 保留：分支名带斜杠时路径必须保留编码（对应 b109aa07 的修复）。
// 由生成器产出的 string 路径参数走 url.PathEscape，这是最易回归的地方。
func TestClient_path_percent_encoding(t *testing.T) {
	var escapedPath, uri string
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		escapedPath = r.URL.EscapedPath()
		uri = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	_, _ = cli.Git().CommitOptions(context.Background(), &git.CommitOptionsRequest{
		GitProjectId: 3,
		Branch:       "feature/x",
	})
	if !strings.Contains(escapedPath, "feature%2Fx") {
		t.Errorf("escaped path = %q, want contains feature%%2Fx", escapedPath)
	}
	if !strings.Contains(uri, "feature%2Fx") {
		t.Errorf("request uri = %q, want contains feature%%2Fx", uri)
	}
}

// encodeQuery：repeated 展开、零值跳过、特殊字符转义。
func Test_encodeQuery(t *testing.T) {
	// 空请求 → 空串
	if got := encodeQuery(&namespace.ListRequest{}); got != "" {
		t.Errorf("empty = %q", got)
	}
	// 零值跳过：page=0 不出现，只出 pageSize
	if got := encodeQuery(&namespace.ListRequest{PageSize: ptr[int32](10)}); got != "pageSize=10" {
		t.Errorf("zero-skip = %q", got)
	}
	// repeated 展开为 a=v1&a=v2，@ 被百分号转义（字段顺序不保证，做包含断言）
	got := encodeQuery(&namespace.SyncMembersRequest{
		Id:     3,
		Emails: []string{"a@x.com", "b@x.com"},
	})
	for _, want := range []string{"id=3", "emails=a%40x.com", "emails=b%40x.com"} {
		if !strings.Contains(got, want) {
			t.Errorf("repeated = %q, want contains %q", got, want)
		}
	}
}

// stream：grpc-gateway 输出 NDJSON（每行一个 JSON 对象，chunked），
// Recv 逐条解出、io.EOF 表示流正常结束。
func TestClient_StreamContainerLog_NDJSON(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/containers/namespaces/devops/pods/p-1/containers/app/stream_logs" {
			t.Errorf("path = %q", r.URL.Path)
		}
		fl, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("no flusher")
		}
		for i := 1; i <= 3; i++ {
			fmt.Fprintf(w, "{\"namespace\":\"devops\",\"podName\":\"p-1\",\"containerName\":\"app\",\"log\":\"line%d\"}\n", i)
			fl.Flush()
		}
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	stream, err := cli.Container().StreamContainerLog(context.Background(), &container.LogRequest{
		Namespace: "devops",
		Pod:       "p-1",
		Container: "app",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	var got []string
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, msg.Log)
	}
	if want := []string{"line1", "line2", "line3"}; !reflect.DeepEqual(got, want) {
		t.Errorf("logs = %v, want %v", got, want)
	}
}

// stream：标准 SSE（data: 块，空行分隔，多行 data 拼接），Recv 同样逐条解出。
func TestClient_StreamContainerLog_SSE(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fl, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("no flusher")
		}
		events := []string{
			"event: log\n: comment\nid: 1\ndata: {\"log\":\"e1\"}\n\n",
			"data: {\"log\":\"e2\"}\n\n",
			// 一个 JSON 拆成两行 data，Recv 按 \n 拼接后再解（逗号边界处 JSON 合法）。
			"data: {\"log\":\"e3\",\ndata: \"namespace\":\"dev\"}\n\n",
		}
		for _, e := range events {
			_, _ = w.Write([]byte(e))
			fl.Flush()
		}
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	stream, err := cli.Container().StreamContainerLog(context.Background(), &container.LogRequest{
		Namespace: "devops",
		Pod:       "p-1",
		Container: "app",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	var got []string
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, msg.Log)
	}
	if want := []string{"e1", "e2", "e3"}; !reflect.DeepEqual(got, want) {
		t.Errorf("logs = %v, want %v", got, want)
	}
}

// stream：grpc-gateway v2 把每个 server-streaming 消息包成 {"result": <msg>}（handler.go
// ForwardResponseStream），必须解包后才能解出业务消息。
func TestClient_StreamContainerLog_gateway_envelope(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fl, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("no flusher")
		}
		// 服务端真实输出形态：{"result":{...}} 每行一条，log 内含转义 \n。
		for _, line := range []string{
			`{"result":{"namespace":"devops","podName":"p-1","containerName":"app","log":"l1\n"}}`,
			`{"result":{"namespace":"devops","podName":"p-1","containerName":"app","log":"l2\n"}}`,
		} {
			_, _ = w.Write([]byte(line + "\n"))
			fl.Flush()
		}
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	stream, err := cli.Container().StreamContainerLog(context.Background(), &container.LogRequest{
		Namespace: "devops",
		Pod:       "p-1",
		Container: "app",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	var got []string
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, msg.Log)
	}
	if want := []string{"l1\n", "l2\n"}; !reflect.DeepEqual(got, want) {
		t.Errorf("logs = %q, want %q", got, want)
	}
}

// stream：流中途错误以 {"error": <google.rpc.Status>} 内联在 body，还原成 codes.Error。
func TestClient_StreamContainerLog_stream_error(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		fl, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("no flusher")
		}
		_, _ = w.Write([]byte("{\"result\":{\"namespace\":\"devops\",\"podName\":\"p-1\",\"containerName\":\"app\",\"log\":\"ok\\n\"}}\n"))
		fl.Flush()
		_, _ = w.Write([]byte("{\"error\":{\"code\":13,\"message\":\"容器退出\"}}\n"))
		fl.Flush()
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	stream, err := cli.Container().StreamContainerLog(context.Background(), &container.LogRequest{})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	first, err := stream.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if first.Log != "ok\n" {
		t.Errorf("first log = %q", first.Log)
	}
	if _, err = stream.Recv(); err == nil {
		t.Fatal("expected stream error")
	} else if status.Code(err) != codes.Internal {
		t.Errorf("code = %v, want Internal", status.Code(err))
	} else if got := status.Convert(err).Message(); got != "容器退出" {
		t.Errorf("message = %q", got)
	}
}

// stream：非 2xx 映射为 gateway 错误体对应的 codes.Error，与 unary 一致。
func TestClient_StreamContainerLog_error(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code":13,"message":"内部错误"}`))
	})
	cli, err := NewHTTPClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	_, err = cli.Container().StreamContainerLog(context.Background(), &container.LogRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	if status.Code(err) != codes.Internal {
		t.Errorf("code = %v, want Internal", status.Code(err))
	}
	if got := status.Convert(err).Message(); got != "内部错误" {
		t.Errorf("message = %q", got)
	}
}

// --- Option 错误分支覆盖（WithTokenAutoRefresh / refreshToken / doReq 的失败路径） ---

// refreshToken 无凭据时直接报错（WithTokenAutoRefresh 未配 WithAuth 时不会进入刷新）。
func TestRefreshToken_NoCredentials(t *testing.T) {
	c := &Client{}
	if err := c.refreshToken(); err == nil {
		t.Fatal("无凭据 refreshToken 应返回错误")
	}
}

// refreshToken 登录失败：flight fn 内 Login 返回错误应透出。
func TestRefreshToken_LoginFailure(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":16,"message":"bad credentials"}`))
	})
	c := &Client{baseURL: srv.URL, username: "admin", password: "bad", flights: &flight.Group{}}
	c.hc = &http.Client{}
	if err := c.refreshToken(); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("refreshToken err = %v, want Unauthenticated", err)
	}
}

// 业务请求 401 → 自动刷新但重登失败 → 返回刷新错误，不吞错。
func TestDo_RefreshLoginFails_ReturnsError(t *testing.T) {
	var logins int
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/login" {
			logins++
			if logins > 1 { // 构造期成功，刷新阶段失败
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"code":16,"message":"bad creds"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"token":"t1"}`))
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":16,"message":"unauthenticated"}`))
	})
	cli, err := NewHTTPClient(srv.URL, WithAuth("a", "b"), WithTokenAutoRefresh())
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	cli.setToken("stale") // 模拟过期 token
	if _, err := cli.Auth().Info(context.Background(), &auth.InfoRequest{}); err == nil {
		t.Fatal("刷新失败时业务请求应返回错误")
	}
	if logins != 2 {
		t.Errorf("logins = %d, want 2（构造 + 一次刷新失败）", logins)
	}
}

// 底层 hc.Do 网络错误（连接拒绝）应原样返回。
func TestDo_NetworkError(t *testing.T) {
	cli, err := NewHTTPClient("http://127.0.0.1:1", WithBearerToken("t")) // 端口 1 无服务
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	if _, err := cli.Auth().Info(context.Background(), &auth.InfoRequest{}); err == nil {
		t.Fatal("连接拒绝应返回错误")
	}
}

// 非 2xx 且非 gateway 错误体 → unexpected status 兜底。
func TestDo_UnexpectedStatus(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom")) // 不是 {"code":..} 错误体
	})
	cli, err := NewHTTPClient(srv.URL, WithBearerToken("t"))
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	_, err = cli.Auth().Info(context.Background(), &auth.InfoRequest{})
	if err == nil || !strings.Contains(err.Error(), "unexpected status 500") {
		t.Fatalf("err = %v, want unexpected status 500", err)
	}
}

// baseURL 含非法字符 → NewRequestWithContext 失败，返回错误而非 panic。
func TestDo_InvalidURL(t *testing.T) {
	c := &Client{baseURL: "http://exa mple.com"} // 空格使 URL 解析失败
	if err := c.do(context.Background(), "GET", "/api/x", nil, nil); err == nil {
		t.Fatal("非法 baseURL 应返回错误")
	}
}

// 构造期 WithAuth 登录失败：NewHTTPClient 应返回 error，而非静默继续。
func TestNewHTTPClient_LoginFails_ReturnsError(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":16,"message":"bad credentials"}`))
	})
	cli, err := NewHTTPClient(srv.URL, WithAuth("admin", "bad"))
	if err == nil {
		t.Fatal("构造期登录失败 NewHTTPClient 应返回 error")
	}
	if cli != nil {
		_ = cli.Close()
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("err = %v, want Unauthenticated", err)
	}
}
