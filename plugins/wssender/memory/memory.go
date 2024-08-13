package memory

import (
	"context"
	"fmt"
	"sync"

	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/project"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/plugins/wssender"
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

type memorySender struct {
	sync.RWMutex

	conns map[string]*Conn

	roomMu sync.RWMutex
	// nsID => map[projectID]map[socketID]selectors
	roomIDs map[int32]map[int32]map[string][]labels.Selector
	// socketID => map[nsID]struct{}
	idRooms map[string]map[int32]struct{}
	logger  mlog.Logger
	db      *ent.Client
}

func (ms *memorySender) Add(uid, id string) {
	ms.logger.Debugf("Register: %s, %s", uid, id)
	if uid == "" || id == "" {
		return
	}

	ms.Lock()
	defer ms.Unlock()
	st := &Conn{id: id, uid: uid, ch: make(chan []byte, wssender.MessageChSize)}
	if _, ok := ms.conns[id]; !ok {
		ms.conns[id] = st
	}
}

func (ms *memorySender) Delete(uid string, id string) {
	ms.Lock()
	defer ms.Unlock()
	delete(ms.conns, id)
}

func (ms *memorySender) Name() string {
	return memorySenderName
}

func (ms *memorySender) Initialize(app application.App, args map[string]any) error {
	ms.conns = map[string]*Conn{}
	ms.idRooms = make(map[string]map[int32]struct{})
	ms.roomIDs = make(map[int32]map[int32]map[string][]labels.Selector)
	ms.logger = app.Logger().WithModule("plugins/ws_sender_memory")
	ms.db = app.DB()
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
}

func (p *memoryPubSub) Run(ctx context.Context) error {
	return nil
}

func (p *memoryPubSub) Publish(nsID int64, pod *corev1.Pod) error {
	p.manager.roomMu.RLock()
	defer p.manager.roomMu.RUnlock()
	projectMap, ok := p.manager.roomIDs[int32(nsID)]

	if ok {
		for pid, socketIDMap := range projectMap {
			for socketID, selectors := range socketIDMap {
				func(socketID string, selectors []labels.Selector) {
					p.manager.RLock()
					defer p.manager.RUnlock()
					conn, ok := p.manager.conns[socketID]
					if ok {
						for _, selector := range selectors {
							if selector.Matches(labels.Set(pod.Labels)) {
								p.logger.Debugf("publish to: (%d---%d)", nsID, pid)
								conn.ch <- wssender.TransformToResponse(&websocket_pb.WsProjectPodEventResponse{
									Metadata: &websocket_pb.Metadata{
										Id:     socketID,
										Type:   websocket_pb.Type_ProjectPodEvent,
										End:    true,
										Result: websocket_pb.ResultType_Success,
										To:     websocket_pb.To_ToSelf,
									},
									ProjectId: pid,
								})
								return
							}
						}
					}
				}(socketID, selectors)
			}
		}

	}
	return nil
}

func (p *memoryPubSub) Join(projectID int64) error {
	pmodel, err := p.db.Project.Query().WithNamespace().Where(project.ID(int(projectID))).Only(context.TODO())
	if err != nil {
		return err
	}
	p.manager.roomMu.Lock()
	defer p.manager.roomMu.Unlock()
	p.logger.Warningf("Join to: (%d---%d)", pmodel.NamespaceID, projectID)
	var (
		nsID = int64(pmodel.Edges.Namespace.ID)
	)
	var selectors []labels.Selector
	for _, s := range pmodel.PodSelectors {
		parse, _ := labels.Parse(s)
		selectors = append(selectors, parse)
	}
	_, projectMapFound := p.manager.roomIDs[int32(nsID)]
	if !projectMapFound {
		p.manager.roomIDs[int32(nsID)] = make(map[int32]map[string][]labels.Selector)
	}
	_, socketIDMapFound := p.manager.roomIDs[int32(nsID)][int32(projectID)]
	if !socketIDMapFound {
		p.manager.roomIDs[int32(nsID)][int32(projectID)] = make(map[string][]labels.Selector)
	}
	p.manager.roomIDs[int32(nsID)][int32(projectID)][p.id] = selectors
	rooms, nsIDFound := p.manager.idRooms[p.id]
	if nsIDFound {
		rooms[int32(nsID)] = struct{}{}
	} else {
		p.manager.idRooms[p.id] = map[int32]struct{}{int32(nsID): {}}
	}
	return nil
}

func (p *memoryPubSub) Leave(nsID, projectID int64) error {
	p.manager.roomMu.Lock()
	defer p.manager.roomMu.Unlock()
	p.logger.Warningf("Leave to: (%d---%d)", nsID, projectID)
	rooms, ok := p.manager.idRooms[p.id]
	if ok {
		delete(rooms, int32(nsID))
	}
	if len(p.manager.idRooms[p.id]) == 0 {
		delete(p.manager.idRooms, p.id)
	}
	mm, ok := p.manager.roomIDs[int32(nsID)]
	if ok {
		delete(mm, int32(projectID))
	}
	if len(p.manager.roomIDs[int32(nsID)]) == 0 {
		delete(p.manager.roomIDs, int32(nsID))
	}
	return nil
}

func (p *memoryPubSub) Info() any {
	p.manager.RLock()
	defer p.manager.RUnlock()
	return p.manager.conns
}

func (p *memoryPubSub) Uid() string {
	return p.uid
}

func (p *memoryPubSub) ID() string {
	return p.id
}

func (p *memoryPubSub) ToSelf(wsResponse application.WebsocketMessage) error {
	p.manager.RLock()
	defer p.manager.RUnlock()
	conn, ok := p.manager.conns[p.id]
	if ok {
		conn.ch <- wssender.TransformToResponse(wsResponse)
	}
	return nil
}

func (p *memoryPubSub) ToAll(wsResponse application.WebsocketMessage) error {
	p.manager.RLock()
	defer p.manager.RUnlock()

	for _, s := range p.manager.conns {
		s.ch <- wssender.TransformToResponse(wsResponse)
	}
	return nil
}

func (p *memoryPubSub) ToOthers(wsResponse application.WebsocketMessage) error {
	p.manager.RLock()
	defer p.manager.RUnlock()

	for _, s := range p.manager.conns {
		if s.id != p.id {
			s.ch <- wssender.TransformToResponse(wsResponse)
		}
	}
	return nil
}

func (p *memoryPubSub) Close() error {
	p.logger.Debugf(fmt.Sprintf("[Websocket]: Closed, uid: %s, id: %s", p.uid, p.id))
	p.manager.Delete(p.uid, p.id)
	return nil
}

func (p *memoryPubSub) Subscribe() <-chan []byte {
	p.manager.RLock()
	defer p.manager.RUnlock()
	return p.manager.conns[p.ID()].ch
}
