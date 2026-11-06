package aichat

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStreamPlainText 验证请求契约（模型、stream、鉴权头、消息）与文本增量回吐。
func TestStreamPlainText(t *testing.T) {
	var received llmRequest
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "text/event-stream", r.Header.Get("Accept"))
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&received))

		writeSSEJSON(t, w, contentChunk("你"))
		writeSSEJSON(t, w, contentChunk("好"))
		writeSSEDone(t, w)
	}))
	defer upstream.Close()

	client := newLLMClient(testConfig(upstream.URL).AI, testLogger())
	var deltas []string
	text, calls, err := client.stream(context.Background(),
		[]chatMessage{{Role: roleUser, Content: "hi"}},
		[]toolDefinition{{Type: "function", Function: toolDeclaration{Name: "list_projects"}}},
		func(delta string) error {
			deltas = append(deltas, delta)
			return nil
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "你好", text)
	assert.Equal(t, []string{"你", "好"}, deltas)
	assert.Empty(t, calls)
	assert.Equal(t, "test-model", received.Model)
	assert.True(t, received.Stream)
	assert.Len(t, received.Tools, 1)
	assert.Equal(t, "hi", received.Messages[0].Content)
}

// TestStreamToolCallsAggregate 验证跨帧的工具调用分片按 index 聚合。
func TestStreamToolCallsAggregate(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeSSEJSON(t, w, toolChunk(toolDelta(0, "call_1", "list_projects", `{"namespace":`)))
		writeSSEJSON(t, w, toolChunk(toolDelta(0, "", "", `"ns"}`), toolDelta(1, "call_2", "get_project", `{}`)))
		writeSSEJSON(t, w, toolChunk(toolDelta(0, "", "", ` `)))
		writeSSEDone(t, w)
	}))
	defer upstream.Close()

	client := newLLMClient(testConfig(upstream.URL).AI, testLogger())
	text, calls, err := client.stream(context.Background(),
		[]chatMessage{{Role: roleUser, Content: "hi"}}, nil, func(string) error { return nil })
	require.NoError(t, err)
	assert.Empty(t, text)
	require.Len(t, calls, 2)
	assert.Equal(t, "call_1", calls[0].ID)
	assert.Equal(t, "list_projects", calls[0].Function.Name)
	assert.JSONEq(t, `{"namespace":"ns"} `, calls[0].Function.Arguments)
	assert.Equal(t, "call_2", calls[1].ID)
	assert.Equal(t, "get_project", calls[1].Function.Name)
	assert.Equal(t, "{}", calls[1].Function.Arguments)
}

// TestStreamHTTPError 验证上游非 2xx 返回错误（响应体不回传给调用方）。
func TestStreamHTTPError(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "upstream secret detail")
	}))
	defer upstream.Close()

	client := newLLMClient(testConfig(upstream.URL).AI, testLogger())
	_, _, err := client.stream(context.Background(),
		[]chatMessage{{Role: roleUser, Content: "hi"}}, nil, func(string) error { return nil })
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "upstream secret detail")
	assert.Contains(t, err.Error(), "500")
}

// TestStreamMalformedChunk 验证无法解析的 SSE 帧返回错误。
func TestStreamMalformedChunk(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "data: not-json\n\n")
	}))
	defer upstream.Close()

	client := newLLMClient(testConfig(upstream.URL).AI, testLogger())
	_, _, err := client.stream(context.Background(),
		[]chatMessage{{Role: roleUser, Content: "hi"}}, nil, func(string) error { return nil })
	require.Error(t, err)
	assert.Contains(t, err.Error(), "解析上游流式响应")
}

// TestStreamDeltaError 验证 onDelta 的错误向上传播（客户端断开时停止读取）。
func TestStreamDeltaError(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeSSEJSON(t, w, contentChunk("x"))
	}))
	defer upstream.Close()

	boom := errors.New("delta boom")
	client := newLLMClient(testConfig(upstream.URL).AI, testLogger())
	_, _, err := client.stream(context.Background(),
		[]chatMessage{{Role: roleUser, Content: "hi"}}, nil, func(string) error { return boom })
	assert.ErrorIs(t, err, boom)
}

