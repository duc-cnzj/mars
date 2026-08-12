package event

//go:generate go tool mockgen -destination ./mock_event.go -package event github.com/duc-cnzj/mars/v6/internal/event Dispatcher
import "github.com/google/wire"

// WireEvent 提供 event.Dispatcher 的装配集。
var WireEvent = wire.NewSet(NewDispatcher)
