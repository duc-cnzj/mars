package aichat

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// scriptedUpstream 是按调用序号回放 SSE 的上游模型替身，并记录每次请求体。
type scriptedUpstream struct {
	mu       sync.Mutex
	calls    int
	requests []llmRequest
	handler  func(call int, w http.ResponseWriter, request llmRequest)
}

// ServeHTTP 记录请求并按序号委派给测试脚本。
func (s *scriptedUpstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := llmRequest{}
	_ = json.NewDecoder(r.Body).Decode(&request)

	s.mu.Lock()
	s.calls++
	call := s.calls
	s.requests = append(s.requests, request)
	s.mu.Unlock()

	s.handler(call, w, request)
}

// CallCount 返回已处理的请求数。
func (s *scriptedUpstream) CallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

// Requests 返回已记录的请求体副本。
func (s *scriptedUpstream) Requests() []llmRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]llmRequest(nil), s.requests...)
}

// testSpecs 返回一组只读 + 写工具定义。
func testSpecs() []toolSpec {
	return []toolSpec{
		{Name: "list_projects", Description: "列出项目", InputSchema: json.RawMessage(`{"type":"object"}`), ReadOnly: true},
		{Name: "get_project", Description: "项目详情", InputSchema: json.RawMessage(`{"type":"object"}`), ReadOnly: true},
		{Name: "restart_pod", Description: "重启 Pod", InputSchema: json.RawMessage(`{"type":"object"}`)},
		{Name: "exec_command", Description: "容器命令", InputSchema: json.RawMessage(`{"type":"object"}`)},
	}
}

// newTestLoop 构造一套带响应记录器的对话循环。
func newTestLoop(t *testing.T, upstreamURL string, session toolSession) (*chatLoop, *httptest.ResponseRecorder, *ticketSigner) {
	t.Helper()
	recorder := newRecorder()
	sse, err := newSSEWriter(recorder)
	require.NoError(t, err)
	tickets := newTicketSigner("test-key")
	client := newLLMClient(testConfig(upstreamURL).AI, testLogger())
	loop := newChatLoop(
		client,
		&toolbox{session: session, skills: mustLoadEmbeddedSkills()},
		sse,
		tickets,
		testLogger(),
		"u@example.com",
		"system prompt",
		testSpecs(),
	)
	return loop, recorder, tickets
}

// newScriptedServer 启动一个按脚本回放的上游服务。
func newScriptedServer(t *testing.T, handler func(call int, w http.ResponseWriter, request llmRequest)) (*scriptedUpstream, *httptest.Server) {
	t.Helper()
	upstream := &scriptedUpstream{handler: handler}
	server := httptest.NewServer(upstream)
	t.Cleanup(server.Close)
	return upstream, server
}

// TestChatLoopPlainText 验证纯文本对话：meta → delta → done，且不调用工具。
func TestChatLoopPlainText(t *testing.T) {
	upstream, server := newScriptedServer(t, func(_ int, w http.ResponseWriter, _ llmRequest) {
		writeSSEJSON(t, w, contentChunk("你好"))
		writeSSEDone(t, w)
	})
	session := &fakeToolSession{}
	loop, recorder, _ := newTestLoop(t, server.URL, session)

	loop.run(context.Background(), []chatMessage{{Role: roleUser, Content: "hi"}})

	frames := parseFrames(t, recorder.Body.String())
	assert.Len(t, eventsOf(frames, eventMeta), 1)
	require.Len(t, eventsOf(frames, eventDelta), 1)
	var delta deltaEvent
	decodeData(t, eventsOf(frames, eventDelta)[0], &delta)
	assert.Equal(t, "你好", delta.Text)

	var done doneEvent
	require.Len(t, eventsOf(frames, eventDone), 1)
	decodeData(t, eventsOf(frames, eventDone)[0], &done)
	assert.Equal(t, 1, done.Rounds)

	assert.Empty(t, session.calls)
	assert.Equal(t, 1, upstream.CallCount())
	// system prompt 必须拼在对话最前面。
	require.Len(t, upstream.Requests(), 1)
	assert.Equal(t, roleSystem, upstream.Requests()[0].Messages[0].Role)
}

