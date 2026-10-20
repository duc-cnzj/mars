package services

// reverseproxy.go 收口 services 层「面向流式上游」的反向代理构造：传输参数
// （响应头超时、SSE 不缓冲）与错误处理（客户端取消不刷日志、上游故障映射 502）
// 收成一处，供 LLM 代理等流式转发复用，避免实现随迭代漂移。

import (
	"context"
	"errors"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

// newStreamingReverseProxy 构造 SSE 友好的反向代理：
//   - ResponseHeaderTimeout 限制等待上游响应头的时间，避免上游卡死时连接无限挂起；
//     流式响应体的读取不受此限制。
//   - FlushInterval=-1 表示不缓冲，逐帧写回（LLM 增量输出）。
//   - 客户端主动断开（context.Canceled）不写响应、不刷错误日志；其余上游故障
//     统一记 Error 并返回 errorMsg（502）。
//
// rewrite 负责改写目标地址与剥离 mars 凭据，由调用方按各自的上游契约实现。
func newStreamingReverseProxy(
	headerTimeout time.Duration,
	logger mlog.Logger,
	errorLogMsg string,
	errorMsg string,
	rewrite func(*httputil.ProxyRequest),
) *httputil.ReverseProxy {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = headerTimeout

	return &httputil.ReverseProxy{
		Transport:     transport,
		FlushInterval: -1,
		Rewrite:       rewrite,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			// 客户端主动断开不算服务端故障，不刷错误日志。
			if errors.Is(err, context.Canceled) {
				return
			}
			logger.ErrorCtx(r.Context(), errorLogMsg, err)
			http.Error(w, errorMsg, http.StatusBadGateway)
		},
	}
}
