package aichat

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// signalWriter 在每次写入时向 channel 发信号，避免心跳测试依赖固定 sleep 时长。
type signalWriter struct {
	mu      sync.Mutex
	buf     bytes.Buffer
	written chan struct{}
}

// Header 返回响应头。
func (w *signalWriter) Header() http.Header {
	return http.Header{}
}

// Write 记录内容并发出信号（非阻塞）。
func (w *signalWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf.Write(p)
	select {
	case w.written <- struct{}{}:
	default:
	}
	return len(p), nil
}

// WriteHeader 空实现。
func (w *signalWriter) WriteHeader(int) {}

// Flush 空实现。
func (w *signalWriter) Flush() {}

// String 返回已写入内容。
func (w *signalWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

// TestNewSSEWriterHeaders 验证 SSE 响应头完整且状态码为 200。
func TestNewSSEWriterHeaders(t *testing.T) {
	recorder := newRecorder()
	_, err := newSSEWriter(recorder)
	require.NoError(t, err)

	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "text/event-stream; charset=utf-8", recorder.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", recorder.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", recorder.Header().Get("Connection"))
	assert.Equal(t, "no", recorder.Header().Get("X-Accel-Buffering"))
}

// TestNewSSEWriterWithoutFlusher 验证不支持 Flush 的写入器被明确拒绝。
func TestNewSSEWriterWithoutFlusher(t *testing.T) {
	_, err := newSSEWriter(&noFlushWriter{})
	assert.Error(t, err)
}

// TestSSEWriteEvent 验证事件编码格式。
func TestSSEWriteEvent(t *testing.T) {
	recorder := newRecorder()
	sse, err := newSSEWriter(recorder)
	require.NoError(t, err)

	require.NoError(t, sse.writeEvent("delta", deltaEvent{Text: "你好"}))
	assert.Contains(t, recorder.Body.String(), "event: delta\n")
	assert.Contains(t, recorder.Body.String(), `data: {"text":"你好"}`)
}

// TestSSEWriteError 验证底层写入失败时错误向上返回。
func TestSSEWriteError(t *testing.T) {
	boom := errors.New("boom")
	sse, err := newSSEWriter(&errorFlusherWriter{err: boom})
	require.NoError(t, err)

	assert.ErrorIs(t, sse.writeEvent("delta", deltaEvent{Text: "x"}), boom)
}

// TestSSEWriteEventMarshalError 验证不可序列化的负载返回编码错误。
func TestSSEWriteEventMarshalError(t *testing.T) {
	sse, err := newSSEWriter(newRecorder())
	require.NoError(t, err)

	err = sse.writeEvent("delta", make(chan int))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "序列化 SSE 事件 delta")
}

// TestSSEClosed 验证关闭后的写入返回 errSSEClosed。
func TestSSEClosed(t *testing.T) {
	sse, err := newSSEWriter(newRecorder())
	require.NoError(t, err)
	sse.close()
	assert.ErrorIs(t, sse.writeEvent("delta", deltaEvent{Text: "x"}), errSSEClosed)
	assert.ErrorIs(t, sse.writeComment("ping"), errSSEClosed)
}

// TestSSEHeartbeat 验证心跳按间隔写出注释帧，且停止函数幂等、不泄漏协程。
func TestSSEHeartbeat(t *testing.T) {
	original := sseHeartbeatInterval
	sseHeartbeatInterval = time.Millisecond
	t.Cleanup(func() { sseHeartbeatInterval = original })

	writer := &signalWriter{written: make(chan struct{}, 1)}
	sse, err := newSSEWriter(writer)
	require.NoError(t, err)

	stop := sse.startHeartbeat(context.Background())
	select {
	case <-writer.written:
	case <-time.After(2 * time.Second):
		t.Fatal("心跳未在预期时间内写出")
	}
	stop()
	stop()
	sse.close()

	assert.Contains(t, writer.String(), ": ping")
}

// TestSSEHeartbeatContextDone 验证 ctx 取消后心跳协程退出。直接驱动心跳主体并
// 预先取消 ctx：此时只有 ctx.Done 就绪（ticker 未到、done 未关），分支命中确定。
func TestSSEHeartbeatContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sse, err := newSSEWriter(newRecorder())
	require.NoError(t, err)

	sse.heartbeatLoop(ctx, make(chan struct{}))
}