// TestChatLoopReadOnlyTool 验证只读工具直接执行，并把结果回填后再请求模型。
func TestChatLoopReadOnlyTool(t *testing.T) {
	upstream, server := newScriptedServer(t, func(call int, w http.ResponseWriter, _ llmRequest) {
		if call == 1 {
			writeSSEJSON(t, w, toolChunk(toolDelta(0, "call_1", "list_projects", `{"namespace":"ns"}`)))
			writeSSEDone(t, w)
			return
		}
		writeSSEJSON(t, w, contentChunk("完成"))
		writeSSEDone(t, w)
	})
	session := &fakeToolSession{callFn: func(context.Context, string, json.RawMessage) (toolResult, error) {
		return toolResult{Text: "projects"}, nil
	}}
	loop, recorder, _ := newTestLoop(t, server.URL, session)

	loop.run(context.Background(), []chatMessage{{Role: roleUser, Content: "列项目"}})

	frames := parseFrames(t, recorder.Body.String())
	assert.Len(t, eventsOf(frames, eventToolCall), 1)
	var callEvent toolCallEvent
	decodeData(t, eventsOf(frames, eventToolCall)[0], &callEvent)
	assert.True(t, callEvent.ReadOnly)
	assert.Equal(t, "list_projects", callEvent.Name)

	var resultEvent toolResultEvent
	require.Len(t, eventsOf(frames, eventToolResult), 1)
	decodeData(t, eventsOf(frames, eventToolResult)[0], &resultEvent)
	assert.True(t, resultEvent.OK)
	assert.Equal(t, "projects", resultEvent.Content)
	assert.Len(t, eventsOf(frames, eventDone), 1)

	assert.Equal(t, []string{"list_projects"}, session.calls)
	// 第二次请求应带上 assistant tool_calls 与 tool 结果。
	requests := upstream.Requests()
	require.Len(t, requests, 2)
	tail := requests[1].Messages[len(requests[1].Messages)-2:]
	assert.Equal(t, roleAssistant, tail[0].Role)
	assert.Len(t, tail[0].ToolCalls, 1)
	assert.Equal(t, roleTool, tail[1].Role)
	assert.Equal(t, "call_1", tail[1].ToolCallID)
}

// TestChatLoopWriteToolPauses 验证遇到写工具立即暂停：不下发 done、不执行工具、
// 票据绑定调用与参数。
func TestChatLoopWriteToolPauses(t *testing.T) {
	upstream, server := newScriptedServer(t, func(_ int, w http.ResponseWriter, _ llmRequest) {
		writeSSEJSON(t, w, toolChunk(toolDelta(0, "call_1", "restart_pod", `{"namespace":"ns","pod":"p"}`)))
		writeSSEDone(t, w)
	})
	session := &fakeToolSession{}
	loop, recorder, tickets := newTestLoop(t, server.URL, session)

	loop.run(context.Background(), []chatMessage{{Role: roleUser, Content: "重启"}})

	frames := parseFrames(t, recorder.Body.String())
	assert.Empty(t, eventsOf(frames, eventDone))
	require.Len(t, eventsOf(frames, eventNeedConfirm), 1)
	var confirm needConfirmEvent
	decodeData(t, eventsOf(frames, eventNeedConfirm)[0], &confirm)
	assert.Equal(t, "call_1", confirm.ID)
	assert.Equal(t, "restart_pod", confirm.Name)
	assert.Contains(t, confirm.Summary, "restart_pod")
	assert.Len(t, confirm.Messages, 2)

	payload, err := tickets.verify(confirm.Ticket)
	require.NoError(t, err)
	assert.Equal(t, "u@example.com", payload.Email)
	assert.Equal(t, "restart_pod", payload.ToolName)

	assert.Empty(t, session.calls)
	assert.Equal(t, 1, upstream.CallCount())
}

