package services

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

var _ app.HttpHandler = (*httpHandlerImpl)(nil)

// httpHandlerImpl 是 HTTP 传输层的门面：内嵌 websocket 服务器（WsHttpServer）承载
// /ws 与 /api/ws_info，swagger 与文件路由分别委托给 swaggerHandler/fileHandler。
// 自身只保留 ws 路由注册与优雅关闭，不再持有文件边界的任何依赖。
type httpHandlerImpl struct {
	app.WsHttpServer
	logger  mlog.Logger
	swagger *swaggerHandler
	files   *fileHandler
}

// HttpHandlerDeps 收口 NewHttpHandler 的构造依赖，由 wire 按字段注入。
type HttpHandlerDeps struct {
	WsHttpServer app.WsHttpServer
	Logger       mlog.Logger
	Uploader     uploader.Uploader
	AuthBiz      biz.AuthBiz
	UserBiz      biz.UserBiz
	EventBiz     biz.EventBiz
	FileBiz      biz.FileBiz
	Timer        timer.Timer
	K8sBiz       biz.K8sBiz
	ContainerBiz biz.ContainerBiz
	AccessBiz    biz.AccessBiz
}

// NewHttpHandler 收口 HTTP 传输层的构造依赖：websocket 服务器（WsHttpServer）、
// 上传适配器（uploader）与计时器（timer）是 HTTP 边界专属的输出端口，鉴权、
// 文件、事件、k8s 各 biz 由 wire 注入；swagger 与文件处理器在内部按关注点拆分为
// 独立子处理器。返回 app.HttpHandler 供服务装配。
func NewHttpHandler(deps HttpHandlerDeps) app.HttpHandler {
	logger := deps.Logger.WithModule("services/httpHandler")
	return &httpHandlerImpl{
		WsHttpServer: deps.WsHttpServer,
		logger:       logger,
		swagger:      newSwaggerHandler(logger),
		files:        newFileHandler(deps),
	}
}

// RegisterWsRoute 注册 websocket 相关路由：/api/ws_info 返回连接信息，
// /ws 是 websocket 升级端点，两者最终都走 WsHttpServer 的 Handler。
func (h *httpHandlerImpl) RegisterWsRoute(mux *mux.Router) {
	mux.HandleFunc("/api/ws_info", h.Info).Name("ws_info")
	mux.HandleFunc("/ws", h.Serve).Name("ws")
}

// RegisterSwaggerUIRoute 注册 swagger 文档路由，委托给 swaggerHandler。
func (h *httpHandlerImpl) RegisterSwaggerUIRoute(mux *mux.Router) {
	h.swagger.Register(mux)
}

// RegisterFileRoute 注册文件相关 HTTP 路由（上传/下载/从 pod 拷贝），委托给 fileHandler。
func (h *httpHandlerImpl) RegisterFileRoute(mux *runtime.ServeMux) {
	h.files.RegisterFileRoute(mux)
}

// Shutdown 带超时优雅关闭 websocket 服务器，超时或失败只记 Warning 不阻断主流程。
func (h *httpHandlerImpl) Shutdown(ctx context.Context) error {
	ctx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc()
	err := h.WsHttpServer.Shutdown(ctx)
	if err != nil {
		h.logger.Warning("shutdown ws server error: ", err.Error())
	}
	return err
}
