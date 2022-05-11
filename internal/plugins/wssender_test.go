package plugins

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/duc-cnzj/mars-client/v4/websocket"
	"github.com/duc-cnzj/mars/internal/app/instance"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestDecodeMessage(t *testing.T) {
	m := Message{
		Data: []byte("111"),
		To:   1,
		ID:   "id",
	}

	s, _ := json.Marshal(&m)
	message, _ := DecodeMessage(s)
	assert.Equal(t, m, message)
}

func TestMessage_Marshal(t *testing.T) {
	m := Message{
		Data: []byte("111"),
		To:   1,
		ID:   "id",
	}

	s, _ := json.Marshal(&m)
	assert.Equal(t, s, m.Marshal())
}

func TestProtoToMessage(t *testing.T) {
	m := websocket.Metadata{
		Message: "aa",
	}
	marshal, _ := proto.Marshal(&m)
	assert.Equal(t, Message{
		Data: marshal,
		To:   1,
		ID:   "idx",
	}, ProtoToMessage(&m, 1, "idx"))
}

func TestEmptyPubSub_Close(t *testing.T) {
	assert.Nil(t, (&EmptyPubSub{}).Close())
}

func TestEmptyPubSub_ID(t *testing.T) {
	assert.Equal(t, "", (&EmptyPubSub{}).ID())
}

func TestEmptyPubSub_Info(t *testing.T) {
	assert.Nil(t, (&EmptyPubSub{}).Info())
}

func TestEmptyPubSub_Subscribe(t *testing.T) {
	assert.Nil(t, (&EmptyPubSub{}).Subscribe())
}

func TestEmptyPubSub_ToAll(t *testing.T) {
	assert.Nil(t, (&EmptyPubSub{}).ToAll(nil))
}

func TestEmptyPubSub_ToOthers(t *testing.T) {
	assert.Nil(t, (&EmptyPubSub{}).ToOthers(nil))
}

func TestEmptyPubSub_ToSelf(t *testing.T) {
	assert.Nil(t, (&EmptyPubSub{}).ToSelf(nil))
}

func TestEmptyPubSub_Uid(t *testing.T) {
	assert.Equal(t, "", (&EmptyPubSub{}).Uid())
}

type senderPlugin struct {
	called bool
	WsSender
}

func (s *senderPlugin) Initialize(args map[string]any) error {
	s.called = true
	return nil
}

func TestGetWsSender(t *testing.T) {
	p := &senderPlugin{}
	ma := &mockApp{
		p: map[string]contracts.PluginInterface{"sender": p},
	}
	instance.SetInstance(ma)
	wsSenderOnce = sync.Once{}
	GetWsSender()
	assert.Equal(t, 1, ma.callback)
	assert.True(t, p.called)
}
