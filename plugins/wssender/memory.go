package wssender

import (
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/plugins"
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
	conns map[string]map[string]*Conn
}

func (ms *memorySender) Add(uid, id string) {
	if uid == "" || id == "" {
		return
	}

	ms.Lock()
	defer ms.Unlock()
	st := &Conn{id: id, uid: uid, ch: make(chan []byte, messageChSize)}
	if _, ok := ms.conns[uid]; ok {
		ms.conns[uid][id] = st
	} else {
		ms.conns[uid] = map[string]*Conn{id: st}
	}
}

func (ms *memorySender) Delete(uid string, id string) {
	ms.Lock()
	defer ms.Unlock()
	if m, ok := ms.conns[uid]; ok {
		if conn, ok := m[id]; ok {
			close(conn.ch)
			delete(m, id)
		}

		if len(m) == 0 {
			delete(ms.conns, uid)
		}
	}
}

func (ms *memorySender) Name() string {
	return memorySenderName
}

func (ms *memorySender) Initialize(args map[string]interface{}) error {
	ms.conns = map[string]map[string]*Conn{}
	mlog.Info("[Plugin]: " + ms.Name() + " plugin Initialize...")
	return nil
}

func (ms *memorySender) Destroy() error {
	mlog.Info("[Plugin]: " + ms.Name() + " plugin Destroy...")
	return nil
}

func (ms *memorySender) New(uid, id string) plugins.PubSub {
	ms.Add(uid, id)
	return &memoryPubSub{uid: uid, id: id, manager: ms}
}

type memoryPubSub struct {
	manager *memorySender
	uid     string
	id      string
}

func (p *memoryPubSub) Info() interface{} {
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

func (p *memoryPubSub) ToSelf(wsResponse proto.Message) error {
	p.manager.RLock()
	defer p.manager.RUnlock()
	marshal, _ := proto.Marshal(wsResponse)
	if pp, ok := p.manager.conns[p.uid]; ok {
		if c, ok := pp[p.id]; ok {
			c.ch <- marshal
		}
	}
	return nil
}

func (p *memoryPubSub) ToAll(wsResponse proto.Message) error {
	p.manager.RLock()
	defer p.manager.RUnlock()
	marshal, _ := proto.Marshal(wsResponse)

	for _, m := range p.manager.conns {
		for _, s := range m {
			s.ch <- marshal
		}
	}
	return nil
}

func (p *memoryPubSub) ToOthers(wsResponse proto.Message) error {
	p.manager.RLock()
	defer p.manager.RUnlock()
	marshal, _ := proto.Marshal(wsResponse)

	for _, m := range p.manager.conns {
		for _, s := range m {
			if s.id != p.id {
				s.ch <- marshal
			}
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
	m := p.manager.conns[p.Uid()]
	s := m[p.ID()]
	return s.ch
}
