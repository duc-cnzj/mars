package websocket

//go:generate go tool mockgen -destination ./mock_websocket_test.go -package websocket github.com/duc-cnzj/mars/v6/internal/services/websocket Conn,PtyHandler,TaskManager,GorillaWs,SessionMapper
import (
	"github.com/duc-cnzj/mars/v6/internal/util/counter"
	"github.com/google/wire"
)

// WireWebsocket 提供 websocket 传输层（WsHttpServer 实现 + 连接计数）。
// WebsocketManagerDeps 由 wire.Struct 全字段注入，与 14 个 gRPC service 的
// XxxSvcDeps 构造模式对齐。
var WireWebsocket = wire.NewSet(NewWebsocketManager, wire.Struct(new(WebsocketManagerDeps), "*"), counter.NewCounter)
