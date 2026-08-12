package eventhandler

import (
	"context"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/event"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"google.golang.org/protobuf/proto"
)

// EventCoordinator 是领域事件用例层的唯一监听入口（entry handler）：
// 一方面编排跨聚合生命周期事件（namespace/project → 注入 TLS 证书、广播
// ws reload、落变更日志），另一方面承接 audit 日志的落库订阅（事件用例层
// 负责监听，data 层只负责生产 Dispatch 与持久化 HandleAuditLog）。所有
// dispatcher 上的业务监听统一在此注册，避免散落。
//
// 插件能力（PluginDeps）以惰性闭包注入：构造发生在 wire 期（插件未加载，
// pm.Domain()/pm.Ws() 恒为 nil），而事件触发在运行期（插件已加载），闭包在
// 触发时实时解析 pm，无需 BeforeServerRunHooks 二次刷新。闭包由组合根（cmd）
// 捕获 PluginManager 构造，本包零依赖 app。
type EventCoordinator struct {
	logger      mlog.Logger
	dispatcher  event.Dispatcher
	getCerts    func() (name, key, crt string)
	toAll       func(msg proto.Message) error
	projectRepo biz.ProjectRepo
	k8sRepo     biz.K8sRepo
	clRepo      biz.ChangelogRepo
	eventRepo   biz.EventRepo
}

// PluginDeps 事件用例的插件能力集合。字段为惰性闭包：wire 期构造（插件未加载），
// 触发时实时解析已加载插件，不依赖服务器启动前二次刷新。
type PluginDeps struct {
	// GetCerts 返回域名插件提供的 TLS 证书名称、密钥与证书。
	GetCerts func() (name, key, crt string)
	// ToAll 向全部 ws 客户端广播一条 websocket 消息。
	ToAll func(msg proto.Message) error
}

// NewEventCoordinator 构造协调器并在 dispatcher 上注册全部业务事件监听
// （4 个跨域生命周期 + 1 个 audit 落库订阅）。deps 为惰性插件闭包集合
// （wire 期由组合根注入），返回的具体类型由组合根（wire）注入 app，保证
// 对象图可达，监听不丢失。
func NewEventCoordinator(
	logger mlog.Logger,
	dispatcher event.Dispatcher,
	projectRepo biz.ProjectRepo,
	k8sRepo biz.K8sRepo,
	clRepo biz.ChangelogRepo,
	eventRepo biz.EventRepo,
	deps *PluginDeps,
) *EventCoordinator {
	c := &EventCoordinator{
		logger:      logger.WithModule("event/coordinator"),
		dispatcher:  dispatcher,
		getCerts:    deps.GetCerts,
		toAll:       deps.ToAll,
		projectRepo: projectRepo,
		k8sRepo:     k8sRepo,
		clRepo:      clRepo,
		eventRepo:   eventRepo,
	}
	c.listen()
	return c
}

// listen 注册业务事件监听：每个监听回调把 event.Event 还原为 biz.EventKey
// 后转交对应 handler。audit 监听在此注册（事件用例层），持久化实现在
// data.eventRepo.HandleAuditLog，监听与落库跨层解耦。
func (c *EventCoordinator) listen() {
	c.dispatcher.Listen(event.Event(biz.EventNamespaceCreated), func(d any, e event.Event) error {
		return c.HandleInjectTlsSecret(d, biz.EventKey(e))
	})
	c.dispatcher.Listen(event.Event(biz.EventNamespaceDeleted), func(d any, e event.Event) error {
		return c.HandleNamespaceDeleted(d, biz.EventKey(e))
	})
	c.dispatcher.Listen(event.Event(biz.EventProjectChanged), func(d any, e event.Event) error {
		return c.HandleProjectChanged(d, biz.EventKey(e))
	})
	c.dispatcher.Listen(event.Event(biz.EventProjectDeleted), func(d any, e event.Event) error {
		return c.HandleProjectDeleted(d, biz.EventKey(e))
	})
	c.dispatcher.Listen(event.Event(biz.AuditLogEvent), func(d any, e event.Event) error {
		return c.HandleAuditLog(d, biz.EventKey(e))
	})
}

