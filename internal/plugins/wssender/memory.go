package wssender

import (
	"fmt"
	"sync"

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
	ch  chan string
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
	st := &Conn{id: id, uid: uid, ch: make(chan string, messageChSize)}
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

func (ms *memorySender) Initialize() error {
	ms.conns = map[string]map[string]*Conn{}
	mlog.Info(ms.Name() + " plugin Initialize...")
	return nil
}

func (ms *memorySender) Destroy() error {
	mlog.Info(ms.Name() + " plugin Destroy...")
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

func (p *memoryPubSub) ToSelf(wsResponse *plugins.WsResponse) error {
	p.manager.RLock()
	defer p.manager.RUnlock()
	wsResponse.To = plugins.ToSelf
	p.manager.conns[p.uid][p.id].ch <- wsResponse.EncodeToString()
	return nil
}

func (p *memoryPubSub) ToAll(wsResponse *plugins.WsResponse) error {
	p.manager.RLock()
	defer p.manager.RUnlock()
	wsResponse.To = plugins.ToAll
	for _, m := range p.manager.conns {
		for _, s := range m {
			s.ch <- wsResponse.EncodeToString()
		}
	}
	return nil
}

func (p *memoryPubSub) ToOthers(wsResponse *plugins.WsResponse) error {
	p.manager.RLock()
	defer p.manager.RUnlock()
	wsResponse.To = plugins.ToOthers
	for _, m := range p.manager.conns {
		for _, s := range m {
			if s.id != p.id {
				s.ch <- wsResponse.EncodeToString()
			}
		}
	}
	return nil
}

func (p *memoryPubSub) Close() error {
	mlog.Debugf(fmt.Sprintf("[Websocket] Closed, uid: %s, id: %s", p.uid, p.id))
	p.manager.Delete(p.uid, p.id)
	return nil
}

func (p *memoryPubSub) Subscribe() <-chan string {
	p.manager.RLock()
	defer p.manager.RUnlock()
	m := p.manager.conns[p.Uid()]
	s := m[p.ID()]
	return s.ch
}
