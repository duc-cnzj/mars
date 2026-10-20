package services

// aiconfig.go 向沙箱侧 mars-agent 暴露 mars 侧的集中配置。mars 是运行时参数的
// 唯一事实来源：agent chart 只携带网关地址与会话令牌，模型等参数在 agent 启动时
// 经本接口拉取，避免同一份配置在 mars 与每个沙箱的 values.yaml 里各维护一份。

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/server/middlewares"
	"github.com/gorilla/mux"
)

// aiConfigPath 是沙箱 agent 拉取运行时配置的路径。
// 与 agent 侧 config.go 的拉取路径是同一契约，改动需两边同步。
const aiConfigPath = "/api/ai/config"

// aiConfigResponse 是 /api/ai/config 的 200 出参。
type aiConfigResponse struct {
	// Model 是 agent 使用的模型 id（裸 id，agent 会在配置里写成 mars/<id>）。
	Model string `json:"model"`
}

// aiConfigHandler 提供 agent 运行时配置的只读接口，路由本身套 LoginHTTP 鉴权。
type aiConfigHandler struct {
	cfg      *config.Config
	activity biz.AgentActivity
	logger   mlog.Logger
	authMW   func(http.Handler) http.Handler
}

// newAIConfigHandler 构造 agent 配置处理器：鉴权核心与 gRPC 拦截器、MCP、文件路由
// 共用 biz.Authenticate，杜绝第二套鉴权实现。每次拉取配置同时刷新沙箱活跃度
// （agent 启动即代表有人在用）。
func newAIConfigHandler(cfg *config.Config, authBiz biz.AuthBiz, activity biz.AgentActivity, logger mlog.Logger) *aiConfigHandler {
	logger = logger.WithModule("services/aiConfigHandler")
	return &aiConfigHandler{
		cfg:      cfg,
		activity: activity,
		logger:   logger,
		authMW: middlewares.LoginHTTP(func(ctx context.Context, token string) (context.Context, error) {
			return biz.Authenticate(ctx, authBiz, token)
		}, logger),
	}
}

// Register 把配置端点挂到 mars 网关路由上。必须在前端 SPA 兜底与 /api/ 前缀
// 路由之前注册，否则 /api/ai/config 会被 grpc-gateway 的 PathPrefix 吞掉
// （与 /api/ws_info 的注册顺序约束一致，见 internal/server/http.go 的 initServer）。
func (h *aiConfigHandler) Register(mux *mux.Router) {
	mux.Handle(aiConfigPath, h.authMW(http.HandlerFunc(h.serve))).Methods(http.MethodGet).Name("ai_config")
}

// serve 回吐 agent 运行时配置。未配置 ai.model 时返回 503：mars 侧配置缺失属于
// 部署问题，不能给 agent 下发一个空模型让它带着坏配置起来。
// 每次拉取配置同时刷新沙箱活跃度（agent 启动即代表有人在用）。
func (h *aiConfigHandler) serve(w http.ResponseWriter, r *http.Request) {
	// agent 启动即代表有人在用：按用户派生的 agent 项目刷新活跃度。
	if user, err := biz.GetUser(r.Context()); err == nil {
		biz.TouchAgentActivity(r.Context(), h.activity, h.logger, user.Email)
	}
	if h.cfg.AI.Model == "" {
		http.Error(w, "mars 未配置 ai.model，无法向沙箱下发模型", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(aiConfigResponse{Model: h.cfg.AI.Model}); err != nil {
		h.logger.ErrorCtx(r.Context(), "write ai config response: ", err)
	}
}
