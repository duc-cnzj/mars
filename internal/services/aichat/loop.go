package aichat

// loop.go 是服务端对话的核心：模型 → 工具 → 模型 的编排循环。
//
// 设计约束：
//   - 只执行模型本轮新生成的工具调用；客户端历史里的 tool_calls 仅作为上下文，绝不执行。
//   - 遇到第一个写工具立即暂停，签发确认票据后结束本次 SSE；确认接口从该调用继续。
//   - 工具结果原样回吐，不做服务端截断（上下文超限由上游模型自行报错）。

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

const (
	// maxToolRounds 限制单次请求的「模型调用 + 工具执行」轮数，防止模型在工具间死循环。
	maxToolRounds = 30
	// maxSummaryRunes 是确认条摘要的最大字符数，只用于展示，不改变实际参数。
	maxSummaryRunes = 200
	// upstreamErrorMessage 是回给浏览器的通用错误文案：上游原始错误只进服务端日志，
	// 避免把上游地址、响应体等内部信息泄漏出去。
	upstreamErrorMessage = "AI 服务暂时不可用，请稍后重试"
)

// 对话循环回吐的事件类型（SSE 的 event 名）。
const (
	eventMeta        = "meta"
	eventDelta       = "delta"
	eventToolCall    = "tool_call"
	eventToolResult  = "tool_result"
	eventNeedConfirm = "need_confirm"
	eventError       = "error"
	eventDone        = "done"
)

// metaEvent 是流开始帧。
type metaEvent struct {
	Model string `json:"model"`
}

// deltaEvent 是模型增量文本。
type deltaEvent struct {
	Text string `json:"text"`
}

// toolCallEvent 是开始执行工具的通知（只读工具与已批准的写工具都会发）。
type toolCallEvent struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
	ReadOnly  bool   `json:"read_only"`
}

// toolResultEvent 是工具执行结果（内容原样透传）。
type toolResultEvent struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Content string `json:"content"`
}

// needConfirmEvent 是写工具的暂停帧：携带签名票据与完整对话状态，
// 前端原样回传给 /api/ai/chat/confirm。
type needConfirmEvent struct {
	Ticket    string        `json:"ticket"`
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Arguments string        `json:"arguments"`
	Summary   string        `json:"summary"`
	Messages  []chatMessage `json:"messages"`
}

// errorEvent 是终止帧。
type errorEvent struct {
	Message string `json:"message"`
}

// doneEvent 是正常结束帧。
type doneEvent struct {
	Rounds    int   `json:"rounds"`
	ElapsedMS int64 `json:"elapsed_ms"`
}

// toolbox 组合 MCP 工具与本地 load_skill 工具，对循环暴露同一套 List/Call 语义。
type toolbox struct {
	session toolSession
	skills  *skillRegistry
}

// list 返回当前身份可见的全部工具（MCP 工具 + 本地 load_skill）。
func (t *toolbox) list(ctx context.Context) ([]toolSpec, error) {
	specs, err := t.session.List(ctx)
	if err != nil {
		return nil, err
	}
	return append(specs, t.skills.toolSpec()), nil
}

// call 分发一次工具调用：load_skill 走本地 skills，其余交给 MCP 会话。
func (t *toolbox) call(ctx context.Context, name string, args json.RawMessage) (toolResult, error) {
	if name == skillToolName {
		return t.callSkill(args), nil
	}
	return t.session.Call(ctx, name, args)
}

// callSkill 执行本地 load_skill：名称非法时以业务错误（isError）回吐，
// 让模型看到明确失败而不是拿到空串继续猜。
func (t *toolbox) callSkill(args json.RawMessage) toolResult {
	var input struct {
		Name string `json:"name"`
	}
	if len(bytes.TrimSpace(args)) > 0 {
		if err := json.Unmarshal(args, &input); err != nil {
			return toolResult{Text: "load_skill 的参数不是合法 JSON", IsError: true}
		}
	}
	body, err := t.skills.load(input.Name)
	if err != nil {
		return toolResult{Text: err.Error(), IsError: true}
	}
	return toolResult{Text: body}
}

// confirmation 是一次用户确认结果：payload 来自已校验的票据，approved 是用户选择。
type confirmation struct {
	payload  ticketPayload
	approved bool
}

// chatLoop 承载一轮 SSE 请求的全部状态。
type chatLoop struct {
	llm     *llmClient
	tools   *toolbox
	sse     *sseWriter
	tickets *ticketSigner
	logger  mlog.Logger
	// email 是当前登录用户，写工具票据绑定该身份。
	email string
	// system 是服务端注入的 system prompt（客户端永远拿不到）。
	system string
	// specs 是当前身份可见的工具集合，specByName 用于判定只读性。
	specs      []toolSpec
	specByName map[string]toolSpec
}

