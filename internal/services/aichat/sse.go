package aichat

// sse.go 是服务端向浏览器写 text/event-stream 的最小实现：事件编码、逐帧 flush、
// 互斥写入与心跳。心跳用于等待上游首 token 或工具执行期间维持连接，防止网关
// 因长时间无数据而断开。

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// sseHeartbeatInterval 是心跳注释帧的发送间隔。声明为变量以便测试把间隔调小，
// 生产路径不会改写它。
var sseHeartbeatInterval = 15 * time.Second

// errSSEClosed 表示流已结束，后续写入应被丢弃（例如心跳与主流程的收尾竞争）。
var errSSEClosed = errors.New("SSE 连接已关闭")

// sseWriter 串行化所有写入：主流程的 delta 与心跳协程共用同一个 http.ResponseWriter，
// 不加锁会把两帧交错在一起，客户端将解析失败。
type sseWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
	mu      sync.Mutex
	closed  bool
}

// newSSEWriter 写出 SSE 响应头。调用后不能再改状态码，因此所有入参校验与
// 上游会话初始化都必须在此之前完成。
func newSSEWriter(w http.ResponseWriter) (*sseWriter, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("http.ResponseWriter 不支持流式刷新")
	}
	header := w.Header()
	header.Set("Content-Type", "text/event-stream; charset=utf-8")
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	// 关闭 nginx 等反向代理的响应缓冲，保证逐帧到达浏览器。
	header.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	return &sseWriter{w: w, flusher: flusher}, nil
}

// writeEvent 编码并写出一帧 `event: <name>` + JSON data。
func (s *sseWriter) writeEvent(event string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化 SSE 事件 %s: %w", event, err)
	}
	return s.write(fmt.Sprintf("event: %s\ndata: %s\n\n", event, data))
}

// writeComment 写出 SSE 注释帧（以 ':' 开头），客户端会忽略其内容。
func (s *sseWriter) writeComment(text string) error {
	return s.write(": " + text + "\n\n")
}

// write 在锁内写一帧并立即 flush；流关闭后返回 errSSEClosed。
func (s *sseWriter) write(frame string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errSSEClosed
	}
	if _, err := io.WriteString(s.w, frame); err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

// heartbeatLoop 是心跳协程主体：直到 ctx 取消或 done 关闭才退出，期间按间隔写注释帧。
func (s *sseWriter) heartbeatLoop(ctx context.Context, done <-chan struct{}) {
	ticker := time.NewTicker(sseHeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
			_ = s.writeComment("ping")
		}
	}
}

// startHeartbeat 启动心跳协程，返回停止函数。停止函数幂等且会等待协程退出，
// 保证响应结束后不再有 goroutine 向底层 writer 写入。
func (s *sseWriter) startHeartbeat(ctx context.Context) func() {
	done := make(chan struct{})
	exited := make(chan struct{})
	go func() {
		defer close(exited)
		s.heartbeatLoop(ctx, done)
	}()

	var once sync.Once
	return func() {
		once.Do(func() { close(done) })
		<-exited
	}
}

// close 标记流已结束，后续写入直接失败而不是继续往已完成的响应写数据。
func (s *sseWriter) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
}
