package aichat

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/services/mcp"
	"github.com/gorilla/mux"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// postJSON 构造一个携带登录用户的 JSON 请求。
func postJSON(t *testing.T, body any, user *biz.UserInfo) *http.Request {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	return withUser(httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(raw)), user)
}

// newHandlerServer 构造一个用 fakeOpener 替换 MCP 会话的 Server。
func newHandlerServer(t *testing.T, cfg *config.Config, opener toolSessionOpener) *Server {
	t.Helper()
	server := NewServer(Deps{Config: cfg, Logger: testLogger()})
	server.opener = opener
	return server
}

// TestServerRegister 验证三条路由都挂载到预期方法与名称上。
func TestServerRegister(t *testing.T) {
	server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{})
	router := mux.NewRouter()
	server.Register(router)

	got := map[string]string{}
	require.NoError(t, router.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		name := route.GetName()
		if name == "" {
			return nil
		}
		methods, _ := route.GetMethods()
		got[name] = strings.Join(methods, ",")
		return nil
	}))
	assert.Equal(t, map[string]string{
		"ai_chat_status":  http.MethodGet,
		"ai_chat":         http.MethodPost,
		"ai_chat_confirm": http.MethodPost,
	}, got)
}

// TestServeStatus 验证启用状态与模型名的出参。
func TestServeStatus(t *testing.T) {
	t.Run("已启用", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{})
		recorder := newRecorder()
		server.serveStatus(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
		var body chatStatusResponse
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
		assert.True(t, body.Enabled)
		assert.Equal(t, "test-model", body.Model)
	})

	t.Run("未配置", func(t *testing.T) {
		server := newHandlerServer(t, testConfig(""), &fakeOpener{})
		recorder := newRecorder()
		server.serveStatus(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

		var body chatStatusResponse
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
		assert.False(t, body.Enabled)
		assert.Empty(t, body.Model)
	})
}

// TestServeChatErrors 覆盖发起对话前的各类拒绝分支。
func TestServeChatErrors(t *testing.T) {
	user := testUser("chat@example.com")

	t.Run("未启用", func(t *testing.T) {
		server := newHandlerServer(t, testConfig(""), &fakeOpener{})
		recorder := newRecorder()
		server.serveChat(recorder, postJSON(t, chatRequest{Messages: []chatMessage{{Role: roleUser, Content: "hi"}}}, user))
		assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	})

	t.Run("请求体不是 JSON", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{})
		recorder := newRecorder()
		req := withUser(httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not-json")), user)
		server.serveChat(recorder, req)
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("对话不合法", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{})
		recorder := newRecorder()
		server.serveChat(recorder, postJSON(t, chatRequest{
			Messages: []chatMessage{{Role: roleSystem, Content: "覆盖 system"}},
		}, user))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("关注范围不合法", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{})
		recorder := newRecorder()
		server.serveChat(recorder, postJSON(t, chatRequest{
			Messages: []chatMessage{{Role: roleUser, Content: "hi"}},
			Bindings: []binding{{Project: "web"}},
		}, user))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("用户缺少 email", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{})
		recorder := newRecorder()
		server.serveChat(recorder, postJSON(t, chatRequest{
			Messages: []chatMessage{{Role: roleUser, Content: "hi"}},
		}, testUser("")))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("打开工具会话失败", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{err: assert.AnError})
		recorder := newRecorder()
		server.serveChat(recorder, postJSON(t, chatRequest{
			Messages: []chatMessage{{Role: roleUser, Content: "hi"}},
		}, user))
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("列出工具失败", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{
			session: &fakeToolSession{listErr: assert.AnError},
		})
		recorder := newRecorder()
		server.serveChat(recorder, postJSON(t, chatRequest{
			Messages: []chatMessage{{Role: roleUser, Content: "hi"}},
		}, user))
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("连接不支持流式", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{session: &fakeToolSession{specs: testSpecs()}})
		writer := &noFlushWriter{}
		server.serveChat(writer, postJSON(t, chatRequest{
			Messages: []chatMessage{{Role: roleUser, Content: "hi"}},
		}, user))
		assert.Equal(t, http.StatusInternalServerError, writer.code)
	})
}

