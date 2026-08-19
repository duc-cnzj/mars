package websocket

// controller.go 定义 ws 传输层入口 websocketManager：
// 负责 HTTP 升级握手、read/write 双循环、消息分发（dispatchEvent）、
// 集群健康同步（TickClusterHealth）与连接计数/优雅关闭（Shutdown）。

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/deploy"
	"github.com/duc-cnzj/mars/v6/internal/locker"
	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"github.com/duc-cnzj/mars/v6/internal/util/counter"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

// upgrader 将 HTTP 请求升级为 websocket 连接；CheckOrigin 恒真以放行任意来源。
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var _ app.WsHttpServer = (*websocketManager)(nil)

// HandleRequestFunc 是单个协议类型的处理器签名（入站消息分发到它）。
type HandleRequestFunc func(ctx context.Context, c Conn, t websocket_pb.Type, message []byte)

// websocketManager 是 ws 传输层入口：持有业务依赖与处理器路由表（handlers），
// 实现 app.WsHttpServer（Serve/Info/Shutdown + TickClusterHealth）。
type websocketManager struct {
	healthTickDuration time.Duration

	timer         timer.Timer
	config        *config.Config
	logger        mlog.Logger
	pluginManager app.PluginManager
	authBiz       biz.AuthBiz
	locker        locker.Locker
	jobManager    deploy.JobManager

	projBiz   biz.ProjectBiz
	k8sRepo   biz.K8sRepo
	fileRepo  biz.FileRepo
	nsRepo    biz.NamespaceRepo
	eventRepo biz.EventRepo
	repoBiz   biz.RepoBiz
	gitBiz    biz.GitBiz
	accessBiz biz.AccessBiz

	counter *counter.Counter

	handlers map[websocket_pb.Type]HandleRequestFunc
}

// WebsocketManagerDeps 是 ws 传输层的依赖集合，wire 经 wire.Struct 全字段注入。
// 与 14 个 gRPC service 的 XxxSvcDeps 模式对齐，消灭 15 个位置参数的 god 构造器。
type WebsocketManagerDeps struct {
	Timer         timer.Timer
	Logger        mlog.Logger
	Counter       *counter.Counter
	ProjBiz       biz.ProjectBiz
	RepoBiz       biz.RepoBiz
	GitBiz        biz.GitBiz
	NsRepo        biz.NamespaceRepo
	AccessBiz     biz.AccessBiz
	JobManager    deploy.JobManager
	Config        *config.Config
	PluginManager app.PluginManager
	AuthBiz       biz.AuthBiz
	Locker        locker.Locker
	ClusterRepo   biz.K8sRepo
	EventRepo     biz.EventRepo
	FileRepo      biz.FileRepo
}

// NewWebsocketManager 组装 WebsocketManager：注入依赖、构建协议类型→处理器路由表。
func NewWebsocketManager(deps WebsocketManagerDeps) app.WsHttpServer {
	mgr := &websocketManager{
		timer:              deps.Timer,
		projBiz:            deps.ProjBiz,
		nsRepo:             deps.NsRepo,
		counter:            deps.Counter,
		repoBiz:            deps.RepoBiz,
		gitBiz:             deps.GitBiz,
		accessBiz:          deps.AccessBiz,
		jobManager:         deps.JobManager,
		fileRepo:           deps.FileRepo,
		healthTickDuration: 15 * time.Second,
		logger:             deps.Logger.WithModule("socket/websocket"),
		pluginManager:      deps.PluginManager,
		authBiz:            deps.AuthBiz,
		locker:             deps.Locker,
		k8sRepo:            deps.ClusterRepo,
		eventRepo:          deps.EventRepo,
		config:             deps.Config,
	}
	mgr.handlers = map[websocket_pb.Type]HandleRequestFunc{
		WsAuthorize:          mgr.HandleAuthorize,
		WsHandleExecShell:    mgr.HandleStartShell,
		WsHandleExecShellMsg: mgr.HandleShellMessage,
		WsHandleCloseShell:   mgr.HandleCloseShell,
		WsCancel:             mgr.HandleCancelDeploy,
		ProjectPodEvent:      mgr.HandleJoinRoom,
		WsCreateProject:      mgr.HandleCreateProject,
		WsUpdateProject:      mgr.HandleUpdateProject,
	}
	return mgr
}

