package services

// llmproxy.go 把沙箱 agent 的 LLM 请求代理到 mars 配置的上游（如 DeepSeek）。
//
// 设计要点：
//   - agent 只认识 <gateway>/llm/v1（见 agent 侧 config.go 的 llmProxyPath），
//     不感知上游厂商；换 DeepSeek/其他 OpenAI 兼容端点只改 mars 配置。
//   - 上游 URL 与 API Key 只存在于 mars，不下发给沙箱：沙箱令牌在代理入口
//     校验后即被剥离，上游请求头统一换成 mars 配置的 provider key。
//   - 走 httputil.ReverseProxy，天然支持 SSE 流式响应（FlushInterval=-1）。

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/server/middlewares"
	"github.com/gorilla/mux"
)

// llmProxyPath 是沙箱 agent 调用 LLM 的路径前缀。与 agent 侧 config.go 的
// llmProxyPath 是同一契约，改动需两边同步。
const llmProxyPath = "/llm/v1"

// llmProxyUpstreamHeaderTimeout 是等待上游响应头的上限：防止上游卡死时
// 连接无限挂起；流式响应体的读取不受此限制。
const llmProxyUpstreamHeaderTimeout = 60 * time.Second

// llmProxyHandler 是 OpenAI 兼容接口的反向代理，路由本身套 LoginHTTP 鉴权。
type llmProxyHandler struct {
	logger   mlog.Logger
	activity biz.AgentActivity
	authMW   func(http.Handler) http.Handler
	// 准入校验依赖：AI 功能必须启用且当前用户已部署自己的 agent 项目，
	// 否则沙箱令牌（或任何 mars token）都能白嫖平台的模型额度。
	repoBiz biz.RepoBiz
	nsBiz   biz.NamespaceBiz
	projBiz biz.ProjectBiz
	// proxy 在上游未配置或地址非法时为 nil，serve 据此返回 503。
	proxy *httputil.ReverseProxy
}

// newLLMProxyHandler 依据 mars 配置构造 LLM 代理：校验沙箱令牌与 AI 功能准入后，
// 把 <gateway>/llm/v1/* 转发到 ai.base_url/*，并把 Authorization 换成
// ai.api_key——沙箱令牌绝不透传给上游。每次已鉴权请求同时刷新沙箱活跃度，
// 用户对话期间沙箱不会被空闲关闭任务缩容。
func newLLMProxyHandler(
	cfg *config.Config,
	authBiz biz.AuthBiz,
	repoBiz biz.RepoBiz,
	nsBiz biz.NamespaceBiz,
	projBiz biz.ProjectBiz,
	activity biz.AgentActivity,
	logger mlog.Logger,
) *llmProxyHandler {
	logger = logger.WithModule("services/llmProxyHandler")
	h := &llmProxyHandler{
		logger:   logger,
		activity: activity,
		repoBiz:  repoBiz,
		nsBiz:    nsBiz,
		projBiz:  projBiz,
		authMW: middlewares.LoginHTTP(func(ctx context.Context, token string) (context.Context, error) {
			return biz.Authenticate(ctx, authBiz, token)
		}, logger),
	}
	proxy := newLLMReverseProxy(cfg.AI.BaseURL, cfg.AI.APIKey, logger)
	if proxy == nil {
		return h
	}
	h.proxy = proxy
	return h
}

// Register 把 LLM 代理挂到 mars 网关路由上。必须在 SPA 兜底之前注册，
// 否则 /llm/v1/* 会被 /{any:.*} 吞成 index.html。
func (h *llmProxyHandler) Register(mux *mux.Router) {
	mux.PathPrefix(llmProxyPath + "/").Handler(h.authMW(h.guard(h.proxyOrError()))).Name("llm_proxy")
}