// TestServeChatSuccess 验证完整的一轮对话以 SSE 回吐 meta/delta/done。
func TestServeChatSuccess(t *testing.T) {
	_, server := newScriptedServer(t, func(_ int, w http.ResponseWriter, _ llmRequest) {
		writeSSEJSON(t, w, contentChunk("你好"))
		writeSSEDone(t, w)
	})
	handler := newHandlerServer(t, testConfig(server.URL), &fakeOpener{
		session: &fakeToolSession{specs: testSpecs()},
	})

	recorder := newRecorder()
	handler.serveChat(recorder, postJSON(t, chatRequest{
		Messages: []chatMessage{{Role: roleUser, Content: "hi"}},
		Bindings: []binding{{Namespace: "ns", Project: "web"}},
	}, testUser("chat@example.com")))

	frames := parseFrames(t, recorder.Body.String())
	assert.Len(t, eventsOf(frames, eventMeta), 1)
	assert.Len(t, eventsOf(frames, eventDelta), 1)
	assert.Len(t, eventsOf(frames, eventDone), 1)
	assert.Equal(t, "text/event-stream; charset=utf-8", recorder.Header().Get("Content-Type"))
}

// writeToolConversation 构造一段「模型已请求 restart_pod、等待确认」的对话。
func writeToolConversation(t *testing.T, server *Server, args string) ([]chatMessage, string) {
	t.Helper()
	call := testCall("call_1", "restart_pod", args)
	ticket := server.tickets.issue("confirm@example.com", call)
	return []chatMessage{
		{Role: roleUser, Content: "重启"},
		{Role: roleAssistant, ToolCalls: []toolCall{call}},
	}, ticket
}

// TestServeConfirmErrors 覆盖确认接口的各类拒绝分支。
func TestServeConfirmErrors(t *testing.T) {
	user := testUser("confirm@example.com")

	t.Run("未启用", func(t *testing.T) {
		server := newHandlerServer(t, testConfig(""), &fakeOpener{})
		recorder := newRecorder()
		server.serveConfirm(recorder, postJSON(t, confirmRequest{Ticket: "x"}, user))
		assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	})

	t.Run("请求体不是 JSON", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{})
		recorder := newRecorder()
		server.serveConfirm(recorder, withUser(
			httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not-json")), user))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("缺少票据", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{})
		recorder := newRecorder()
		server.serveConfirm(recorder, postJSON(t, confirmRequest{
			Messages: []chatMessage{{Role: roleUser, Content: "hi"}},
		}, user))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "缺少确认票据")
	})

	t.Run("对话不合法", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{})
		recorder := newRecorder()
		server.serveConfirm(recorder, postJSON(t, confirmRequest{
			Ticket:   "x",
			Messages: []chatMessage{{Role: "unknown", Content: "hi"}},
		}, user))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("关注范围不合法", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{})
		recorder := newRecorder()
		server.serveConfirm(recorder, postJSON(t, confirmRequest{
			Ticket:   "x",
			Messages: []chatMessage{{Role: roleUser, Content: "hi"}},
			Bindings: []binding{{}},
		}, user))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("用户缺少 email", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{})
		recorder := newRecorder()
		server.serveConfirm(recorder, postJSON(t, confirmRequest{
			Ticket:   "x",
			Messages: []chatMessage{{Role: roleUser, Content: "hi"}},
		}, testUser("")))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("票据无效", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{})
		recorder := newRecorder()
		server.serveConfirm(recorder, postJSON(t, confirmRequest{
			Ticket:   "bad.ticket",
			Messages: []chatMessage{{Role: roleUser, Content: "hi"}},
		}, user))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("票据与用户不匹配", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{})
		recorder := newRecorder()
		ticket := server.tickets.issue("other@example.com", testCall("call_1", "restart_pod", "{}"))
		server.serveConfirm(recorder, postJSON(t, confirmRequest{
			Ticket:   ticket,
			Messages: []chatMessage{{Role: roleUser, Content: "hi"}},
		}, user))
		assert.Equal(t, http.StatusForbidden, recorder.Code)
	})

	t.Run("对话与票据不匹配", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{})
		recorder := newRecorder()
		ticket := server.tickets.issue("confirm@example.com", testCall("call_1", "restart_pod", "{}"))
		server.serveConfirm(recorder, postJSON(t, confirmRequest{
			Ticket:   ticket,
			Messages: []chatMessage{{Role: roleUser, Content: "hi"}},
		}, user))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "对话状态与确认票据不匹配")
	})

	t.Run("打开工具会话失败", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{err: assert.AnError})
		recorder := newRecorder()
		messages, ticket := writeToolConversation(t, server, "{}")
		server.serveConfirm(recorder, postJSON(t, confirmRequest{
			Ticket: ticket, Approved: true, Messages: messages,
		}, user))
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("列出工具失败", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{
			session: &fakeToolSession{listErr: assert.AnError},
		})
		recorder := newRecorder()
		messages, ticket := writeToolConversation(t, server, "{}")
		server.serveConfirm(recorder, postJSON(t, confirmRequest{
			Ticket: ticket, Approved: true, Messages: messages,
		}, user))
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("票据指向只读工具", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{
			session: &fakeToolSession{specs: testSpecs()},
		})
		recorder := newRecorder()
		call := testCall("call_1", "list_projects", "{}")
		ticket := server.tickets.issue("confirm@example.com", call)
		messages := []chatMessage{
			{Role: roleUser, Content: "列出"},
			{Role: roleAssistant, ToolCalls: []toolCall{call}},
		}
		server.serveConfirm(recorder, postJSON(t, confirmRequest{
			Ticket: ticket, Approved: true, Messages: messages,
		}, user))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "不是可执行的写工具")
	})

	t.Run("连接不支持流式", func(t *testing.T) {
		server := newHandlerServer(t, testConfig("http://example.com"), &fakeOpener{
			session: &fakeToolSession{specs: testSpecs()},
		})
		messages, ticket := writeToolConversation(t, server, "{}")
		writer := &noFlushWriter{}
		server.serveConfirm(writer, postJSON(t, confirmRequest{
			Ticket: ticket, Approved: true, Messages: messages,
		}, user))
		assert.Equal(t, http.StatusInternalServerError, writer.code)
	})
}