// TickClusterHealth 周期性同步集群信息：抢锁成功后向所有连接广播
// WsClusterInfoSync 帧，done 关闭时退出循环。
func (wc *websocketManager) TickClusterHealth(done <-chan struct{}) {
	ticker := time.NewTicker(wc.healthTickDuration)
	lock := wc.locker
	sub := wc.pluginManager.Ws().New("", "")
	defer sub.Close()

	defer wc.logger.HandlePanic("TickClusterHealth")
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if lock.Acquire("TickClusterHealth", 5) {
				func() {
					defer lock.Release("TickClusterHealth")
					info := wc.k8sRepo.ClusterInfo()
					sub.ToAll(&websocket_pb.WsHandleClusterResponse{
						Metadata: &websocket_pb.Metadata{
							Type: WsClusterInfoSync,
						},
						Info: transformer.FromClusterInfo(info),
					})
				}()
			}
		case <-done:
			wc.logger.Info("[Websocket]: app shutdown and stop WsClusterInfoSync")
			return
		}
	}
}

// Info 以 JSON 返回发布订阅实例的统计信息（供调试探活）。
func (wc *websocketManager) Info(writer http.ResponseWriter, request *http.Request) {
	sub := wc.pluginManager.Ws().New("", "")
	defer sub.Close()
	marshal, _ := json.Marshal(sub.Info())
	writer.Header().Set("Content-Type", "application/json")
	writer.Write(marshal)
}

// Serve 是 ws 入口：HTTP 升级握手后按 errgroup 拉起三个 goroutine
// （pubsub 运行、write 循环、read 循环），任一退出即整体收尾。
func (wc *websocketManager) Serve(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		wc.logger.Error(err)
		return
	}

	var (
		id  = uuid.New().String()
		uid = uuid.New().String()

		inputUid = r.URL.Query().Get("uid")
	)
	if inputUid != "" {
		uid = inputUid
	}

	wsConn := wc.newWsConn(uid, id, c, NewTaskManager(wc.logger), NewSessionMap(wc.logger))
	g, ctx := errgroup.WithContext(r.Context())

	defer func() {
		wc.logger.Debugf("[Websocket]: Serve exit")
		wsConn.CloseAndClean(r.Context())
		wc.counter.Dec()
	}()

	g.Go(func() error {
		defer wc.logger.HandlePanic("[ProjectPodEventSubscriber]: Run")
		return wsConn.PubSub().Run(ctx)
	})

	g.Go(func() error {
		var err error
		defer func() {
			wc.logger.Debugf("[Websocket]: go write exit, err: %v", err)
		}()
		defer wc.logger.HandlePanic("Websocket: Write, err %v")

		return wc.write(ctx, wsConn)
	})

	newMessageSender(wsConn, WsSetUid).SendMsg(wsConn.UID())

	g.Go(func() error {
		var err error
		defer func() {
			wc.logger.Debugf("[Websocket]: go read exit, err: %v", err)
		}()
		defer wc.logger.HandlePanic("[Websocket]: read recovery")
		return wc.read(ctx, wsConn)
	})

	if err = g.Wait(); err != nil {
		// 三个 goroutine 各自的 defer 已打详细退出日志，这里只记录组级错误，
		// 避免重复刷屏；read/write 的 errgroup 返回值本质是"连接已结束"信号。
		wc.logger.Debugf("[Websocket]: Serve goroutines exit with err: %v", err)
		return
	}
}

// Shutdown 等待全部活跃连接结束（counter.Wait），供应用优雅退出。
func (wc *websocketManager) Shutdown(ctx context.Context) error {
	return wc.counter.Wait(ctx)
}