// resumeSetup 构造一次待确认的写工具对话。
func resumeSetup(t *testing.T, tickets *ticketSigner, name, args string) ([]chatMessage, []toolCall, ticketPayload) {
	t.Helper()
	call := testCall("call_1", name, args)
	ticket := tickets.issue("u@example.com", call)
	payload, err := tickets.verify(ticket)
	require.NoError(t, err)
	conversation := []chatMessage{
		{Role: roleUser, Content: "执行"},
		{Role: roleAssistant, ToolCalls: []toolCall{call}},
	}
	return conversation, []toolCall{call}, payload
}

// TestChatLoopResumeApprove 验证批准后执行写工具并继续请求模型。
func TestChatLoopResumeApprove(t *testing.T) {
	_, server := newScriptedServer(t, func(_ int, w http.ResponseWriter, _ llmRequest) {
		writeSSEJSON(t, w, contentChunk("已重启"))
		writeSSEDone(t, w)
	})
	session := &fakeToolSession{}
	loop, recorder, tickets := newTestLoop(t, server.URL, session)
	conversation, calls, payload := resumeSetup(t, tickets, "restart_pod", `{"namespace":"ns","pod":"p"}`)

	loop.resume(context.Background(), conversation, calls, 0, &confirmation{payload: payload, approved: true})

	frames := parseFrames(t, recorder.Body.String())
	require.Len(t, eventsOf(frames, eventToolCall), 1)
	var callEvent toolCallEvent
	decodeData(t, eventsOf(frames, eventToolCall)[0], &callEvent)
	assert.False(t, callEvent.ReadOnly)
	require.Len(t, eventsOf(frames, eventToolResult), 1)
	assert.Len(t, eventsOf(frames, eventDone), 1)
	assert.Equal(t, []string{"restart_pod"}, session.calls)
}

// TestChatLoopResumeReject 验证拒绝后不执行工具，只是把拒绝结果回填给模型。
func TestChatLoopResumeReject(t *testing.T) {
	_, server := newScriptedServer(t, func(_ int, w http.ResponseWriter, _ llmRequest) {
		writeSSEJSON(t, w, contentChunk("已取消"))
		writeSSEDone(t, w)
	})
	session := &fakeToolSession{}
	loop, recorder, tickets := newTestLoop(t, server.URL, session)
	conversation, calls, payload := resumeSetup(t, tickets, "restart_pod", `{}`)

	loop.resume(context.Background(), conversation, calls, 0, &confirmation{payload: payload, approved: false})

	frames := parseFrames(t, recorder.Body.String())
	require.Len(t, eventsOf(frames, eventToolResult), 1)
	var resultEvent toolResultEvent
	decodeData(t, eventsOf(frames, eventToolResult)[0], &resultEvent)
	assert.False(t, resultEvent.OK)
	assert.Contains(t, resultEvent.Content, "拒绝")
	assert.Len(t, eventsOf(frames, eventDone), 1)
	assert.Empty(t, session.calls)
}

// TestChatLoopSecondWritePauses 验证批准第一个写工具后，同批第二个写工具会重新确认。
func TestChatLoopSecondWritePauses(t *testing.T) {
	upstream, server := newScriptedServer(t, func(_ int, w http.ResponseWriter, _ llmRequest) {
		t.Error("不应该继续请求上游")
	})
	session := &fakeToolSession{}
	loop, recorder, tickets := newTestLoop(t, server.URL, session)

	first := testCall("call_1", "restart_pod", `{}`)
	second := testCall("call_2", "exec_command", `{"command":["ls"]}`)
	ticket := tickets.issue("u@example.com", first)
	payload, err := tickets.verify(ticket)
	require.NoError(t, err)
	conversation := []chatMessage{
		{Role: roleUser, Content: "执行"},
		{Role: roleAssistant, ToolCalls: []toolCall{first, second}},
	}

	loop.resume(context.Background(), conversation, []toolCall{first, second}, 0,
		&confirmation{payload: payload, approved: true})

	frames := parseFrames(t, recorder.Body.String())
	require.Len(t, eventsOf(frames, eventNeedConfirm), 1)
	var confirm needConfirmEvent
	decodeData(t, eventsOf(frames, eventNeedConfirm)[0], &confirm)
	assert.Equal(t, "call_2", confirm.ID)
	assert.Equal(t, []string{"restart_pod"}, session.calls)
	assert.Equal(t, 0, upstream.CallCount())
}

