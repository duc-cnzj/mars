package services

import (
	"net/http"

	"github.com/duc-cnzj/mars/v6/doc"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/server/middlewares"
	swagger_ui "github.com/duc-cnzj/mars/v6/third_party/swagger-ui"
	"github.com/gorilla/mux"
)

// swaggerHandler 仅负责 swagger 文档路由的注册与响应，依赖面最小（只有 logger），
// 与 websocket/文件路由关注点解耦。
type swaggerHandler struct {
	logger mlog.Logger
}

// newSwaggerHandler 创建 swagger 文档处理器。
func newSwaggerHandler(logger mlog.Logger) *swaggerHandler {
	return &swaggerHandler{logger: logger.WithModule("services/swagger")}
}

// Register 注册 swagger 文档路由：/doc/swagger.json 输出 OpenAPI 原始 JSON，
// /docs/ 托管 swagger-ui 静态页；两者都套 HttpCache 缓存中间件。
func (s *swaggerHandler) Register(mux *mux.Router) {
	subrouter := mux.PathPrefix("").Subrouter()
	subrouter.Use(middlewares.HttpCache)

	subrouter.Handle("/doc/swagger.json",
		http.HandlerFunc(s.swaggerJSON),
	)

	subrouter.PathPrefix("/docs/").Handler(
		http.StripPrefix("/docs/", http.FileServer(http.FS(swagger_ui.SwaggerUI))),
	)
}

// swaggerJSON 把编译期嵌入的 OpenAPI 文档原样写回响应。
func (s *swaggerHandler) swaggerJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(doc.SwaggerJson); err != nil {
		s.logger.Debug("write swagger json: ", err)
	}
}
