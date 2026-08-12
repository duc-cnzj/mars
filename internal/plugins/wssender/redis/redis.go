package redis

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/plugins/wssender"
	"github.com/go-redis/redis/v8"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// redisSenderName 插件注册名。
var redisSenderName = "ws_sender_redis"

func init() {
	dr := &redisSender{}
	app.RegisterPlugin(dr.Name(), dr)
}

// ---------------------------------------------------------------------------
// 共享订阅 redisSender（全部 ws 消息共用一条 Redis PubSub 连接）
// ---------------------------------------------------------------------------

// redisSender 是 Redis PubSub 版 WsSender：全实例共享一条订阅连接，经 dispatcher 扇出到本地订阅者。
type redisSender struct {
	rds         *redis.Client
	logger      mlog.Logger
	projectRepo biz.ProjectRepo

	// 一条共享 PubSub 连接，而非每个连接一条。
	wsPubSub *redis.PubSub
	msgCh    <-chan *redis.Message

	mu   sync.RWMutex
	subs map[string]*subEntry // id → local subscriber

	ctx    context.Context
	cancel context.CancelFunc
}

// subEntry 是 dispatcher 的本地订阅项：ch 为消息缓冲通道，uid/id 标识连接。
type subEntry struct {
	ch  chan []byte
	uid string
	id  string
}

// Name 返回插件名 ws_sender_redis。
func (p *redisSender) Name() string {
	return redisSenderName
}

// Initialize 从 args 读取 addr/password/db 建立 Redis 客户端，
// 订阅广播房间并启动 dispatcher 扇出 goroutine。
func (p *redisSender) Initialize(pluginApp app.PluginApp, args map[string]any) error {
	addr, ok := args["addr"].(string)
	if !ok || addr == "" {
		return errors.New("redisSender need valid addr")
	}
	pwd, _ := args["password"].(string)
	db, _ := args["db"].(int)

	p.projectRepo = pluginApp.ProjectRepo()
	p.logger = pluginApp.Logger()

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       db,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return err
	}
	p.rds = rdb

	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.subs = make(map[string]*subEntry)

	// 一条共享 PubSub 订阅广播房间。绑定 p.ctx：Destroy 取消时订阅随之关闭，不再飘在空里。
	p.wsPubSub = p.rds.Subscribe(p.ctx)
	// Subscribe 仅写 SUBSCRIBE 命令：Ping 刚成功、连接健康，写入不会失败；后续连接故障
	// 经 Channel() 关闭由 dispatcher 退出兜底，故此处不检查错误。
	p.wsPubSub.Subscribe(p.ctx, wssender.BroadcastRoom)
	p.msgCh = p.wsPubSub.Channel()
	go p.dispatcher()

	p.logger.Info("[Plugin]: " + p.Name() + " plugin Initialize...")
	return nil
}

// Destroy 取消 dispatcher 并关闭订阅与客户端连接。
func (p *redisSender) Destroy() error {
	p.cancel()
	p.wsPubSub.Close()
	p.rds.Close()
	p.logger.Info("[Plugin]: " + p.Name() + " plugin Destroy...")
	return nil
}

// dispatcher 从共享 PubSub 通道读取消息并扇出到本地订阅者。
// 全部连接共用一个 goroutine，而非每连接一个 goroutine。
func (p *redisSender) dispatcher() {
	defer p.logger.HandlePanic("[PubSub]: dispatcher")
	for {
		select {
		case <-p.ctx.Done():
			return
		case msg, ok := <-p.msgCh:
			if !ok {
				return
			}
			message, err := wssender.DecodeMessage([]byte(msg.Payload))
			if err != nil {
				p.logger.Error(err)
				continue
			}

			p.mu.RLock()
			switch message.To {
			case websocket_pb.To_ToSelf:
				if sub, ok := p.subs[message.ID]; ok {
					wssender.SendOrDrop(sub.ch, message.Data, p.logger, message.ID)
				}
			case websocket_pb.To_ToAll:
				for _, sub := range p.subs {
					wssender.SendOrDrop(sub.ch, message.Data, p.logger, sub.id)
				}
			// To_ToOthers：跳过消息来源连接。ToOthers 方法已删除（无生产调用方），
			// 此分支保留为防御代码，供未来投递语义复用，仍需测试覆盖。
			case websocket_pb.To_ToOthers:
				for _, sub := range p.subs {
					if sub.id != message.ID {
						wssender.SendOrDrop(sub.ch, message.Data, p.logger, sub.id)
					}
				}
			}
			p.mu.RUnlock()
		}
	}
}