// guard 是 LLM 代理的准入闸门：只有「AI 功能已启用 + 当前用户已部署 agent 项目」
// 的请求才允许转发。通过后刷新沙箱活跃度、计数并记录访问日志。
// 鉴权已在 LoginHTTP 完成，这里拿不到用户说明路由装配错误，一律拒绝。
func (h *llmProxyHandler) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, err := biz.GetUser(ctx)
		if err != nil || user.Email == "" {
			http.Error(w, "当前用户缺少 email，无法使用 AI 功能", http.StatusForbidden)
			return
		}
		if status, msg := h.checkAgentReady(ctx, user.Email); status != 0 {
			http.Error(w, msg, status)
			return
		}

		// 经网关的真实 LLM 调用代表用户正在使用沙箱，刷新活跃度防缩容。
		biz.TouchAgentActivity(ctx, h.activity, h.logger, user.Email)
		metrics.AgentLLMProxyRequestsTotal.Inc()
		h.logger.InfoCtxf(ctx, "LLM 代理访问: user=%s method=%s path=%s", user.Email, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// checkAgentReady 校验 AI 功能与用户 agent 项目：返回 0 表示放行；否则返回
// 应回吐的 HTTP 状态码与可操作提示。NotFound 是「尚未启用」的业务语义（403），
// 其余后端错误按 500 处理并留在本层统一打日志。
func (h *llmProxyHandler) checkAgentReady(ctx context.Context, email string) (status int, msg string) {
	repo, err := h.repoBiz.GetAIAgent(ctx)
	if err != nil {
		if errs.IsNotFound(err) {
			return http.StatusForbidden, "AI 功能未启用：尚未配置 AI Agent 仓库"
		}
		h.logger.ErrorCtx(ctx, "查询 AI Agent 仓库失败: ", err)
		return http.StatusInternalServerError, "查询 AI Agent 仓库失败"
	}
	if !repo.Enabled {
		return http.StatusForbidden, "AI Agent 仓库已禁用"
	}

	ns, err := h.nsBiz.FindByName(ctx, biz.AgentNamespace)
	if err != nil {
		if errs.IsNotFound(err) {
			return http.StatusForbidden, "AI Agent 空间尚未创建，请先启用 AI 助手"
		}
		h.logger.ErrorCtx(ctx, "查询 AI Agent 空间失败: ", err)
		return http.StatusInternalServerError, "查询 AI Agent 空间失败"
	}
	if _, err := h.projBiz.FindByName(ctx, biz.AgentProjectName(email), ns.ID); err != nil {
		if errs.IsNotFound(err) {
			return http.StatusForbidden, "AI 助手尚未启用，请先调用 /api/ai/agent/enable"
		}
		h.logger.ErrorCtx(ctx, "查询 AI Agent 项目失败: ", err)
		return http.StatusInternalServerError, "查询 AI Agent 项目失败"
	}
	return 0, ""
}

// proxyOrError 返回实际代理；上游未配置时返回 503，让 agent 侧拿到明确原因
// 而不是一个难以定位的连接错误。
func (h *llmProxyHandler) proxyOrError() http.Handler {
	if h.proxy != nil {
		return h.proxy
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "mars 未配置 ai.base_url，无法代理 LLM 请求", http.StatusServiceUnavailable)
	})
}

// newLLMReverseProxy 构造反向代理；baseURL 为空或非法时返回 nil（fail fast
// 到 503），apiKey 为空时删除 Authorization，避免把沙箱令牌透传给上游；
// Cookie 无论是否配置 key 都剥离。
func newLLMReverseProxy(baseURL, apiKey string, logger mlog.Logger) *httputil.ReverseProxy {
	if baseURL == "" {
		return nil
	}
	target, err := url.Parse(baseURL)
	if err != nil || target.Scheme == "" || target.Host == "" {
		return nil
	}

	// SSE 逐帧写回与上游错误映射复用统一的反向代理构造。
	return newStreamingReverseProxy(
		llmProxyUpstreamHeaderTimeout,
		logger,
		"proxy llm request: ",
		"LLM 上游不可用",
		func(pr *httputil.ProxyRequest) {
			req := pr.Out
			incomingPath := req.URL.Path
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = joinURLPath(target.Path, strings.TrimPrefix(incomingPath, llmProxyPath))
			// 路径已被重写，清除可能存在的转义形式，避免按旧 RawPath 发送。
			req.URL.RawPath = ""
			req.Host = target.Host
			// Rewrite 模式不再自动补 X-Forwarded-For，显式补齐（同时带 Host/Proto，
			// 均为标准转发头；进入 Rewrite 前网关已剥离客户端伪造的同名头）。
			pr.SetXForwarded()
			// Authorization 换成 provider key；无上游 key（如本地无鉴权端点）时删除，
			// 防止沙箱令牌泄露。Cookie 属于浏览器与 mars 的会话，必须一律剥离，
			// 否则会随代理请求发到模型服务。
			if apiKey != "" {
				req.Header.Set("Authorization", "Bearer "+apiKey)
			} else {
				req.Header.Del("Authorization")
			}
			req.Header.Del("Cookie")
		},
	)
}

// joinURLPath 拼接上游 base path 与被剥离前缀后的请求路径：
// base 为 /v1、请求路径为 /chat/completions 时得到 /v1/chat/completions。
func joinURLPath(base, path string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")
}
