package socket

import (
	"sync"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/util/closeable"
)

type SafeWriteMessageChan interface {
	Close()
	Chan() <-chan MessageItem
	Send(m MessageItem)
}

var _ SafeWriteMessageChan = (*safeWriteMessageCh)(nil)

type safeWriteMessageCh struct {
	logger    mlog.Logger
	closeable closeable.Closeable

	chMu sync.Mutex
	ch   chan MessageItem
	once sync.Once
}

func NewSafeWriteMessageCh(logger mlog.Logger, chSize int) SafeWriteMessageChan {
	return &safeWriteMessageCh{
		logger: logger,
		ch:     make(chan MessageItem, chSize),
	}
}

func (s *safeWriteMessageCh) Close() {
	s.once.Do(func() {
		s.logger.Debug("safeWriteMessageCh closed")
		s.closeable.Close()
		s.chMu.Lock()
		defer s.chMu.Unlock()
		close(s.ch)
	})
}

func (s *safeWriteMessageCh) Chan() <-chan MessageItem {
	return s.ch
}

func (s *safeWriteMessageCh) Send(m MessageItem) {
	s.chMu.Lock()
	defer s.chMu.Unlock()

	if s.closeable.IsClosed() {
		s.logger.Debugf("[Websocket]: Drop %s type %s", m.Msg, m.Type)
		return
	}

	s.logger.Debug("Send message to channel", m.Msg)

	select {
	case s.ch <- m:
	default:
		s.logger.Warningf("Channel is full, dropping message: %s type %s", m.Msg, m.Type)
	}
}
