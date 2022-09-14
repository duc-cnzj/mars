package memory

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	websocket_pb "github.com/duc-cnzj/mars-client/v4/websocket"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/plugins/wssender"
)

var memorySenderName = "ws_sender_memory"

func init() {
	dr := &memorySender{}
	plugins.RegisterPlugin(dr.Name(), dr)
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
	roomIDs map[int64]map[int64]map[string][]labels.Selector
	// socketID => map[nsID]struct{}
	idRooms map[string]map[int64]struct{}
}

func (ms *memorySender) Add(uid, id string) {
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

func (ms *memorySender) Initialize(args map[string]any) error {
	ms.conns = map[string]*Conn{}
	ms.idRooms = make(map[string]map[int64]struct{})
	ms.roomIDs = make(map[int64]map[int64]map[string][]labels.Selector)
	mlog.Info("[Plugin]: " + ms.Name() + " plugin Initialize...")
	return nil
}

func (ms *memorySender) Destroy() error {
	mlog.Info("[Plugin]: " + ms.Name() + " plugin Destroy...")
	return nil
}

func (ms *memorySender) New(uid, id string) contracts.PubSub {
	ms.Add(uid, id)
	return &memoryPubSub{uid: uid, id: id, manager: ms}
}

type memoryPubSub struct {
	manager *memorySender
	uid     string
	id      string
}

func (p *memoryPubSub) Run(ctx context.Context) error {
	return nil
}

func (p *memoryPubSub) Publish(nsID int64, pod *corev1.Pod) error {
	p.manager.roomMu.RLock()
	defer p.manager.roomMu.RUnlock()
	projectMap, ok := p.manager.roomIDs[nsID]

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
								marshal, _ := proto.Marshal(&websocket_pb.WsProjectPodEventResponse{
									Metadata: &websocket_pb.Metadata{
										Id:     socketID,
										Type:   websocket_pb.Type_ProjectPodEvent,
										End:    true,
										Result: websocket_pb.ResultType_Success,
										To:     plugins.ToSelf,
									},
									ProjectId: pid,
								})
								mlog.Debugf("publish to: (%d---%d)", nsID, pid)
								conn.ch <- marshal
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
	var pmodel models.Project
	app.DB().Select("id", "namespace_id", "pod_selectors").First(&pmodel, projectID)
	p.manager.roomMu.Lock()
	defer p.manager.roomMu.Unlock()
	mlog.Warningf("Join to: (%d---%d)", pmodel.NamespaceId, projectID)
	var (
		nsID = int64(pmodel.NamespaceId)
	)
	var selectors []labels.Selector
	for _, s := range pmodel.GetPodSelectors() {
		parse, _ := labels.Parse(s)
		selectors = append(selectors, parse)
	}
	_, projectMapFound := p.manager.roomIDs[nsID]
	if !projectMapFound {
		p.manager.roomIDs[nsID] = make(map[int64]map[string][]labels.Selector)
	}
	_, socketIDMapFound := p.manager.roomIDs[nsID][projectID]
	if !socketIDMapFound {
		p.manager.roomIDs[nsID][projectID] = make(map[string][]labels.Selector)
	}
	p.manager.roomIDs[nsID][projectID][p.id] = selectors
	rooms, nsIDFound := p.manager.idRooms[p.id]
	if nsIDFound {
		rooms[nsID] = struct{}{}
	} else {
		p.manager.idRooms[p.id] = map[int64]struct{}{nsID: {}}
	}
	return nil
}

func (p *memoryPubSub) Leave(nsID, projectID int64) error {
	p.manager.roomMu.Lock()
	defer p.manager.roomMu.Unlock()
	mlog.Warningf("Leave to: (%d---%d)", nsID, projectID)
	rooms, ok := p.manager.idRooms[p.id]
	if ok {
		delete(rooms, nsID)
	}
	if len(p.manager.idRooms[p.id]) == 0 {
		delete(p.manager.idRooms, p.id)
	}
	mm, ok := p.manager.roomIDs[nsID]
	if ok {
		delete(mm, projectID)
	}
	if len(p.manager.roomIDs[nsID]) == 0 {
		delete(p.manager.roomIDs, nsID)
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

func (p *memoryPubSub) ToSelf(wsResponse contracts.WebsocketMessage) error {
	p.manager.RLock()
	defer p.manager.RUnlock()
	marshal, _ := proto.Marshal(wsResponse)
	conn, ok := p.manager.conns[p.id]
	if ok {
		conn.ch <- marshal
	}
	return nil
}

func (p *memoryPubSub) ToAll(wsResponse contracts.WebsocketMessage) error {
	p.manager.RLock()
	defer p.manager.RUnlock()
	marshal, _ := proto.Marshal(wsResponse)

	for _, s := range p.manager.conns {
		s.ch <- marshal
	}
	return nil
}

func (p *memoryPubSub) ToOthers(wsResponse contracts.WebsocketMessage) error {
	p.manager.RLock()
	defer p.manager.RUnlock()
	marshal, _ := proto.Marshal(wsResponse)

	for _, s := range p.manager.conns {
		if s.id != p.id {
			s.ch <- marshal
		}
	}
	return nil
}

func (p *memoryPubSub) Close() error {
	mlog.Debugf(fmt.Sprintf("[Websocket]: Closed, uid: %s, id: %s", p.uid, p.id))
	p.manager.Delete(p.uid, p.id)
	return nil
}

func (p *memoryPubSub) Subscribe() <-chan []byte {
	p.manager.RLock()
	defer p.manager.RUnlock()
	return p.manager.conns[p.ID()].ch
}
