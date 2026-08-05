package http

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/duc-cnzj/mars/api/v6/http/transport"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// openStream 打开一条 server-streaming 流。grpc-gateway 对带 http 注解的
// server-streaming 方法输出 NDJSON（JSONPb delimiter="\n"），chunked 传输；
// 部分网关/中间件可能转成标准 SSE（text/event-stream）。eventStream 两者都解析。
// 流建立请求遇 401 时自动刷新重试一次（与 unary doReq 对齐）。
func (c *Client) openStream(ctx context.Context, method, path string, req proto.Message) (transport.RawStream, error) {
	return c.openStreamRefresh(ctx, method, path, req, true)
}

// openStreamRefresh 建立 server-streaming 流。allowRefresh 为 true 时，若建立请求
// 返回 401（token 过期）且配置了 WithAuth + WithTokenAutoRefresh，则刷新 token 后
// 重试一次（不递归），与 unary doReq 的自动刷新行为一致。
func (c *Client) openStreamRefresh(ctx context.Context, method, path string, req proto.Message, allowRefresh bool) (transport.RawStream, error) {
	u := c.baseURL + path
	if method == http.MethodGet || method == http.MethodDelete {
		if q := encodeQuery(req); q != "" {
			u += "?" + q
		}
	} else if req != nil {
		return nil, fmt.Errorf("http: openStream 暂只支持 GET/DELETE query 绑定")
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Accept", "text/event-stream, application/json")
	if tok := c.authToken(); tok != "" {
		httpReq.Header.Set("Authorization", tok)
	}

	resp, err := c.hc.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if allowRefresh && c.autoRefresh && c.username != "" {
			if ge, ok := parseGatewayError(data); ok && codes.Code(ge.Code) == codes.Unauthenticated {
				if gerr := c.refreshToken(); gerr != nil {
					return nil, gerr
				}
				return c.openStreamRefresh(ctx, method, path, req, false)
			}
		}
		return nil, c.errFromStatus(resp.StatusCode, data)
	}
	return &eventStream{rc: resp.Body, br: bufio.NewReader(resp.Body)}, nil
}

// eventStream 逐条消费 gateway 的流式输出。
type eventStream struct {
	rc io.ReadCloser
	br *bufio.Reader
}

// Recv 读取下一条消息反序列化进 out；io.EOF 表示流正常结束。
// 兼容两种格式：
//   - NDJSON：每行一个 JSON 对象（grpc-gateway JSONPb 流式输出，delimiter="\n"）；
//   - 标准 SSE：data: {...} 事件块，空行分隔，多行 data 拼接。
func (s *eventStream) Recv(out proto.Message) error {
	var data strings.Builder
	inEvent := false
	for {
		line, err := s.br.ReadBytes('\n')
		trimmed := strings.TrimRight(string(line), "\r\n")
		switch {
		case trimmed == "":
			// 空行 = SSE 事件结束；NDJSON 中的孤立空行忽略。
			if inEvent && data.Len() > 0 {
				return decodeStreamEvent(out, data.String())
			}
		case strings.HasPrefix(trimmed, ":"):
			// SSE 注释行。
		case strings.HasPrefix(trimmed, "data:"):
			inEvent = true
			payload := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
			if data.Len() > 0 {
				data.WriteByte('\n')
			}
			data.WriteString(payload)
		case strings.HasPrefix(trimmed, "event:") || strings.HasPrefix(trimmed, "id:") || strings.HasPrefix(trimmed, "retry:"):
			inEvent = true // SSE 元字段，忽略值
		case strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "["):
			// NDJSON 消息行。
			return decodeStreamEvent(out, trimmed)
		}
		if err != nil {
			if err == io.EOF {
				if inEvent && data.Len() > 0 {
					return decodeStreamEvent(out, data.String())
				}
				return io.EOF
			}
			return err
		}
	}
}

// decodeStreamEvent 解出单条流消息。
//
// grpc-gateway v2 的 ForwardResponseStream（runtime/handler.go）把每个 server-streaming
// 消息包成 {"result": <msg>}，流中途错误包成 {"error": <google.rpc.Status>}；
// 无包装的裸消息（自定义 marshaler / 老版本）直接解整行。三种形态都兼容。
func decodeStreamEvent(out proto.Message, payload string) error {
	var env struct {
		Result json.RawMessage `json:"result"`
		Error  json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal([]byte(payload), &env); err == nil && (env.Result != nil || env.Error != nil) {
		if env.Error != nil {
			return streamErrorFromEnvelope(env.Error)
		}
		return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(env.Result, out)
	}
	return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal([]byte(payload), out)
}

// streamErrorFromEnvelope 把 {"error": {"code":..,"message":..}} 还原成 codes.Error，
// 让流中途错误与 unary 错误码通用。
func streamErrorFromEnvelope(raw json.RawMessage) error {
	var se struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(raw, &se); err != nil {
		return fmt.Errorf("http: 解析流错误体失败: %w", err)
	}
	return status.Error(codes.Code(se.Code), se.Message)
}

func (s *eventStream) Close() error { return s.rc.Close() }
