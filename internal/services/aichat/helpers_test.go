package aichat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testLogger 返回一个不写文件的测试日志器。
func testLogger() mlog.Logger {
	return mlog.NewForConfig(nil)
}

// testConfig 构造一份可用的 AI 配置；baseURL 传空串即得到「未启用」状态。
func testConfig(baseURL string) *config.Config {
	return &config.Config{AI: config.AIConfig{
		BaseURL: baseURL,
		APIKey:  "test-api-key",
		Model:   "test-model",
	}}
}

// testUser 构造一个带 email 的登录用户。
func testUser(email string) *biz.UserInfo {
	return &biz.UserInfo{Email: email}
}

// withUser 把用户注入请求 ctx（跳过 HTTP 鉴权中间件，直接驱动 handler）。
func withUser(req *http.Request, user *biz.UserInfo) *http.Request {
	return req.WithContext(biz.SetUser(req.Context(), user))
}

// fakeToolSession 是 toolSession 的内存替身。
type fakeToolSession struct {
	specs   []toolSpec
	listErr error
	callFn  func(ctx context.Context, name string, args json.RawMessage) (toolResult, error)
	calls   []string
	args    []json.RawMessage
}

// List 返回预设的工具清单或错误。
func (f *fakeToolSession) List(context.Context) ([]toolSpec, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.specs, nil
}

// Call 记录调用并委派给 callFn；callFn 为空时返回空成功结果。
func (f *fakeToolSession) Call(ctx context.Context, name string, args json.RawMessage) (toolResult, error) {
	f.calls = append(f.calls, name)
	f.args = append(f.args, args)
	if f.callFn == nil {
		return toolResult{Text: "ok"}, nil
	}
	return f.callFn(ctx, name, args)
}

// fakeOpener 是 toolSessionOpener 的测试替身。
type fakeOpener struct {
	session toolSession
	err     error
}

// open 返回预设会话或错误；会话无外部资源，cleanup 为空实现。
func (o *fakeOpener) open(context.Context) (toolSession, func(), error) {
	if o.err != nil {
		return nil, nil, o.err
	}
	return o.session, func() {}, nil
}

// sseFrame 是解析后的 SSE 帧（忽略注释帧）。
type sseFrame struct {
	Event string
	Data  string
}

// parseFrames 把响应的 SSE 文本拆成帧列表。
func parseFrames(t *testing.T, body string) []sseFrame {
	t.Helper()
	frames := make([]sseFrame, 0)
	for _, block := range strings.Split(body, "\n\n") {
		block = strings.TrimSpace(block)
		if block == "" || strings.HasPrefix(block, ":") {
			continue
		}
		var frame sseFrame
		for _, line := range strings.Split(block, "\n") {
			switch {
			case strings.HasPrefix(line, "event: "):
				frame.Event = strings.TrimPrefix(line, "event: ")
			case strings.HasPrefix(line, "data: "):
				frame.Data = strings.TrimPrefix(line, "data: ")
			}
		}
		frames = append(frames, frame)
	}
	return frames
}

// eventsOf 返回指定事件名的全部帧。
func eventsOf(frames []sseFrame, name string) []sseFrame {
	matched := make([]sseFrame, 0)
	for _, frame := range frames {
		if frame.Event == name {
			matched = append(matched, frame)
		}
	}
	return matched
}

// decodeData 把帧的 data 字段反序列化到 dst。
func decodeData(t *testing.T, frame sseFrame, dst any) {
	t.Helper()
	require.NoError(t, json.Unmarshal([]byte(frame.Data), dst))
}

// writeSSEJSON 向上游响应写一帧 SSE 数据；在 httptest handler 协程中调用，
// 因此只能用 assert 记录失败，不能触发 FailNow。
func writeSSEJSON(t *testing.T, w http.ResponseWriter, payload any) {
	t.Helper()
	raw, err := json.Marshal(payload)
	assert.NoError(t, err)
	_, err = fmt.Fprintf(w, "data: %s\n\n", raw)
	assert.NoError(t, err)
}

// writeSSEDone 向上游响应写 [DONE] 结束标记（同 writeSSEJSON，只允许 assert）。
func writeSSEDone(t *testing.T, w http.ResponseWriter) {
	t.Helper()
	_, err := io.WriteString(w, "data: [DONE]\n\n")
	assert.NoError(t, err)
}

// contentChunk 构造只含文本增量的上游帧。
func contentChunk(text string) map[string]any {
	return map[string]any{
		"choices": []any{map[string]any{"delta": map[string]any{"content": text}}},
	}
}

// toolDelta 构造一个工具调用增量分片。
func toolDelta(index int, id, name, args string) map[string]any {
	return map[string]any{
		"index": index,
		"id":    id,
		"type":  "function",
		"function": map[string]any{
			"name":      name,
			"arguments": args,
		},
	}
}

// toolChunk 构造携带工具调用分片的上游帧。
func toolChunk(deltas ...map[string]any) map[string]any {
	items := make([]any, 0, len(deltas))
	for _, delta := range deltas {
		items = append(items, delta)
	}
	return map[string]any{
		"choices": []any{map[string]any{"delta": map[string]any{"tool_calls": items}}},
	}
}

// noFlushWriter 是不实现 http.Flusher 的响应写入器，用于覆盖 SSE 初始化失败分支。
type noFlushWriter struct {
	header http.Header
	code   int
}

// Header 返回响应头。
func (w *noFlushWriter) Header() http.Header {
	if w.header == nil {
		w.header = http.Header{}
	}
	return w.header
}

// Write 丢弃响应体。
func (w *noFlushWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// WriteHeader 记录状态码。
func (w *noFlushWriter) WriteHeader(code int) {
	w.code = code
}

// errorFlusherWriter 是「写响应体必然失败」的写入器，用于覆盖 SSE 写失败分支。
type errorFlusherWriter struct {
	header http.Header
	err    error
}

// Header 返回响应头。
func (w *errorFlusherWriter) Header() http.Header {
	if w.header == nil {
		w.header = http.Header{}
	}
	return w.header
}

// Write 恒返回注入的错误。
func (w *errorFlusherWriter) Write([]byte) (int, error) {
	return 0, w.err
}

// WriteHeader 空实现。
func (w *errorFlusherWriter) WriteHeader(int) {}

// Flush 空实现（满足 http.Flusher）。
func (w *errorFlusherWriter) Flush() {}

// newRecorder 返回一个带 Flusher 的响应记录器。
func newRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// failAfterWriter 允许前 remaining 次写入，之后恒失败；用于覆盖 SSE 写失败分支。
type failAfterWriter struct {
	*httptest.ResponseRecorder
	remaining int
}

// Write 在额度耗尽后返回错误。
func (w *failAfterWriter) Write(p []byte) (int, error) {
	if w.remaining <= 0 {
		return 0, errors.New("write boom")
	}
	w.remaining--
	return w.ResponseRecorder.Write(p)
}

// WriteString 复用 Write 的额度控制：io.WriteString 会优先走 io.StringWriter，
// 底层的 ResponseRecorder.WriteString 会绕过 Write 的拦截。
func (w *failAfterWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}
