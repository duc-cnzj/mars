package memory

import (
	"context"
	"sync"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/ent"
	"github.com/duc-cnzj/mars/v5/internal/ent/project"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/plugins/wssender"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

var memorySenderName = "ws_sender_memory"

func init() {
	dr := &memorySender{}
	application.RegisterPlugin(dr.Name(), dr)
}

type Conn struct {
	id  string
	uid string
	ch  chan []byte
}

// Memory struct types for room subscription tracking.
type (
	socketSubscriptions  map[string][]labels.Selector     // socketID -> selectors per project
	projectSubscriptions map[int32]socketSubscriptions     // projectID -> socket subs
	namespaceRooms       map[int32]projectSubscriptions    // nsID -> project subs
)

type memorySender struct {
	connMu sync.RWMutex
	conns  map[string]*Conn

	roomMu  sync.RWMutex
	rooms   namespaceRooms                 // nsID -> projectID -> socketID -> selectors
	idRooms map[string]map[int32]struct{}  // socketID -> set of nsIDs

	logger mlog.Logger
	db     *ent.Client
}

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

func (ms *memorySender) Delete(uid string, id string) {
	ms.connMu.Lock()
	defer ms.connMu.Unlock()
	if c, ok := ms.conns[id]; ok {
		close(c.ch)
		delete(ms.conns, id)
	}
}

func (ms *memorySender) Name() string {
	return memorySenderName
}

func (ms *memorySender) Initialize(app application.App, args map[string]any) error {
	ms.db = app.DB()
	ms.conns = make(map[string]*Conn)
	ms.idRooms = make(map[string]map[int32]struct{})
	ms.rooms = make(namespaceRooms)
	ms.logger = app.Logger().WithModule("plugins/ws_sender_memory")
	ms.logger.Info("[Plugin]: " + ms.Name() + " plugin Initialize...")
	return nil
}

func (ms *memorySender) Destroy() error {
	ms.logger.Info("[Plugin]: " + ms.Name() + " plugin Destroy...")
	return nil
}

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

type memoryPubSub struct {
	db      *ent.Client
	manager *memorySender
	uid     string
	id      string
	logger  mlog.Logger

	closeOnce sync.Once
}

func (p *memoryPubSub) Run(ctx context.Context) error {
	return nil
}

func (p *memoryPubSub) Publish(nsID int64, pod *corev1.Pod) error {
	// Phase 1: collect pidSelectors per socket under roomMu (fast, read-only).
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

	// Phase 2: match selectors and send under connMu.
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

func (p *memoryPubSub) Info() any {
	p.manager.connMu.RLock()
	defer p.manager.connMu.RUnlock()
	infos := make(map[string]string, len(p.manager.conns))
	for id, conn := range p.manager.conns {
		infos[id] = conn.uid
	}
	return infos
}

func (p *memoryPubSub) Uid() string {
	return p.uid
}

func (p *memoryPubSub) ID() string {
	return p.id
}

func (p *memoryPubSub) ToSelf(wsResponse application.WebsocketMessage) error {
	p.manager.connMu.RLock()
	defer p.manager.connMu.RUnlock()
	conn, ok := p.manager.conns[p.id]
	if ok {
		wssender.SendOrDrop(conn.ch, wssender.TransformToResponse(wsResponse), p.logger, p.id)
	}
	return nil
}

func (p *memoryPubSub) ToAll(wsResponse application.WebsocketMessage) error {
	p.manager.connMu.RLock()
	defer p.manager.connMu.RUnlock()

	data := wssender.TransformToResponse(wsResponse)
	for _, s := range p.manager.conns {
		wssender.SendOrDrop(s.ch, data, p.logger, s.id)
	}
	return nil
}

func (p *memoryPubSub) ToOthers(wsResponse application.WebsocketMessage) error {
	p.manager.connMu.RLock()
	defer p.manager.connMu.RUnlock()

	data := wssender.TransformToResponse(wsResponse)
	for _, s := range p.manager.conns {
		if s.id != p.id {
			wssender.SendOrDrop(s.ch, data, p.logger, s.id)
		}
	}
	return nil
}

func (p *memoryPubSub) Close() error {
	p.closeOnce.Do(func() {
		p.logger.Debugf("[Websocket]: Closed, uid: %s, id: %s", p.uid, p.id)

		// Clean up room subscriptions so sender-level rooms/idRooms don't leak.
		// Without this, Close() without prior Leave() for each joined project
		// leaves stale entries in the singleton sender's room maps forever.
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