// TestServeConfirmSuccess 验证批准后执行写工具并继续对话。
func TestServeConfirmSuccess(t *testing.T) {
	_, server := newScriptedServer(t, func(_ int, w http.ResponseWriter, _ llmRequest) {
		writeSSEJSON(t, w, contentChunk("已重启"))
		writeSSEDone(t, w)
	})
	session := &fakeToolSession{specs: testSpecs()}
	handler := newHandlerServer(t, testConfig(server.URL), &fakeOpener{session: session})
	messages, ticket := writeToolConversation(t, handler, `{"namespace":"ns"}`)

	recorder := newRecorder()
	handler.serveConfirm(recorder, postJSON(t, confirmRequest{
		Ticket: ticket, Approved: true, Messages: messages,
	}, testUser("confirm@example.com")))

	frames := parseFrames(t, recorder.Body.String())
	require.Len(t, eventsOf(frames, eventToolCall), 1)
	require.Len(t, eventsOf(frames, eventToolResult), 1)
	var result toolResultEvent
	decodeData(t, eventsOf(frames, eventToolResult)[0], &result)
	assert.True(t, result.OK)
	assert.Len(t, eventsOf(frames, eventDone), 1)
	assert.Equal(t, []string{"restart_pod"}, session.calls)
}

// TestValidateConversation 覆盖消息结构校验的全部分支。
func TestValidateConversation(t *testing.T) {
	t.Run("通过", func(t *testing.T) {
		assert.NoError(t, validateConversation([]chatMessage{
			{Role: roleUser, Content: "hi"},
			{Role: roleAssistant, ToolCalls: []toolCall{testCall("c1", "list_projects", "{}")}},
			{Role: roleTool, ToolCallID: "c1", Content: "ok"},
		}))
	})

	cases := map[string]struct {
		messages []chatMessage
	}{
		"空列表":        {nil},
		"system 角色":  {[]chatMessage{{Role: roleSystem}}},
		"tool 缺少 id": {[]chatMessage{{Role: roleTool}}},
		"调用缺少 id":    {[]chatMessage{{Role: roleAssistant, ToolCalls: []toolCall{{Function: toolFunction{Name: "x"}}}}}},
		"调用缺少名称":     {[]chatMessage{{Role: roleAssistant, ToolCalls: []toolCall{{ID: "c1"}}}}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Error(t, validateConversation(tc.messages))
		})
	}
}

