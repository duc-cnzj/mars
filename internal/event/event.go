package event

import (
	"context"
	"sync"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

// Event 是经 Dispatcher 分发的具名应用事件。
type Event string

// String 实现 fmt.Stringer，返回事件名。
func (e Event) String() string {
	return string(e)
}

// Listener 处理分发的事件负载；返回的错误由 dispatcher 记录，不阻断同事件其余监听器。
type Listener func(any, Event) error

// eventChannelBuffer is the capacity of the dispatcher's event channel.
const eventChannelBuffer = 800

// Dispatcher 是进程内即发即忘（fire-and-forget）的事件总线。
type Dispatcher interface {
	// Listen 为事件注册监听器。
	Listen(Event, Listener)

	// Dispatch 异步入队事件，永不阻塞：内部缓冲满时丢弃事件并告警。
	Dispatch(Event, any)

	// GetListeners 返回事件已注册监听器的拷贝。
	GetListeners(Event) []Listener

	// Run 在后台 goroutine 启动处理循环；必须恰好调用一次。
	Run(context.Context) error

	// Shutdown 停止处理循环。
	Shutdown(context.Context) error

	// List 返回按事件分组的全部监听器拷贝。
	List() map[Event][]Listener
}

// eventBody is a queued event and its payload.
type eventBody struct {
	event   Event
	payload any
}

// dispatcher is the in-memory Dispatcher implementation. All access to the
// listeners map is guarded by the embedded sync.RWMutex.
type dispatcher struct {
	sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	ch        chan *eventBody
	logger    mlog.Logger
	listeners map[Event][]Listener
}

var _ Dispatcher = (*dispatcher)(nil)

// NewDispatcher 构造基于 eventChannelBuffer 容量事件通道的 Dispatcher。
func NewDispatcher(logger mlog.Logger) Dispatcher {
	ctx, cancelFunc := context.WithCancel(context.TODO())

	return &dispatcher{
		ctx:       ctx,
		cancel:    cancelFunc,
		ch:        make(chan *eventBody, eventChannelBuffer),
		logger:    logger.WithModule("event/dispatcher"),
		listeners: map[Event][]Listener{},
	}
}

// Run 在后台 goroutine 启动处理循环并返回 nil：循环排空事件通道、逐个执行各事件
// 监听器；dispatcher/caller 的 ctx 结束或通道关闭时停止。错误返回值仅用于满足
// app.Server 接口（启动不可能失败）。调用方必须恰好调用一次。
func (d *dispatcher) Run(ctx context.Context) error {
	d.logger.Info("[Event]: dispatcher running")
	go func() {
		for {
			select {
			case <-d.ctx.Done():
				d.logger.Warning("[Shutdown]: event dispatcher context done")
				return
			case <-ctx.Done():
				d.logger.Warning("event dispatcher context done")
				return
			case obj, ok := <-d.ch:
				if !ok {
					d.logger.Warning("event dispatcher channel closed")
					return
				}
				go func() {
					defer d.logger.HandlePanic("event dispatcher")
					for _, fn := range d.GetListeners(obj.event) {
						if err := fn(obj.payload, obj.event); err != nil {
							d.logger.Error(err)
						}
					}
				}()
			}
		}
	}()
	return nil
}

// Shutdown 取消 dispatcher 内部 ctx，停止 Run 启动的处理循环；立即返回，
// 入参 ctx 暂不消费（为将来优雅排空预留）。
func (d *dispatcher) Shutdown(ctx context.Context) error {
	d.logger.Info("[Event]: dispatcher shutdown")
	d.cancel()
	return nil
}

// Listen 为事件注册监听器。
func (d *dispatcher) Listen(event Event, listener Listener) {
	d.Lock()
	defer d.Unlock()

	// append on a nil slice yields []Listener{listener}, covering the first
	// registration for an event.
	d.listeners[event] = append(d.listeners[event], listener)
}

// Dispatch 入队事件，非阻塞：缓冲满时丢弃事件并告警。
func (d *dispatcher) Dispatch(event Event, payload any) {
	select {
	case d.ch <- &eventBody{
		event:   event,
		payload: payload,
	}:
	default:
		d.logger.Warningf("event dispatcher channel full drop event %v %v", event.String(), payload)
	}
}

// GetListeners 返回事件已注册监听器的拷贝，调用方无法改动内部注册。
func (d *dispatcher) GetListeners(event Event) []Listener {
	d.RLock()
	defer d.RUnlock()

	return append([]Listener(nil), d.listeners[event]...)
}

// List 返回按事件分组的全部监听器拷贝，调用方无法改动内部注册。
func (d *dispatcher) List() map[Event][]Listener {
	d.RLock()
	defer d.RUnlock()

	out := make(map[Event][]Listener, len(d.listeners))
	for event, listeners := range d.listeners {
		out[event] = append([]Listener(nil), listeners...)
	}
	return out
}