// newChatLoop 构造对话循环并建立工具名索引。
func newChatLoop(
	llm *llmClient,
	tools *toolbox,
	sse *sseWriter,
	tickets *ticketSigner,
	logger mlog.Logger,
	email string,
	system string,
	specs []toolSpec,
) *chatLoop {
	byName := make(map[string]toolSpec, len(specs))
	for _, spec := range specs {
		byName[spec.Name] = spec
	}
	return &chatLoop{
		llm:        llm,
		tools:      tools,
		sse:        sse,
		tickets:    tickets,
		logger:     logger,
		email:      email,
		system:     system,
		specs:      specs,
		specByName: byName,
	}
}

// run 从一段新对话开始执行循环。conversation 不含 system 消息（服务端每次请求
// 时临时拼到最前面），因此暂停时回传给前端的历史也不会带出 system prompt。
func (l *chatLoop) run(ctx context.Context, conversation []chatMessage) {
	l.execute(ctx, conversation, nil, 0, nil)
}

// resume 从确认票据指向的工具调用继续执行：pending 是该 assistant 消息里的
// 全部调用，start 是票据指向调用的下标。
func (l *chatLoop) resume(ctx context.Context, conversation []chatMessage, pending []toolCall, start int, approval *confirmation) {
	l.execute(ctx, conversation, pending, start, approval)
}

// execute 是 run/resume 的公共主体：先处理待确认调用（resume），再进入
// 「请求模型 → 派发工具」循环。
func (l *chatLoop) execute(ctx context.Context, conversation []chatMessage, pending []toolCall, start int, approval *confirmation) {
	started := time.Now()
	if err := l.sse.writeEvent(eventMeta, metaEvent{Model: l.llm.model}); err != nil {
		return
	}

	if approval != nil {
		cont, err := l.processCalls(ctx, &conversation, pending, start, approval)
		if err != nil {
			l.fail(ctx, err, upstreamErrorMessage)
			return
		}
		if !cont {
			return
		}
	}

	firstToken := false
	onDelta := func(text string) error {
		if !firstToken {
			firstToken = true
			metrics.AIChatFirstTokenSeconds.Observe(time.Since(started).Seconds())
		}
		return l.sse.writeEvent(eventDelta, deltaEvent{Text: text})
	}

	for round := 1; round <= maxToolRounds; round++ {
		text, calls, err := l.llm.stream(ctx, l.upstreamMessages(conversation), l.toolDefinitions(), onDelta)
		if err != nil {
			l.fail(ctx, err, upstreamErrorMessage)
			return
		}
		if len(calls) == 0 {
			l.finish(round, started)
			return
		}
		conversation = append(conversation, chatMessage{Role: roleAssistant, Content: text, ToolCalls: calls})
		cont, err := l.processCalls(ctx, &conversation, calls, 0, nil)
		if err != nil {
			l.fail(ctx, err, upstreamErrorMessage)
			return
		}
		if !cont {
			return
		}
	}

	// 直到轮数上限仍在请求工具：如实终止，不编造完成。
	metrics.AIChatRequestsTotal.WithLabelValues(metrics.AIChatResultError).Inc()
	_ = l.sse.writeEvent(eventError, errorEvent{Message: "工具调用轮数超过上限，已停止"})
}

// processCalls 从 calls[start] 起顺序处理工具调用：
//   - 只读工具直接执行；
//   - 写工具在 approval 指向它时按用户选择执行/拒绝，否则下发 need_confirm 并暂停。
//
// 返回 false 表示已暂停（已下发 need_confirm），调用方应结束本次 SSE。
func (l *chatLoop) processCalls(
	ctx context.Context,
	conversation *[]chatMessage,
	calls []toolCall,
	start int,
	approval *confirmation,
) (bool, error) {
	for i := start; i < len(calls); i++ {
		call := calls[i]
		spec, ok := l.specByName[call.Function.Name]
		if !ok {
			l.reportToolResult(call, toolResult{
				Text:    fmt.Sprintf("未知工具：%s", call.Function.Name),
				IsError: true,
			}, conversation)
			continue
		}

		if spec.ReadOnly {
			_ = l.sse.writeEvent(eventToolCall, toolCallEvent{
				ID: call.ID, Name: call.Function.Name, Arguments: call.Function.Arguments, ReadOnly: true,
			})
			l.reportToolResult(call, l.executeTool(ctx, call), conversation)
			continue
		}

		if approval == nil || approval.payload.ToolCallID != call.ID {
			return false, l.pause(call, *conversation)
		}

		if !approval.approved {
			l.reportToolResult(call, toolResult{Text: "用户拒绝执行该操作", IsError: true}, conversation)
			// 一张票据只授权一个调用：后续写工具必须重新确认。
			approval = nil
			continue
		}

		_ = l.sse.writeEvent(eventToolCall, toolCallEvent{
			ID: call.ID, Name: call.Function.Name, Arguments: call.Function.Arguments, ReadOnly: false,
		})
		l.reportToolResult(call, l.executeTool(ctx, call), conversation)
		approval = nil
	}
	return true, nil
}

