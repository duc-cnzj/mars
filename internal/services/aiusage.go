package services

// aiusage.go 接收前端上报的 AI 使用漏斗事件。事件名走服务端白名单校验，
// 写入低基数指标；协议码与业务无关，失败一律 400，不触碰任何业务依赖。

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/server/middlewares"
	"github.com/gorilla/mux"
)

// aiUsagePath 是前端上报 AI 使用漏斗事件的路径。
const aiUsagePath = "/api/ai/usage"

// AI 使用漏斗事件名：前端上报、服务端白名单校验后写入低基数指标。
// 新增事件时在常量块追加常量并在 isAIUsageEvent 中补 case，否则接口按未知事件拒绝。
const (
	// aiUsageEntryClick 记录用户点击 AI 入口的次数。
	aiUsageEntryClick = "entry_click"
	// aiUsageChatFirstToken 记录对话首段文本到达（用户开始看到输出）。
	aiUsageChatFirstToken = "chat_first_token"
	// aiUsageChatDone 记录一轮对话正常结束。
	aiUsageChatDone = "chat_done"
	// aiUsageChatError 记录一轮对话失败终止。
	aiUsageChatError = "chat_error"
)

// isAIUsageEvent 判断事件名是否命中受限白名单：指标 label 不能由客户端任意决定，
// 否则每个新字符串都会生成一条时间序列，造成基数爆炸。新增事件时在常量块加常量、
// 并在 switch 中补一个 case，两处都指向同一组常量，不存在值重复维护。
func isAIUsageEvent(name string) bool {
	switch name {
	case aiUsageEntryClick,
		aiUsageChatFirstToken,
		aiUsageChatDone,
		aiUsageChatError:
		return true
	default:
		return false
	}
}

// aiUsageRequest 是 POST /api/ai/usage 的请求体。
type aiUsageRequest struct {
	// Event 必须命中 isAIUsageEvent 白名单。
	Event string `json:"event"`
}

// aiUsageHandler 是 AI 使用事件的上报入口，路由本身套 LoginHTTP 鉴权。
type aiUsageHandler struct {
	authMW func(http.Handler) http.Handler
}

// newAIUsageHandler 构造上报处理器：鉴权核心与 gRPC 拦截器、MCP、文件路由共用
// biz.Authenticate，杜绝第二套鉴权实现。
func newAIUsageHandler(authBiz biz.AuthBiz, logger mlog.Logger) *aiUsageHandler {
	logger = logger.WithModule("services/aiUsageHandler")
	return &aiUsageHandler{
		authMW: middlewares.LoginHTTP(func(ctx context.Context, token string) (context.Context, error) {
			return biz.Authenticate(ctx, authBiz, token)
		}, logger),
	}
}

// Register 把上报端点挂到 mars 网关路由上。必须在前端 SPA 兜底与 /api/ 前缀
// 路由之前注册，否则会被 grpc-gateway 的 PathPrefix 吞掉。
func (h *aiUsageHandler) Register(mux *mux.Router) {
	mux.Handle(aiUsagePath, h.authMW(http.HandlerFunc(h.serveUsage))).Methods(http.MethodPost).Name("ai_usage")
}

// serveUsage 接收前端上报的 AI 使用事件。事件名走白名单，未知名直接 400：
// 指标 label 不能由客户端任意决定，否则每个新字符串都会生成一条时间序列。
func (h *aiUsageHandler) serveUsage(w http.ResponseWriter, r *http.Request) {
	var req aiUsageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的 AI 使用事件", http.StatusBadRequest)
		return
	}
	if !isAIUsageEvent(req.Event) {
		http.Error(w, "未知的 AI 使用事件", http.StatusBadRequest)
		return
	}
	metrics.AIUsageEventTotal.WithLabelValues(req.Event).Inc()
	w.WriteHeader(http.StatusNoContent)
}
