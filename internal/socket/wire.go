package socket

//go:generate mockgen -destination ./mock_socket.go -package socket github.com/duc-cnzj/mars/v4/internal/socket JobManager,Job,Percentable
import (
	"github.com/duc-cnzj/mars/v4/internal/util/counter"
	"github.com/google/wire"
)

var WireSocket = wire.NewSet(NewWebsocketManager, NewJobManager, counter.NewCounter, NewReleaseInstaller)