// TestChatLoopMaxRounds 验证达到轮数上限后如实报错，不编造完成。
func TestChatLoopMaxRounds(t *testing.T) {
	upstream, server := newScriptedServer(t, func(_ int, w http.ResponseWriter, _ llmRequest) {
		writeSSEJSON(t, w, toolChunk(toolDelta(0, "call_1", "list_projects", `{}`)))
		writeSSEDone(t, w)
	})
	session := &fakeToolSession{}
	loop, recorder, _ := newTestLoop(t, server.URL, session)

	loop.run(context.Background(), []chatMessage{{Role: roleUser, Content: "循环"}})

	frames := parseFrames(t, recorder.Body.String())
	assert.Empty(t, eventsOf(frames, eventDone))
	require.Len(t, eventsOf(frames, eventError), 1)
	var errorEvent errorEvent
	decodeData(t, eventsOf(frames, eventError)[0], &errorEvent)
	assert.Contains(t, errorEvent.Message, "上限")
	assert.Equal(t, maxToolRounds, upstream.CallCount())
	assert.Len(t, session.calls, maxToolRounds)
}

// TestChatLoopUpstreamError 验证上游失败只回吐通用错误，不泄漏上游响应体。
func TestChatLoopUpstreamError(t *testing.T) {
	_, server := newScriptedServer(t, func(_ int, w http.ResponseWriter, _ llmRequest) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "upstream secret detail")
	})
	loop, recorder, _ := newTestLoop(t, server.URL, &fakeToolSession{})

	loop.run(context.Background(), []chatMessage{{Role: roleUser, Content: "hi"}})

	frames := parseFrames(t, recorder.Body.String())
	require.Len(t, eventsOf(frames, eventError), 1)
	var errorEvent errorEvent
	decodeData(t, eventsOf(frames, eventError)[0], &errorEvent)
	assert.Equal(t, upstreamErrorMessage, errorEvent.Message)
	assert.NotContains(t, recorder.Body.String(), "upstream secret detail")
}

// TestChatLoopMalformedUpstream 验证畸形 SSE 被收敛为通用错误。
func TestChatLoopMalformedUpstream(t *testing.T) {
	_, server := newScriptedServer(t, func(_ int, w http.ResponseWriter, _ llmRequest) {
		_, _ = io.WriteString(w, "data: not-json\n\n")
	})
	loop, recorder, _ := newTestLoop(t, server.URL, &fakeToolSession{})

	loop.run(context.Background(), []chatMessage{{Role: roleUser, Content: "hi"}})

	frames := parseFrames(t, recorder.Body.String())
	require.Len(t, eventsOf(frames, eventError), 1)
}

// TestChatLoopToolProtocolError 验证工具协议错误收敛为 isError 结果，对话继续。
func TestChatLoopToolProtocolError(t *testing.T) {
	upstream, server := newScriptedServer(t, func(call int, w http.ResponseWriter, _ llmRequest) {
		if call == 1 {
			writeSSEJSON(t, w, toolChunk(toolDelta(0, "call_1", "list_projects", `{}`)))
			writeSSEDone(t, w)
			return
		}
		writeSSEJSON(t, w, contentChunk("继续"))
		writeSSEDone(t, w)
	})
	session := &fakeToolSession{callFn: func(context.Context, string, json.RawMessage) (toolResult, error) {
		return toolResult{}, errors.New("boom")
	}}
	loop, recorder, _ := newTestLoop(t, server.URL, session)

	loop.run(context.Background(), []chatMessage{{Role: roleUser, Content: "hi"}})

	frames := parseFrames(t, recorder.Body.String())
	require.Len(t, eventsOf(frames, eventToolResult), 1)
	var resultEvent toolResultEvent
	decodeData(t, eventsOf(frames, eventToolResult)[0], &resultEvent)
	assert.False(t, resultEvent.OK)
	assert.Contains(t, resultEvent.Content, "服务端内部错误")
	assert.Len(t, eventsOf(frames, eventDone), 1)
	assert.Equal(t, 2, upstream.CallCount())
}

