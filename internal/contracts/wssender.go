package contracts

//go:generate mockgen -destination ../mock/mock_wssender_pubsub.go -package mock github.com/duc-cnzj/mars/internal/contracts PubSub

type PubSub interface {
	Info() any
	Uid() string
	ID() string
	ToSelf(WebsocketMessage) error
	ToAll(WebsocketMessage) error
	ToOthers(WebsocketMessage) error
	Subscribe() <-chan []byte
	Close() error
}