// New 注册本地订阅项与用户直连频道，返回 rdsPubSub 实例。
func (p *redisSender) New(uid, id string) app.PubSub {
	ch := make(chan []byte, wssender.MessageChSize)

	// 注册到共享 dispatcher。
	p.mu.Lock()
	p.subs[id] = &subEntry{ch: ch, uid: uid, id: id}
	p.mu.Unlock()

	// 共享连接订阅该用户的直连频道（供跨实例 ToSelf 使用）。
	p.wsPubSub.Subscribe(p.ctx, id)

	pem := &podEventManagers{
		ctx:          p.ctx,
		logger:       p.logger.WithModule("plugins/ws_sender_redis"),
		projectRepo:  p.projectRepo,
		ch:           ch,
		id:           id,
		uid:          uid,
		rds:          p.rds,
		channelRefs:  make(map[string]int),
		pubSub:       p.rds.Subscribe(p.ctx),
		pidSelectors: make(map[int32][]labels.Selector),
	}

	return &rdsPubSub{
		logger:                    p.logger,
		ch:                        ch,
		rds:                       p.rds,
		uid:                       uid,
		id:                        id,
		manager:                   p,
		ProjectPodEventSubscriber: pem,
		ProjectPodEventPublisher:  pem,
	}
}

// ---------------------------------------------------------------------------
// rdsPubSub —— 薄包装，无每连接独立的 Redis 订阅
// ---------------------------------------------------------------------------

type rdsPubSub struct {
	logger    mlog.Logger
	rds       *redis.Client
	manager   *redisSender
	uid, id   string
	ch        chan []byte
	closeOnce sync.Once

	app.ProjectPodEventSubscriber
	app.ProjectPodEventPublisher
}

// ID 返回连接标识。
func (p *rdsPubSub) ID() string {
	return p.id
}

// Uid 返回连接对应用户标识。
func (p *rdsPubSub) Uid() string {
	return p.uid
}

// Info 返回订阅者数量与当前连接 id。
func (p *rdsPubSub) Info() any {
	p.manager.mu.RLock()
	defer p.manager.mu.RUnlock()
	return map[string]any{
		"subscribers": len(p.manager.subs),
		"id":          p.id,
	}
}

// Close 将当前连接从 dispatcher 移除并退订用户直连频道，保证只执行一次。
func (p *rdsPubSub) Close() error {
	p.logger.Debugf("[Websocket]: Closed, uid: %v id: %v", p.uid, p.id)

	p.closeOnce.Do(func() {
		// 先从共享 dispatcher 移除，保证不再有新消息路由到本连接。
		p.manager.mu.Lock()
		delete(p.manager.subs, p.id)
		p.manager.mu.Unlock()

		p.manager.wsPubSub.Unsubscribe(p.manager.ctx, p.id)

		// 不要 close(p.ch)：dispatcher 与 podEventManagers.Run 两个 goroutine 都可能
		// 向 p.ch 发送，send-on-closed-channel 会 panic；消费者（websocket write）已
		// 监听 ctx.Done() 退出，不依赖 channel 关闭信号。写者全部退出后 channel 由 GC 回收。
	})
	return nil
}

// ToSelf 发布到用户直连频道，仅当前连接可收到。
func (p *rdsPubSub) ToSelf(wsResponse app.WebsocketMessage) error {
	return p.to(wsResponse, websocket_pb.To_ToSelf)
}

// ToAll 发布到广播频道，全部订阅者均可收到。
func (p *rdsPubSub) ToAll(wsResponse app.WebsocketMessage) error {
	return p.to(wsResponse, websocket_pb.To_ToAll)
}

// to 发布消息到 Redis：共享 dispatcher 负责扇出。不修改 wsResponse。
func (p *rdsPubSub) to(response app.WebsocketMessage, to websocket_pb.To) error {
	room := wssender.BroadcastRoom
	if to == websocket_pb.To_ToSelf {
		room = p.id
	}
	return p.rds.Publish(context.Background(), room, wssender.ProtoToMessage(response, p.id, to).Marshal()).Err()
}

// Subscribe 返回本地通道，无每连接的 Redis 订阅。
func (p *rdsPubSub) Subscribe() <-chan []byte {
	return p.ch
}

// ---------------------------------------------------------------------------
// Pod 事件类型（保持不变，仍是每连接独立的 Redis 订阅）
// ---------------------------------------------------------------------------

