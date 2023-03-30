package plugins

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/stretchr/testify/assert"
)

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
	err error
}

func (s *senderPlugin) Initialize(args map[string]any) error {
	s.called = true
	return s.err
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
func TestGetWsSender2(t *testing.T) {
	p := &senderPlugin{
		err: errors.New("xx"),
	}
	ma := &mockApp{
		p: map[string]contracts.PluginInterface{"sender": p},
	}
	instance.SetInstance(ma)
	wsSenderOnce = sync.Once{}
	assert.Panics(t, func() {
		GetWsSender()
	})
	assert.Equal(t, 0, ma.callback)
	assert.True(t, p.called)
}

func TestEmptyPubSub_Join(t *testing.T) {
	assert.Nil(t, (&EmptyPubSub{}).Join(1))
}

func TestEmptyPubSub_Leave(t *testing.T) {
	assert.Nil(t, (&EmptyPubSub{}).Leave(1, 1))
}

func TestEmptyPubSub_Run(t *testing.T) {
	assert.Nil(t, (&EmptyPubSub{}).Run(context.TODO()))
}

func TestEmptyPubSub_Publish(t *testing.T) {
	assert.Nil(t, (&EmptyPubSub{}).Publish(1, nil))
}