// read 是读循环：解析协议帧，合法消息异步分发给对应 handler，
// 非法帧回 WsInternalError；对端断开时返回错误终止循环。
func (wc *websocketManager) read(ctx context.Context, wsconn Conn) error {
	wsconn.SetReadLimit(maxMessageSize)
	wsconn.SetReadDeadline(wc.timer.Now().Add(pongWait))
	wsconn.SetPongHandler(func(string) error {
		wsconn.SetReadDeadline(wc.timer.Now().Add(pongWait))
		return nil
	})
	for {
		var wsRequest websocket_pb.WsRequestMetadata
		_, message, err := wsconn.ReadMessage()
		if err != nil {
			wc.logger.Debugf("[Websocket]: read error: %v", err)
			return err
		}
		if err := proto.Unmarshal(message, &wsRequest); err != nil {
			newMessageSender(wsconn, WsInternalError).SendEndError(err)

			continue
		}

		go wc.dispatchEvent(ctx, wsconn, &wsRequest, message)
	}
}

// write 是写循环：订阅发布订阅频道，把消息帧/心跳 ping 写入对端，
// 频道关闭或 ctx 取消时收尾并调用 CloseAndClean。
func (wc *websocketManager) write(ctx context.Context, wsconn Conn) (err error) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		wc.logger.Debugf("[Websocket]: go write exit, %v", err)
		ticker.Stop()
		wsconn.CloseAndClean(ctx)
	}()
	wc.logger.Debug(wsconn.PubSub().ID(), wsconn.PubSub().Uid())
	ch := wsconn.PubSub().Subscribe()
	var w io.WriteCloser
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case message, ok := <-ch:
			if !ok {
				return wsconn.WriteMessage(websocket.CloseMessage, []byte{})
			}

			wsconn.SetWriteDeadline(wc.timer.Now().Add(writeWait))
			w, err = wsconn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return err
			}
			w.Write(message)

			if err = w.Close(); err != nil {
				return err
			}
		case <-ticker.C:
			wsconn.SetWriteDeadline(wc.timer.Now().Add(writeWait))
			if err = wsconn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return err
			}
		}
	}
}

// dispatchEvent 把单条请求分发给注册的 handler：未认证（除授权外）先回
// 提示帧，同时统计成功/失败/panic 指标，panic 交由外层恢复。
func (wc *websocketManager) dispatchEvent(ctx context.Context, wsconn Conn, wsRequest *websocket_pb.WsRequestMetadata, message []byte) {
	if handler, ok := wc.handlers[wsRequest.Type]; ok {
		defer wc.logger.HandlePanicWithCallback(wsRequest.Type.String(), func(err error) {
			metrics.WebsocketPanicCount.With(prometheus.Labels{"method": wsRequest.Type.String()}).Inc()
		})
		defer func(t time.Time) {
			metrics.WebsocketRequestLatency.With(prometheus.Labels{"method": wsRequest.Type.String()}).Observe(wc.timer.Since(t).Seconds())
			e := recover()
			if e == nil {
				metrics.WebsocketRequestTotalSuccess.With(prometheus.Labels{"method": wsRequest.Type.String()}).Inc()
			} else {
				metrics.WebsocketRequestTotalFail.With(prometheus.Labels{"method": wsRequest.Type.String()}).Inc()
				panic(e)
			}
		}(wc.timer.Now())

		wc.logger.Debugf("wsType: %v, message: %v", wsRequest.Type.String(), string(message))

		// websocket.onopen 事件不一定是最早发出来的，所以要等 onopen 的认证结束后才能进行后面的操作
		if wsconn.GetUser() == nil && wsRequest.Type != websocket_pb.Type_HandleAuthorize {
			newMessageSender(wsconn, WsAuthorize).SendMsg("认证中，请稍等~")
			return
		}
		// 已认证连接的 user 在 Conn 上（SetUser），不在 ctx；分发前把用户物化进 ctx，
		// 使 handler 内直接调用的 AccessBiz 门卫（内部走 MustGetUser）与 gRPC 侧一致取值。
		// 未认证（仅授权帧）时 user 为 nil，保持裸 ctx：HandleAuthorize 不触达任何门卫。
		if user := wsconn.GetUser(); user != nil {
			ctx = biz.SetUser(ctx, user)
		}
		handler(ctx, wsconn, wsRequest.Type, message)
	}
}