// TestChatLoopUnknownTool 验证模型调用不存在的工具时回吐业务错误。
func TestChatLoopUnknownTool(t *testing.T) {
	upstream, server := newScriptedServer(t, func(call int, w http.ResponseWriter, _ llmRequest) {
		if call == 1 {
			writeSSEJSON(t, w, toolChunk(toolDelta(0, "call_x", "no_such_tool", `{}`)))
			writeSSEDone(t, w)
			return
		}
		writeSSEJSON(t, w, contentChunk("ok"))
		writeSSEDone(t, w)
	})
	session := &fakeToolSession{}
	loop, recorder, _ := newTestLoop(t, server.URL, session)

	loop.run(context.Background(), []chatMessage{{Role: roleUser, Content: "hi"}})

	frames := parseFrames(t, recorder.Body.String())
	require.Len(t, eventsOf(frames, eventToolResult), 1)
	var resultEvent toolResultEvent
	decodeData(t, eventsOf(frames, eventToolResult)[0], &resultEvent)
	assert.False(t, resultEvent.OK)
	assert.Contains(t, resultEvent.Content, "未知工具")
	assert.Empty(t, session.calls)
	assert.Equal(t, 2, upstream.CallCount())
}

// TestChatLoopInvalidToolArguments 验证非法参数 JSON 不触发工具调用。
func TestChatLoopInvalidToolArguments(t *testing.T) {
	upstream, server := newScriptedServer(t, func(call int, w http.ResponseWriter, _ llmRequest) {
		if call == 1 {
			writeSSEJSON(t, w, toolChunk(toolDelta(0, "call_1", "list_projects", "not-json")))
			writeSSEDone(t, w)
			return
		}
		writeSSEJSON(t, w, contentChunk("ok"))
		writeSSEDone(t, w)
	})
	session := &fakeToolSession{}
	loop, recorder, _ := newTestLoop(t, server.URL, session)

	loop.run(context.Background(), []chatMessage{{Role: roleUser, Content: "hi"}})

	frames := parseFrames(t, recorder.Body.String())
	require.Len(t, eventsOf(frames, eventToolResult), 1)
	var resultEvent toolResultEvent
	decodeData(t, eventsOf(frames, eventToolResult)[0], &resultEvent)
	assert.False(t, resultEvent.OK)
	assert.Contains(t, resultEvent.Content, "不是合法 JSON")
	assert.Empty(t, session.calls)
	assert.Equal(t, 2, upstream.CallCount())
}

// TestChatLoopClosedStream 验证流已关闭时立即返回，不向上游发起请求。
func TestChatLoopClosedStream(t *testing.T) {
	upstream, server := newScriptedServer(t, func(_ int, w http.ResponseWriter, _ llmRequest) {
		t.Error("流已关闭，不应请求上游")
	})
	loop, recorder, _ := newTestLoop(t, server.URL, &fakeToolSession{})
	loop.sse.close()

	loop.run(context.Background(), []chatMessage{{Role: roleUser, Content: "hi"}})

	assert.Empty(t, recorder.Body.String())
	assert.Equal(t, 0, upstream.CallCount())
}

