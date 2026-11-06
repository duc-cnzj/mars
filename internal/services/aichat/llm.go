package aichat

// llm.go 是 OpenAI 兼容上游（/chat/completions）的流式客户端：手写 net/http +
// SSE 解析，不引入厂商 SDK。只覆盖对话循环需要的子集——文本增量、tool_calls
// 增量聚合与 [DONE] 结束标记。
//
// 该文件同时定义服务端对话的消息模型（chatMessage/toolCall）：请求体、上游
// 请求与 SSE 回传的 messages 共用同一套 JSON 结构，避免多处投影漂移。

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

const (
	// llmResponseHeaderTimeout 是等待上游响应头的上限；流式响应体的读取不受此限制。
	llmResponseHeaderTimeout = 60 * time.Second
	// llmErrorBodyLimit 只限制上游错误体写入日志的大小，避免把整页 HTML 打进日志；
	// 不影响任何回传给客户端或模型的数据。
	llmErrorBodyLimit = 4 << 10
)

// 对话消息角色。客户端请求只允许 user/assistant/tool：system 由服务端统一注入，
// 不允许客户端携带（否则可以覆盖服务端的安全约定）。
const (
	roleSystem    = "system"
	roleUser      = "user"
	roleAssistant = "assistant"
	roleTool      = "tool"
)

// chatMessage 是对话消息的线上结构（请求体、上游请求与 SSE 回传共用）。
type chatMessage struct {
	// Role 是消息角色：system/user/assistant/tool。
	Role string `json:"role"`
	// Content 是文本内容；assistant 带 tool_calls 时可能为空。
	Content string `json:"content,omitempty"`
	// ToolCalls 仅 assistant 消息携带：模型请求调用的工具列表。
	ToolCalls []toolCall `json:"tool_calls,omitempty"`
	// ToolCallID 仅 tool 消息携带：对应哪一次工具调用。
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// toolCall 是模型发起的一次工具调用。
type toolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function toolFunction `json:"function"`
}

// toolFunction 是工具调用的名称与参数原文（JSON 字符串）。
type toolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// toolDefinition 是发给上游的工具体定义（OpenAI function calling 格式）。
type toolDefinition struct {
	Type     string          `json:"type"`
	Function toolDeclaration `json:"function"`
}

// toolDeclaration 是 function 字段：名称、说明与 JSON Schema 入参。
type toolDeclaration struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

// llmRequest 是上游 /chat/completions 的请求体（只含用到的字段）。
type llmRequest struct {
	Model    string           `json:"model"`
	Messages []chatMessage    `json:"messages"`
	Tools    []toolDefinition `json:"tools,omitempty"`
	Stream   bool             `json:"stream"`
}

// chatChunk 是流式响应的一帧。
type chatChunk struct {
	Choices []chatChoice `json:"choices"`
}

// chatChoice 是单帧中的一个候选。
type chatChoice struct {
	Delta chatDelta `json:"delta"`
}

// chatDelta 是增量内容：文本或工具调用的分片。
type chatDelta struct {
	Content   string          `json:"content"`
	ToolCalls []toolCallDelta `json:"tool_calls"`
}

