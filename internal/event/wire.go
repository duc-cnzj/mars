package event

//go:generate go tool mockgen -destination ./mock_event.go -package event github.com/duc-cnzj/mars/v6/internal/event Dispatcher
import "github.com/google/wire"

var WireEvent = wire.NewSet(NewDispatcher)