// TestValidateConversation_NoSizeLimit 验证消息内容、工具参数与消息条数不设大小上限。
func TestValidateConversation_NoSizeLimit(t *testing.T) {
	content := strings.Repeat("你", 64<<10)
	args := strings.Repeat("a", 64<<10)
	messages := make([]chatMessage, 0, 301)
	for i := 0; i < 300; i++ {
		messages = append(messages, chatMessage{Role: roleUser, Content: content})
	}
	messages = append(messages, chatMessage{
		Role:      roleAssistant,
		ToolCalls: []toolCall{{ID: "c1", Function: toolFunction{Name: "x", Arguments: args}}},
	})

	require.NoError(t, validateConversation(messages))
	assert.Equal(t, content, messages[0].Content)
	assert.Equal(t, args, messages[300].ToolCalls[0].Function.Arguments)
}

// TestValidateBindings 覆盖关注范围校验的全部分支。
func TestValidateBindings(t *testing.T) {
	assert.NoError(t, validateBindings(nil))
	assert.NoError(t, validateBindings([]binding{{Namespace: "ns"}}))
	many := make([]binding, 20)
	for i := range many {
		many[i] = binding{Namespace: "ns"}
	}
	assert.NoError(t, validateBindings(many))
	assert.Error(t, validateBindings([]binding{{Project: "web"}}))
}

// TestFindConfirmedCall 覆盖票据与对话匹配的各类情形。
func TestFindConfirmedCall(t *testing.T) {
	first := testCall("call_1", "restart_pod", `{"a":1}`)
	second := testCall("call_2", "exec_command", `{}`)
	payload := ticketPayload{
		ToolCallID: "call_2",
		ToolName:   "exec_command",
		ArgsSHA256: argsHash("{}"),
	}
	messages := []chatMessage{
		{Role: roleUser, Content: "执行"},
		{Role: roleAssistant, ToolCalls: []toolCall{first, second}},
		{Role: roleTool, ToolCallID: "call_1", Content: "ok"},
	}

	t.Run("命中", func(t *testing.T) {
		assistantIndex, callIndex, ok := findConfirmedCall(messages, payload)
		assert.True(t, ok)
		assert.Equal(t, 1, assistantIndex)
		assert.Equal(t, 1, callIndex)
	})

	t.Run("前序调用未完成", func(t *testing.T) {
		_, _, ok := findConfirmedCall(messages[:2], payload)
		assert.False(t, ok)
	})

	t.Run("名称不匹配", func(t *testing.T) {
		_, _, ok := findConfirmedCall(messages, ticketPayload{
			ToolCallID: "call_2", ToolName: "restart_pod", ArgsSHA256: argsHash("{}"),
		})
		assert.False(t, ok)
	})

	t.Run("参数被替换", func(t *testing.T) {
		_, _, ok := findConfirmedCall(messages, ticketPayload{
			ToolCallID: "call_2", ToolName: "exec_command", ArgsSHA256: argsHash(`{"changed":true}`),
		})
		assert.False(t, ok)
	})

	t.Run("找不到调用", func(t *testing.T) {
		_, _, ok := findConfirmedCall(messages, ticketPayload{ToolCallID: "missing"})
		assert.False(t, ok)
	})
}

// TestPrecedingCallsResolved 验证前序调用为空与缺失两种情形。
func TestPrecedingCallsResolved(t *testing.T) {
	assert.True(t, precedingCallsResolved(nil, nil))
	assert.False(t, precedingCallsResolved(
		[]chatMessage{{Role: roleUser, Content: "hi"}},
		[]toolCall{testCall("c1", "x", "{}")},
	))
}

// TestIsWritableTool 验证工具存在性与只读标记的判定。
func TestIsWritableTool(t *testing.T) {
	specs := testSpecs()
	assert.True(t, isWritableTool(specs, "restart_pod"))
	assert.False(t, isWritableTool(specs, "list_projects"))
	assert.False(t, isWritableTool(specs, "unknown"))
}

