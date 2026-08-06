package memory

import (
	"context"
	"sync"

	"github.com/duc-cnzj/mars/v6/internal/application"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/project"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/plugins/wssender"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// memorySenderName 插件注册名。
var memorySenderName = "ws_sender_memory"

func init() {
	dr := &memorySender{}
	application.RegisterPlugin(dr.Name(), dr)
}

// Conn 表示一个已注册的 websocket 连接：id 为连接标识，uid 为用户标识，ch 为消息缓冲通道。
type Conn struct {
	id  string
	uid string
	ch  chan []byte
}

// 房间订阅的嵌套表结构：namespace → project → socket → selectors。
type (
	socketSubscriptions  map[string][]labels.Selector   // socketID -> selectors per project
	projectSubscriptions map[int32]socketSubscriptions  // projectID -> socket subs
	namespaceRooms       map[int32]projectSubscriptions // nsID -> project subs
)

// memorySender 是内存版 WsSender：维护连接表与订阅房间，同一进程内广播消息。
type memorySender struct {
	connMu sync.RWMutex
	conns  map[string]*Conn

	roomMu  sync.RWMutex
	rooms   namespaceRooms                // nsID -> projectID -> socketID -> selectors
	idRooms map[string]map[int32]struct{} // socketID -> set of nsIDs

	logger mlog.Logger
	db     *ent.Client
}

// Add 注册一个连接；uid 或 id 为空或已存在时静默忽略。
func (ms *memorySender) Add(uid, id string) {
	ms.logger.Debugf("Register: %s, %s", uid, id)
	if uid == "" || id == "" {
		return
	}

	ms.connMu.Lock()
	defer ms.connMu.Unlock()
	if _, ok := ms.conns[id]; ok {
		return
	}
	ms.conns[id] = &Conn{id: id, uid: uid, ch: make(chan []byte, wssender.MessageChSize)}
}

// Delete 移除并关闭指定连接的通道。
func (ms *memorySender) Delete(uid string, id string) {
	ms.connMu.Lock()
	defer ms.connMu.Unlock()
	if c, ok := ms.conns[id]; ok {
		close(c.ch)
		delete(ms.conns, id)
	}
}

// Name 返回插件名 ws_sender_memory。
func (ms *memorySender) Name() string {
	return memorySenderName
}

// Initialize 初始化连接表、房间表与日志器。
func (ms *memorySender) Initialize(app application.PluginApp, args map[string]any) error {
	ms.db = app.Data().DB()
	ms.conns = make(map[string]*Conn)
	ms.idRooms = make(map[string]map[int32]struct{})
	ms.rooms = make(namespaceRooms)
	ms.logger = app.Logger().WithModule("plugins/ws_sender_memory")
	ms.logger.Info("[Plugin]: " + ms.Name() + " plugin Initialize...")
	return nil
}

// Destroy 输出销毁日志。
func (ms *memorySender) Destroy() error {
	ms.logger.Info("[Plugin]: " + ms.Name() + " plugin Destroy...")
	return nil
}

// New 注册连接并返回对应的 memoryPubSub 实例。
func (ms *memorySender) New(uid, id string) application.PubSub {
	ms.Add(uid, id)
	return &memoryPubSub{
		db:      ms.db,
		manager: ms,
		uid:     uid,
		id:      id,
		logger:  ms.logger,
	}
}

// memoryPubSub 是单个连接对应的 PubSub：消息经 sender 的连接表在进程内分发。
type memoryPubSub struct {
	db      *ent.Client
	manager *memorySender
	uid     string
	id      string
	logger  mlog.Logger

	closeOnce sync.Once
}

// Run 内存模式下无需后台分发，直接返回 nil。
func (p *memoryPubSub) Run(ctx context.Context) error {
	return nil
}

// Publish 将 pod 事件按订阅选择器匹配并投递到对应连接通道。
func (p *memoryPubSub) Publish(nsID int64, pod *corev1.Pod) error {
	// 第一阶段：在 roomMu 下收集每个 socket 的 pidSelectors（只读，快）。
	p.manager.roomMu.RLock()
	projectMap, ok := p.manager.rooms[int32(nsID)]
	if !ok {
		p.manager.roomMu.RUnlock()
		return nil
	}

	socketPIDs := make(map[string]map[int32][]labels.Selector)
	for pid, socketIDMap := range projectMap {
		for socketID, selectors := range socketIDMap {
			if _, ok := socketPIDs[socketID]; !ok {
				socketPIDs[socketID] = make(map[int32][]labels.Selector)
			}
			socketPIDs[socketID][pid] = selectors
		}
	}
	p.manager.roomMu.RUnlock()

	// 第二阶段：在 connMu 下匹配选择器并投递。
	p.manager.connMu.RLock()
	defer p.manager.connMu.RUnlock()

	podLabels := labels.Set(pod.Labels)
	for socketID, pidSelectors := range socketPIDs {
		conn, ok := p.manager.conns[socketID]
		if !ok {
			continue
		}
		p.logger.Debugf("publish to: (%d---%s)", nsID, socketID)
		wssender.MatchSelectorsAndSend(conn.ch, podLabels, pidSelectors, socketID, conn.uid, p.logger)
	}
	return nil
}

// Join 将当前连接加入项目对应的房间并登记其 pod 选择器。
func (p *memoryPubSub) Join(projectID int64) error {
	pmodel, err := p.db.Project.Query().WithNamespace().Where(project.ID(int(projectID))).Only(context.TODO())
	if err != nil {
		return err
	}

	var selectors []labels.Selector
	for _, s := range pmodel.PodSelectors {
		parse, err := labels.Parse(s)
		if err != nil {
			p.logger.Errorf("[Memory] invalid pod selector %q: %v", s, err)
			continue
		}
		selectors = append(selectors, parse)
	}

	nsID := int64(pmodel.Edges.Namespace.ID)

	p.manager.roomMu.Lock()
	defer p.manager.roomMu.Unlock()

	p.logger.Debugf("Join to: (%d---%d)", nsID, projectID)

	if _, ok := p.manager.rooms[int32(nsID)]; !ok {
		p.manager.rooms[int32(nsID)] = make(projectSubscriptions)
	}
	if _, ok := p.manager.rooms[int32(nsID)][int32(projectID)]; !ok {
		p.manager.rooms[int32(nsID)][int32(projectID)] = make(socketSubscriptions)
	}
	p.manager.rooms[int32(nsID)][int32(projectID)][p.id] = selectors

	if _, ok := p.manager.idRooms[p.id]; !ok {
		p.manager.idRooms[p.id] = make(map[int32]struct{})
	}
	p.manager.idRooms[p.id][int32(nsID)] = struct{}{}

	return nil
}

// Leave 将当前连接移出项目房间，并清理空房间与空 idRooms。
func (p *memoryPubSub) Leave(nsID, projectID int64) error {
	p.manager.roomMu.Lock()
	defer p.manager.roomMu.Unlock()

	p.logger.Warningf("Leave to: (%d---%d)", nsID, projectID)

	if rooms, ok := p.manager.idRooms[p.id]; ok {
		delete(rooms, int32(nsID))
		if len(p.manager.idRooms[p.id]) == 0 {
			delete(p.manager.idRooms, p.id)
		}
	}

	if nsRoom, ok := p.manager.rooms[int32(nsID)]; ok {
		if subs, ok := nsRoom[int32(projectID)]; ok {
			delete(subs, p.id)
			if len(subs) == 0 {
				delete(nsRoom, int32(projectID))
			}
		}
		if len(p.manager.rooms[int32(nsID)]) == 0 {
			delete(p.manager.rooms, int32(nsID))
		}
	}

	return nil
}

// Info 返回 id→uid 的连接快照。
func (p *memoryPubSub) Info() any {
	p.manager.connMu.RLock()
	defer p.manager.connMu.RUnlock()
	infos := make(map[string]string, len(p.manager.conns))
	for id, conn := range p.manager.conns {
		infos[id] = conn.uid
	}
	return infos
}

// Uid 返回连接对应用户标识。
func (p *memoryPubSub) Uid() string {
	return p.uid
}

// ID 返回连接标识。
func (p *memoryPubSub) ID() string {
	return p.id
}

// ToSelf 只向当前连接发送消息。
func (p *memoryPubSub) ToSelf(wsResponse application.WebsocketMessage) error {
	p.manager.connMu.RLock()
	defer p.manager.connMu.RUnlock()
	conn, ok := p.manager.conns[p.id]
	if ok {
		wssender.SendOrDrop(conn.ch, wssender.TransformToResponse(wsResponse), p.logger, p.id)
	}
	return nil
}

// ToAll 向全部连接广播消息。
func (p *memoryPubSub) ToAll(wsResponse application.WebsocketMessage) error {
	p.manager.connMu.RLock()
	defer p.manager.connMu.RUnlock()

	data := wssender.TransformToResponse(wsResponse)
	for _, s := range p.manager.conns {
		wssender.SendOrDrop(s.ch, data, p.logger, s.id)
	}
	return nil
}

// Close 清理当前连接的房间订阅并从 sender 移除连接，保证只执行一次。
func (p *memoryPubSub) Close() error {
	p.closeOnce.Do(func() {
		p.logger.Debugf("[Websocket]: Closed, uid: %s, id: %s", p.uid, p.id)

		// 清理房间订阅，避免 sender 级 rooms/idRooms 泄漏：若 Close() 之前未对每个
		// 已加入的项目调用 Leave()，不清理会令单例 sender 的房间表残留过期条目。
		p.manager.roomMu.Lock()
		if nsIDs, ok := p.manager.idRooms[p.id]; ok {
			for nsID := range nsIDs {
				if nsRoom, ok := p.manager.rooms[nsID]; ok {
					for pid := range nsRoom {
						delete(nsRoom[pid], p.id)
						if len(nsRoom[pid]) == 0 {
							delete(nsRoom, pid)
						}
					}
					if len(p.manager.rooms[nsID]) == 0 {
						delete(p.manager.rooms, nsID)
					}
				}
			}
			delete(p.manager.idRooms, p.id)
		}
		p.manager.roomMu.Unlock()

		p.manager.Delete(p.uid, p.id)
	})
	return nil
}

// Subscribe 返回当前连接的接收通道；未知 id 时返回已关闭通道。
func (p *memoryPubSub) Subscribe() <-chan []byte {
	p.manager.connMu.RLock()
	defer p.manager.connMu.RUnlock()
	conn, ok := p.manager.conns[p.ID()]
	if !ok {
		p.logger.Warningf("[Websocket]: Subscribe called for unknown id: %s", p.id)
		ch := make(chan []byte)
		close(ch)
		return ch
	}
	return conn.ch
}
