package event

//go:generate mockgen -destination ./mock_event.go -package event github.com/duc-cnzj/mars/v4/internal/event Dispatcher
import "github.com/google/wire"

var WireEvent = wire.NewSet(NewDispatcher)