// TestBuildSystemPrompt 验证 system prompt 按身份、目录、关注范围、界面语言与
// 本次会话的工具规格拼装。
func TestBuildSystemPrompt(t *testing.T) {
	specs := testSpecs()
	userPrompt := buildSystemPrompt(false, "", nil, "zh-CN", specs)
	assert.NotContains(t, userPrompt, "管理员专属工具")
	assert.Contains(t, userPrompt, "## 会话约定")
	assert.Contains(t, userPrompt, "简体中文")
	// 写工具清单来自本次规格：只列真实存在的 restart_pod / exec_command。
	// （MCP 指令正文本身会提到 apply_project，断言只看生成的那一行。）
	assert.Contains(t, userPrompt, "- 写工具（restart_pod / exec_command）会暂停并等待用户确认")
	assert.NotContains(t, userPrompt, "- 写工具（apply_project")

	adminPrompt := buildSystemPrompt(true, "", nil, "", specs)
	assert.Contains(t, adminPrompt, "管理员专属工具")
	// 未传 locale 时按中文处理（默认语言）。
	assert.Contains(t, adminPrompt, "简体中文")

	withCatalog := buildSystemPrompt(false, "- crashloop-triage：演示\n", nil, "zh-CN", specs)
	assert.Contains(t, withCatalog, "## 可用 Skills")
	assert.Contains(t, withCatalog, "crashloop-triage")

	withBindings := buildSystemPrompt(false, "", []binding{
		{Namespace: "ns-a"},
		{Namespace: "ns-b", Project: "web"},
	}, "zh-CN", specs)
	assert.Contains(t, withBindings, "## 当前关注范围")
	assert.Contains(t, withBindings, "空间 ns-a")
	assert.Contains(t, withBindings, "空间 ns-b 下的项目 web")

	english := buildSystemPrompt(false, "", nil, "en-US", specs)
	assert.Contains(t, english, "## 回答语言")
	assert.Contains(t, english, "英文")
	assert.NotContains(t, english, "简体中文：始终")

	// 只读身份（无写工具）：生成段不出现写工具约定，避免误导模型去调用不存在的工具。
	readOnlyOnly := buildSystemPrompt(false, "", nil, "zh-CN", []toolSpec{
		{Name: "list_projects", ReadOnly: true},
	})
	assert.NotContains(t, readOnlyOnly, "- 写工具（")
}

// TestWriteToolNames 验证写工具筛选只取非只读规格并保持清单顺序。
func TestWriteToolNames(t *testing.T) {
	assert.Equal(t, []string{"restart_pod", "exec_command"}, writeToolNames(testSpecs()))
	assert.Empty(t, writeToolNames(nil))
}

// TestMCPOpener 验证生产适配器能打开进程内会话并列出工具。
func TestMCPOpener(t *testing.T) {
	server := mcp.NewServer(mcp.Deps{Logger: testLogger()})
	ctx := biz.SetUser(context.Background(), testUser("opener@example.com"))

	session, cleanup, err := mcpOpener{server: server}.open(ctx)
	require.NoError(t, err)
	defer cleanup()

	specs, err := session.List(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, specs)
}

// fakeSessionServer 让进程内会话建立恒失败。
type fakeSessionServer struct {
	err error
}

// OpenSession 恒返回注入的错误。
func (f fakeSessionServer) OpenSession(context.Context) (*sdk.ClientSession, func(), error) {
	return nil, nil, f.err
}

// TestMCPOpenerError 验证会话建立失败时错误向上返回。
func TestMCPOpenerError(t *testing.T) {
	_, _, err := mcpOpener{server: fakeSessionServer{err: assert.AnError}}.open(context.Background())
	assert.ErrorIs(t, err, assert.AnError)
}

// TestAuthMiddleware 验证路由鉴权复用 biz.Authenticate，并把用户注入 ctx。
func TestAuthMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	authBiz := biz.NewMockAuthBiz(ctrl)
	authBiz.EXPECT().VerifyToken(gomock.Any(), "Bearer tok").Return(testUser("auth@example.com"), nil)
	authBiz.EXPECT().EffectiveRoles(gomock.Any(), "auth@example.com", gomock.Any()).
		Return([]string{"mars_user"}, nil)

	server := NewServer(Deps{Config: testConfig("http://example.com"), AuthBiz: authBiz, Logger: testLogger()})
	server.opener = &fakeOpener{}
	router := mux.NewRouter()
	server.Register(router)

	req := httptest.NewRequest(http.MethodGet, chatStatusPath, nil)
	req.Header.Set("Authorization", "Bearer tok")
	recorder := newRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
}