// TestChatLoopEmptyToolArguments 验证空参数被归一化为 {} 后再交给工具。
func TestChatLoopEmptyToolArguments(t *testing.T) {
	upstream, server := newScriptedServer(t, func(call int, w http.ResponseWriter, _ llmRequest) {
		if call == 1 {
			writeSSEJSON(t, w, toolChunk(toolDelta(0, "call_1", "list_projects", "")))
			writeSSEDone(t, w)
			return
		}
		writeSSEJSON(t, w, contentChunk("ok"))
		writeSSEDone(t, w)
	})
	session := &fakeToolSession{}
	loop, recorder, _ := newTestLoop(t, server.URL, session)

	loop.run(context.Background(), []chatMessage{{Role: roleUser, Content: "hi"}})

	require.Len(t, session.args, 1)
	assert.JSONEq(t, "{}", string(session.args[0]))
	assert.Len(t, eventsOf(parseFrames(t, recorder.Body.String()), eventDone), 1)
	assert.Equal(t, 2, upstream.CallCount())
}

// TestChatLoopPauseWriteFailure 验证写通道断开时暂停帧写出失败不会继续请求上游。
func TestChatLoopPauseWriteFailure(t *testing.T) {
	upstream, server := newScriptedServer(t, func(_ int, w http.ResponseWriter, _ llmRequest) {
		writeSSEJSON(t, w, toolChunk(toolDelta(0, "call_1", "restart_pod", `{}`)))
		writeSSEDone(t, w)
	})
	writer := &failAfterWriter{ResponseRecorder: newRecorder(), remaining: 1}
	sse, err := newSSEWriter(writer)
	require.NoError(t, err)
	loop := newChatLoop(
		newLLMClient(testConfig(server.URL).AI, testLogger()),
		&toolbox{session: &fakeToolSession{}, skills: mustLoadEmbeddedSkills()},
		sse, newTicketSigner("test-key"), testLogger(), "u@example.com", "system", testSpecs(),
	)

	loop.run(context.Background(), []chatMessage{{Role: roleUser, Content: "重启"}})

	body := writer.Body.String()
	assert.Contains(t, body, "event: meta")
	assert.NotContains(t, body, eventNeedConfirm)
	assert.Equal(t, 1, upstream.CallCount())
}

// TestChatLoopResumePauseWriteFailure 验证确认续跑时后续写工具的暂停帧写出失败即终止。
func TestChatLoopResumePauseWriteFailure(t *testing.T) {
	upstream, server := newScriptedServer(t, func(_ int, w http.ResponseWriter, _ llmRequest) {
		t.Error("暂停帧写出失败后不应再请求上游")
	})
	writer := &failAfterWriter{ResponseRecorder: newRecorder(), remaining: 3}
	sse, err := newSSEWriter(writer)
	require.NoError(t, err)
	tickets := newTicketSigner("test-key")
	session := &fakeToolSession{}
	loop := newChatLoop(
		newLLMClient(testConfig(server.URL).AI, testLogger()),
		&toolbox{session: session, skills: mustLoadEmbeddedSkills()},
		sse, tickets, testLogger(), "u@example.com", "system", testSpecs(),
	)

	first := testCall("call_1", "restart_pod", `{}`)
	second := testCall("call_2", "exec_command", `{}`)
	payload, err := tickets.verify(tickets.issue("u@example.com", first))
	require.NoError(t, err)
	conversation := []chatMessage{
		{Role: roleUser, Content: "执行"},
		{Role: roleAssistant, ToolCalls: []toolCall{first, second}},
	}

	loop.resume(context.Background(), conversation, []toolCall{first, second}, 0,
		&confirmation{payload: payload, approved: true})

	assert.Equal(t, []string{"restart_pod"}, session.calls)
	assert.Equal(t, 0, upstream.CallCount())
}

// TestSummarizeToolCall 验证确认条摘要的空参数与超长截断。
func TestSummarizeToolCall(t *testing.T) {
	assert.Equal(t, "restart_pod", summarizeToolCall(testCall("1", "restart_pod", "")))
	assert.Equal(t, "restart_pod {}", summarizeToolCall(testCall("1", "restart_pod", "{}")))

	long := strings.Repeat("x", maxSummaryRunes+10)
	summary := summarizeToolCall(testCall("1", "exec_command", long))
	assert.True(t, strings.HasSuffix(summary, "..."))
	assert.LessOrEqual(t, len(summary), maxSummaryRunes+len("exec_command ")+3)
}