// toolCallDelta 是工具调用的增量分片：id/name 通常只在首帧出现，
// arguments 会跨多帧拼接。
type toolCallDelta struct {
	Index    int    `json:"index"`
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// llmClient 是 OpenAI 兼容上游的流式客户端。
type llmClient struct {
	// endpoint 是完整的 /chat/completions 地址（base_url + 路径）。
	endpoint string
	// apiKey 只存在于服务端，不会下发给浏览器。
	apiKey string
	// model 是上游模型 id。
	model string
	// httpClient 等待响应头上限 60s；流式响应体读取不受此限制。
	httpClient *http.Client
	logger     mlog.Logger
}

// newLLMClient 依据 mars 的 AI 配置构造客户端。base_url 为空时 endpoint 无意义，
// 但对话入口会先做 enabled 判断，不会走到请求路径。
func newLLMClient(cfg config.AIConfig, logger mlog.Logger) *llmClient {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = llmResponseHeaderTimeout
	return &llmClient{
		endpoint:   joinURL(cfg.BaseURL, "chat/completions"),
		apiKey:     cfg.APIKey,
		model:      cfg.Model,
		httpClient: &http.Client{Transport: transport},
		logger:     logger,
	}
}

// stream 发起一次流式请求：文本增量经 onDelta 实时回吐，返回累计文本与聚合后的
// 工具调用（assistant 消息需要保存完整文本）。上游非 2xx 与流解析失败都返回 error，
// 由调用方统一转成通用错误帧（不回传上游原文）。
func (c *llmClient) stream(
	ctx context.Context,
	messages []chatMessage,
	tools []toolDefinition,
	onDelta func(text string) error,
) (string, []toolCall, error) {
	// llmRequest 的字段全部是可序列化类型，编码不会失败。
	payload, _ := json.Marshal(llmRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    tools,
		Stream:   true,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", nil, fmt.Errorf("构造上游请求: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("请求上游模型服务: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, llmErrorBodyLimit))
		c.logger.ErrorCtxf(ctx, "上游模型服务返回 %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		return "", nil, fmt.Errorf("上游模型服务返回状态码 %d", resp.StatusCode)
	}

	return parseStream(resp.Body, onDelta)
}

// parseStream 解析上游 SSE：忽略非 data 行，data: [DONE] 结束，其余按帧聚合
// 文本增量与工具调用分片。按行读取不设单行长度上限——工具参数可能很长，
// 超长内容原样透传，不在解析层做防御性截断。
func parseStream(body io.Reader, onDelta func(text string) error) (string, []toolCall, error) {
	reader := bufio.NewReader(body)
	accumulator := &toolCallAccumulator{byIndex: map[int]*toolCall{}}
	var text strings.Builder
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "[DONE]" {
				return text.String(), accumulator.result(), nil
			}
			if data != "" {
				var chunk chatChunk
				if err := json.Unmarshal([]byte(data), &chunk); err != nil {
					return "", nil, fmt.Errorf("解析上游流式响应: %w", err)
				}
				for _, choice := range chunk.Choices {
					if choice.Delta.Content != "" {
						text.WriteString(choice.Delta.Content)
						if err := onDelta(choice.Delta.Content); err != nil {
							return "", nil, err
						}
					}
					accumulator.add(choice.Delta.ToolCalls)
				}
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", nil, fmt.Errorf("读取上游流式响应: %w", err)
		}
	}
	return text.String(), accumulator.result(), nil
}

// toolCallAccumulator 按 index 聚合工具调用分片：id/name/type 取首个非空值，
// arguments 逐帧拼接。
type toolCallAccumulator struct {
	order   []int
	byIndex map[int]*toolCall
}

// add 合并一帧中的工具调用分片。
func (a *toolCallAccumulator) add(deltas []toolCallDelta) {
	for _, delta := range deltas {
		call, ok := a.byIndex[delta.Index]
		if !ok {
			call = &toolCall{Type: "function"}
			a.byIndex[delta.Index] = call
			a.order = append(a.order, delta.Index)
		}
		if delta.ID != "" {
			call.ID = delta.ID
		}
		if delta.Type != "" {
			call.Type = delta.Type
		}
		call.Function.Name += delta.Function.Name
		call.Function.Arguments += delta.Function.Arguments
	}
}

// result 按首次出现顺序返回聚合后的工具调用。
func (a *toolCallAccumulator) result() []toolCall {
	calls := make([]toolCall, 0, len(a.order))
	for _, index := range a.order {
		calls = append(calls, *a.byIndex[index])
	}
	return calls
}

// joinURL 拼接上游 base 与路径：base 为 .../v1 时得到 .../v1/chat/completions。
func joinURL(base, path string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")
}