// podEventManagers 管理项目 pod 事件的订阅与发布：每个连接独立的 PubSub 订阅对应命名空间频道。
type podEventManagers struct {
	// ctx 是插件生命周期 ctx（New 时从 redisSender 注入）：订阅/退订绑它，
	// 插件 Destroy 时这些 Redis 操作可取消，而非飘在空里的 TODO。
	// Publish 是 fire-and-forget 广播，不绑 ctx（见 to()/Publish 注释）。
	ctx         context.Context
	projectRepo biz.ProjectRepo
	logger      mlog.Logger
	id          string
	uid         string
	rds         *redis.Client
	pubSub      *redis.PubSub

	ch chan []byte

	mu          sync.RWMutex
	channelRefs map[string]int // reference count for Join/Leave symmetry

	pmu          sync.RWMutex
	pidSelectors map[int32][]labels.Selector
}

// Publish 将 pod 事件序列化后发布到对应命名空间频道。
func (p *podEventManagers) Publish(nsID int64, pod *v1.Pod) error {
	channel := wssender.GetProjectPodEventRoom(nsID)
	// ProjectPodEventObj 的字段全部可序列化（nil Pod 序列化为 null），Marshal 恒不失败。
	marshal, _ := json.Marshal(&wssender.ProjectPodEventObj{
		Channel:     channel,
		NamespaceID: nsID,
		Pod:         pod,
	})
	return p.rds.Publish(context.Background(), channel, marshal).Err()
}

// Join 订阅项目所在命名空间频道（首个引用时），并登记该项目的 pod 选择器。
func (p *podEventManagers) Join(projectID int64) error {
	pmodel, err := p.projectRepo.Show(context.TODO(), int(projectID))
	if err != nil {
		return err
	}

	channel := wssender.GetProjectPodEventRoom(pmodel.Namespace.ID)

	p.mu.Lock()
	p.channelRefs[channel]++
	if p.channelRefs[channel] == 1 {
		// 首个引用才真正订阅。
		if err := p.pubSub.Subscribe(p.ctx, channel); err != nil {
			p.channelRefs[channel]--
			p.mu.Unlock()
			return err
		}
	}
	p.mu.Unlock()

	p.pmu.Lock()
	var selectors []labels.Selector
	for _, s := range pmodel.PodSelectors {
		parse, err := labels.Parse(s)
		if err != nil {
			p.logger.Errorf("[Redis] invalid pod selector %q: %v", s, err)
			continue
		}
		selectors = append(selectors, parse)
	}
	p.pidSelectors[int32(projectID)] = selectors
	p.pmu.Unlock()

	return nil
}

// Leave 退订命名空间频道（引用计数归零时），并移除项目选择器。
func (p *podEventManagers) Leave(nsID int64, projectID int64) error {
	channel := wssender.GetProjectPodEventRoom(nsID)

	p.mu.Lock()
	p.channelRefs[channel]--
	if p.channelRefs[channel] <= 0 {
		delete(p.channelRefs, channel)
		if err := p.pubSub.Unsubscribe(p.ctx, channel); err != nil {
			p.mu.Unlock()
			return err
		}
	}
	p.mu.Unlock()

	p.pmu.Lock()
	delete(p.pidSelectors, int32(projectID))
	p.pmu.Unlock()

	return nil
}

// Run 消费 pod 事件频道，按订阅选择器匹配并投递到连接通道；ctx 取消时退出。
func (p *podEventManagers) Run(ctx context.Context) error {
	defer p.pubSub.Close()
	ch := p.pubSub.Channel()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case data, ok := <-ch:
			if !ok {
				return errors.New("podEventManagers ch closed")
			}

			p.mu.RLock()
			_, subscribed := p.channelRefs[data.Channel]
			p.mu.RUnlock()
			if !subscribed {
				continue
			}

			var obj wssender.ProjectPodEventObj
			if err := json.Unmarshal([]byte(data.Payload), &obj); err != nil {
				p.logger.Error(err)
				continue
			}
			// nil Pod 没有 labels 可匹配；payload 来自 Redis 外部边界，需防御性跳过。
			if obj.Pod == nil {
				p.logger.Debugf("[Redis] pod event without pod, skip: %s", data.Channel)
				continue
			}

			p.pmu.RLock()
			wssender.MatchSelectorsAndSend(p.ch, labels.Set(obj.Pod.Labels), p.pidSelectors, p.id, p.uid, p.logger)
			p.pmu.RUnlock()
		}
	}
}