// executeTool 校验参数 JSON 并执行工具；协议层失败收敛为 isError 结果，
// 让模型能看到失败并自行调整，而不是中断整轮对话。
func (l *chatLoop) executeTool(ctx context.Context, call toolCall) toolResult {
	raw := json.RawMessage(call.Function.Arguments)
	if len(bytes.TrimSpace(raw)) == 0 {
		raw = json.RawMessage("{}")
	}
	if !json.Valid(raw) {
		return toolResult{Text: "工具参数不是合法 JSON", IsError: true}
	}

	result, err := l.tools.call(ctx, call.Function.Name, raw)
	if err != nil {
		l.logger.ErrorCtx(ctx, "工具执行失败: ", err)
		result = toolResult{Text: "工具执行失败（服务端内部错误）", IsError: true}
	}

	label := metrics.AIChatResultOK
	if result.IsError {
		label = metrics.AIChatResultError
	}
	metrics.AIChatToolCallsTotal.WithLabelValues(call.Function.Name, label).Inc()
	return result
}

// reportToolResult 回吐结果帧并把 tool 消息追加进对话。
func (l *chatLoop) reportToolResult(call toolCall, result toolResult, conversation *[]chatMessage) {
	_ = l.sse.writeEvent(eventToolResult, toolResultEvent{
		ID:      call.ID,
		Name:    call.Function.Name,
		OK:      !result.IsError,
		Content: result.Text,
	})
	*conversation = append(*conversation, chatMessage{
		Role:       roleTool,
		ToolCallID: call.ID,
		Content:    result.Text,
	})
}

// pause 签发确认票据并下发 need_confirm；返回写帧错误（连接已断时由调用方终止）。
func (l *chatLoop) pause(call toolCall, conversation []chatMessage) error {
	ticket := l.tickets.issue(l.email, call)
	metrics.AIChatRequestsTotal.WithLabelValues(metrics.AIChatResultOK).Inc()
	return l.sse.writeEvent(eventNeedConfirm, needConfirmEvent{
		Ticket:    ticket,
		ID:        call.ID,
		Name:      call.Function.Name,
		Arguments: call.Function.Arguments,
		Summary:   summarizeToolCall(call),
		Messages:  conversation,
	})
}

// finish 回吐 done 帧并记账成功请求。
func (l *chatLoop) finish(rounds int, started time.Time) {
	metrics.AIChatRequestsTotal.WithLabelValues(metrics.AIChatResultOK).Inc()
	_ = l.sse.writeEvent(eventDone, doneEvent{
		Rounds:    rounds,
		ElapsedMS: time.Since(started).Milliseconds(),
	})
}

// fail 记录真实错误并回吐通用错误帧。客户端已断开（或写入失败）时只做尽力而为的
// 收尾，不把连接中断记成服务端错误。
func (l *chatLoop) fail(ctx context.Context, err error, message string) {
	if ctx.Err() == nil && !errors.Is(err, errSSEClosed) {
		metrics.AIChatRequestsTotal.WithLabelValues(metrics.AIChatResultError).Inc()
		l.logger.ErrorCtx(ctx, "AI 对话失败: ", err)
	}
	_ = l.sse.writeEvent(eventError, errorEvent{Message: message})
}

// upstreamMessages 在对话前拼上服务端 system prompt（只发给上游，不回传前端）。
func (l *chatLoop) upstreamMessages(conversation []chatMessage) []chatMessage {
	messages := make([]chatMessage, 0, len(conversation)+1)
	messages = append(messages, chatMessage{Role: roleSystem, Content: l.system})
	return append(messages, conversation...)
}

// toolDefinitions 把工具描述转成 OpenAI function calling 格式。
func (l *chatLoop) toolDefinitions() []toolDefinition {
	defs := make([]toolDefinition, 0, len(l.specs))
	for _, spec := range l.specs {
		defs = append(defs, toolDefinition{
			Type: "function",
			Function: toolDeclaration{
				Name:        spec.Name,
				Description: spec.Description,
				Parameters:  spec.InputSchema,
			},
		})
	}
	return defs
}

// summarizeToolCall 生成确认条展示用的一行摘要（只用于展示，不改变实际参数）。
func summarizeToolCall(call toolCall) string {
	args := oneLine(call.Function.Arguments)
	if args == "" {
		return call.Function.Name
	}
	runes := []rune(args)
	if len(runes) > maxSummaryRunes {
		args = string(runes[:maxSummaryRunes]) + "..."
	}
	return call.Function.Name + " " + args
}
