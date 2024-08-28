package socket

import (
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
)

func TestSafeWriteMessageChSendWhenNotClosed(t *testing.T) {
	logger := mlog.NewLogger(nil)
	ch := NewSafeWriteMessageCh(logger, 1)

	ch.Send(MessageItem{Msg: "test", Type: MessageSuccess})

	select {
	case msg := <-ch.Chan():
		assert.Equal(t, "test", msg.Msg)
		assert.Equal(t, MessageSuccess, msg.Type)
	default:
		t.Fatal("Expected message to be sent")
	}
}

func TestSafeWriteMessageChSendWhenClosed(t *testing.T) {
	logger := mlog.NewLogger(nil)
	ch := NewSafeWriteMessageCh(logger, 1)

	ch.Close()
	ch.Send(MessageItem{Msg: "test", Type: MessageText})

	select {
	case _, ok := <-ch.Chan():
		if ok {
			t.Fatal("Expected no message to be sent")
		}
	default:
	}
}

func TestSafeWriteMessageChSendWhenFull(t *testing.T) {
	logger := mlog.NewLogger(nil)
	ch := NewSafeWriteMessageCh(logger, 1)

	ch.Send(MessageItem{Msg: "test1", Type: MessageError})
	ch.Send(MessageItem{Msg: "test2", Type: MessageError})

	select {
	case msg := <-ch.Chan():
		assert.Equal(t, "test1", msg.Msg)
		assert.Equal(t, MessageError, msg.Type)
	default:
		t.Fatal("Expected message to be sent")
	}

	select {
	case <-ch.Chan():
		t.Fatal("Expected no message to be sent")
	default:
	}
}