// HandleInjectTlsSecret 处理 namespace 创建事件：域名插件配置了证书时，
// 向该 namespace 注入 TLS secret。异步监听场景错误无法上抛，原地打日志。
func (c *EventCoordinator) HandleInjectTlsSecret(d any, e biz.EventKey) error {
	if createdData, ok := d.(biz.NamespaceCreatedData); ok {
		name, key, crt := c.getCerts()
		if name != "" && key != "" && crt != "" {
			ns := createdData.NsK8sObj.Name
			if _, err := c.k8sRepo.AddTlsSecret(ns, name, key, crt); err != nil {
				c.logger.Error(err)
			}
		}
	}
	return nil
}

// HandleNamespaceDeleted 处理 namespace 删除事件：向所有 ws 客户端广播 reload 消息。
func (c *EventCoordinator) HandleNamespaceDeleted(d any, e biz.EventKey) error {
	c.toAll(&websocket_pb.WsReloadProjectsResponse{
		Metadata:    &websocket_pb.Metadata{Type: websocket_pb.Type_ReloadProjects},
		NamespaceId: int32(d.(biz.NamespaceDeletedData).ID),
	})
	c.logger.Debug("event handled: ", e.String())
	return nil
}

// HandleProjectChanged 处理项目变更事件：读取最新项目快照，对比上一条变更记录
// 的 Config/GitCommit 判定配置是否变化，落一条新的变更日志。
func (c *EventCoordinator) HandleProjectChanged(d any, e biz.EventKey) error {
	if changedData, ok := d.(*biz.ProjectChangedData); ok {
		proj, err := c.projectRepo.Show(context.TODO(), changedData.ID)
		if err != nil {
			c.logger.Error("[HandleProjectChanged]: ", err)
			return err
		}
		var configChanged bool
		if lastChange, err := c.clRepo.FindLastChangeByProjectID(context.TODO(), changedData.ID); err == nil {
			c.logger.Debug(lastChange, "lastChange")
			configChanged = lastChange.Config != proj.Config || lastChange.GitCommit != proj.GitCommit
		}
		if _, err := c.clRepo.Create(context.TODO(), &biz.CreateChangeLogInput{
			Version:          proj.Version,
			Username:         changedData.Username,
			Config:           proj.Config,
			GitBranch:        proj.GitBranch,
			GitCommit:        proj.GitCommit,
			DockerImage:      proj.DockerImage,
			EnvValues:        proj.EnvValues,
			ExtraValues:      proj.ExtraValues,
			FinalExtraValues: proj.FinalExtraValues,
			GitCommitWebURL:  proj.GitCommitWebURL,
			GitCommitTitle:   proj.GitCommitTitle,
			GitCommitAuthor:  proj.GitCommitAuthor,
			GitCommitDate:    proj.GitCommitDate,
			ConfigChanged:    configChanged,
			ProjectID:        changedData.ID,
		}); err != nil {
			c.logger.Error(err)
		}
	}
	return nil
}

// HandleProjectDeleted 处理项目删除事件：向所有 ws 客户端广播 reload 消息。
func (c *EventCoordinator) HandleProjectDeleted(d any, e biz.EventKey) error {
	input := d.(*biz.ProjectDeletedPayload)
	c.toAll(&websocket_pb.WsReloadProjectsResponse{
		Metadata:    &websocket_pb.Metadata{Type: websocket_pb.Type_ReloadProjects},
		NamespaceId: int32(input.NamespaceID),
	})
	c.logger.Debug("event handled: ", e.String(), d)
	return nil
}

// HandleAuditLog 处理审计日志事件：委托 data.eventRepo 落库。监听注册在
// 事件用例层（listen），持久化实现保留在 data 层，错误原样返回由
// dispatcher 统一消费。
func (c *EventCoordinator) HandleAuditLog(d any, e biz.EventKey) error {
	return c.eventRepo.HandleAuditLog(d, e)
}