// TestParseStreamScanError 验证底层读取失败被包装返回。
func TestParseStreamScanError(t *testing.T) {
	boom := errors.New("read boom")
	_, _, err := parseStream(iotest.ErrReader(boom), func(string) error { return nil })
	require.Error(t, err)
	assert.Contains(t, err.Error(), "读取上游流式响应")
}

// TestStreamRequestErrors 覆盖请求构造与发送失败两条分支。
func TestStreamRequestErrors(t *testing.T) {
	t.Run("非法地址", func(t *testing.T) {
		client := newLLMClient(testConfig("://bad").AI, testLogger())
		_, _, err := client.stream(context.Background(),
			[]chatMessage{{Role: roleUser, Content: "hi"}}, nil, func(string) error { return nil })
		require.Error(t, err)
		assert.Contains(t, err.Error(), "构造上游请求")
	})

	t.Run("上游不可达", func(t *testing.T) {
		upstream := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		url := upstream.URL
		upstream.Close()

		client := newLLMClient(testConfig(url).AI, testLogger())
		_, _, err := client.stream(context.Background(),
			[]chatMessage{{Role: roleUser, Content: "hi"}}, nil, func(string) error { return nil })
		require.Error(t, err)
		assert.Contains(t, err.Error(), "请求上游模型服务")
	})
}

// TestParseStreamIgnoresNonDataLines 验证注释行与空 data 行被忽略。
func TestParseStreamIgnoresNonDataLines(t *testing.T) {
	body := strings.NewReader(": ping\n\ndata: \n\n" +
		`data: {"choices":[{"delta":{"content":"ok"}}]}` + "\n\ndata: [DONE]\n\n")
	var deltas []string
	text, calls, err := parseStream(body, func(delta string) error {
		deltas = append(deltas, delta)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, "ok", text)
	assert.Equal(t, []string{"ok"}, deltas)
	assert.Empty(t, calls)
}

// TestParseStreamWithoutDone 验证上游未发送 [DONE] 就断流时按已收内容正常返回。
func TestParseStreamWithoutDone(t *testing.T) {
	body := strings.NewReader(`data: {"choices":[{"delta":{"content":"半"}}]}` + "\n\n")
	text, calls, err := parseStream(body, func(string) error { return nil })
	require.NoError(t, err)
	assert.Equal(t, "半", text)
	assert.Empty(t, calls)
}

// TestParseStreamCRLF 验证 Windows 风格行尾（\r\n）不影响解析。
func TestParseStreamCRLF(t *testing.T) {
	body := strings.NewReader("data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\r\n\r\ndata: [DONE]\r\n")
	text, calls, err := parseStream(body, func(string) error { return nil })
	require.NoError(t, err)
	assert.Equal(t, "ok", text)
	assert.Empty(t, calls)
}

// TestParseStreamLastLineWithoutNewline 验证末尾没有换行符的数据行仍被解析。
func TestParseStreamLastLineWithoutNewline(t *testing.T) {
	body := strings.NewReader(`data: {"choices":[{"delta":{"content":"尾"}}]}`)
	text, _, err := parseStream(body, func(string) error { return nil })
	require.NoError(t, err)
	assert.Equal(t, "尾", text)
}

// TestParseStreamLongLine 验证超长单行不被截断（原 4 MiB 行上限已移除）。
func TestParseStreamLongLine(t *testing.T) {
	long := strings.Repeat("x", 5<<20)
	body := strings.NewReader(`data: {"choices":[{"delta":{"content":"` + long + `"}}]}` + "\n")
	var received int
	text, _, err := parseStream(body, func(delta string) error {
		received += len(delta)
		return nil
	})
	require.NoError(t, err)
	assert.Len(t, text, 5<<20)
	assert.Equal(t, 5<<20, received)
}

// TestJoinURL 验证 base 尾部斜杠与路径前导斜杠都被归一化。
func TestJoinURL(t *testing.T) {
	assert.Equal(t, "https://api.example.com/v1/chat/completions",
		joinURL("https://api.example.com/v1/", "/chat/completions"))
	assert.Equal(t, "https://api.example.com/chat/completions",
		joinURL("https://api.example.com", "chat/completions"))
}
